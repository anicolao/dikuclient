package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestDetectAndParseTell verifies that tell messages are correctly detected and parsed
func TestDetectAndParseTell(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTells int
		expectedEntry string
	}{
		{
			name:          "basic tell",
			input:         "Alice tells you 'hello there'",
			expectedTells: 1,
			expectedEntry: "Alice: hello there",
		},
		{
			name:          "tell with spaces in name",
			input:         "Bob Smith tells you 'how are you?'",
			expectedTells: 1,
			expectedEntry: "Bob Smith: how are you?",
		},
		{
			name:          "tell with punctuation in content",
			input:         "Charlie tells you 'Hello! How are you doing?'",
			expectedTells: 1,
			expectedEntry: "Charlie: Hello! How are you doing?",
		},
		{
			name:          "tell with ANSI codes",
			input:         "\x1b[32mDave tells you 'test message'\x1b[0m",
			expectedTells: 1,
			expectedEntry: "Dave: test message",
		},
		{
			name:          "not a tell - missing format",
			input:         "Someone says 'hello'",
			expectedTells: 0,
			expectedEntry: "",
		},
		{
			name:          "not a tell - wrong format",
			input:         "tells you 'hello'",
			expectedTells: 0,
			expectedEntry: "",
		},
		{
			name:          "tell with empty content",
			input:         "Eve tells you ''",
			expectedTells: 1,
			expectedEntry: "Eve: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				tells:    []string{},
				worldMap: mapper.NewMap(),
			}

			m.detectAndParseTell(tt.input)

			if len(m.tells) != tt.expectedTells {
				t.Errorf("Expected %d tells, got %d", tt.expectedTells, len(m.tells))
			}

			if tt.expectedTells > 0 && len(m.tells) > 0 {
				if m.tells[0] != tt.expectedEntry {
					t.Errorf("Expected tell entry '%s', got '%s'", tt.expectedEntry, m.tells[0])
				}
			}
		})
	}
}

// TestTellsListLimit verifies that tells list is limited to 50 entries
func TestTellsListLimit(t *testing.T) {
	m := Model{
		tells:    []string{},
		worldMap: mapper.NewMap(),
	}

	// Add 60 tell messages
	for i := 0; i < 60; i++ {
		m.detectAndParseTell("Player tells you 'message'")
	}

	// Should only keep the last 50
	if len(m.tells) != 50 {
		t.Errorf("Expected tells list to be limited to 50, got %d", len(m.tells))
	}
}

// TestTellsRendering verifies that tells panel is rendered correctly
func TestTellsRendering(t *testing.T) {
	m := Model{
		tells:    []string{"Alice: hello", "Bob: hi there"},
		worldMap: mapper.NewMap(),
		width:    80,
		height:   24,
	}

	// Initialize the tells viewport
	m.tellsViewport.Width = 56
	m.tellsViewport.Height = 2

	sidebar := m.renderSidebar(60, 24)

	// Check that sidebar contains "Tells" header
	if !strings.Contains(sidebar, "Tells") {
		t.Error("Sidebar should contain 'Tells' header")
	}

	// Check that sidebar does not contain "Character Stats"
	if strings.Contains(sidebar, "Character Stats") {
		t.Error("Sidebar should not contain 'Character Stats' (replaced by Tells)")
	}
}

// TestTellsRenderingEmpty verifies that tells panel shows placeholder when empty
func TestTellsRenderingEmpty(t *testing.T) {
	m := Model{
		tells:    []string{},
		worldMap: mapper.NewMap(),
		width:    80,
		height:   24,
	}

	// Initialize the tells viewport
	m.tellsViewport.Width = 56
	m.tellsViewport.Height = 2

	sidebar := m.renderSidebar(60, 24)

	// Check that sidebar contains "Tells" header
	if !strings.Contains(sidebar, "Tells") {
		t.Error("Sidebar should contain 'Tells' header")
	}

	// Check that sidebar contains placeholder text
	if !strings.Contains(sidebar, "no tells yet") {
		t.Error("Sidebar should contain 'no tells yet' when tells list is empty")
	}
}

// TestStripANSI verifies that ANSI codes are properly removed
func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "simple color code",
			input:    "\x1b[32mgreen text\x1b[0m",
			expected: "green text",
		},
		{
			name:     "multiple codes",
			input:    "\x1b[1m\x1b[31mbold red\x1b[0m normal",
			expected: "bold red normal",
		},
		{
			name:     "mixed text and codes",
			input:    "before\x1b[36mcolor\x1b[0mafter",
			expected: "beforecolorafter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
