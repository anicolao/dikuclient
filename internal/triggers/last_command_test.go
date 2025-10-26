package triggers

import (
	"testing"
)

// TestLastCommandSubstitution tests that <last_command> can be used in trigger actions
func TestLastCommandSubstitution(t *testing.T) {
	m := NewManager()
	
	// Add a trigger that uses <last_command>
	_, err := m.Add("Huh?!", "/ai <last_command>")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}
	
	// Test that the trigger matches
	actions := m.Match("Huh?!")
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	
	// The action should still contain <last_command> - it will be substituted by the TUI layer
	expectedAction := "/ai <last_command>"
	if actions[0] != expectedAction {
		t.Errorf("Expected action '%s', got '%s'", expectedAction, actions[0])
	}
}

// TestVariableSubstitutionStillWorks tests that normal variable substitution still works
func TestVariableSubstitutionStillWorks(t *testing.T) {
	m := NewManager()
	
	// Add a trigger with a normal variable
	_, err := m.Add("The <subject> dies", "get <subject>")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}
	
	// Test that the trigger matches and substitutes
	actions := m.Match("The goblin dies")
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	
	expectedAction := "get goblin"
	if actions[0] != expectedAction {
		t.Errorf("Expected action '%s', got '%s'", expectedAction, actions[0])
	}
}
