package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
	"github.com/charmbracelet/bubbles/viewport"
)

// TestTellsIntegration demonstrates the full tells feature workflow
func TestTellsIntegration(t *testing.T) {
	// Create a model
	m := Model{
		tells:         []string{},
		worldMap:      mapper.NewMap(),
		width:         80,
		height:        24,
		tellsViewport: viewport.New(56, 2),
		output:        []string{},
	}

	// Simulate receiving various tell messages as the MUD server would send them
	mudOutput := []string{
		"Welcome to the MUD!",
		"> ",
		"Alice tells you 'Hello, how are you?'",
		"> ",
		"You say, 'I'm doing well, thanks!'",
		"> ",
		"Bob tells you 'Can you help me with this quest?'",
		"> ",
		"Charlie tells you 'Thanks for the tip!'",
		"> ",
	}

	// Process each line as if it came from the MUD
	for _, line := range mudOutput {
		m.detectAndParseTell(line)
		m.output = append(m.output, line)
	}

	// Verify tells were captured
	if len(m.tells) != 3 {
		t.Errorf("Expected 3 tells to be captured, got %d", len(m.tells))
	}

	// Verify tells are formatted correctly (Player: content format)
	expectedTells := []string{
		"Alice: Hello, how are you?",
		"Bob: Can you help me with this quest?",
		"Charlie: Thanks for the tip!",
	}

	for i, expected := range expectedTells {
		if i >= len(m.tells) {
			t.Errorf("Missing tell at index %d", i)
			continue
		}
		if m.tells[i] != expected {
			t.Errorf("Tell %d: expected '%s', got '%s'", i, expected, m.tells[i])
		}
	}

	// Verify that the main output still contains the full tell messages
	// (tells should appear in BOTH the main window AND the tells panel)
	fullTellCount := 0
	for _, line := range m.output {
		if strings.Contains(line, "tells you") {
			fullTellCount++
		}
	}
	if fullTellCount != 3 {
		t.Errorf("Expected 3 full tell messages in main output, got %d", fullTellCount)
	}

	// Render the sidebar to verify "Tells" panel appears
	sidebar := m.renderSidebar(60, 24)

	// Verify "Tells" header is present
	if !strings.Contains(sidebar, "Tells") {
		t.Error("Sidebar should contain 'Tells' header")
	}

	// Verify "Character Stats" is NOT present (it was replaced)
	if strings.Contains(sidebar, "Character Stats") {
		t.Error("Sidebar should NOT contain 'Character Stats' (replaced by Tells)")
	}

	// Verify at least one tell appears in the sidebar
	// (They may be truncated or formatted differently in the viewport)
	hasAlice := strings.Contains(sidebar, "Alice")
	hasBob := strings.Contains(sidebar, "Bob")
	hasCharlie := strings.Contains(sidebar, "Charlie")

	if !hasAlice && !hasBob && !hasCharlie {
		t.Error("Sidebar should contain at least one player name from tells")
	}
}

// TestTellsWithANSICodes verifies that tells with ANSI codes are handled correctly
func TestTellsWithANSICodes(t *testing.T) {
	m := Model{
		tells:         []string{},
		worldMap:      mapper.NewMap(),
		tellsViewport: viewport.New(56, 2),
	}

	// Simulate tell with ANSI color codes
	coloredTell := "\x1b[32mAlice tells you 'Hello!'\x1b[0m"
	m.detectAndParseTell(coloredTell)

	// Verify the tell was captured and ANSI codes were stripped for parsing
	if len(m.tells) != 1 {
		t.Errorf("Expected 1 tell to be captured, got %d", len(m.tells))
	}

	if len(m.tells) > 0 {
		expected := "Alice: Hello!"
		if m.tells[0] != expected {
			t.Errorf("Expected tell '%s', got '%s'", expected, m.tells[0])
		}
	}
}
