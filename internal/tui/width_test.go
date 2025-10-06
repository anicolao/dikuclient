package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestLayoutWidth verifies that the TUI layout uses the full terminal width correctly
func TestLayoutWidth(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)

	// Simulate various terminal widths
	testCases := []struct {
		width        int
		sidebarWidth int
	}{
		{120, 60},
		{100, 50},
		{140, 70},
		{80, 40},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			m.width = tc.width
			m.height = 40
			m.sidebarWidth = tc.sidebarWidth

			// Trigger window resize to update viewport widths
			_, _ = m.Update(tea.WindowSizeMsg{Width: tc.width, Height: 40})

			// Render the main content
			content := m.renderMainContent()

			// Measure the rendered width
			renderedWidth := lipgloss.Width(content)

			// The rendered content should match the terminal width
			if renderedWidth != tc.width {
				t.Errorf("Width mismatch: terminal=%d, rendered=%d, diff=%d",
					tc.width, renderedWidth, tc.width-renderedWidth)
			}
		})
	}
}

// TestMainPanelWidth verifies that the main panel width calculation is correct
func TestMainPanelWidth(t *testing.T) {
	// Main panel has left/top/bottom borders (no right border)
	// So: Width(X) renders to X + 1 total width

	testWidth := 59
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		BorderStyle(border).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(false). // No right border (welded to sidebar)
		BorderBottom(true).
		Width(testWidth)

	rendered := style.Render("Test content")
	renderedWidth := lipgloss.Width(rendered)
	expectedWidth := testWidth + 1 // +1 for left border only

	if renderedWidth != expectedWidth {
		t.Errorf("Main panel width: set Width(%d), expected rendered width %d, got %d",
			testWidth, expectedWidth, renderedWidth)
	}
}

// TestSidebarPanelWidth verifies that sidebar panel width calculation is correct
func TestSidebarPanelWidth(t *testing.T) {
	// Sidebar panels have all 4 borders and padding
	// Width(X) with borders renders to X + 2 total width

	targetWidth := 60
	testWidth := targetWidth - 2 // Subtract 2 to account for borders

	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().
		BorderStyle(border).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false). // No bottom for middle panels
		Padding(1).
		Width(testWidth)

	rendered := style.Render("Test content")
	renderedWidth := lipgloss.Width(rendered)

	if renderedWidth != targetWidth {
		t.Errorf("Sidebar panel width: set Width(%d), expected rendered width %d, got %d",
			testWidth, targetWidth, renderedWidth)
	}
}

// TestFullLayoutRendering tests that a full layout renders to the expected width
func TestFullLayoutRendering(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)

	// Test with 120x40 terminal
	m.width = 120
	m.height = 40
	m.sidebarWidth = 60

	// Initialize viewports with window size
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Get the full view
	view := m.View()

	// Measure dimensions
	viewWidth := lipgloss.Width(view)
	viewHeight := lipgloss.Height(view)

	// Log for debugging
	t.Logf("Terminal: 120x40")
	t.Logf("Rendered view: %dx%d", viewWidth, viewHeight)

	// The view should match the terminal width exactly
	if viewWidth != 120 {
		t.Errorf("View width mismatch: expected 120, got %d (off by %d)", viewWidth, 120-viewWidth)
	}

	// Note: Height may differ due to other layout factors, we only test width here
}

// TestMainPanelSidebarHeightMatch verifies that the main panel and sidebar have the same height
func TestMainPanelSidebarHeightMatch(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)

	// Test all terminal heights from 40 to 80 to ensure the fix works for all cases
	// Generate test cases in a loop for readability
	testCases := []struct {
		width  int
		height int
		name   string
	}{}
	
	for height := 40; height <= 80; height++ {
		testCases = append(testCases, struct {
			width  int
			height int
			name   string
		}{
			width:  120,
			height: height,
			name:   fmt.Sprintf("height %d", height),
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m.width = tc.width
			m.height = tc.height
			m.sidebarWidth = 60
			m.connected = true

			// Trigger window resize to update viewport heights
			_, _ = m.Update(tea.WindowSizeMsg{Width: tc.width, Height: tc.height})

			// Calculate what the expected heights are after the fix
			headerHeight := 5
			contentHeight := tc.height - headerHeight
			panelHeight := contentHeight / 4
			// Both sidebar and main panel should have height = 4 * panelHeight + 5
			// (Sidebar: 4 panels with borders add 3*1 + 1*2 = 5)
			// (Main: actualContentHeight = 4 * panelHeight + 3, + 2 for borders = 4 * panelHeight + 5)
			expectedHeight := 4*panelHeight + 5
			if expectedHeight > tc.height {
				t.Errorf("Expected height %d exceeds terminal height %d", expectedHeight, tc.height)
			}

			// Render the main content which includes both main panel and sidebar
			content := m.renderMainContent()

			// The rendered content height should match the expected height
			renderedHeight := lipgloss.Height(content)

			// The main panel and sidebar should be exactly the same height
			if renderedHeight != expectedHeight {
				t.Errorf("Rendered height mismatch: terminal height=%d, contentHeight=%d, panelHeight=%d, expected=%d, rendered=%d",
					tc.height, contentHeight, panelHeight, expectedHeight, renderedHeight)
			}

			// Verify gameOutput and sidebar heights match expected height exactly
			gameOutputHeight := lipgloss.Height(m.lastRenderedGameOutput)
			sidebarHeight := lipgloss.Height(m.lastRenderedSidebar)
			
			if gameOutputHeight != expectedHeight {
				t.Errorf("GameOutput height mismatch: terminal height=%d, expected=%d, gameOutput=%d",
					tc.height, expectedHeight, gameOutputHeight)
			}
			
			if sidebarHeight != expectedHeight {
				t.Errorf("Sidebar height mismatch: terminal height=%d, expected=%d, sidebar=%d",
					tc.height, expectedHeight, sidebarHeight)
			}
			
			// Ensure gameOutput and sidebar have the same height
			if gameOutputHeight != sidebarHeight {
				t.Errorf("GameOutput and Sidebar heights don't match: terminal height=%d, gameOutput=%d, sidebar=%d",
					tc.height, gameOutputHeight, sidebarHeight)
			}
		})
	}
}
