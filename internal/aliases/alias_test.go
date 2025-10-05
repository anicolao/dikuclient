package aliases

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAliasExpansion(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []string
		expected string
	}{
		{
			name:     "Single placeholder with one arg",
			template: "give all <target>",
			args:     []string{"mary"},
			expected: "give all mary",
		},
		{
			name:     "Single placeholder with multiple args",
			template: "give all <target>",
			args:     []string{"mary", "jane"},
			expected: "give all mary jane",
		},
		{
			name:     "Two placeholders with two args",
			template: "give <object> <target>",
			args:     []string{"sword", "mary"},
			expected: "give sword mary",
		},
		{
			name:     "Two placeholders with one arg",
			template: "give <object> <target>",
			args:     []string{"sword"},
			expected: "give sword ",
		},
		{
			name:     "Args placeholder with multiple remaining",
			template: "cast <spell> <args>",
			args:     []string{"fireball", "at", "goblin", "with", "power"},
			expected: "cast fireball at goblin with power",
		},
		{
			name:     "Args placeholder with one arg only",
			template: "cast <spell> <args>",
			args:     []string{"fireball"},
			expected: "cast fireball ",
		},
		{
			name:     "Three placeholders with args",
			template: "tell <target> <arg1> <arg2> <args>",
			args:     []string{"mary", "hello", "there", "how", "are", "you"},
			expected: "tell mary hello there how are you",
		},
		{
			name:     "Three placeholders with three args",
			template: "tell <target> <arg1> <arg2> <args>",
			args:     []string{"mary", "hello", "there"},
			expected: "tell mary hello there ",
		},
		{
			name:     "Two args plus remainder",
			template: "perform <arg1> <arg2> <args>",
			args:     []string{"action", "one", "and", "more", "stuff"},
			expected: "perform action one and more stuff",
		},
		{
			name:     "No placeholders",
			template: "look",
			args:     []string{"extra", "args"},
			expected: "look",
		},
		{
			name:     "Empty args with placeholder",
			template: "give all <target>",
			args:     []string{},
			expected: "give all ",
		},
	}

	manager := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.expandTemplate(tt.template, tt.args)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestManagerExpand(t *testing.T) {
	manager := NewManager()
	
	// Add some aliases
	manager.Add("gat", "give all <target>")
	manager.Add("gt", "give <object> <target>")
	manager.Add("k", "kill <target>")

	tests := []struct {
		name     string
		command  string
		expected string
		expanded bool
	}{
		{
			name:     "Expand gat alias",
			command:  "gat mary",
			expected: "give all mary",
			expanded: true,
		},
		{
			name:     "Expand gt alias",
			command:  "gt sword mary",
			expected: "give sword mary",
			expanded: true,
		},
		{
			name:     "Expand k alias",
			command:  "k goblin",
			expected: "kill goblin",
			expanded: true,
		},
		{
			name:     "Non-alias command",
			command:  "look",
			expected: "look",
			expanded: false,
		},
		{
			name:     "Alias with no args",
			command:  "gat",
			expected: "give all ",
			expanded: true,
		},
		{
			name:     "Multiple args to single placeholder",
			command:  "k big scary goblin",
			expected: "kill big scary goblin",
			expanded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, expanded := manager.Expand(tt.command)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
			if expanded != tt.expanded {
				t.Errorf("Expected expanded=%v, got %v", tt.expanded, expanded)
			}
		})
	}
}

func TestManagerAddRemove(t *testing.T) {
	manager := NewManager()

	// Add an alias
	alias, err := manager.Add("gat", "give all <target>")
	if err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	if len(manager.Aliases) != 1 {
		t.Errorf("Expected 1 alias, got %d", len(manager.Aliases))
	}

	if alias.Name != "gat" || alias.Template != "give all <target>" {
		t.Errorf("Alias not added correctly")
	}

	// Try to add duplicate
	_, err = manager.Add("gat", "something else")
	if err == nil {
		t.Errorf("Expected error when adding duplicate alias")
	}

	// Remove the alias
	err = manager.Remove(0)
	if err != nil {
		t.Fatalf("Failed to remove alias: %v", err)
	}

	if len(manager.Aliases) != 0 {
		t.Errorf("Expected 0 aliases, got %d", len(manager.Aliases))
	}
}

func TestValidation(t *testing.T) {
	manager := NewManager()

	// Test invalid alias names
	invalidNames := []string{
		"gat mary",    // Contains space
		"give-all",    // Contains hyphen
		"give.all",    // Contains dot
		"",            // Empty
	}

	for _, name := range invalidNames {
		_, err := manager.Add(name, "give all <target>")
		if err == nil {
			t.Errorf("Expected error for invalid alias name '%s'", name)
		}
	}

	// Test valid alias names
	validNames := []string{
		"gat",
		"gt",
		"k",
		"alias123",
		"MyAlias",
	}

	for _, name := range validNames {
		manager = NewManager() // Reset for each test
		_, err := manager.Add(name, "test <target>")
		if err != nil {
			t.Errorf("Unexpected error for valid alias name '%s': %v", name, err)
		}
	}
}

func TestPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	aliasesPath := filepath.Join(tempDir, "aliases.json")

	// Create and save a manager
	manager := NewManager()
	manager.filePath = aliasesPath
	manager.Add("gat", "give all <target>")
	manager.Add("gt", "give <object> <target>")
	manager.Add("k", "kill <target>")

	err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save aliases: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(aliasesPath); os.IsNotExist(err) {
		t.Fatalf("Aliases file was not created")
	}

	// Load the manager
	loadedManager, err := LoadFromPath(aliasesPath)
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	if len(loadedManager.Aliases) != 3 {
		t.Errorf("Expected 3 aliases, got %d", len(loadedManager.Aliases))
	}

	// Test that loaded aliases work
	result, expanded := loadedManager.Expand("gat mary")
	if !expanded || result != "give all mary" {
		t.Errorf("Loaded alias did not expand correctly: got '%s', expanded=%v", result, expanded)
	}

	result, expanded = loadedManager.Expand("gt sword john")
	if !expanded || result != "give sword john" {
		t.Errorf("Loaded alias did not expand correctly: got '%s', expanded=%v", result, expanded)
	}
}

func TestComplexParameterSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		template string
		command  string
		expected string
	}{
		{
			name:     "Target with args - problem statement example",
			template: "command <target> <args>",
			command:  "alias john hello there friend",
			expected: "command john hello there friend",
		},
		{
			name:     "Target arg1 arg2 args - four placeholders",
			template: "complex <target> <arg1> <arg2> <args>",
			command:  "alias bob first second the rest here",
			expected: "complex bob first second the rest here",
		},
		{
			name:     "Arg1 arg2 args - three placeholders",
			template: "action <arg1> <arg2> <args>",
			command:  "alias one two three four five",
			expected: "action one two three four five",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh manager for each test
			m := NewManager()
			m.Add("alias", tt.template)
			
			result, expanded := m.Expand(tt.command)
			if !expanded {
				t.Errorf("Expected alias to be expanded")
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
