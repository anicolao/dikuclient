package tui

import (
	"os"
	"testing"
	"time"
)

// TestTickTimerDetection tests that the tick timer detects T:NN in prompts
func TestTickTimerDetection(t *testing.T) {
	// Create a model with tick timer
	m := NewModel("test.mud", 4000, nil, nil)

	// Simulate receiving a prompt with T:24
	promptLine := "101H 132V 54710X 49.60% 570C [Hero:Good] [enemy:Bad] T:24 Exits:NS>"

	// Detect tick time
	m.detectTickPrompt(promptLine)

	// Check that tick timer was updated
	if m.tickTimerManager.LastSeenValue != 24 {
		t.Errorf("Expected last seen value 24, got %d", m.tickTimerManager.LastSeenValue)
	}

	if m.tickTimerManager.TickInterval == 0 {
		t.Error("Expected tick interval to be set")
	}

	// Wait a moment and check current tick time
	time.Sleep(100 * time.Millisecond)
	currentTick := m.tickTimerManager.GetCurrentTickTime()
	if currentTick < 23 || currentTick > 24 {
		t.Errorf("Expected current tick time around 24, got %d", currentTick)
	}
}

// TestTickTriggerCommand tests the /ticktrigger command
func TestTickTriggerCommand(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval (simulate detecting a prompt)
	m.tickTimerManager.TickInterval = 75
	m.tickTimerManager.UpdateFromPrompt(24)

	// Add a tick trigger
	command := "ticktrigger 5 \"cast 'heal'\""
	m.handleTickTriggerCommand(command)

	// Check that trigger was added
	if len(m.tickTimerManager.TickTriggers) != 1 {
		t.Fatalf("Expected 1 tick trigger, got %d", len(m.tickTimerManager.TickTriggers))
	}

	trigger := m.tickTimerManager.TickTriggers[0]
	if trigger.TickTime != 5 {
		t.Errorf("Expected tick time 5, got %d", trigger.TickTime)
	}

	if trigger.Commands != "cast 'heal'" {
		t.Errorf("Expected commands 'cast 'heal'', got %s", trigger.Commands)
	}
}

// TestTickTriggerFiring tests that tick triggers fire at the correct time
func TestTickTriggerFiring(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 10 // Short interval for testing

	// Add a tick trigger at T:5
	m.tickTimerManager.AddTrigger(5, "test command")

	// Set tick time to 6 (not yet time to fire)
	m.tickTimerManager.UpdateFromPrompt(6)

	// Check that no triggers fire
	commands := m.tickTimerManager.GetTriggersToFire(0)
	if len(commands) != 0 {
		t.Errorf("Expected no triggers to fire at T:6, got %d", len(commands))
	}

	// Wait 1 second to reach T:5
	time.Sleep(1 * time.Second)

	// Check that trigger fires
	commands = m.tickTimerManager.GetTriggersToFire(0)
	if len(commands) != 1 {
		t.Errorf("Expected 1 trigger to fire at T:5, got %d", len(commands))
	} else if commands[0] != "test command" {
		t.Errorf("Expected 'test command', got '%s'", commands[0])
	}
}

// TestTickTimerMessage tests that tickTimerMsg is handled correctly
func TestTickTimerMessage(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval and timer
	m.tickTimerManager.TickInterval = 10 // Short interval for testing
	m.tickTimerManager.AddTrigger(5, "test command")
	m.tickTimerManager.UpdateFromPrompt(6)

	// Wait 1 second to reach T:5
	time.Sleep(1 * time.Second)

	// Send tickTimerMsg
	msg := tickTimerMsg{}
	_, cmd := m.Update(msg)

	// Check that command was scheduled
	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	// Check that commands were queued
	if len(m.pendingCommands) != 1 {
		t.Errorf("Expected 1 command in queue, got %d", len(m.pendingCommands))
	} else if m.pendingCommands[0] != "test command" {
		t.Errorf("Expected 'test command' in queue, got '%s'", m.pendingCommands[0])
	}

	// Check that command queue was activated
	if !m.commandQueueActive {
		t.Error("Expected command queue to be active")
	}
}

// TestTickTriggersListCommand tests the /ticktriggers list command
func TestTickTriggersListCommand(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 75
	m.tickTimerManager.UpdateFromPrompt(24)

	// Add some tick triggers
	m.tickTimerManager.AddTrigger(5, "cast 'heal'")
	m.tickTimerManager.AddTrigger(10, "cast 'bless'")

	// List triggers
	m.handleTickTriggersListCommand()

	// Check that output contains the triggers
	found5 := false
	found10 := false
	for _, line := range m.output {
		if contains(line, "T:5") && contains(line, "cast 'heal'") {
			found5 = true
		}
		if contains(line, "T:10") && contains(line, "cast 'bless'") {
			found10 = true
		}
	}

	if !found5 {
		t.Error("Expected to find T:5 trigger in output")
	}
	if !found10 {
		t.Error("Expected to find T:10 trigger in output")
	}
}

// TestTickTriggersRemoveCommand tests the /ticktriggers remove command
func TestTickTriggersRemoveCommand(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 75
	m.tickTimerManager.UpdateFromPrompt(24)

	// Add some tick triggers
	m.tickTimerManager.AddTrigger(5, "cast 'heal'")
	m.tickTimerManager.AddTrigger(10, "cast 'bless'")

	// Remove first trigger (1-based index)
	m.handleTickTriggersRemoveCommand(1)

	// Check that only one trigger remains
	if len(m.tickTimerManager.TickTriggers) != 1 {
		t.Fatalf("Expected 1 trigger after removal, got %d", len(m.tickTimerManager.TickTriggers))
	}

	// Check that the remaining trigger is the second one
	if m.tickTimerManager.TickTriggers[0].TickTime != 10 {
		t.Errorf("Expected remaining trigger at T:10, got T:%d", m.tickTimerManager.TickTriggers[0].TickTime)
	}
}

// TestTickPromptWithANSI tests that tick detection works with ANSI color codes
func TestTickPromptWithANSI(t *testing.T) {
	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Simulate receiving a prompt with ANSI codes and T:15
	promptLine := "\x1b[1;32m101H\x1b[0m \x1b[1;36m132V\x1b[0m T:15 Exits:NS>"

	// Detect tick time
	m.detectTickPrompt(promptLine)

	// Check that tick timer was updated
	if m.tickTimerManager.LastSeenValue != 15 {
		t.Errorf("Expected last seen value 15, got %d", m.tickTimerManager.LastSeenValue)
	}
}

// TestMultipleTickTriggersSameTime tests that multiple triggers at the same time all fire
func TestMultipleTickTriggersSameTime(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 10
	m.tickTimerManager.AddTrigger(5, "command1")
	m.tickTimerManager.AddTrigger(5, "command2")
	m.tickTimerManager.AddTrigger(5, "command3")

	// Set tick time to 6
	m.tickTimerManager.UpdateFromPrompt(6)

	// Wait 1 second to reach T:5
	time.Sleep(1 * time.Second)

	// Check that all triggers fire
	commands := m.tickTimerManager.GetTriggersToFire(0)
	if len(commands) != 3 {
		t.Errorf("Expected 3 triggers to fire, got %d", len(commands))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
