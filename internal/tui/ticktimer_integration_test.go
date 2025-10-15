package tui

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestTickTimerIntegration simulates receiving prompts with T:NN from a MUD server
func TestTickTimerIntegration(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Simulate receiving combat prompts with tick times
	mudOutput := []string{
		"101H 132V 54710X 49.60% 570C [Hero:Good] [goblin:Bad] T:24 Exits:NS>",
		"You hit the goblin.",
		"The goblin hits you.",
		"101H 130V 54710X 49.60% 570C [Hero:Good] [goblin:V.Bad] T:23 Exits:NS>",
		"You hit the goblin.",
		"The goblin is dead! R.I.P.",
		"You receive 45 experience.",
		"101H 130V 54755X 49.60% 570C T:22 Exits:NS>",
	}

	// Process each line as if coming from the MUD
	for _, line := range mudOutput {
		m.output = append(m.output, line)
		m.detectTickPrompt(line)
	}

	// Verify that tick timer was updated with T:22 (the last seen value)
	if m.tickTimerManager.LastSeenValue != 22 {
		t.Errorf("Expected last seen tick value 22, got %d", m.tickTimerManager.LastSeenValue)
	}

	// Verify that tick interval was set (should default to 75)
	if m.tickTimerManager.TickInterval == 0 {
		t.Error("Tick interval should be set after detecting tick in prompt")
	}

	// Verify that current tick time is close to 22
	currentTick := m.tickTimerManager.GetCurrentTickTime()
	if currentTick < 21 || currentTick > 23 {
		t.Errorf("Expected current tick time around 22, got %d", currentTick)
	}
}

// TestTickTriggerIntegration simulates a complete tick trigger scenario
func TestTickTriggerIntegration(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Set up tick timer with a short interval for testing
	m.tickTimerManager.TickInterval = 10 // 10 second interval

	// Add a tick trigger at T:5
	err := m.tickTimerManager.AddTrigger(5, "cast 'heal'")
	if err != nil {
		t.Fatalf("Failed to add tick trigger: %v", err)
	}

	// Simulate receiving a prompt with T:6
	prompt := "100H 100V T:6 Exits:NS>"
	m.detectTickPrompt(prompt)

	// Verify trigger was added
	if len(m.tickTimerManager.TickTriggers) != 1 {
		t.Fatalf("Expected 1 tick trigger, got %d", len(m.tickTimerManager.TickTriggers))
	}

	// Wait 1 second for tick time to reach T:5
	time.Sleep(1 * time.Second)

	// Simulate tickTimerMsg
	msg := tickTimerMsg{}
	_, cmd := m.Update(msg)

	// Verify that command was scheduled
	if cmd == nil {
		t.Error("Expected command to be scheduled")
	}

	// Verify that commands were queued
	if len(m.pendingCommands) != 1 {
		t.Errorf("Expected 1 command in queue, got %d", len(m.pendingCommands))
	} else if m.pendingCommands[0] != "cast 'heal'" {
		t.Errorf("Expected 'cast 'heal'' in queue, got '%s'", m.pendingCommands[0])
	}
}

// TestTickTriggerMultiCommand simulates tick trigger with multiple commands
func TestTickTriggerMultiCommand(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Set up tick timer with a short interval
	m.tickTimerManager.TickInterval = 10

	// Add a tick trigger with multiple commands
	err := m.tickTimerManager.AddTrigger(5, "cast 'heal';cast 'bless';say Ready!")
	if err != nil {
		t.Fatalf("Failed to add tick trigger: %v", err)
	}

	// Simulate receiving a prompt with T:6
	m.detectTickPrompt("100H 100V T:6 Exits:NS>")

	// Wait for tick time to reach T:5
	time.Sleep(1 * time.Second)

	// Simulate tickTimerMsg
	msg := tickTimerMsg{}
	_, cmd := m.Update(msg)

	// Verify that all commands were queued
	if len(m.pendingCommands) != 3 {
		t.Errorf("Expected 3 commands in queue, got %d", len(m.pendingCommands))
	}

	expectedCommands := []string{"cast 'heal'", "cast 'bless'", "say Ready!"}
	for i, expected := range expectedCommands {
		if i >= len(m.pendingCommands) {
			t.Errorf("Missing command %d: %q", i, expected)
			continue
		}
		if m.pendingCommands[i] != expected {
			t.Errorf("Command %d: expected %q, got %q", i, expected, m.pendingCommands[i])
		}
	}

	// Verify command was scheduled
	if cmd == nil {
		t.Error("Expected command to be scheduled")
	}
}

// TestTickTriggerPersistence verifies that tick triggers are saved and loaded
func TestTickTriggerPersistence(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 75
	m.detectTickPrompt("100H 100V T:50 Exits:NS>")

	// Add some tick triggers
	m.tickTimerManager.AddTrigger(5, "cast 'heal'")
	m.tickTimerManager.AddTrigger(10, "cast 'bless'")

	// Save tick timer
	err := m.tickTimerManager.Save()
	if err != nil {
		t.Fatalf("Failed to save tick timer: %v", err)
	}

	// Create a new model (simulating a restart)
	m2 := NewModel("test.mud", 4000, nil, nil)

	// Verify that tick triggers were loaded
	if len(m2.tickTimerManager.TickTriggers) != 2 {
		t.Errorf("Expected 2 tick triggers after reload, got %d", len(m2.tickTimerManager.TickTriggers))
	}

	// Verify trigger details
	if m2.tickTimerManager.TickTriggers[0].TickTime != 5 {
		t.Errorf("Expected first trigger at T:5, got T:%d", m2.tickTimerManager.TickTriggers[0].TickTime)
	}

	if m2.tickTimerManager.TickTriggers[1].TickTime != 10 {
		t.Errorf("Expected second trigger at T:10, got T:%d", m2.tickTimerManager.TickTriggers[1].TickTime)
	}
}

// TestTickTriggerOutputFormatting tests the output messages for tick triggers
func TestTickTriggerOutputFormatting(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Initialize tick interval
	m.tickTimerManager.TickInterval = 75
	m.detectTickPrompt("100H 100V T:50 Exits:NS>")

	// Add a tick trigger
	command := "ticktrigger 5 \"cast 'heal'\""
	m.handleTickTriggerCommand(command)

	// Check that success message was added
	found := false
	for _, line := range m.output {
		if strings.Contains(line, "Tick trigger added") && strings.Contains(line, "T:5") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected success message in output")
	}

	// Clear output
	m.output = []string{}

	// List tick triggers
	m.handleTickTriggersListCommand()

	// Check that list output is formatted correctly
	foundHeader := false
	foundTrigger := false
	for _, line := range m.output {
		if strings.Contains(line, "Active Tick Triggers") && strings.Contains(line, "Interval: 75s") {
			foundHeader = true
		}
		if strings.Contains(line, "T:5") && strings.Contains(line, "cast 'heal'") {
			foundTrigger = true
		}
	}

	if !foundHeader {
		t.Error("Expected header with interval in list output")
	}

	if !foundTrigger {
		t.Error("Expected trigger details in list output")
	}
}

// TestTickTimerWithRealPrompts tests with realistic MUD prompts
func TestTickTimerWithRealPrompts(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a model
	m := NewModel("test.mud", 4000, nil, nil)

	// Test various prompt formats that might appear in real MUDs
	testPrompts := []struct {
		prompt      string
		expectedVal int
	}{
		{"101H 132V 54710X 49.60% 570C T:24 Exits:NS>", 24},
		{"[100H 200M 300V] T:15 >", 15},
		{"\x1b[1;32m101H\x1b[0m \x1b[1;36m132V\x1b[0m T:5 Exits:NS>", 5},
		{"HP:100 MP:50 T:60 >", 60},
		{"< 100hp 50ma 75mv > T:1 [NS]", 1},
	}

	for _, test := range testPrompts {
		m.detectTickPrompt(test.prompt)
		if m.tickTimerManager.LastSeenValue != test.expectedVal {
			t.Errorf("For prompt %q: expected tick value %d, got %d",
				test.prompt, test.expectedVal, m.tickTimerManager.LastSeenValue)
		}
	}
}
