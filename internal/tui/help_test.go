package tui

import (
	"strings"
	"testing"
)

// TestHelpCommand tests the general help command
func TestHelpCommand(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Execute the /help command
	model.handleHelpCommand(nil)

	// Check that the output contains expected sections
	found := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "Client Commands") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected /help output to include 'Client Commands' section")
	}

	// Check for the new help text
	foundHelpText := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "Use /help <command> for detailed help") {
			foundHelpText = true
			break
		}
	}

	if !foundHelpText {
		t.Error("Expected /help output to include 'Use /help <command>' text")
	}
}

// TestHelpCommandWithArgument tests help for specific commands
func TestHelpCommandWithArgument(t *testing.T) {
	tests := []struct {
		command string
		expect  string
	}{
		{"point", "Show Next Direction"},
		{"wayfind", "Show Full Path"},
		{"go", "Auto-Walk to Room"},
		{"stop", "Stop Auto-Walk"},
		{"map", "Show Map Information"},
		{"rooms", "List Known Rooms"},
		{"nearby", "List Nearby Rooms"},
		{"legend", "List Rooms on Map"},
		{"trigger", "Automated Responses"},
		{"alias", "Command Shortcuts"},
		{"share", "Share Web Session"},
		{"help", "Show Help Information"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			model := NewModel("localhost", 4000, nil, nil)
			model.width = 80
			model.height = 24

			// Execute the /help <command> command
			model.handleHelpCommand([]string{tt.command})

			// Check that the output contains the expected text
			found := false
			for _, line := range model.output {
				cleanLine := stripANSI(line)
				if strings.Contains(cleanLine, tt.expect) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected /help %s output to include '%s'", tt.command, tt.expect)
				t.Logf("Output was: %v", model.output)
			}
		})
	}
}

// TestHelpCommandWithUnknownArgument tests help for unknown command
func TestHelpCommandWithUnknownArgument(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Execute the /help unknown command
	model.handleHelpCommand([]string{"unknown"})

	// Check that the output contains error message
	found := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "Unknown command: unknown") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected /help unknown output to include error message")
	}

	// Check that it shows available commands
	foundAvailable := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "Available commands for detailed help") {
			foundAvailable = true
			break
		}
	}

	if !foundAvailable {
		t.Error("Expected /help unknown output to include list of available commands")
	}
}

// TestHelpCommandPluralForms tests that both singular and plural forms work
func TestHelpCommandPluralForms(t *testing.T) {
	tests := []struct {
		command string
		expect  string
	}{
		{"trigger", "Automated Responses"},
		{"triggers", "Automated Responses"},
		{"alias", "Command Shortcuts"},
		{"aliases", "Command Shortcuts"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			model := NewModel("localhost", 4000, nil, nil)
			model.width = 80
			model.height = 24

			// Execute the /help <command> command
			model.handleHelpCommand([]string{tt.command})

			// Check that the output contains the expected text
			found := false
			for _, line := range model.output {
				cleanLine := stripANSI(line)
				if strings.Contains(cleanLine, tt.expect) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected /help %s output to include '%s'", tt.command, tt.expect)
			}
		})
	}
}
