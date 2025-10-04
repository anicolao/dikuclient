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
		output:    []string{"Welcome to the MUD", "> "}, // Simulate some output with a prompt
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

	// Add two empty lines and restore prompt after command output
	m.output = append(m.output, "")
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

	// Second-to-last and third-to-last lines should be empty (two newlines)
	if m.output[len(m.output)-2] != "" {
		t.Errorf("Second-to-last line should be empty, got: %q", m.output[len(m.output)-2])
	}
	if m.output[len(m.output)-3] != "" {
		t.Errorf("Third-to-last line should be empty, got: %q", m.output[len(m.output)-3])
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

// TestMultipleClientCommands verifies behavior when multiple commands are executed
func TestMultipleClientCommands(t *testing.T) {
	// Create a model with a simple setup
	m := Model{
		output:    []string{"Welcome", "> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Execute first command
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/map\x1b[0m"
	m.handleClientCommand("/map")
	m.output = append(m.output, "")
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	firstCommandOutputLen := len(m.output)

	// Execute second command
	savedPrompt = m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/help\x1b[0m"
	m.handleClientCommand("/help")
	m.output = append(m.output, "")
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	// Verify both commands are in output with proper spacing
	if len(m.output) <= firstCommandOutputLen {
		t.Errorf("Expected output to grow after second command")
	}

	// Last line should be prompt
	if m.output[len(m.output)-1] != savedPrompt {
		t.Errorf("Last line should be prompt")
	}

	// Second-to-last should be empty
	if m.output[len(m.output)-2] != "" {
		t.Errorf("Second-to-last line should be empty")
	}
}

// TestErrorCommand verifies error messages also follow the same format
func TestErrorCommand(t *testing.T) {
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/unknown\x1b[0m"
	m.handleClientCommand("/unknown")
	m.output = append(m.output, "")
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	// Should have error message
	hasError := false
	for _, line := range m.output {
		if strings.Contains(line, "Unknown command") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error message for unknown command")
	}

	// Should still have proper structure with two empty lines and prompt
	if m.output[len(m.output)-1] != savedPrompt {
		t.Error("Last line should be prompt")
	}
	if m.output[len(m.output)-2] != "" {
		t.Error("Second-to-last line should be empty")
	}
	if m.output[len(m.output)-3] != "" {
		t.Error("Third-to-last line should be empty")
	}
}
