package triggers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTriggerMatching(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		action   string
		input    string
		expected string
	}{
		{
			name:     "Simple pattern match",
			pattern:  "hungry",
			action:   "eat bread",
			input:    "You are hungry",
			expected: "eat bread",
		},
		{
			name:     "No match",
			pattern:  "hungry",
			action:   "eat bread",
			input:    "You are tired",
			expected: "",
		},
		{
			name:     "Variable substitution single word",
			pattern:  "The <subject> cries a sad death cry",
			action:   "get all <subject>.corpse",
			input:    "The spider cries a sad death cry",
			expected: "get all spider.corpse",
		},
		{
			name:     "Variable substitution multiple words with space-to-dot",
			pattern:  "The <subject> cries a sad death cry",
			action:   "get all <subject>.corpse",
			input:    "The small spider cries a sad death cry",
			expected: "get all small.spider.corpse",
		},
		{
			name:     "Multiple variables",
			pattern:  "<actor> hits <target> hard",
			action:   "say <actor> attacked <target>",
			input:    "The warrior hits the goblin hard",
			expected: "say The.warrior attacked the.goblin",
		},
		{
			name:     "Pattern at start of line",
			pattern:  "<name> arrives",
			action:   "greet <name>",
			input:    "John Smith arrives",
			expected: "greet John.Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := &Trigger{
				ID:      "test",
				Pattern: tt.pattern,
				Action:  tt.action,
			}

			err := trigger.compilePattern()
			if err != nil {
				t.Fatalf("Failed to compile pattern: %v", err)
			}

			result := trigger.match(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestManagerAddRemove(t *testing.T) {
	manager := NewManager()

	// Add a trigger
	trigger, err := manager.Add("hungry", "eat bread")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}

	if len(manager.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(manager.Triggers))
	}

	if trigger.Pattern != "hungry" || trigger.Action != "eat bread" {
		t.Errorf("Trigger not added correctly")
	}

	// Remove the trigger
	err = manager.Remove(0)
	if err != nil {
		t.Fatalf("Failed to remove trigger: %v", err)
	}

	if len(manager.Triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(manager.Triggers))
	}
}

func TestManagerMatch(t *testing.T) {
	manager := NewManager()

	// Add multiple triggers
	manager.Add("hungry", "eat bread")
	manager.Add("thirsty", "drink water")
	manager.Add("The <subject> dies", "get <subject>")

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Match first trigger",
			input:    "You are hungry",
			expected: []string{"eat bread"},
		},
		{
			name:     "Match second trigger",
			input:    "You are thirsty",
			expected: []string{"drink water"},
		},
		{
			name:     "Match third trigger with variable",
			input:    "The orc dies",
			expected: []string{"get orc"},
		},
		{
			name:     "Match multiple triggers",
			input:    "You are hungry and thirsty",
			expected: []string{"eat bread", "drink water"},
		},
		{
			name:     "No matches",
			input:    "You are fine",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := manager.Match(tt.input)
			if len(results) != len(tt.expected) {
				t.Errorf("Expected %d matches, got %d", len(tt.expected), len(results))
				return
			}

			for i, expected := range tt.expected {
				if results[i] != expected {
					t.Errorf("Match %d: expected '%s', got '%s'", i, expected, results[i])
				}
			}
		})
	}
}

func TestPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	triggersPath := filepath.Join(tempDir, "triggers.json")

	// Create and save a manager
	manager := NewManager()
	manager.filePath = triggersPath
	manager.Add("hungry", "eat bread")
	manager.Add("The <subject> dies", "get <subject>")

	err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save triggers: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(triggersPath); os.IsNotExist(err) {
		t.Fatalf("Triggers file was not created")
	}

	// Load the manager
	loadedManager, err := LoadFromPath(triggersPath)
	if err != nil {
		t.Fatalf("Failed to load triggers: %v", err)
	}

	if len(loadedManager.Triggers) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(loadedManager.Triggers))
	}

	// Test that loaded triggers work
	results := loadedManager.Match("You are hungry")
	if len(results) != 1 || results[0] != "eat bread" {
		t.Errorf("Loaded trigger did not match correctly")
	}

	results = loadedManager.Match("The goblin dies")
	if len(results) != 1 || results[0] != "get goblin" {
		t.Errorf("Loaded trigger with variable did not match correctly")
	}
}
