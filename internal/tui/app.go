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
	"github.com/anicolao/dikuclient/internal/xpstats"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Model represents the application state
type Model struct {
	conn                  *client.Connection
	viewport              viewport.Model
	output                []string
	currentInput          string
	cursorPos             int
	width                 int
	height                int
	connected             bool
	host                  string
	port                  int
	sidebarWidth          int
	err                   error
	mudLogFile            *os.File
	tuiLogFile            *os.File
	telnetDebugLog        *os.File // Debug log for telnet/UTF-8 processing
	echoSuppressed        bool     // Server has disabled echo (e.g., for passwords)
	username              string
	password              string
	autoLoginState        int                // 0=idle, 1=sent username, 2=sent password
	worldMap              *mapper.Map        // World map for navigation
	recentOutput          []string           // Buffer for recent output to detect rooms
	pendingMovement       string             // Last movement command sent
	mapDebug              bool               // Enable mapper debug output
	autoWalking           bool               // Currently auto-walking with /go
	autoWalkPath          []string           // Path to auto-walk
	autoWalkIndex         int                // Current step in auto-walk
	lastRoomSearch        []*mapper.Room     // Last room search results for disambiguation
	triggerManager        *triggers.Manager  // Trigger manager
	inventory             []string           // Current inventory items
	inventoryTime         time.Time          // Time when inventory was last updated
	inventoryViewport     viewport.Model     // Viewport for scrollable inventory
	tells                 []string           // Recent tells received
	tellsViewport         viewport.Model     // Viewport for scrollable tells
	skipNextRoomDetection bool               // Skip next room detection (e.g., after recall teleport)
	autoWalkTarget        string             // Target room title for auto-walk (for recovery)
	mapLegend             map[string]int     // Room ID to number mapping for map legend display
	mapLegendRooms        []*mapper.Room     // Rooms in the current legend (for /go command)
	xpTracking            map[string]*XPStat // XP/s tracking per creature (current session)
	pendingKill           string             // Last kill command target
	killTime              time.Time          // Time when kill command was sent
	xpViewport            viewport.Model     // Viewport for scrollable XP stats
	xpStatsManager        *xpstats.Manager   // Persistent XP stats manager
	webSessionID          string             // Web session ID for sharing (empty if not in web mode)
	webServerURL          string             // Web server URL for sharing (empty if not in web mode)
	isSplit               bool               // Whether the main viewport is split
	splitViewport         viewport.Model     // Second viewport for tracking live output when split
}

// XPStat represents XP per second statistics for a creature
type XPStat struct {
	CreatureName string
	XP           int
	Seconds      float64
	XPPerSecond  float64
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

	// Load or create XP stats manager
	xpStatsManager, err := xpstats.Load()
	if err != nil {
		// If we can't load XP stats, create a new manager
		xpStatsManager = xpstats.NewManager()
	}

	inventoryVp := viewport.New(0, 0)
	tellsVp := viewport.New(0, 0)
	xpVp := viewport.New(0, 0)
	splitVp := viewport.New(0, 0)

	// Read web session information from environment variables
	webSessionID := os.Getenv("DIKUCLIENT_WEB_SESSION_ID")
	webServerURL := os.Getenv("DIKUCLIENT_WEB_SERVER_URL")

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
		xpTracking:        make(map[string]*XPStat),
		xpViewport:        xpVp,
		xpStatsManager:    xpStatsManager,
		webSessionID:      webSessionID,
		webServerURL:      webServerURL,
		isSplit:           false,
		splitViewport:     splitVp,
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

		case tea.KeyPgUp:
			// Scroll up in the main viewport
			m.viewport.ViewUp()
			// Check if we need to enable split mode
			if !m.isSplit && !m.viewport.AtBottom() {
				m.isSplit = true
			}
			return m, nil

		case tea.KeyPgDown:
			// Scroll down in the main viewport
			m.viewport.ViewDown()
			// Check if we should disable split mode
			if m.isSplit && m.viewport.AtBottom() {
				m.isSplit = false
			}
			return m, nil

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
					// Clear map legend on movement
					m.mapLegend = nil
					m.mapLegendRooms = nil
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

		// Set up split viewport dimensions (1/3 of main viewport height)
		m.splitViewport.Width = mainWidth
		m.splitViewport.Height = (m.height - headerHeight - 2) / 3

		// Update viewport sizes for 4 panels
		panelHeight := (m.height - headerHeight - 2 - 8) / 4
		m.inventoryViewport.Width = sidebarWidth - 4 // Account for borders and padding
		m.inventoryViewport.Height = panelHeight - 4 // Account for header, timestamp, and borders

		// Update tells viewport size
		m.tellsViewport.Width = sidebarWidth - 4 // Account for borders and padding
		m.tellsViewport.Height = panelHeight - 4 // Account for header and borders

		// Update XP viewport size
		m.xpViewport.Width = sidebarWidth - 4 // Account for borders and padding
		m.xpViewport.Height = panelHeight - 4 // Account for header and borders

		m.updateViewport()
		return m, nil

	case tea.MouseMsg:
		// Handle mouse wheel scrolling on main viewport
		if msg.Action == tea.MouseActionPress {
			if msg.Button == tea.MouseButtonWheelUp {
				m.viewport.ViewUp()
				// Check if we need to enable split mode
				if !m.isSplit && !m.viewport.AtBottom() {
					m.isSplit = true
				}
				return m, nil
			} else if msg.Button == tea.MouseButtonWheelDown {
				m.viewport.ViewDown()
				// Check if we should disable split mode
				if m.isSplit && m.viewport.AtBottom() {
					m.isSplit = false
				}
				return m, nil
			}
		}
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

			// Check for combat prompt to track XP/s
			m.detectCombatPrompt(line)

			// Check for XP tracking events (death message and XP gain)
			m.detectXPEvents(line)

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

	// Check if viewport is at bottom BEFORE setting new content
	wasAtBottom := m.viewport.AtBottom()
	
	m.viewport.SetContent(content)
	
	// If not in split mode or if viewport is already at bottom, go to bottom
	// This preserves scroll position when in split mode
	if !m.isSplit {
		m.viewport.GotoBottom()
	} else if wasAtBottom {
		// If user was already at bottom and new content arrived, exit split mode
		m.isSplit = false
		m.viewport.GotoBottom()
	}

	// Update split viewport content (always stays at bottom for live tracking)
	m.splitViewport.SetContent(content)
	m.splitViewport.GotoBottom()

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
	mainWidth := m.width - sidebarWidth - 4
	contentHeight := m.height - headerHeight

	// Build title for main window with current room and exits
	mainTitle := ""
	if m.worldMap != nil {
		currentRoom := m.worldMap.GetCurrentRoom()
		if currentRoom != nil {
			exitList := make([]string, 0, len(currentRoom.Exits))
			for dir := range currentRoom.Exits {
				exitList = append(exitList, dir)
			}
			sort.Strings(exitList)
			exitsStr := strings.Join(exitList, ", ")
			if exitsStr == "" {
				exitsStr = "none"
			}
			mainTitle = currentRoom.Title + " [" + exitsStr + "]"
		}
	}

	// Create custom border with title embedded in top border
	customBorder := lipgloss.RoundedBorder()
	if mainTitle != "" {
		// Build a custom top border line with the title embedded
		// Format: "─ Title ─────────────..."
		titleWithSpaces := "── " + mainTitle + " ──"
		availableWidth := mainWidth
		if len(titleWithSpaces) < availableWidth {
			// Fill remaining space with border characters
			remainingChars := availableWidth - len(titleWithSpaces)
			customBorder.Top = titleWithSpaces + strings.Repeat("─", remainingChars+10)
		} else {
			// Title is too long, truncate it
			customBorder.Top = titleWithSpaces[:availableWidth]
		}
	}

	var gameOutput string

	if m.isSplit {
		// Split mode: 2/3 for user scrolled position, 1/3 for live output at bottom
		topHeight := (contentHeight * 2) / 3
		bottomHeight := contentHeight - topHeight

		// Top viewport (user's scrolled position)
		topBorderStyle := lipgloss.NewStyle().
			BorderStyle(customBorder).
			BorderForeground(lipgloss.Color("62")).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(false).
			BorderBottom(false)

		topView := topBorderStyle.
			Width(mainWidth).
			Height(topHeight).
			Render(m.viewport.View())

		// Bottom viewport (live output - always at bottom)
		bottomBorder := lipgloss.RoundedBorder()
		bottomBorder.Top = strings.Repeat("─", mainWidth+10)

		bottomBorderStyle := lipgloss.NewStyle().
			BorderStyle(bottomBorder).
			BorderForeground(lipgloss.Color("62")).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(false).
			BorderBottom(true)

		bottomView := bottomBorderStyle.
			Width(mainWidth).
			Height(bottomHeight).
			Render(m.splitViewport.View())

		gameOutput = lipgloss.JoinVertical(lipgloss.Left, topView, bottomView)
	} else {
		// Normal mode: single viewport
		mainBorderStyle := lipgloss.NewStyle().
			BorderStyle(customBorder).
			BorderForeground(lipgloss.Color("62")).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(false).
			BorderBottom(true)

		gameOutput = mainBorderStyle.
			Width(mainWidth).
			Height(contentHeight).
			Render(m.viewport.View())
	}

	// Sidebar with panels
	sidebar := m.renderSidebar(sidebarWidth, contentHeight)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		gameOutput,
		sidebar,
	)
}

// Helper function to create a custom border with title embedded in top border
// position: "top" for top panel (Tells), "middle" for middle panels, "bottom" for bottom panel (Map)
func createBorderWithTitle(title string, panelWidth int, position string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	if title != "" {
		titleWithSpaces := "── " + title + " "
		availableWidth := panelWidth
		if len(titleWithSpaces) < availableWidth {
			remainingChars := availableWidth - len(titleWithSpaces)
			border.Top = titleWithSpaces + strings.Repeat("─", remainingChars)
		} else {
			border.Top = titleWithSpaces[:availableWidth]
		}
	}

	// Set appropriate corners based on position in stack
	switch position {
	case "top":
		// Top panel: use ┬ for top-left to weld with main panel
		border.TopLeft = "┬"
	case "middle":
		// Middle panels: use ├ and ┤ for vertical welding
		border.TopLeft = "├"
		border.TopRight = "┤"
	case "bottom":
		// Bottom panel: use ├ for top corners and ┴ for bottom-left to weld with main panel
		border.TopLeft = "├"
		border.TopRight = "┤"
		border.BottomLeft = "┴"
	}

	return border
}

func (m *Model) renderSidebar(width, height int) string {
	panelHeight := height / 4

	// Tells panel with scrollable viewport
	var tellsContent string
	if len(m.tells) > 0 {
		tellsContent = strings.Join(m.tells, "\n")
	} else {
		tellsContent = emptyPanelStyle.Render("(no tells yet)")
	}
	m.tellsViewport.SetContent(tellsContent)

	tellsBorder := createBorderWithTitle("Tells", width, "top") // Top panel uses ┬ for top-right corner
	tellsStyle := lipgloss.NewStyle().
		BorderStyle(tellsBorder).
		BorderForeground(lipgloss.Color("62")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		Padding(1)

	tellsPanel := tellsStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(m.tellsViewport.View())

	// XP/s panel with scrollable viewport - shows persistent averaged stats
	var xpContent string
	if m.xpStatsManager != nil && len(m.xpStatsManager.GetAllStats()) > 0 {
		// Convert map to slice and sort by XP/s (best to worst)
		allStats := m.xpStatsManager.GetAllStats()
		stats := make([]*xpstats.XPStat, 0, len(allStats))
		for _, stat := range allStats {
			stats = append(stats, stat)
		}
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].XPPerSecond > stats[j].XPPerSecond
		})

		// Create a lipgloss table for XP stats
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true).
			BorderHeader(true).
			BorderColumn(true).
			BorderRow(false).
			Headers("Creature", "XP/s", "Samples").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Padding(0, 1).Align(lipgloss.Center)
				}
				// Left align creature names, right align numbers
				if col == 0 {
					return lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
				}
				return lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Right)
			})

		// Add data rows
		for _, stat := range stats {
			t.Row(stat.CreatureName, fmt.Sprintf("%.1f", stat.XPPerSecond), fmt.Sprintf("%d", stat.SampleCount))
		}

		xpContent = t.String()
	} else {
		xpContent = emptyPanelStyle.Render("(no kills yet)")
	}
	m.xpViewport.SetContent(xpContent)

	xpBorder := createBorderWithTitle("XP/s (avg)", width, "middle") // Middle panel uses T-junction corners
	xpStyle := lipgloss.NewStyle().
		BorderStyle(xpBorder).
		BorderForeground(lipgloss.Color("62")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		Padding(1)

	xpPanel := xpStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(m.xpViewport.View())

	// Inventory panel with scrollable viewport
	var inventoryContent string
	inventoryTitle := "Inventory"
	if len(m.inventory) > 0 {
		timeStr := m.inventoryTime.Format("15:04:05")
		inventoryTitle = "Inventory (" + timeStr + ")"
		inventoryContent = strings.Join(m.inventory, "\n")
	} else {
		inventoryContent = emptyPanelStyle.Render("(not populated)")
	}
	m.inventoryViewport.SetContent(inventoryContent)

	inventoryBorder := createBorderWithTitle(inventoryTitle, width, "middle") // Middle panel uses T-junction corners
	inventoryStyle := lipgloss.NewStyle().
		BorderStyle(inventoryBorder).
		BorderForeground(lipgloss.Color("62")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		Padding(1)

	inventoryPanel := inventoryStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(m.inventoryViewport.View())

	// Map panel
	var mapContent string
	mapTitle := "Map"

	if m.worldMap == nil {
		mapContent = emptyPanelStyle.Render("(not implemented)")
	} else {
		currentRoom := m.worldMap.GetCurrentRoom()
		if currentRoom == nil {
			mapContent = emptyPanelStyle.Render("(exploring...)")
		} else {
			mapTitle = currentRoom.Title
			// Calculate available height for map content
			mapHeight := panelHeight - 2
			mapContent = m.worldMap.FormatMapPanelWithLegend(width-4, mapHeight, m.mapLegend)
		}
	}

	mapBorder := createBorderWithTitle(mapTitle, width, "bottom") // Bottom panel uses ┴ for bottom-left corner
	mapStyle := lipgloss.NewStyle().
		BorderStyle(mapBorder).
		BorderForeground(lipgloss.Color("62")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		Padding(1)

	mapPanel := mapStyle.
		Width(width - 2).
		Height(panelHeight).
		Render(mapContent)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tellsPanel,
		xpPanel,
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

// combatPromptRegex matches combat prompts in format: [Hero:Status] [Target:Status]
// Example: 101H 132V 54710X 49.60% 570C [Osric:V.Bad] [a goblin scout:Good] T:24 Exits:NS>
var combatPromptRegex = regexp.MustCompile(`\[([^:]+):[^\]]+\]\s*\[([^:]+):[^\]]+\]`)

// deathMessageRegex matches death messages in format: The <target> is dead!
var deathMessageRegex = regexp.MustCompile(`^(The|A|An)\s+(.+?)\s+is dead!`)

// xpGainRegex matches XP gain messages in format: You receive [0-9]+ experience.
var xpGainRegex = regexp.MustCompile(`^You receive (\d+) experience\.`)

// detectCombatPrompt detects combat status in the prompt
func (m *Model) detectCombatPrompt(line string) {
	cleanLine := stripANSI(line)
	matches := combatPromptRegex.FindStringSubmatch(cleanLine)
	if matches != nil && len(matches) == 3 {
		// matches[1] is the hero name, matches[2] is the target name
		target := strings.ToLower(strings.TrimSpace(matches[2]))

		// Only start tracking if we don't have a pending kill or if this is a new target
		if m.pendingKill == "" || m.pendingKill != target {
			m.pendingKill = target
			m.killTime = time.Now()
		}
	}
}

// detectXPEvents detects death messages and XP gains to calculate XP/s
func (m *Model) detectXPEvents(line string) {
	cleanLine := stripANSI(line)

	// Check for death message
	if m.pendingKill != "" {
		matches := deathMessageRegex.FindStringSubmatch(cleanLine)
		if matches != nil && len(matches) == 3 {
			// matches[1] is the article (The/A/An), matches[2] is the creature name
			creatureName := strings.ToLower(strings.TrimSpace(matches[2]))
			// Check if this matches our pending kill
			if strings.Contains(creatureName, m.pendingKill) {
				// Store the death time, but don't finalize yet - wait for XP gain
				m.pendingKill = creatureName
			}
		}
	}

	// Check for XP gain
	if m.pendingKill != "" {
		matches := xpGainRegex.FindStringSubmatch(cleanLine)
		if matches != nil && len(matches) == 2 {
			xp := 0
			fmt.Sscanf(matches[1], "%d", &xp)

			// Calculate time elapsed
			deathTime := time.Now()
			seconds := deathTime.Sub(m.killTime).Seconds()

			// Calculate XP/s
			xpPerSecond := 0.0
			if seconds > 0 {
				xpPerSecond = float64(xp) / seconds
			}

			// Store in current session tracking
			m.xpTracking[m.pendingKill] = &XPStat{
				CreatureName: m.pendingKill,
				XP:           xp,
				Seconds:      seconds,
				XPPerSecond:  xpPerSecond,
			}

			// Update persistent stats with EMA
			if m.xpStatsManager != nil {
				m.xpStatsManager.UpdateStat(m.pendingKill, xpPerSecond)
				// Save to disk (ignore errors to not disrupt gameplay)
				_ = m.xpStatsManager.Save()
			}

			// Clear pending kill
			m.pendingKill = ""
		}
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

	// Clear map legend unless we're executing nearby or legend commands
	if cmd != "nearby" && cmd != "legend" {
		m.mapLegend = nil
		m.mapLegendRooms = nil
	}

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
	case "nearby":
		m.handleNearbyCommand()
		return nil
	case "legend":
		m.handleLegendCommand()
		return nil
	case "go":
		return m.handleGoCommand(args)
	case "trigger":
		m.handleTriggerCommand(command)
		return nil
	case "triggers":
		m.handleTriggersCommand(args)
		return nil
	case "share":
		m.handleShareCommand()
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

// handleShareCommand generates a shareable URL for web sessions
func (m *Model) handleShareCommand() {
	if m.webSessionID == "" || m.webServerURL == "" {
		m.output = append(m.output, "\x1b[91mError: /share command is only available in web mode\x1b[0m")
		m.output = append(m.output, "\x1b[90mStart the client with --web flag to enable session sharing\x1b[0m")
		return
	}

	shareURL := fmt.Sprintf("%s/?id=%s", m.webServerURL, m.webSessionID)
	m.output = append(m.output, "\x1b[92m=== Share This Session ===\x1b[0m")
	m.output = append(m.output, fmt.Sprintf("\x1b[96m%s\x1b[0m", shareURL))
	m.output = append(m.output, "")
	m.output = append(m.output, "\x1b[90mAnyone who opens this URL will see and control the same session\x1b[0m")
}

// handleHelpCommand shows available client commands
func (m *Model) handleHelpCommand() {
	m.output = append(m.output, "\x1b[92m=== Client Commands ===\x1b[0m")
	m.output = append(m.output, "  \x1b[96m/point <room>\x1b[0m            - Show next direction to reach a room")
	m.output = append(m.output, "  \x1b[96m/wayfind <room>\x1b[0m         - Show full path to reach a room")
	m.output = append(m.output, "  \x1b[96m/go <room>\x1b[0m              - Auto-walk to a room (one step per second)")
	m.output = append(m.output, "  \x1b[96m/map\x1b[0m                    - Show map information")
	m.output = append(m.output, "  \x1b[96m/rooms [filter]\x1b[0m         - List all known rooms (optionally filtered)")
	m.output = append(m.output, "  \x1b[96m/nearby\x1b[0m                 - List all rooms within 5 steps")
	m.output = append(m.output, "  \x1b[96m/legend\x1b[0m                 - List all rooms currently on the map")
	m.output = append(m.output, "  \x1b[96m/trigger \"pat\" \"act\"\x1b[0m - Add a trigger (pattern can use <var>)")
	m.output = append(m.output, "  \x1b[96m/triggers list\x1b[0m          - List all triggers")
	m.output = append(m.output, "  \x1b[96m/triggers remove <n>\x1b[0m    - Remove trigger by number")
	m.output = append(m.output, "  \x1b[96m/share\x1b[0m                  - Get shareable URL (web mode only)")
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

	// Sort rooms by durable room number for consistent display
	sort.Slice(roomsToDisplay, func(i, j int) bool {
		numI := m.worldMap.GetRoomNumber(roomsToDisplay[i].ID)
		numJ := m.worldMap.GetRoomNumber(roomsToDisplay[j].ID)
		return numI < numJ
	})

	// Store results for later disambiguation
	m.lastRoomSearch = roomsToDisplay

	// Display rooms with durable numbers
	for _, room := range roomsToDisplay {
		exitList := make([]string, 0, len(room.Exits))
		for dir := range room.Exits {
			exitList = append(exitList, dir)
		}
		sort.Strings(exitList)

		exitsStr := strings.Join(exitList, ", ")
		if exitsStr == "" {
			exitsStr = "none"
		}

		// Use durable room number
		roomNum := m.worldMap.GetRoomNumber(room.ID)
		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m \x1b[90m[%s]\x1b[0m", roomNum, room.Title, exitsStr))
	}
}

// handleNearbyCommand lists all rooms within 5 steps of current location
func (m *Model) handleNearbyCommand() {
	currentRoom := m.worldMap.GetCurrentRoom()
	if currentRoom == nil {
		m.output = append(m.output, "\x1b[91mNo current room. You need to be in a mapped location.\x1b[0m")
		return
	}

	nearby := m.worldMap.FindNearbyRooms(5)

	if len(nearby) == 0 {
		m.output = append(m.output, "\x1b[93mNo nearby rooms found within 5 steps.\x1b[0m")
		return
	}

	// Get visible room IDs on the map display
	// Use typical sidebar dimensions for determining visibility
	visibleRoomIDs := m.worldMap.GetVisibleRoomIDs(30, 15)
	visibleSet := make(map[string]bool)
	for _, id := range visibleRoomIDs {
		visibleSet[id] = true
	}

	// Filter nearby rooms to only those visible on the map
	var filteredNearby []mapper.NearbyRoom
	for _, nr := range nearby {
		if visibleSet[nr.Room.ID] {
			filteredNearby = append(filteredNearby, nr)
		}
	}

	if len(filteredNearby) == 0 {
		m.output = append(m.output, "\x1b[93mNo nearby rooms are currently visible on the map.\x1b[0m")
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92m=== Nearby Rooms (%d visible on map) ===\x1b[0m", len(filteredNearby)))

	// Build room legend mapping for map display and store rooms for /go
	m.mapLegend = make(map[string]int)
	m.mapLegendRooms = make([]*mapper.Room, 0, len(filteredNearby))

	currentDistance := -1
	for i, nr := range filteredNearby {
		// Show distance header when it changes
		if nr.Distance != currentDistance {
			currentDistance = nr.Distance
			stepLabel := "step"
			if currentDistance > 1 {
				stepLabel = "steps"
			}
			m.output = append(m.output, fmt.Sprintf("\x1b[93m%d %s away:\x1b[0m", currentDistance, stepLabel))
		}

		// Get exits for display
		exitList := make([]string, 0, len(nr.Room.Exits))
		for dir := range nr.Room.Exits {
			exitList = append(exitList, dir)
		}
		sort.Strings(exitList)

		exitsStr := strings.Join(exitList, ", ")
		if exitsStr == "" {
			exitsStr = "none"
		}

		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m \x1b[90m[%s]\x1b[0m", i+1, nr.Room.Title, exitsStr))

		// Add to legend mapping and store room
		m.mapLegend[nr.Room.ID] = i + 1
		m.mapLegendRooms = append(m.mapLegendRooms, nr.Room)
	}
}

// handleLegendCommand lists all rooms currently on the map using durable room numbers
func (m *Model) handleLegendCommand() {
	allRooms := m.worldMap.GetAllRooms()

	if len(allRooms) == 0 {
		m.output = append(m.output, "\x1b[93mNo rooms have been explored yet.\x1b[0m")
		return
	}

	// Get visible room IDs on the map display
	// Use typical sidebar dimensions for determining visibility
	visibleRoomIDs := m.worldMap.GetVisibleRoomIDs(30, 15)
	visibleSet := make(map[string]bool)
	for _, id := range visibleRoomIDs {
		visibleSet[id] = true
	}

	// Build list of visible rooms with their durable numbers
	type roomWithNumber struct {
		room   *mapper.Room
		number int
	}
	var visibleRooms []roomWithNumber

	// Iterate through room numbering to maintain order
	for _, roomID := range m.worldMap.RoomNumbering {
		if visibleSet[roomID] {
			if room := m.worldMap.Rooms[roomID]; room != nil {
				number := m.worldMap.GetRoomNumber(roomID)
				visibleRooms = append(visibleRooms, roomWithNumber{room: room, number: number})
			}
		}
	}

	if len(visibleRooms) == 0 {
		m.output = append(m.output, "\x1b[93mNo rooms are currently visible on the map.\x1b[0m")
		return
	}

	m.output = append(m.output, fmt.Sprintf("\x1b[92m=== Rooms on Map (%d visible) ===\x1b[0m", len(visibleRooms)))

	// Build room legend mapping for map display using durable numbers
	m.mapLegend = make(map[string]int)
	m.mapLegendRooms = make([]*mapper.Room, 0)

	// Track mapping from display position to room for /go command
	displayPosToRoom := make(map[int]*mapper.Room)

	for i, rn := range visibleRooms {
		// Get exits for display
		exitList := make([]string, 0, len(rn.room.Exits))
		for dir := range rn.room.Exits {
			exitList = append(exitList, dir)
		}
		sort.Strings(exitList)

		exitsStr := strings.Join(exitList, ", ")
		if exitsStr == "" {
			exitsStr = "none"
		}

		// Display with durable room number
		m.output = append(m.output, fmt.Sprintf("  \x1b[96m%d. %s\x1b[0m \x1b[90m[%s]\x1b[0m", rn.number, rn.room.Title, exitsStr))

		// Use durable number for legend mapping
		m.mapLegend[rn.room.ID] = rn.number

		// Map display position (1-indexed) to room for /go command
		displayPosToRoom[i+1] = rn.room
	}

	// Store rooms in display order for /go to use
	for i := 1; i <= len(displayPosToRoom); i++ {
		if room, ok := displayPosToRoom[i]; ok {
			m.mapLegendRooms = append(m.mapLegendRooms, room)
		}
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

		// If only a number is provided, try different sources
		if len(args) == 1 {
			// Try mapLegendRooms first (from /nearby - temporary numbers)
			if len(m.mapLegendRooms) > 0 {
				if index < 1 || index > len(m.mapLegendRooms) {
					m.output = append(m.output, fmt.Sprintf("\x1b[91mInvalid room number. Must be between 1 and %d.\x1b[0m", len(m.mapLegendRooms)))
					return nil
				}
				rooms = []*mapper.Room{m.mapLegendRooms[index-1]}
			} else if len(m.lastRoomSearch) > 0 {
				// Check if this is a durable number from /rooms or /legend
				// Try to find the room by its durable number in lastRoomSearch
				found := false
				for _, room := range m.lastRoomSearch {
					if m.worldMap.GetRoomNumber(room.ID) == index {
						rooms = []*mapper.Room{room}
						found = true
						break
					}
				}
				if !found {
					m.output = append(m.output, fmt.Sprintf("\x1b[91mRoom number %d not found in previous search results.\x1b[0m", index))
					return nil
				}
			} else {
				// Try durable room number lookup as last resort
				room := m.worldMap.GetRoomByNumber(index)
				if room != nil {
					rooms = []*mapper.Room{room}
				} else {
					m.output = append(m.output, "\x1b[91mNo previous room search to select from. Use /rooms, /nearby, or /legend to see room listings.\x1b[0m")
					return nil
				}
			}
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
