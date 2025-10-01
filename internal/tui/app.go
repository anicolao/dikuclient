package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/anicolao/dikuclient/internal/client"
	"github.com/anicolao/dikuclient/internal/mapper"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	conn              *client.Connection
	viewport          viewport.Model
	output            []string
	currentInput      string
	cursorPos         int
	width             int
	height            int
	connected         bool
	host              string
	port              int
	sidebarWidth      int
	err               error
	mudLogFile        *os.File
	tuiLogFile        *os.File
	telnetDebugLog    *os.File // Debug log for telnet/UTF-8 processing
	echoSuppressed    bool     // Server has disabled echo (e.g., for passwords)
	username          string
	password          string
	autoLoginState    int               // 0=idle, 1=sent username, 2=sent password
	worldMap          *mapper.Map       // World map for navigation
	recentOutput      []string          // Buffer for recent output to detect rooms
	pendingMovement   string            // Last movement command sent
	mapDebug          bool              // Enable mapper debug output
	autoWalking       bool              // Currently auto-walking with /go
	autoWalkPath      []string          // Path to auto-walk
	autoWalkIndex     int               // Current step in auto-walk
}

var (
	mainStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1)

	emptyPanelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

type mudMsg string
type errMsg error
type echoStateMsg bool // true if echo suppressed (password mode)
type autoWalkTickMsg struct{}

// NewModel creates a new application model
func NewModel(host string, port int, mudLogFile, tuiLogFile *os.File) Model {
	return NewModelWithAuth(host, port, "", "", mudLogFile, tuiLogFile, nil, false)
}

// NewModelWithAuth creates a new application model with authentication credentials
func NewModelWithAuth(host string, port int, username, password string, mudLogFile, tuiLogFile, telnetDebugLog *os.File, mapDebug bool) Model {
	vp := viewport.New(0, 0)
	// Don't apply any style to viewport - let ANSI codes pass through naturally

	// Load or create world map
	worldMap, err := mapper.Load()
	if err != nil {
		// If we can't load the map, create a new one
		worldMap = mapper.NewMap()
	}

	return Model{
		viewport:       vp,
		output:         []string{},
		currentInput:   "",
		cursorPos:      0,
		host:           host,
		port:           port,
		sidebarWidth:   30,
		mudLogFile:     mudLogFile,
		tuiLogFile:     tuiLogFile,
		telnetDebugLog: telnetDebugLog,
		username:       username,
		password:       password,
		autoLoginState: 0,
		worldMap:       worldMap,
		recentOutput:   []string{},
		mapDebug:       mapDebug,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return m.connect
}

// connect establishes a connection to the MUD server
func (m Model) connect() tea.Msg {
	conn, err := client.NewConnectionWithDebug(m.host, m.port, m.telnetDebugLog)
	if err != nil {
		return errMsg(err)
	}
	return conn
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if m.conn != nil && m.connected {
				command := m.currentInput

				// Check if this is a client command (starts with /)
				if strings.HasPrefix(command, "/") {
					clientCmd := m.handleClientCommand(command)
					m.currentInput = ""
					m.cursorPos = 0
					m.updateViewport()
					return m, clientCmd
				}

				// Check if this is a movement command
				if movement := mapper.DetectMovement(command); movement != "" {
					m.pendingMovement = movement
				}

				// Send command to MUD server
				m.conn.Send(command)

				// Don't modify m.output here - let the server echo if it wants to
				// Or we can store the command for display purposes
				if !m.echoSuppressed && command != "" {
					// Add the command as a new line in output (it will show on the prompt line)
					// This preserves it even when new output arrives
					if len(m.output) > 0 {
						// Modify the last line to include the command
						m.output[len(m.output)-1] = m.output[len(m.output)-1] + "\x1b[93m" + command + "\x1b[0m"
					}
				}

				// Reset input
				m.currentInput = ""
				m.cursorPos = 0
				// Update display immediately
				m.updateViewport()
			}
			return m, nil

		case tea.KeyBackspace:
			if m.cursorPos > 0 {
				m.currentInput = m.currentInput[:m.cursorPos-1] + m.currentInput[m.cursorPos:]
				m.cursorPos--
				m.updateViewport()
			}
			return m, nil

		case tea.KeyLeft:
			if m.cursorPos > 0 {
				m.cursorPos--
			}
			return m, nil

		case tea.KeyRight:
			if m.cursorPos < len(m.currentInput) {
				m.cursorPos++
			}
			return m, nil

		case tea.KeyHome:
			m.cursorPos = 0
			return m, nil

		case tea.KeyEnd:
			m.cursorPos = len(m.currentInput)
			return m, nil

		case tea.KeySpace:
			// Explicitly handle space key
			m.currentInput = m.currentInput[:m.cursorPos] + " " + m.currentInput[m.cursorPos:]
			m.cursorPos++
			m.updateViewport()
			return m, nil

		default:
			// Handle regular character input
			if msg.Type == tea.KeyRunes {
				// Insert character at cursor position
				m.currentInput = m.currentInput[:m.cursorPos] + string(msg.Runes) + m.currentInput[m.cursorPos:]
				m.cursorPos += len(msg.Runes)
				m.updateViewport()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		sidebarWidth := m.sidebarWidth
		mainWidth := m.width - sidebarWidth - 6

		m.viewport.Width = mainWidth
		m.viewport.Height = m.height - headerHeight - 2
		// Don't apply viewport style - let ANSI codes pass through

		m.updateViewport()
		return m, nil

	case *client.Connection:
		m.conn = msg
		m.connected = true
		m.output = append(m.output, fmt.Sprintf("Connected to %s:%d", m.host, m.port))
		m.updateViewport()
		return m, m.listenForMessages

	case mudMsg:
		// Add message to output - it already has proper line endings
		msgStr := string(msg)

		// Log raw MUD output if logging enabled
		if m.mudLogFile != nil {
			fmt.Fprintf(m.mudLogFile, "[%s] %s", time.Now().Format("15:04:05.000"), msgStr)
			m.mudLogFile.Sync()
		}

		// Split into lines and add them individually to preserve formatting
		lines := strings.Split(msgStr, "\n")
		for i, line := range lines {
			// Don't add empty line at the end if message ended with \n
			if i == len(lines)-1 && line == "" {
				continue
			}
			m.output = append(m.output, line)
			m.recentOutput = append(m.recentOutput, line)
		}

		// Keep recentOutput to last 30 lines for room detection
		if len(m.recentOutput) > 30 {
			m.recentOutput = m.recentOutput[len(m.recentOutput)-30:]
		}

		// Try to detect room information from recent output
		m.detectAndUpdateRoom()

		// Check for auto-login prompts
		if m.username != "" && m.autoLoginState < 2 {
			lastLine := ""
			if len(m.output) > 0 {
				lastLine = strings.ToLower(strings.TrimSpace(m.output[len(m.output)-1]))
			}

			// Check for username prompt
			if m.autoLoginState == 0 && (strings.Contains(lastLine, "name") ||
				strings.Contains(lastLine, "login") ||
				strings.Contains(lastLine, "account") ||
				strings.Contains(lastLine, "character")) {
				// Send username automatically
				m.conn.Send(m.username)
				m.autoLoginState = 1
				m.output = append(m.output, fmt.Sprintf("\x1b[90m[Auto-login: sending username '%s']\x1b[0m", m.username))
			}
		}

		if m.password != "" && m.autoLoginState == 1 {
			lastLine := ""
			if len(m.output) > 0 {
				lastLine = strings.ToLower(strings.TrimSpace(m.output[len(m.output)-1]))
			}

			// Check for password prompt
			if strings.Contains(lastLine, "password") || strings.Contains(lastLine, "pass") {
				// Send password automatically
				m.conn.Send(m.password)
				m.autoLoginState = 2
				m.output = append(m.output, "\x1b[90m[Auto-login: sending password]\x1b[0m")
			}
		}

		m.updateViewport()
		return m, m.listenForMessages

	case echoStateMsg:
		// Update echo suppression state (true = suppressed/password mode)
		m.echoSuppressed = bool(msg)
		m.updateViewport()
		return m, m.listenForMessages

	case errMsg:
		m.err = msg
		m.output = append(m.output, fmt.Sprintf("Error: %v", msg))
		m.updateViewport()
		return m, nil
		
	case autoWalkTickMsg:
		// Process next step in auto-walk
		if m.autoWalking && m.autoWalkIndex < len(m.autoWalkPath) {
			direction := m.autoWalkPath[m.autoWalkIndex]
			m.autoWalkIndex++
			
			// Send the movement command
			if m.conn != nil && m.connected {
				m.conn.Send(direction)
				m.pendingMovement = direction
				m.output = append(m.output, fmt.Sprintf("\x1b[90m[Auto-walk: %s (%d/%d)]\x1b[0m", direction, m.autoWalkIndex, len(m.autoWalkPath)))
				m.updateViewport()
			}
			
			// If more steps remain, schedule next tick
			if m.autoWalkIndex < len(m.autoWalkPath) {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return autoWalkTickMsg{}
				})
			} else {
				// Auto-walk complete
				m.autoWalking = false
				m.autoWalkPath = nil
				m.autoWalkIndex = 0
				m.output = append(m.output, "\x1b[92m[Auto-walk complete!]\x1b[0m")
				m.updateViewport()
			}
		}
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateViewport updates the viewport content with output and current input
func (m *Model) updateViewport() {
	// Always append input to the last line (all lines are treated as potential prompts)
	var content string
	if len(m.output) > 0 {
		lastLine := m.output[len(m.output)-1]

		if (m.currentInput != "" || m.connected) && !m.echoSuppressed {
			// Build input line with cursor (only if echo is not suppressed)
			inputLine := m.currentInput
			if m.cursorPos < len(m.currentInput) {
				// Show cursor in the middle of text
				inputLine = m.currentInput[:m.cursorPos] + "█" + m.currentInput[m.cursorPos:]
			} else {
				// Show cursor at the end
				inputLine = m.currentInput + "█"
			}

			// Append input inline to the last line with yellow color
			// Use bright yellow (93) for better visibility
			lines := make([]string, len(m.output)-1)
			copy(lines, m.output[:len(m.output)-1])
			lines = append(lines, lastLine+"\x1b[93m"+inputLine+"\x1b[0m")
			content = strings.Join(lines, "\n")
		} else if m.echoSuppressed && m.connected {
			// In password mode, just show the prompt without the input
			// Show a cursor to indicate user can type
			lines := make([]string, len(m.output)-1)
			copy(lines, m.output[:len(m.output)-1])
			lines = append(lines, lastLine+"█")
			content = strings.Join(lines, "\n")
		} else {
			content = strings.Join(m.output, "\n")
		}
	} else {
		// No output yet, just show cursor if connected
		if m.currentInput != "" || m.connected {
			if !m.echoSuppressed {
				inputLine := m.currentInput
				if m.cursorPos < len(m.currentInput) {
					inputLine = m.currentInput[:m.cursorPos] + "█" + m.currentInput[m.cursorPos:]
				} else {
					inputLine = m.currentInput + "█"
				}
				// Use bright yellow for better visibility
				content = "\x1b[93m" + inputLine + "\x1b[0m"
			} else {
				// Password mode - just show cursor
				content = "█"
			}
		}
	}

	m.viewport.SetContent(content)
	m.viewport.GotoBottom()

	// Log TUI content if logging enabled
	if m.tuiLogFile != nil {
		fmt.Fprintf(m.tuiLogFile, "[%s] === TUI Update ===\n%s\n\n", time.Now().Format("15:04:05.000"), content)
		m.tuiLogFile.Sync()
	}
}

// listenForMessages listens for messages from the MUD server
func (m Model) listenForMessages() tea.Msg {
	if m.conn == nil || m.conn.IsClosed() {
		return nil
	}

	select {
	case msg := <-m.conn.Receive():
		return mudMsg(msg)
	case echoSuppressed := <-m.conn.EchoState():
		return echoStateMsg(echoSuppressed)
	case err := <-m.conn.Errors():
		return errMsg(err)
	}
}

// View renders the application
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Status bar
	status := m.renderStatusBar()

	// Main content area (game output + sidebar)
	mainContent := m.renderMainContent()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		mainContent,
	)
}

func (m Model) renderStatusBar() string {
	statusText := "Disconnected"
	if m.connected {
		statusText = fmt.Sprintf("Connected to %s:%d", m.host, m.port)
	}

	status := statusStyle.Render(statusText)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(status)))
	return lipgloss.JoinHorizontal(lipgloss.Left, status, line)
}

func (m Model) renderMainContent() string {
	headerHeight := 3
	sidebarWidth := m.sidebarWidth
	mainWidth := m.width - sidebarWidth - 6
	contentHeight := m.height - headerHeight - 2

	// Game output viewport (already has dark background style applied)
	// Wrap in border
	gameOutput := mainStyle.
		Width(mainWidth).
		Height(contentHeight).
		Render(m.viewport.View())

	// Sidebar with empty panels
	sidebar := m.renderSidebar(sidebarWidth, contentHeight)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		gameOutput,
		sidebar,
	)
}

func (m Model) renderSidebar(width, height int) string {
	panelHeight := (height - 6) / 3

	// Character Stats panel (empty placeholder)
	statsPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("Character Stats"),
				"",
				emptyPanelStyle.Render("(not implemented)"),
			),
		)

	// Inventory panel (empty placeholder)
	inventoryPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("Inventory"),
				"",
				emptyPanelStyle.Render("(not implemented)"),
			),
		)

	// Map panel (empty placeholder)
	mapPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("Map"),
				"",
				emptyPanelStyle.Render("(not implemented)"),
			),
		)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statsPanel,
		inventoryPanel,
		mapPanel,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// detectAndUpdateRoom tries to parse room information from recent output
func (m *Model) detectAndUpdateRoom() {
	// Only detect rooms when we have a pending movement (user just moved)
	if m.pendingMovement == "" {
		return
	}
	
	if len(m.recentOutput) < 3 {
		return // Need at least a few lines to detect a room
	}

	// Try to parse room info from recent output
	roomInfo := mapper.ParseRoomInfo(m.recentOutput, m.mapDebug)
	
	// Only display debug info if mapDebug flag is enabled
	if m.mapDebug && roomInfo != nil && roomInfo.DebugInfo != "" {
		// Add debug info to output
		debugLines := strings.Split(strings.TrimSpace(roomInfo.DebugInfo), "\n")
		for _, line := range debugLines {
			m.output = append(m.output, "\x1b[90m"+line+"\x1b[0m") // Gray color for debug
		}
	}
	
	if roomInfo == nil || roomInfo.Title == "" {
		return // No valid room detected
	}

	// Create or update room in map
	room := mapper.NewRoom(roomInfo.Title, roomInfo.Description, roomInfo.Exits)
	
	// Set the movement direction
	m.worldMap.SetLastDirection(m.pendingMovement)
	m.pendingMovement = ""
	
	m.worldMap.AddOrUpdateRoom(room)
	
	// Save map periodically (every room visit)
	m.worldMap.Save()
	
	// Notify user that room was added (only if debug enabled)
	if m.mapDebug {
		m.output = append(m.output, fmt.Sprintf("\x1b[92m[Mapper: Added room '%s' with exits: %v]\x1b[0m", room.Title, roomInfo.Exits))
	}
}

// handleClientCommand processes client-side commands starting with /
func (m *Model) handleClientCommand(command string) tea.Cmd {
	command = strings.TrimSpace(command)
	if !strings.HasPrefix(command, "/") {
		return nil
	}

	// Remove the leading /
	command = strings.TrimPrefix(command, "/")
	parts := strings.Fields(command)
	
	if len(parts) == 0 {
		m.output = append(m.output, "\x1b[91mError: Empty command\x1b[0m")
		return nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "point":
		m.handlePointCommand(args)
		return nil
	case "wayfind":
		m.handleWayfindCommand(args)
		return nil
	case "map":
		m.handleMapCommand(args)
		return nil
	case "rooms":
		m.handleRoomsCommand(args)
		return nil
	case "go":
		return m.handleGoCommand(args)
	case "help":
		m.handleHelpCommand()
		return nil
	default:
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError: Unknown command '/%s'. Type /help for available commands.\x1b[0m", cmd))
		return nil
	}
}

// handlePointCommand shows the next direction to reach a destination
func (m *Model) handlePointCommand(args []string) {
	if len(args) == 0 {
		m.output = append(m.output, "\x1b[91mUsage: /point <room search terms>\x1b[0m")
		return
	}

	query := strings.Join(args, " ")
	rooms := m.worldMap.FindRooms(query)

	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return
	}

	if len(rooms) > 1 {
		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific.\x1b[0m")
		return
	}

	// Find path to the room
	targetRoom := rooms[0]
	path := m.worldMap.FindPath(targetRoom.ID)

	if path == nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo path found to '%s'\x1b[0m", targetRoom.Title))
		return
	}

	if len(path) == 0 {
		m.output = append(m.output, "\x1b[92mYou are already at that location!\x1b[0m")
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92mTo reach '%s', go: %s\x1b[0m", targetRoom.Title, path[0]))
}

// handleWayfindCommand shows the full path to reach a destination
func (m *Model) handleWayfindCommand(args []string) {
	if len(args) == 0 {
		m.output = append(m.output, "\x1b[91mUsage: /wayfind <room search terms>\x1b[0m")
		return
	}

	query := strings.Join(args, " ")
	rooms := m.worldMap.FindRooms(query)

	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return
	}

	if len(rooms) > 1 {
		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific.\x1b[0m")
		return
	}

	// Find path to the room
	targetRoom := rooms[0]
	path := m.worldMap.FindPath(targetRoom.ID)

	if path == nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo path found to '%s'\x1b[0m", targetRoom.Title))
		return
	}

	if len(path) == 0 {
		m.output = append(m.output, "\x1b[92mYou are already at that location!\x1b[0m")
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92mPath to '%s' (%d steps):\x1b[0m", targetRoom.Title, len(path)))
	m.output = append(m.output, fmt.Sprintf("  \x1b[96m%s\x1b[0m", strings.Join(path, " -> ")))
}

// handleMapCommand shows information about the current map
func (m *Model) handleMapCommand(args []string) {
	current := m.worldMap.GetCurrentRoom()
	
	m.output = append(m.output, "\x1b[92m=== Map Information ===\x1b[0m")
	m.output = append(m.output, fmt.Sprintf("Total rooms explored: \x1b[96m%d\x1b[0m", len(m.worldMap.Rooms)))
	
	if current != nil {
		m.output = append(m.output, fmt.Sprintf("Current room: \x1b[96m%s\x1b[0m", current.Title))
		if len(current.Exits) > 0 {
			exits := []string{}
			for dir := range current.Exits {
				exits = append(exits, dir)
			}
			m.output = append(m.output, fmt.Sprintf("Exits: \x1b[96m%s\x1b[0m", strings.Join(exits, ", ")))
		}
	} else {
		m.output = append(m.output, "\x1b[90mNo current room detected yet\x1b[0m")
	}
}

// handleHelpCommand shows available client commands
func (m *Model) handleHelpCommand() {
	m.output = append(m.output, "\x1b[92m=== Client Commands ===\x1b[0m")
	m.output = append(m.output, "  \x1b[96m/point <room>\x1b[0m    - Show next direction to reach a room")
	m.output = append(m.output, "  \x1b[96m/wayfind <room>\x1b[0m - Show full path to reach a room")
	m.output = append(m.output, "  \x1b[96m/go <room>\x1b[0m      - Auto-walk to a room (one step per second)")
	m.output = append(m.output, "  \x1b[96m/map\x1b[0m            - Show map information")
	m.output = append(m.output, "  \x1b[96m/rooms [filter]\x1b[0m - List all known rooms (optionally filtered)")
	m.output = append(m.output, "  \x1b[96m/help\x1b[0m           - Show this help message")
	m.output = append(m.output, "")
	m.output = append(m.output, "\x1b[90mRoom search matches all terms in room title, description, or exits\x1b[0m")
}

// handleRoomsCommand lists all known rooms or filters by search terms
func (m *Model) handleRoomsCommand(args []string) {
	var roomsToDisplay []*mapper.Room
	var headerText string
	
	if len(args) == 0 {
		// No filter - show all rooms
		allRooms := m.worldMap.GetAllRooms()
		
		if len(allRooms) == 0 {
			m.output = append(m.output, "\x1b[93mNo rooms have been explored yet.\x1b[0m")
			return
		}
		
		roomsToDisplay = make([]*mapper.Room, 0, len(allRooms))
		for _, room := range allRooms {
			roomsToDisplay = append(roomsToDisplay, room)
		}
		headerText = fmt.Sprintf("\x1b[92m=== Known Rooms (%d) ===\x1b[0m", len(roomsToDisplay))
	} else {
		// Filter by search terms
		query := strings.Join(args, " ")
		roomsToDisplay = m.worldMap.FindRooms(query)
		
		if len(roomsToDisplay) == 0 {
			m.output = append(m.output, fmt.Sprintf("\x1b[93mNo rooms found matching '%s'\x1b[0m", query))
			return
		}
		
		headerText = fmt.Sprintf("\x1b[92m=== Rooms matching '%s' (%d) ===\x1b[0m", query, len(roomsToDisplay))
	}
	
	m.output = append(m.output, headerText)
	
	// Sort rooms by title for consistent display
	sort.Slice(roomsToDisplay, func(i, j int) bool {
		return roomsToDisplay[i].Title < roomsToDisplay[j].Title
	})
	
	// Display rooms
	for i, room := range roomsToDisplay {
		exitList := make([]string, 0, len(room.Exits))
		for dir := range room.Exits {
			exitList = append(exitList, dir)
		}
		sort.Strings(exitList)
		
		exitsStr := strings.Join(exitList, ", ")
		if exitsStr == "" {
			exitsStr = "none"
		}
		
		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m \x1b[90m[%s]\x1b[0m", i+1, room.Title, exitsStr))
	}
}

// handleGoCommand starts auto-walking to a destination
func (m *Model) handleGoCommand(args []string) tea.Cmd {
	if len(args) == 0 {
		m.output = append(m.output, "\x1b[91mUsage: /go <room search terms>\x1b[0m")
		return nil
	}
	
	// If already auto-walking, stop it
	if m.autoWalking {
		m.autoWalking = false
		m.autoWalkPath = nil
		m.autoWalkIndex = 0
		m.output = append(m.output, "\x1b[93mAuto-walk cancelled.\x1b[0m")
		return nil
	}
	
	query := strings.Join(args, " ")
	rooms := m.worldMap.FindRooms(query)
	
	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return nil
	}
	
	if len(rooms) > 1 {
		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific.\x1b[0m")
		return nil
	}
	
	// Find path to the room
	targetRoom := rooms[0]
	path := m.worldMap.FindPath(targetRoom.ID)
	
	if path == nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo path found to '%s'\x1b[0m", targetRoom.Title))
		return nil
	}
	
	if len(path) == 0 {
		m.output = append(m.output, "\x1b[92mYou are already at that location!\x1b[0m")
		return nil
	}
	
	// Start auto-walking
	m.autoWalking = true
	m.autoWalkPath = path
	m.autoWalkIndex = 0
	m.output = append(m.output, fmt.Sprintf("\x1b[92mAuto-walking to '%s' (%d steps). Type /go to cancel.\x1b[0m", targetRoom.Title, len(path)))
	
	// Return a command that starts the first tick after 1 second
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return autoWalkTickMsg{}
	})
}
