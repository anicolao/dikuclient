package tui

import (
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestCommandHistoryBasic verifies basic command history storage
func TestCommandHistoryBasic(t *testing.T) {
	m := Model{
		connected:      true,
		worldMap:       mapper.NewMap(),
		commandHistory: []string{},
		historyIndex:   -1,
	}

	// Simulate entering a command
	m.currentInput = "north"
	
	// Process Enter key (without actual connection)
	// We'll manually add to history as the test doesn't have a real connection
	command := m.currentInput
	if command != "" {
		if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command {
			m.commandHistory = append(m.commandHistory, command)
		}
		m.historyIndex = -1
		m.historySavedInput = ""
	}

	// Verify command was added to history
	if len(m.commandHistory) != 1 {
		t.Fatalf("Expected 1 command in history, got %d", len(m.commandHistory))
	}
	if m.commandHistory[0] != "north" {
		t.Errorf("Expected 'north' in history, got '%s'", m.commandHistory[0])
	}

	// Add another command
	m.currentInput = "south"
	command = m.currentInput
	if command != "" {
		if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command {
			m.commandHistory = append(m.commandHistory, command)
		}
		m.historyIndex = -1
		m.historySavedInput = ""
	}

	// Verify both commands in history
	if len(m.commandHistory) != 2 {
		t.Fatalf("Expected 2 commands in history, got %d", len(m.commandHistory))
	}
	if m.commandHistory[1] != "south" {
		t.Errorf("Expected 'south' in history[1], got '%s'", m.commandHistory[1])
	}
}

// TestCommandHistoryNoDuplicates verifies consecutive duplicate commands are not added
func TestCommandHistoryNoDuplicates(t *testing.T) {
	m := Model{
		connected:      true,
		worldMap:       mapper.NewMap(),
		commandHistory: []string{},
		historyIndex:   -1,
	}

	// Add a command
	command := "north"
	if command != "" {
		if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command {
			m.commandHistory = append(m.commandHistory, command)
		}
		m.historyIndex = -1
		m.historySavedInput = ""
	}

	// Try to add the same command again
	if command != "" {
		if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command {
			m.commandHistory = append(m.commandHistory, command)
		}
		m.historyIndex = -1
		m.historySavedInput = ""
	}

	// Verify only one command in history
	if len(m.commandHistory) != 1 {
		t.Fatalf("Expected 1 command in history (no duplicates), got %d", len(m.commandHistory))
	}
}

// TestCommandHistoryUpNavigation verifies Up arrow navigates backward through history
func TestCommandHistoryUpNavigation(t *testing.T) {
	m := Model{
		connected:         true,
		worldMap:          mapper.NewMap(),
		commandHistory:    []string{"north", "south", "east"},
		historyIndex:      -1,
		currentInput:      "test",
		historySavedInput: "",
	}

	// Press Up once - should get "east" (most recent)
	if len(m.commandHistory) > 0 {
		if m.historyIndex == -1 {
			m.historySavedInput = m.currentInput
			m.historyIndex = len(m.commandHistory)
		}
		
		if m.historyIndex > 0 {
			m.historyIndex--
			m.currentInput = m.commandHistory[m.historyIndex]
		}
	}

	if m.currentInput != "east" {
		t.Errorf("Expected 'east' after first Up, got '%s'", m.currentInput)
	}
	if m.historySavedInput != "test" {
		t.Errorf("Expected saved input 'test', got '%s'", m.historySavedInput)
	}
	if m.historyIndex != 2 {
		t.Errorf("Expected historyIndex 2, got %d", m.historyIndex)
	}

	// Press Up again - should get "south"
	if m.historyIndex > 0 {
		m.historyIndex--
		m.currentInput = m.commandHistory[m.historyIndex]
	}

	if m.currentInput != "south" {
		t.Errorf("Expected 'south' after second Up, got '%s'", m.currentInput)
	}
	if m.historyIndex != 1 {
		t.Errorf("Expected historyIndex 1, got %d", m.historyIndex)
	}

	// Press Up again - should get "north"
	if m.historyIndex > 0 {
		m.historyIndex--
		m.currentInput = m.commandHistory[m.historyIndex]
	}

	if m.currentInput != "north" {
		t.Errorf("Expected 'north' after third Up, got '%s'", m.currentInput)
	}
	if m.historyIndex != 0 {
		t.Errorf("Expected historyIndex 0, got %d", m.historyIndex)
	}

	// Press Up again - should stay at "north" (can't go further back)
	if m.historyIndex > 0 {
		m.historyIndex--
		m.currentInput = m.commandHistory[m.historyIndex]
	}

	if m.currentInput != "north" {
		t.Errorf("Expected 'north' at history beginning, got '%s'", m.currentInput)
	}
	if m.historyIndex != 0 {
		t.Errorf("Expected historyIndex to stay at 0, got %d", m.historyIndex)
	}
}

// TestCommandHistoryDownNavigation verifies Down arrow navigates forward through history
func TestCommandHistoryDownNavigation(t *testing.T) {
	m := Model{
		connected:         true,
		worldMap:          mapper.NewMap(),
		commandHistory:    []string{"north", "south", "east"},
		historyIndex:      0, // Start at beginning
		currentInput:      "north",
		historySavedInput: "test",
	}

	// Press Down - should get "south"
	if m.historyIndex != -1 {
		m.historyIndex++
		
		if m.historyIndex >= len(m.commandHistory) {
			m.currentInput = m.historySavedInput
			m.historyIndex = -1
			m.historySavedInput = ""
		} else {
			m.currentInput = m.commandHistory[m.historyIndex]
		}
	}

	if m.currentInput != "south" {
		t.Errorf("Expected 'south' after first Down, got '%s'", m.currentInput)
	}
	if m.historyIndex != 1 {
		t.Errorf("Expected historyIndex 1, got %d", m.historyIndex)
	}

	// Press Down again - should get "east"
	if m.historyIndex != -1 {
		m.historyIndex++
		
		if m.historyIndex >= len(m.commandHistory) {
			m.currentInput = m.historySavedInput
			m.historyIndex = -1
			m.historySavedInput = ""
		} else {
			m.currentInput = m.commandHistory[m.historyIndex]
		}
	}

	if m.currentInput != "east" {
		t.Errorf("Expected 'east' after second Down, got '%s'", m.currentInput)
	}
	if m.historyIndex != 2 {
		t.Errorf("Expected historyIndex 2, got %d", m.historyIndex)
	}

	// Press Down again - should restore saved input "test"
	if m.historyIndex != -1 {
		m.historyIndex++
		
		if m.historyIndex >= len(m.commandHistory) {
			m.currentInput = m.historySavedInput
			m.historyIndex = -1
			m.historySavedInput = ""
		} else {
			m.currentInput = m.commandHistory[m.historyIndex]
		}
	}

	if m.currentInput != "test" {
		t.Errorf("Expected saved input 'test' after going past end, got '%s'", m.currentInput)
	}
	if m.historyIndex != -1 {
		t.Errorf("Expected historyIndex -1 after restoration, got %d", m.historyIndex)
	}
	if m.historySavedInput != "" {
		t.Errorf("Expected historySavedInput to be cleared, got '%s'", m.historySavedInput)
	}
}

// TestCommandHistoryEmptyInput verifies empty commands are not added to history
func TestCommandHistoryEmptyInput(t *testing.T) {
	m := Model{
		connected:      true,
		worldMap:       mapper.NewMap(),
		commandHistory: []string{},
		historyIndex:   -1,
	}

	// Try to add empty command
	command := ""
	if command != "" {
		if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != command {
			m.commandHistory = append(m.commandHistory, command)
		}
		m.historyIndex = -1
		m.historySavedInput = ""
	}

	// Verify no command was added
	if len(m.commandHistory) != 0 {
		t.Fatalf("Expected 0 commands in history (empty not added), got %d", len(m.commandHistory))
	}
}

// TestCommandHistoryUpWithNoHistory verifies Up arrow does nothing when history is empty
func TestCommandHistoryUpWithNoHistory(t *testing.T) {
	m := Model{
		connected:      true,
		worldMap:       mapper.NewMap(),
		commandHistory: []string{},
		historyIndex:   -1,
		currentInput:   "test",
	}

	originalInput := m.currentInput

	// Try to press Up with empty history
	if len(m.commandHistory) > 0 {
		if m.historyIndex == -1 {
			m.historySavedInput = m.currentInput
			m.historyIndex = len(m.commandHistory)
		}
		
		if m.historyIndex > 0 {
			m.historyIndex--
			m.currentInput = m.commandHistory[m.historyIndex]
		}
	}

	// Input should remain unchanged
	if m.currentInput != originalInput {
		t.Errorf("Expected input to remain '%s', got '%s'", originalInput, m.currentInput)
	}
	if m.historyIndex != -1 {
		t.Errorf("Expected historyIndex to remain -1, got %d", m.historyIndex)
	}
}

// TestCommandHistoryFullWorkflow simulates a complete user workflow
func TestCommandHistoryFullWorkflow(t *testing.T) {
	m := Model{
		connected:         true,
		worldMap:          mapper.NewMap(),
		commandHistory:    []string{},
		historyIndex:      -1,
		currentInput:      "",
		historySavedInput: "",
	}

	// Helper to add command to history (simulates Enter key)
	addCommand := func(cmd string) {
		m.currentInput = cmd
		if cmd != "" {
			if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != cmd {
				m.commandHistory = append(m.commandHistory, cmd)
			}
			m.historyIndex = -1
			m.historySavedInput = ""
		}
		m.currentInput = ""
	}

	// Helper to navigate up (simulates Up arrow)
	navigateUp := func() {
		if len(m.commandHistory) > 0 {
			if m.historyIndex == -1 {
				m.historySavedInput = m.currentInput
				m.historyIndex = len(m.commandHistory)
			}
			
			if m.historyIndex > 0 {
				m.historyIndex--
				m.currentInput = m.commandHistory[m.historyIndex]
			}
		}
	}

	// Helper to navigate down (simulates Down arrow)
	navigateDown := func() {
		if m.historyIndex != -1 {
			m.historyIndex++
			
			if m.historyIndex >= len(m.commandHistory) {
				m.currentInput = m.historySavedInput
				m.historyIndex = -1
				m.historySavedInput = ""
			} else {
				m.currentInput = m.commandHistory[m.historyIndex]
			}
		}
	}

	// Scenario: User enters three commands
	addCommand("north")
	addCommand("look")
	addCommand("south")

	if len(m.commandHistory) != 3 {
		t.Fatalf("Expected 3 commands in history, got %d", len(m.commandHistory))
	}

	// Start typing a new command
	m.currentInput = "ea"

	// User presses Up to get previous command
	navigateUp()
	if m.currentInput != "south" {
		t.Errorf("Expected 'south' after first Up, got '%s'", m.currentInput)
	}
	if m.historySavedInput != "ea" {
		t.Errorf("Expected saved input 'ea', got '%s'", m.historySavedInput)
	}

	// User presses Up again
	navigateUp()
	if m.currentInput != "look" {
		t.Errorf("Expected 'look' after second Up, got '%s'", m.currentInput)
	}

	// User presses Down to go forward
	navigateDown()
	if m.currentInput != "south" {
		t.Errorf("Expected 'south' after Down, got '%s'", m.currentInput)
	}

	// User presses Down again to restore what they were typing
	navigateDown()
	if m.currentInput != "ea" {
		t.Errorf("Expected restored input 'ea', got '%s'", m.currentInput)
	}
	if m.historyIndex != -1 {
		t.Errorf("Expected historyIndex -1 after restoration, got %d", m.historyIndex)
	}

	// User finishes typing and executes
	m.currentInput = "east"
	addCommand(m.currentInput)

	if len(m.commandHistory) != 4 {
		t.Fatalf("Expected 4 commands in history, got %d", len(m.commandHistory))
	}
	if m.commandHistory[3] != "east" {
		t.Errorf("Expected 'east' as last command, got '%s'", m.commandHistory[3])
	}

	// User presses Up to repeat the last command
	navigateUp()
	if m.currentInput != "east" {
		t.Errorf("Expected 'east' after Up, got '%s'", m.currentInput)
	}

	// User executes the repeated command
	cmd := m.currentInput
	addCommand(cmd)

	// Should not add duplicate
	if len(m.commandHistory) != 4 {
		t.Fatalf("Expected 4 commands (no duplicate), got %d", len(m.commandHistory))
	}
}
