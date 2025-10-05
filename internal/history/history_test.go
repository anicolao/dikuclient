package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Commands == nil {
		t.Error("Commands slice should be initialized")
	}
	if len(m.Commands) != 0 {
		t.Errorf("Expected empty commands, got %d", len(m.Commands))
	}
}

func TestAdd(t *testing.T) {
	m := NewManager()

	// Add a command
	m.Add("north")
	if len(m.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(m.Commands))
	}
	if m.Commands[0] != "north" {
		t.Errorf("Expected 'north', got '%s'", m.Commands[0])
	}

	// Add another command
	m.Add("south")
	if len(m.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(m.Commands))
	}

	// Add empty command (should be ignored)
	m.Add("")
	if len(m.Commands) != 2 {
		t.Errorf("Empty command should not be added, got %d commands", len(m.Commands))
	}

	// Add duplicate consecutive command (should be ignored)
	m.Add("south")
	if len(m.Commands) != 2 {
		t.Errorf("Consecutive duplicate should not be added, got %d commands", len(m.Commands))
	}

	// Add same command after another command (should be added)
	m.Add("east")
	m.Add("south")
	if len(m.Commands) != 4 {
		t.Fatalf("Expected 4 commands, got %d", len(m.Commands))
	}
	if m.Commands[3] != "south" {
		t.Errorf("Expected last command to be 'south', got '%s'", m.Commands[3])
	}
}

func TestGetCommands(t *testing.T) {
	m := NewManager()
	m.Add("north")
	m.Add("south")

	commands := m.GetCommands()
	if len(commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(commands))
	}

	// Modify the returned slice
	commands[0] = "modified"

	// Original should be unchanged
	if m.Commands[0] != "north" {
		t.Error("GetCommands should return a copy, not the original slice")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")

	// Create a manager and add commands
	m1 := NewManager()
	m1.filePath = historyPath
	m1.Add("north")
	m1.Add("south")
	m1.Add("east")

	// Save to disk
	if err := m1.Save(); err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Load from disk
	m2, err := LoadFromPath(historyPath)
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	// Verify loaded data
	if len(m2.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(m2.Commands))
	}
	if m2.Commands[0] != "north" {
		t.Errorf("Expected 'north', got '%s'", m2.Commands[0])
	}
	if m2.Commands[1] != "south" {
		t.Errorf("Expected 'south', got '%s'", m2.Commands[1])
	}
	if m2.Commands[2] != "east" {
		t.Errorf("Expected 'east', got '%s'", m2.Commands[2])
	}
}

func TestLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "nonexistent.json")

	// Load from non-existent file should return empty manager
	m, err := LoadFromPath(historyPath)
	if err != nil {
		t.Fatalf("Loading non-existent file should not error: %v", err)
	}

	if len(m.Commands) != 0 {
		t.Errorf("Expected empty commands for non-existent file, got %d", len(m.Commands))
	}

	if m.filePath != historyPath {
		t.Errorf("Expected filePath to be set to %s, got %s", historyPath, m.filePath)
	}
}

func TestSaveWithoutPath(t *testing.T) {
	// This test requires setting up environment to control where the file is saved
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	m := NewManager()
	m.Add("test command")

	if err := m.Save(); err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Verify file was created in the expected location
	expectedPath := filepath.Join(tempDir, "history.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatal("History file was not created in expected location")
	}
}
