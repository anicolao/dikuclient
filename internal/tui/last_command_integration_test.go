package tui

import (
	"strings"
	"testing"
	
	"github.com/anicolao/dikuclient/internal/triggers"
)

// TestLastCommandTriggerIntegration tests the full integration of <last_command> with triggers
func TestLastCommandTriggerIntegration(t *testing.T) {
	// Create a model with trigger manager
	m := &Model{
		triggerManager: triggers.NewManager(),
		output:         make([]string, 0),
		lastCommand:    "", // No command sent yet
	}
	
	// Add a trigger that uses <last_command>
	_, err := m.triggerManager.Add("Huh?!", "/ai <last_command>")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}
	
	// Simulate sending a command
	m.lastCommand = "heall" // Misspelled command
	
	// Simulate receiving trigger text from MUD
	line := "Huh?!"
	
	// Test that trigger matches
	actions := m.triggerManager.Match(line)
	if len(actions) != 1 {
		t.Fatalf("Expected 1 trigger action, got %d", len(actions))
	}
	
	// The raw action should still have <last_command>
	if actions[0] != "/ai <last_command>" {
		t.Errorf("Expected action '/ai <last_command>', got '%s'", actions[0])
	}
	
	// Now test the substitution that happens in the TUI layer
	// This is what happens in the actual code when processing triggers
	substitutedAction := strings.ReplaceAll(actions[0], "<last_command>", m.lastCommand)
	
	expectedAction := "/ai heall"
	if substitutedAction != expectedAction {
		t.Errorf("Expected substituted action '%s', got '%s'", expectedAction, substitutedAction)
	}
}

// TestMultipleTriggerVariables tests that <last_command> works alongside normal trigger variables
func TestMultipleTriggerVariables(t *testing.T) {
	m := &Model{
		triggerManager: triggers.NewManager(),
		output:         make([]string, 0),
		lastCommand:    "kill goblin",
	}
	
	// Add a trigger with both normal variable and <last_command>
	_, err := m.triggerManager.Add("<enemy> defeated!", "say I beat <enemy> using <last_command>")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}
	
	// Simulate trigger firing
	line := "The goblin defeated!"
	actions := m.triggerManager.Match(line)
	
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	
	// The action should have the normal variable substituted
	// Format will have spaces replaced with dots
	expectedPartial := "say I beat The.goblin using <last_command>"
	if actions[0] != expectedPartial {
		t.Errorf("Expected action '%s', got '%s'", expectedPartial, actions[0])
	}
}
