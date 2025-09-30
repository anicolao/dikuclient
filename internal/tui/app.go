package tui

import (
	"fmt"
	"strings"

	"github.com/anicolao/dikuclient/internal/client"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	conn          *client.Connection
	viewport      viewport.Model
	textarea      textarea.Model
	output        []string
	width         int
	height        int
	connected     bool
	host          string
	port          int
	sidebarWidth  int
	err           error
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
func NewModel(host string, port int) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your command here..."
	ta.Focus()
	ta.CharLimit = 200
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	vp := viewport.New(0, 0)

	return Model{
		textarea:     ta,
		viewport:     vp,
		output:       []string{},
		host:         host,
		port:         port,
		sidebarWidth: 30,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.connect)
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
				value := m.textarea.Value()
				if value != "" {
					m.conn.Send(value)
					m.output = append(m.output, fmt.Sprintf("> %s", value))
					m.textarea.Reset()
					m.viewport.SetContent(strings.Join(m.output, "\n"))
					m.viewport.GotoBottom()
				}
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 5
		sidebarWidth := m.sidebarWidth
		mainWidth := m.width - sidebarWidth - 6

		m.viewport.Width = mainWidth
		m.viewport.Height = m.height - headerHeight - footerHeight - 2
		m.textarea.SetWidth(mainWidth)

		return m, nil

	case *client.Connection:
		m.conn = msg
		m.connected = true
		m.output = append(m.output, fmt.Sprintf("Connected to %s:%d", m.host, m.port))
		m.viewport.SetContent(strings.Join(m.output, "\n"))
		return m, m.listenForMessages

	case mudMsg:
		// Add message to output - it already has proper line endings
		msgStr := string(msg)
		// Split into lines and add them individually to preserve formatting
		lines := strings.Split(msgStr, "\n")
		for i, line := range lines {
			// Don't add empty line at the end if message ended with \n
			if i == len(lines)-1 && line == "" {
				continue
			}
			m.output = append(m.output, line)
		}
		m.viewport.SetContent(strings.Join(m.output, "\n"))
		m.viewport.GotoBottom()
		return m, m.listenForMessages

	case errMsg:
		m.err = msg
		m.output = append(m.output, fmt.Sprintf("Error: %v", msg))
		m.viewport.SetContent(strings.Join(m.output, "\n"))
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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

	// Input area
	inputArea := m.renderInputArea()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		mainContent,
		inputArea,
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
	footerHeight := 5
	sidebarWidth := m.sidebarWidth
	mainWidth := m.width - sidebarWidth - 6
	contentHeight := m.height - headerHeight - footerHeight - 2

	// Game output viewport
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

func (m Model) renderInputArea() string {
	return mainStyle.Render(m.textarea.View())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
