package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/aliases"
	"github.com/anicolao/dikuclient/internal/mapper"
	"github.com/anicolao/dikuclient/internal/triggers"
	tea "github.com/charmbracelet/bubbletea"
)

// TestAliasWithSemicolon tests that aliases with semicolons split into multiple commands
func TestAliasWithSemicolon(t *testing.T) {
	// Create alias manager and add an alias with multiple commands
	aliasManager := aliases.NewManager()
	_, err := aliasManager.Add("test", "north;east;south")
	if err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	// Test that the alias expands correctly
	expanded, ok := aliasManager.Expand("test")
	if !ok {
		t.Fatal("Expected alias to be expanded")
	}
	if expanded != "north;east;south" {
		t.Errorf("Expected expansion to be 'north;east;south', got '%s'", expanded)
	}

	// Test that enqueueCommands works
	m := &Model{
		output:       []string{},
		connected:    true,
		aliasManager: aliasManager,
		worldMap:     mapper.NewMap(),
	}

	// Directly test enqueueCommands with split commands
	commands := strings.Split(expanded, ";")
	for i := range commands {
		commands[i] = strings.TrimSpace(commands[i])
	}
	
	cmd := m.enqueueCommands(commands)
	
	// Check that commands were enqueued
	if len(m.pendingCommands) != 3 {
		t.Errorf("Expected 3 commands in queue, got %d", len(m.pendingCommands))
	}
	
	if len(m.pendingCommands) >= 3 {
		if m.pendingCommands[0] != "north" {
			t.Errorf("Expected first command to be 'north', got '%s'", m.pendingCommands[0])
		}
		if m.pendingCommands[1] != "east" {
			t.Errorf("Expected second command to be 'east', got '%s'", m.pendingCommands[1])
		}
		if m.pendingCommands[2] != "south" {
			t.Errorf("Expected third command to be 'south', got '%s'", m.pendingCommands[2])
		}
	}
	
	// Check that command queue is active
	if !m.commandQueueActive {
		t.Error("Expected command queue to be active")
	}
	
	// Check that a tick command was returned to start processing
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned to start queue processing")
	}
}

// TestTriggerWithSemicolon tests that triggers with semicolons split into multiple commands
func TestTriggerWithSemicolon(t *testing.T) {
	// Create trigger manager and add a trigger with multiple commands
	triggerManager := triggers.NewManager()
	_, err := triggerManager.Add("You are hungry", "eat bread;drink water;rest")
	if err != nil {
		t.Fatalf("Failed to add trigger: %v", err)
	}

	// Test that the trigger matches and returns the action
	actions := triggerManager.Match("You are hungry")
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	action := actions[0]
	if action != "eat bread;drink water;rest" {
		t.Errorf("Expected action to be 'eat bread;drink water;rest', got '%s'", action)
	}

	// Test that enqueueCommands works with the action split on semicolon
	m := &Model{
		output:         []string{},
		connected:      true,
		triggerManager: triggerManager,
		worldMap:       mapper.NewMap(),
	}

	// Split action on semicolon
	commands := strings.Split(action, ";")
	for i := range commands {
		commands[i] = strings.TrimSpace(commands[i])
	}
	
	cmd := m.enqueueCommands(commands)
	
	// Check that commands were enqueued
	if len(m.pendingCommands) != 3 {
		t.Errorf("Expected 3 commands in queue, got %d", len(m.pendingCommands))
	}
	
	if len(m.pendingCommands) >= 3 {
		if m.pendingCommands[0] != "eat bread" {
			t.Errorf("Expected first command to be 'eat bread', got '%s'", m.pendingCommands[0])
		}
		if m.pendingCommands[1] != "drink water" {
			t.Errorf("Expected second command to be 'drink water', got '%s'", m.pendingCommands[1])
		}
		if m.pendingCommands[2] != "rest" {
			t.Errorf("Expected third command to be 'rest', got '%s'", m.pendingCommands[2])
		}
	}
	
	// Check that command queue is active
	if !m.commandQueueActive {
		t.Error("Expected command queue to be active")
	}
	
	// Check that a tick command was returned
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned to start queue processing")
	}
}

// TestStopCommand tests that /stop clears the command queue
func TestStopCommand(t *testing.T) {
	m := &Model{
		output:             []string{},
		connected:          true,
		worldMap:           mapper.NewMap(),
		pendingCommands:    []string{"north", "east", "south"},
		commandQueueActive: true,
		autoWalking:        true,
		autoWalkPath:       []string{"north", "east"},
		autoWalkIndex:      1,
	}

	// Call /stop
	m.handleStopCommand()

	// Check that command queue is cleared
	if len(m.pendingCommands) != 0 {
		t.Errorf("Expected command queue to be empty, got %d commands", len(m.pendingCommands))
	}

	// Check that command queue is not active
	if m.commandQueueActive {
		t.Error("Expected command queue to be inactive")
	}

	// Check that auto-walk is stopped
	if m.autoWalking {
		t.Error("Expected auto-walking to be stopped")
	}

	if m.autoWalkPath != nil {
		t.Error("Expected autoWalkPath to be nil")
	}

	// Check for stop message in output
	foundStopMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "stopped") {
			foundStopMessage = true
			break
		}
	}
	if !foundStopMessage {
		t.Error("Expected stop message in output")
	}
}

// TestStopCommandWhenNotActive tests /stop when nothing is active
func TestStopCommandWhenNotActive(t *testing.T) {
	m := &Model{
		output:             []string{},
		connected:          true,
		worldMap:           mapper.NewMap(),
		commandQueueActive: false,
		autoWalking:        false,
	}

	// Call /stop
	m.handleStopCommand()

	// Check for appropriate message in output
	foundMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "No active") {
			foundMessage = true
			break
		}
	}
	if !foundMessage {
		t.Error("Expected 'No active' message in output")
	}
}

// TestCommandQueueTick tests processing of command queue tick
func TestCommandQueueTick(t *testing.T) {
	m := &Model{
		output:             []string{},
		connected:          true,
		worldMap:           mapper.NewMap(),
		pendingCommands:    []string{"north", "east"},
		commandQueueActive: true,
	}

	// Process first tick - should send "north"
	msg := commandQueueTickMsg{}
	model, cmd := m.Update(msg)
	m = model.(*Model)

	// Check that first command was removed
	if len(m.pendingCommands) != 1 {
		t.Errorf("Expected 1 command remaining in queue, got %d", len(m.pendingCommands))
	}

	if len(m.pendingCommands) >= 1 && m.pendingCommands[0] != "east" {
		t.Errorf("Expected remaining command to be 'east', got '%s'", m.pendingCommands[0])
	}

	// Check that queue is still active
	if !m.commandQueueActive {
		t.Error("Expected command queue to still be active")
	}

	// Check that another tick was scheduled
	if cmd == nil {
		t.Error("Expected another tick to be scheduled")
	}

	// Process second tick - should send "east" and complete
	msg = commandQueueTickMsg{}
	model, cmd = m.Update(msg)
	m = model.(*Model)

	// Check that queue is empty
	if len(m.pendingCommands) != 0 {
		t.Errorf("Expected queue to be empty, got %d commands", len(m.pendingCommands))
	}

	// Check that queue is no longer active
	if m.commandQueueActive {
		t.Error("Expected command queue to be inactive after completion")
	}
}

// TestMultipleTriggersAddToQueue tests that multiple triggers add to the end of the queue
func TestMultipleTriggersAddToQueue(t *testing.T) {
	// Create trigger manager with two triggers
	triggerManager := triggers.NewManager()
	_, err := triggerManager.Add("Pattern A", "cmd1;cmd2")
	if err != nil {
		t.Fatalf("Failed to add trigger A: %v", err)
	}
	_, err = triggerManager.Add("Pattern B", "cmd3;cmd4")
	if err != nil {
		t.Fatalf("Failed to add trigger B: %v", err)
	}

	m := &Model{
		output:         []string{},
		connected:      true,
		triggerManager: triggerManager,
		worldMap:       mapper.NewMap(),
	}

	// Get actions for trigger A and enqueue them
	actionsA := triggerManager.Match("Pattern A")
	if len(actionsA) > 0 {
		commandsA := strings.Split(actionsA[0], ";")
		for i := range commandsA {
			commandsA[i] = strings.TrimSpace(commandsA[i])
		}
		m.enqueueCommands(commandsA)
	}

	// Check that 2 commands were enqueued
	if len(m.pendingCommands) != 2 {
		t.Errorf("Expected 2 commands in queue after first trigger, got %d", len(m.pendingCommands))
	}

	// Get actions for trigger B and enqueue them
	actionsB := triggerManager.Match("Pattern B")
	if len(actionsB) > 0 {
		commandsB := strings.Split(actionsB[0], ";")
		for i := range commandsB {
			commandsB[i] = strings.TrimSpace(commandsB[i])
		}
		m.enqueueCommands(commandsB)
	}

	// Check that 4 commands total are now in queue (2 from A + 2 from B)
	if len(m.pendingCommands) != 4 {
		t.Errorf("Expected 4 commands in queue after both triggers, got %d", len(m.pendingCommands))
	}

	// Verify the order
	expectedCommands := []string{"cmd1", "cmd2", "cmd3", "cmd4"}
	for i, expected := range expectedCommands {
		if i < len(m.pendingCommands) && m.pendingCommands[i] != expected {
			t.Errorf("Expected command %d to be '%s', got '%s'", i, expected, m.pendingCommands[i])
		}
	}
}

// TestDirectCommandWithSemicolon tests that user-typed commands with semicolons are split
func TestDirectCommandWithSemicolon(t *testing.T) {
	m := &Model{
		output:       []string{},
		connected:    true,
		aliasManager: aliases.NewManager(), // Empty alias manager
		worldMap:     mapper.NewMap(),
	}

	// Test that a command with semicolons gets split and enqueued
	command := "north;east;south"
	commands := strings.Split(command, ";")
	for i := range commands {
		commands[i] = strings.TrimSpace(commands[i])
	}
	
	cmd := m.enqueueCommands(commands)
	
	// Check that commands were enqueued
	if len(m.pendingCommands) != 3 {
		t.Errorf("Expected 3 commands in queue, got %d", len(m.pendingCommands))
	}
	
	if len(m.pendingCommands) >= 3 {
		if m.pendingCommands[0] != "north" {
			t.Errorf("Expected first command to be 'north', got '%s'", m.pendingCommands[0])
		}
		if m.pendingCommands[1] != "east" {
			t.Errorf("Expected second command to be 'east', got '%s'", m.pendingCommands[1])
		}
		if m.pendingCommands[2] != "south" {
			t.Errorf("Expected third command to be 'south', got '%s'", m.pendingCommands[2])
		}
	}
	
	// Check that command queue is active
	if !m.commandQueueActive {
		t.Error("Expected command queue to be active")
	}
	
	// Check that a tick command was returned
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned to start queue processing")
	}
}

// TestSingleCommandNotQueued tests that a single command (no semicolon) is sent immediately
func TestSingleCommandNotQueued(t *testing.T) {
	m := &Model{
		output:       []string{},
		connected:    true,
		aliasManager: aliases.NewManager(),
		worldMap:     mapper.NewMap(),
	}

	// Simulate user typing a single command
	m.currentInput = "north"
	
	// Process Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	model, _ := m.Update(msg)
	m = model.(*Model)
	
	// Check that no commands were enqueued (sent immediately)
	if len(m.pendingCommands) != 0 {
		t.Errorf("Expected 0 commands in queue (sent immediately), got %d", len(m.pendingCommands))
	}
	
	// Check that command queue is not active
	if m.commandQueueActive {
		t.Error("Expected command queue to be inactive for single command")
	}
}
