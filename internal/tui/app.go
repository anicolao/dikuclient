package tui

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/anicolao/dikuclient/internal/client"
	"github.com/anicolao/dikuclient/internal/mapper"
	"github.com/anicolao/dikuclient/internal/triggers"
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
	lastRoomSearch    []*mapper.Room    // Last room search results for disambiguation
	triggerManager    *triggers.Manager // Trigger manager
	inventory         []string          // Current inventory items
	inventoryTime     time.Time         // Time when inventory was last updated
	inventoryViewport viewport.Model    // Viewport for scrollable inventory
	tells             []string          // Recent tells received
	tellsViewport     viewport.Model    // Viewport for scrollable tells
	skipNextRoomDetection bool          // Skip next room detection (e.g., after recall teleport)
	autoWalkTarget    string            // Target room title for auto-walk (for recovery)
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

	// Load or create trigger manager
	triggerManager, err := triggers.Load()
	if err != nil {
		// If we can't load triggers, create a new manager
		triggerManager = triggers.NewManager()
	}

	inventoryVp := viewport.New(0, 0)
	tellsVp := viewport.New(0, 0)

	return Model{
		viewport:          vp,
		output:            []string{},
		currentInput:      "",
		cursorPos:         0,
		host:              host,
		port:              port,
		sidebarWidth:      60, // Doubled from 30 to 60
		mudLogFile:        mudLogFile,
		tuiLogFile:        tuiLogFile,
		telnetDebugLog:    telnetDebugLog,
		username:          username,
		password:          password,
		autoLoginState:    0,
		worldMap:          worldMap,
		recentOutput:      []string{},
		mapDebug:          mapDebug,
		triggerManager:    triggerManager,
		inventoryViewport: inventoryVp,
		tellsViewport:     tellsVp,
	}
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return m.connect
}

// connect establishes a connection to the MUD server
func (m *Model) connect() tea.Msg {
	conn, err := client.NewConnectionWithDebug(m.host, m.port, m.telnetDebugLog)
	if err != nil {
		return errMsg(err)
	}
	return conn
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					// Save the current prompt line before executing command
					var savedPrompt string
					if len(m.output) > 0 {
						savedPrompt = m.output[len(m.output)-1]
						// Replace the prompt line with the command
						m.output[len(m.output)-1] = savedPrompt + "\x1b[93m" + command + "\x1b[0m"
					}

					clientCmd := m.handleClientCommand(command)

					// Add two newlines (empty lines) and restore prompt after command output
					m.output = append(m.output, "")
					m.output = append(m.output, "")
					m.output = append(m.output, savedPrompt)

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

		// Update inventory viewport size
		panelHeight := (m.height - headerHeight - 2 - 6) / 3
		m.inventoryViewport.Width = sidebarWidth - 4 // Account for borders and padding
		m.inventoryViewport.Height = panelHeight - 4 // Account for header, timestamp, and borders

		// Update tells viewport size
		m.tellsViewport.Width = sidebarWidth - 4 // Account for borders and padding
		m.tellsViewport.Height = panelHeight - 4 // Account for header and borders

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

		var autoWalkCmd tea.Cmd

		// Split into lines and add them individually to preserve formatting
		lines := strings.Split(msgStr, "\n")
		for i, line := range lines {
			// Don't add empty line at the end if message ended with \n
			if i == len(lines)-1 && line == "" {
				continue
			}
			m.output = append(m.output, line)
			m.recentOutput = append(m.recentOutput, line)

			// Check if this line is a tell message
			m.detectAndParseTell(line)

			// Check for recall command (which causes teleportation)
			// Strip ANSI codes and check if line contains 'recall'
			cleanLine := stripANSI(line)
			if strings.Contains(strings.ToLower(cleanLine), "recall") {
				// Set flag to skip next room detection to avoid creating bad links
				m.skipNextRoomDetection = true
				if m.mapDebug {
					m.output = append(m.output, "\x1b[90m[Mapper: Detected 'recall' - will skip next room detection]\x1b[0m")
				}
			}

			// Check for "Alas, you cannot go that way..." during auto-walk
			if m.autoWalking && (strings.Contains(cleanLine, "Alas, you cannot go that way") || 
				strings.Contains(cleanLine, "cannot go that way")) {
				// Cancel current auto-walk and trigger recovery
				autoWalkCmd = m.handleAutoWalkFailure()
			}

			// Check if this line matches any triggers
			if m.triggerManager != nil && m.conn != nil {
				actions := m.triggerManager.Match(line)
				for _, action := range actions {
					// Send the action as if the user typed it
					m.output = append(m.output, fmt.Sprintf("\x1b[90m[Trigger: %s]\x1b[0m", action))
					m.conn.Send(action)
				}
			}
		}

		// Keep recentOutput to last 30 lines for room detection
		if len(m.recentOutput) > 30 {
			m.recentOutput = m.recentOutput[len(m.recentOutput)-30:]
		}

		// Try to detect room information from recent output
		m.detectAndUpdateRoom()

		// Try to detect inventory information from recent output
		m.detectAndUpdateInventory()

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
		
		// If we have an auto-walk command (from recovery), execute it along with listening
		if autoWalkCmd != nil {
			return m, tea.Batch(m.listenForMessages, autoWalkCmd)
		}
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

	// Update inventory viewport for mouse wheel scrolling
	m.inventoryViewport, cmd = m.inventoryViewport.Update(msg)
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
func (m *Model) listenForMessages() tea.Msg {
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
func (m *Model) View() string {
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

func (m *Model) renderStatusBar() string {
	statusText := "Disconnected"
	if m.connected {
		statusText = fmt.Sprintf("Connected to %s:%d", m.host, m.port)
	}

	status := statusStyle.Render(statusText)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(status)))
	return lipgloss.JoinHorizontal(lipgloss.Left, status, line)
}

func (m *Model) renderMainContent() string {
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

func (m *Model) renderSidebar(width, height int) string {
	panelHeight := (height - 6) / 3

	// Tells panel with scrollable viewport
	var tellsHeader string
	var tellsContent string
	if len(m.tells) > 0 {
		tellsHeader = lipgloss.NewStyle().Bold(true).Render("Tells")
		tellsContent = strings.Join(m.tells, "\n")
	} else {
		tellsHeader = lipgloss.NewStyle().Bold(true).Render("Tells")
		tellsContent = emptyPanelStyle.Render("(no tells yet)")
	}

	// Update viewport content
	m.tellsViewport.SetContent(tellsContent)

	// Render tells panel with header and scrollable content
	tellsPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				tellsHeader,
				"",
				m.tellsViewport.View(),
			),
		)

	// Inventory panel with scrollable viewport
	var inventoryHeader string
	var inventoryContent string
	if len(m.inventory) > 0 {
		// Build header with timestamp
		timeStr := m.inventoryTime.Format("15:04:05")
		inventoryHeader = lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("Inventory"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("("+timeStr+")"),
		)

		// Add all items without truncation - viewport will handle scrolling
		inventoryContent = strings.Join(m.inventory, "\n")
	} else {
		inventoryHeader = lipgloss.NewStyle().Bold(true).Render("Inventory")
		inventoryContent = emptyPanelStyle.Render("(not populated)")
	}

	// Update viewport content
	m.inventoryViewport.SetContent(inventoryContent)

	// Render inventory panel with header and scrollable content
	inventoryPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				inventoryHeader,
				"",
				m.inventoryViewport.View(),
			),
		)

	// Map panel
	var mapHeader string
	var mapContent string
	
	if m.worldMap == nil {
		mapHeader = lipgloss.NewStyle().Bold(true).Render("Map")
		mapContent = emptyPanelStyle.Render("(not implemented)")
	} else {
		currentRoom := m.worldMap.GetCurrentRoom()
		if currentRoom == nil {
			mapHeader = lipgloss.NewStyle().Bold(true).Render("Map")
			mapContent = emptyPanelStyle.Render("(exploring...)")
		} else {
			mapHeader = lipgloss.NewStyle().Bold(true).Render(currentRoom.Title)
			// Calculate available height for map content (subtract header and spacing)
			mapHeight := panelHeight - 2
			mapContent = m.worldMap.FormatMapPanel(width-4, mapHeight)
		}
	}
	
	mapPanel := sidebarStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				mapHeader,
				"",
				mapContent,
			),
		)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tellsPanel,
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

	// Skip room detection if flag is set (e.g., after recall teleport)
	if m.skipNextRoomDetection {
		m.skipNextRoomDetection = false
		m.pendingMovement = "" // Clear pending movement
		if m.mapDebug {
			m.output = append(m.output, "\x1b[90m[Mapper: Skipped room detection due to recall]\x1b[0m")
		}
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

// detectAndUpdateInventory tries to parse inventory information from recent output
func (m *Model) detectAndUpdateInventory() {
	if len(m.recentOutput) < 3 {
		return // Need at least a few lines to detect inventory
	}

	// Try to parse inventory info from recent output
	invInfo := mapper.ParseInventoryInfo(m.recentOutput, false)

	if invInfo == nil {
		return // No valid inventory detected
	}

	// Update inventory and timestamp
	m.inventory = invInfo.Items
	m.inventoryTime = time.Now()
}

// tellRegex matches tell messages in format: <player> tells you '<content>'
var tellRegex = regexp.MustCompile(`^(.+?) tells you '(.*)'$`)

// detectAndParseTell tries to detect and parse a tell message from a line
func (m *Model) detectAndParseTell(line string) {
	// Strip ANSI codes for pattern matching
	cleanLine := stripANSI(line)

	matches := tellRegex.FindStringSubmatch(cleanLine)
	if matches == nil || len(matches) != 3 {
		return // Not a tell message
	}

	player := matches[1]
	content := matches[2]

	// Format as "Player: content" for the tells panel
	tellEntry := fmt.Sprintf("%s: %s", player, content)

	// Add to tells list (keep last 50 tells)
	m.tells = append(m.tells, tellEntry)
	if len(m.tells) > 50 {
		m.tells = m.tells[len(m.tells)-50:]
	}
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
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
	case "trigger":
		m.handleTriggerCommand(command)
		return nil
	case "triggers":
		m.handleTriggersCommand(args)
		return nil
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
		m.output = append(m.output, "\x1b[91mUsage: /point <room search terms> or /point <number> [search terms]\x1b[0m")
		return
	}

	var rooms []*mapper.Room
	var query string

	// Check if first argument is a number for room selection
	if roomNum, err := fmt.Sscanf(args[0], "%d", new(int)); err == nil && roomNum == 1 {
		var index int
		fmt.Sscanf(args[0], "%d", &index)

		// If only a number is provided, use lastRoomSearch
		if len(args) == 1 {
			if len(m.lastRoomSearch) == 0 {
				m.output = append(m.output, "\x1b[91mNo previous room search to select from. Use /rooms to see all rooms.\x1b[0m")
				return
			}
			if index < 1 || index > len(m.lastRoomSearch) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Must be between 1 and %d.\x1b[0m", len(m.lastRoomSearch)))
				return
			}
			rooms = []*mapper.Room{m.lastRoomSearch[index-1]}
		} else {
			// Number followed by search terms - search first, then select by index
			query = strings.Join(args[1:], " ")
			allMatches := m.worldMap.FindRooms(query)

			if len(allMatches) == 0 {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
				return
			}

			if index < 1 || index > len(allMatches) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Found %d rooms matching '%s'. Must be between 1 and %d.\x1b[0m", len(allMatches), query, len(allMatches)))
				return
			}

			rooms = []*mapper.Room{allMatches[index-1]}
			m.lastRoomSearch = allMatches
		}
	} else {
		// Regular search without numeric selection
		query = strings.Join(args, " ")
		rooms = m.worldMap.FindRooms(query)
	}

	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return
	}

	if len(rooms) > 1 {
		// Store results for later disambiguation
		m.lastRoomSearch = rooms

		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific, or use /point <number> to select a room.\x1b[0m")
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
		m.output = append(m.output, "\x1b[91mUsage: /wayfind <room search terms> or /wayfind <number> [search terms]\x1b[0m")
		return
	}

	var rooms []*mapper.Room
	var query string

	// Check if first argument is a number for room selection
	if roomNum, err := fmt.Sscanf(args[0], "%d", new(int)); err == nil && roomNum == 1 {
		var index int
		fmt.Sscanf(args[0], "%d", &index)

		// If only a number is provided, use lastRoomSearch
		if len(args) == 1 {
			if len(m.lastRoomSearch) == 0 {
				m.output = append(m.output, "\x1b[91mNo previous room search to select from. Use /rooms to see all rooms.\x1b[0m")
				return
			}
			if index < 1 || index > len(m.lastRoomSearch) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Must be between 1 and %d.\x1b[0m", len(m.lastRoomSearch)))
				return
			}
			rooms = []*mapper.Room{m.lastRoomSearch[index-1]}
		} else {
			// Number followed by search terms - search first, then select by index
			query = strings.Join(args[1:], " ")
			allMatches := m.worldMap.FindRooms(query)

			if len(allMatches) == 0 {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
				return
			}

			if index < 1 || index > len(allMatches) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Found %d rooms matching '%s'. Must be between 1 and %d.\x1b[0m", len(allMatches), query, len(allMatches)))
				return
			}

			rooms = []*mapper.Room{allMatches[index-1]}
			m.lastRoomSearch = allMatches
		}
	} else {
		// Regular search without numeric selection
		query = strings.Join(args, " ")
		rooms = m.worldMap.FindRooms(query)
	}

	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return
	}

	if len(rooms) > 1 {
		// Store results for later disambiguation
		m.lastRoomSearch = rooms

		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific, or use /wayfind <number> to select a room.\x1b[0m")
		return
	}

	// Find path to the room
	targetRoom := rooms[0]
	pathSteps := m.worldMap.FindPathWithRooms(targetRoom.ID)

	if pathSteps == nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo path found to '%s'\x1b[0m", targetRoom.Title))
		return
	}

	if len(pathSteps) == 0 {
		m.output = append(m.output, "\x1b[92mYou are already at that location!\x1b[0m")
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92mPath to '%s' (%d steps):\x1b[0m", targetRoom.Title, len(pathSteps)))
	for i, step := range pathSteps {
		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s -> %s\x1b[0m", i+1, step.Direction, step.RoomTitle))
	}
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
	m.output = append(m.output, "  \x1b[96m/point <room>\x1b[0m            - Show next direction to reach a room")
	m.output = append(m.output, "  \x1b[96m/wayfind <room>\x1b[0m         - Show full path to reach a room")
	m.output = append(m.output, "  \x1b[96m/go <room>\x1b[0m              - Auto-walk to a room (one step per second)")
	m.output = append(m.output, "  \x1b[96m/map\x1b[0m                    - Show map information")
	m.output = append(m.output, "  \x1b[96m/rooms [filter]\x1b[0m         - List all known rooms (optionally filtered)")
	m.output = append(m.output, "  \x1b[96m/trigger \"pat\" \"act\"\x1b[0m - Add a trigger (pattern can use <var>)")
	m.output = append(m.output, "  \x1b[96m/triggers list\x1b[0m          - List all triggers")
	m.output = append(m.output, "  \x1b[96m/triggers remove <n>\x1b[0m    - Remove trigger by number")
	m.output = append(m.output, "  \x1b[96m/help\x1b[0m                   - Show this help message")
	m.output = append(m.output, "")
	m.output = append(m.output, "\x1b[90mRoom search matches all terms in room title, description, or exits\x1b[0m")
	m.output = append(m.output, "\x1b[90mTriggers match output lines and execute actions (supports <variable> capture)\x1b[0m")
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

	// Store results for later disambiguation
	m.lastRoomSearch = roomsToDisplay

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
	// If already auto-walking, stop it
	if m.autoWalking {
		m.autoWalking = false
		m.autoWalkPath = nil
		m.autoWalkIndex = 0
		m.output = append(m.output, "\x1b[93mAuto-walk cancelled.\x1b[0m")
		return nil
	}

	if len(args) == 0 {
		m.output = append(m.output, "\x1b[91mUsage: /go <room search terms> or /go <number> [search terms]\x1b[0m")
		return nil
	}

	var rooms []*mapper.Room
	var query string

	// Check if first argument is a number for room selection
	if roomNum, err := fmt.Sscanf(args[0], "%d", new(int)); err == nil && roomNum == 1 {
		var index int
		fmt.Sscanf(args[0], "%d", &index)

		// If only a number is provided, use lastRoomSearch
		if len(args) == 1 {
			if len(m.lastRoomSearch) == 0 {
				m.output = append(m.output, "\x1b[91mNo previous room search to select from. Use /rooms to see all rooms.\x1b[0m")
				return nil
			}
			if index < 1 || index > len(m.lastRoomSearch) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Must be between 1 and %d.\x1b[0m", len(m.lastRoomSearch)))
				return nil
			}
			rooms = []*mapper.Room{m.lastRoomSearch[index-1]}
		} else {
			// Number followed by search terms - search first, then select by index
			query = strings.Join(args[1:], " ")
			allMatches := m.worldMap.FindRooms(query)

			if len(allMatches) == 0 {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
				return nil
			}

			if index < 1 || index > len(allMatches) {
				m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Found %d rooms matching '%s'. Must be between 1 and %d.\x1b[0m", len(allMatches), query, len(allMatches)))
				return nil
			}

			rooms = []*mapper.Room{allMatches[index-1]}
			m.lastRoomSearch = allMatches
		}
	} else {
		// Regular search without numeric selection
		query = strings.Join(args, " ")
		rooms = m.worldMap.FindRooms(query)
	}

	if len(rooms) == 0 {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mNo rooms found matching '%s'\x1b[0m", query))
		return nil
	}

	if len(rooms) > 1 {
		// Store results for later disambiguation
		m.lastRoomSearch = rooms

		m.output = append(m.output, fmt.Sprintf("\x1b[93mFound %d rooms matching '%s':\x1b[0m", len(rooms), query))
		for i, room := range rooms {
			if i >= 5 {
				m.output = append(m.output, fmt.Sprintf("  \x1b[90m... and %d more\x1b[0m", len(rooms)-5))
				break
			}
			m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m", i+1, room.Title))
		}
		m.output = append(m.output, "\x1b[93mPlease be more specific, or use /go <number> to select a room.\x1b[0m")
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
	m.autoWalkTarget = targetRoom.Title // Store target for recovery
	m.output = append(m.output, fmt.Sprintf("\x1b[92mAuto-walking to '%s' (%d steps). Type /go to cancel.\x1b[0m", targetRoom.Title, len(path)))

	// Return a command that starts the first tick after 1 second
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return autoWalkTickMsg{}
	})
}

// handleAutoWalkFailure handles a failed movement during auto-walk
func (m *Model) handleAutoWalkFailure() tea.Cmd {
	if !m.autoWalking {
		return nil
	}

	// Get the last direction we tried to go
	lastDirection := ""
	if m.autoWalkIndex > 0 && m.autoWalkIndex <= len(m.autoWalkPath) {
		lastDirection = m.autoWalkPath[m.autoWalkIndex-1]
	}

	m.output = append(m.output, "\x1b[91m[Auto-walk: Movement failed - cannot go that way]\x1b[0m")

	// Remove the exit from the current room first, before replanning
	if lastDirection != "" && m.worldMap.GetCurrentRoom() != nil {
		currentRoom := m.worldMap.GetCurrentRoom()
		m.output = append(m.output, fmt.Sprintf("\x1b[93m[Auto-walk: Removing invalid exit '%s' from current room]\x1b[0m", lastDirection))
		currentRoom.RemoveExit(lastDirection)
		m.worldMap.Save()
	}

	// Stop current auto-walk
	targetTitle := m.autoWalkTarget
	m.autoWalking = false
	m.autoWalkPath = nil
	m.autoWalkIndex = 0
	m.autoWalkTarget = ""

	// Try to replan the route to the same destination
	if targetTitle != "" {
		m.output = append(m.output, fmt.Sprintf("\x1b[93m[Auto-walk: Re-planning route to '%s']\x1b[0m", targetTitle))
		
		// Find the target room again
		rooms := m.worldMap.FindRooms(targetTitle)
		if len(rooms) == 0 {
			m.output = append(m.output, fmt.Sprintf("\x1b[91m[Auto-walk: Cannot find target room '%s']\x1b[0m", targetTitle))
			return nil
		}

		// Find a new path
		targetRoom := rooms[0]
		path := m.worldMap.FindPath(targetRoom.ID)

		if path == nil || len(path) == 0 {
			m.output = append(m.output, fmt.Sprintf("\x1b[91m[Auto-walk: No valid path found to '%s']\x1b[0m", targetTitle))
			return nil
		}

		// Restart auto-walking with the new path
		m.autoWalking = true
		m.autoWalkPath = path
		m.autoWalkIndex = 0
		m.autoWalkTarget = targetTitle
		m.output = append(m.output, fmt.Sprintf("\x1b[92m[Auto-walk: Restarting with new route (%d steps)]\x1b[0m", len(path)))
		
		// Start the first tick after 1 second
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return autoWalkTickMsg{}
		})
	}
	
	return nil
}

// handleTriggerCommand adds a new trigger
func (m *Model) handleTriggerCommand(command string) {
	// Parse the command to extract pattern and action
	// Expected format: /trigger "pattern" "action"

	// Remove "/trigger " prefix
	command = strings.TrimPrefix(command, "trigger ")
	command = strings.TrimSpace(command)

	// Parse quoted strings
	pattern, action, err := parseQuotedArgs(command)
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError: %v\x1b[0m", err))
		m.output = append(m.output, "\x1b[93mUsage: /trigger \"pattern\" \"action\"\x1b[0m")
		m.output = append(m.output, "\x1b[93mExample: /trigger \"hungry\" \"eat bread\"\x1b[0m")
		m.output = append(m.output, "\x1b[93mExample: /trigger \"The <subject> dies\" \"get <subject>\"\x1b[0m")
		return
	}

	// Add the trigger
	trigger, err := m.triggerManager.Add(pattern, action)
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError adding trigger: %v\x1b[0m", err))
		return
	}

	// Save triggers
	if err := m.triggerManager.Save(); err != nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError saving triggers: %v\x1b[0m", err))
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92mTrigger added: \"%s\" -> \"%s\"\x1b[0m", trigger.Pattern, trigger.Action))
}

// handleTriggersCommand handles /triggers list and /triggers remove
func (m *Model) handleTriggersCommand(args []string) {
	if len(args) == 0 {
		// Default to list
		m.handleTriggersListCommand()
		return
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "list":
		m.handleTriggersListCommand()
	case "remove":
		if len(args) < 2 {
			m.output = append(m.output, "\x1b[91mUsage: /triggers remove <index>\x1b[0m")
			return
		}
		var index int
		_, err := fmt.Sscanf(args[1], "%d", &index)
		if err != nil {
			m.output = append(m.output, "\x1b[91mError: Invalid index\x1b[0m")
			return
		}
		m.handleTriggersRemoveCommand(index)
	default:
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError: Unknown subcommand '%s'\x1b[0m", subCmd))
		m.output = append(m.output, "\x1b[93mUsage: /triggers [list|remove <index>]\x1b[0m")
	}
}

// handleTriggersListCommand lists all triggers
func (m *Model) handleTriggersListCommand() {
	if len(m.triggerManager.Triggers) == 0 {
		m.output = append(m.output, "\x1b[93mNo triggers defined.\x1b[0m")
		m.output = append(m.output, "\x1b[93mUse /trigger \"pattern\" \"action\" to add a trigger.\x1b[0m")
		return
	}

	m.output = append(m.output, "\x1b[92m=== Active Triggers ===\x1b[0m")
	for i, trigger := range m.triggerManager.Triggers {
		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. \"%s\" -> \"%s\"\x1b[0m", i+1, trigger.Pattern, trigger.Action))
	}
}

// handleTriggersRemoveCommand removes a trigger by index
func (m *Model) handleTriggersRemoveCommand(index int) {
	// Convert from 1-based to 0-based index
	index--

	if index < 0 || index >= len(m.triggerManager.Triggers) {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError: Invalid trigger index. Use /triggers list to see available triggers.\x1b[0m"))
		return
	}

	trigger := m.triggerManager.Triggers[index]
	if err := m.triggerManager.Remove(index); err != nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError removing trigger: %v\x1b[0m", err))
		return
	}

	// Save triggers
	if err := m.triggerManager.Save(); err != nil {
		m.output = append(m.output, fmt.Sprintf("\x1b[91mError saving triggers: %v\x1b[0m", err))
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92mRemoved trigger: \"%s\" -> \"%s\"\x1b[0m", trigger.Pattern, trigger.Action))
}

// parseQuotedArgs parses two quoted strings from a command
func parseQuotedArgs(input string) (string, string, error) {
	input = strings.TrimSpace(input)

	// Find first quoted string
	if !strings.HasPrefix(input, "\"") {
		return "", "", fmt.Errorf("expected quoted pattern")
	}

	// Find the closing quote for the first string
	endQuote := 1
	for endQuote < len(input) {
		if input[endQuote] == '"' && (endQuote == 1 || input[endQuote-1] != '\\') {
			break
		}
		endQuote++
	}

	if endQuote >= len(input) {
		return "", "", fmt.Errorf("unterminated pattern quote")
	}

	pattern := input[1:endQuote]

	// Find second quoted string
	rest := strings.TrimSpace(input[endQuote+1:])
	if !strings.HasPrefix(rest, "\"") {
		return "", "", fmt.Errorf("expected quoted action")
	}

	// Find the closing quote for the second string
	endQuote = 1
	for endQuote < len(rest) {
		if rest[endQuote] == '"' && (endQuote == 1 || rest[endQuote-1] != '\\') {
			break
		}
		endQuote++
	}

	if endQuote >= len(rest) {
		return "", "", fmt.Errorf("unterminated action quote")
	}

	action := rest[1:endQuote]

	return pattern, action, nil
}
