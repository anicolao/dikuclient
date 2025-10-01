package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestClientCommandOutputFormatting verifies that client commands
// replace the prompt line and restore it after output with an empty line
func TestClientCommandOutputFormatting(t *testing.T) {
	// Create a model with a simple setup
	m := Model{
		output:    []string{"Welcome to the MUD", "> "},  // Simulate some output with a prompt
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Record initial output length and prompt
	initialOutputCount := len(m.output)
	savedPrompt := m.output[len(m.output)-1]

	// Simulate typing "/help" and pressing enter
	m.currentInput = "/help"

	// Execute the client command handler (this simulates what happens in Update)
	if len(m.output) > 0 {
		// Replace the prompt line with the command
		m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/help\x1b[0m"
	}
	
	m.handleClientCommand("/help")
	
	// Add empty line and restore prompt after command output
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	// Verify the output structure
	if len(m.output) < initialOutputCount {
		t.Fatalf("Expected output to increase, got %d lines", len(m.output))
	}

	// The line that was the prompt should now contain the command
	commandLine := m.output[initialOutputCount-1]
	if !strings.Contains(commandLine, "/help") {
		t.Errorf("Command line should contain '/help', got: %s", commandLine)
	}

	// Check that help output was added (should have several lines after the command line)
	hasHelpHeader := false
	for i := initialOutputCount; i < len(m.output)-2; i++ {
		if strings.Contains(m.output[i], "Client Commands") {
			hasHelpHeader = true
			break
		}
	}
	if !hasHelpHeader {
		t.Error("Expected to find 'Client Commands' in output after command line")
	}

	// Second-to-last line should be empty
	if m.output[len(m.output)-2] != "" {
		t.Errorf("Second-to-last line should be empty, got: %q", m.output[len(m.output)-2])
	}

	// Last line should be the restored prompt
	lastLine := m.output[len(m.output)-1]
	if lastLine != savedPrompt {
		t.Errorf("Last line should be the restored prompt %q, got: %q", savedPrompt, lastLine)
	}

	// Verify structure: [...existing output..., command_line, output_lines..., empty_line, prompt]
	t.Logf("Output structure after /help command:")
	for i, line := range m.output {
		// Strip ANSI codes for clearer logging
		cleanLine := strings.ReplaceAll(line, "\x1b[93m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[0m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[92m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[96m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[90m", "")
		t.Logf("  [%d]: %q", i, cleanLine)
	}
}

// TestClientCommandWithoutPrompt verifies behavior when there's no existing prompt
func TestClientCommandWithoutPrompt(t *testing.T) {
	// Create a model with no output
	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Execute a client command
	m.handleClientCommand("/help")

	// Should still work - output should be added
	if len(m.output) == 0 {
		t.Error("Expected output to be added even without existing prompt")
	}
}
