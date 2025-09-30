package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anicolao/dikuclient/internal/client"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	conn          *client.Connection
	viewport      viewport.Model
	output        []string
	currentInput  string
	cursorPos     int
	width         int
	height        int
	connected     bool
	host          string
	port          int
	sidebarWidth  int
	err           error
	mudLogFile    *os.File
	tuiLogFile    *os.File
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

// NewModel creates a new application model
func NewModel(host string, port int, mudLogFile, tuiLogFile *os.File) Model {
	vp := viewport.New(0, 0)
	// Don't apply any style to viewport - let ANSI codes pass through naturally

	return Model{
		viewport:     vp,
		output:       []string{},
		currentInput: "",
		cursorPos:    0,
		host:         host,
		port:         port,
		sidebarWidth: 30,
		mudLogFile:   mudLogFile,
		tuiLogFile:   tuiLogFile,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return m.connect
}

// connect establishes a connection to the MUD server
func (m Model) connect() tea.Msg {
	conn, err := client.NewConnection(m.host, m.port)
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
				// Send command (even if empty - user may want to send blank line)
				m.conn.Send(m.currentInput)
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
		}
		m.updateViewport()
		return m, m.listenForMessages

	case errMsg:
		m.err = msg
		m.output = append(m.output, fmt.Sprintf("Error: %v", msg))
		m.updateViewport()
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
		
		if m.currentInput != "" || m.connected {
			// Build input line with cursor
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
		} else {
			content = strings.Join(m.output, "\n")
		}
	} else {
		// No output yet, just show cursor if connected
		if m.currentInput != "" || m.connected {
			inputLine := m.currentInput
			if m.cursorPos < len(m.currentInput) {
				inputLine = m.currentInput[:m.cursorPos] + "█" + m.currentInput[m.cursorPos:]
			} else {
				inputLine = m.currentInput + "█"
			}
			// Use bright yellow for better visibility
			content = "\x1b[93m" + inputLine + "\x1b[0m"
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
