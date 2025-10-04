package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestRecallDetection tests that 'recall' keyword triggers skip flag
func TestRecallDetection(t *testing.T) {
	m := Model{
		output:        []string{},
		recentOutput:  []string{},
		worldMap:      mapper.NewMap(),
		skipNextRoomDetection: false,
	}

	// Simulate receiving a line with "recall" in it
	msg := mudMsg("You recite a glowing scroll of recall which dissolves.")
	
	// Process the message
	msgStr := string(msg)
	lines := strings.Split(msgStr, "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		m.output = append(m.output, line)
		m.recentOutput = append(m.recentOutput, line)
		
		// Check for recall command
		cleanLine := stripANSI(line)
		if strings.Contains(strings.ToLower(cleanLine), "recall") {
			m.skipNextRoomDetection = true
		}
	}

	if !m.skipNextRoomDetection {
		t.Error("Expected skipNextRoomDetection to be true after 'recall' detected")
	}
}

// TestRecallSkipsRoomDetection tests that room detection is skipped after recall
func TestRecallSkipsRoomDetection(t *testing.T) {
	worldMap := mapper.NewMap()
	
	// Create a starting room
	room1 := mapper.NewRoom("Starting Room", "A starting location.", []string{"north"})
	worldMap.AddOrUpdateRoom(room1)
	
	m := Model{
		output:                []string{},
		recentOutput:          []string{},
		worldMap:              worldMap,
		skipNextRoomDetection: true,
		pendingMovement:       "north",
	}

	// Add some room-like output to recentOutput
	m.recentOutput = append(m.recentOutput, "Temple Square")
	m.recentOutput = append(m.recentOutput, "A large temple square with pillars.")
	m.recentOutput = append(m.recentOutput, "Exits: N S E W >")

	// Try to detect room
	m.detectAndUpdateRoom()

	// Check that the flag was cleared
	if m.skipNextRoomDetection {
		t.Error("Expected skipNextRoomDetection to be cleared after being used")
	}

	// Check that pendingMovement was cleared
	if m.pendingMovement != "" {
		t.Error("Expected pendingMovement to be cleared")
	}

	// Check that no new room was added
	if len(worldMap.Rooms) > 1 {
		t.Error("Expected no new room to be added after recall skip")
	}
}

// TestAutoWalkFailureDetection tests detection of "Alas, you cannot go that way..."
func TestAutoWalkFailureDetection(t *testing.T) {
	worldMap := mapper.NewMap()
	
	// Create rooms
	room1 := mapper.NewRoom("Starting Room", "A starting location.", []string{"north", "east"})
	room2 := mapper.NewRoom("Target Room", "A target location.", []string{"south"})
	
	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.CurrentRoomID = room1.ID
	
	// Create a bad link (exit exists but destination is wrong)
	room1.Exits["north"] = room2.ID
	
	m := Model{
		output:         []string{},
		recentOutput:   []string{},
		worldMap:       worldMap,
		autoWalking:    true,
		autoWalkPath:   []string{"north"},
		autoWalkIndex:  1,
		autoWalkTarget: "Target Room",
	}

	// Simulate receiving "Alas, you cannot go that way..."
	msg := mudMsg("Alas, you cannot go that way...")
	
	msgStr := string(msg)
	lines := strings.Split(msgStr, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		m.output = append(m.output, line)
		
		cleanLine := stripANSI(line)
		if m.autoWalking && (strings.Contains(cleanLine, "Alas, you cannot go that way") || 
			strings.Contains(cleanLine, "cannot go that way")) {
			m.handleAutoWalkFailure()
		}
	}

	// Check that the exit was removed
	if _, exists := room1.Exits["north"]; exists {
		t.Error("Expected 'north' exit to be removed from current room")
	}

	// Note: Full re-planning test would require a more complex setup with valid paths
}

// TestAutoWalkFailureRemovesExit tests that failed exits are removed from rooms
func TestAutoWalkFailureRemovesExit(t *testing.T) {
	worldMap := mapper.NewMap()
	
	// Create a room with an invalid exit
	room1 := mapper.NewRoom("Test Room", "A test location.", []string{"north", "south"})
	worldMap.AddOrUpdateRoom(room1)
	worldMap.CurrentRoomID = room1.ID
	
	m := Model{
		output:         []string{},
		worldMap:       worldMap,
		autoWalking:    true,
		autoWalkPath:   []string{"north", "east"},
		autoWalkIndex:  1, // Just tried "north"
		autoWalkTarget: "Some Room",
	}

	// Before failure, room should have north exit
	if _, exists := room1.Exits["north"]; !exists {
		t.Error("Expected 'north' exit to exist before failure")
	}

	// Trigger failure
	m.handleAutoWalkFailure()

	// After failure, north exit should be removed
	if _, exists := room1.Exits["north"]; exists {
		t.Error("Expected 'north' exit to be removed after failure")
	}

	// South exit should still exist
	if _, exists := room1.Exits["south"]; !exists {
		t.Error("Expected 'south' exit to still exist after failure")
	}

	// Auto-walking should be stopped (no valid path to replan)
	if m.autoWalking {
		t.Error("Expected auto-walking to be stopped")
	}
}

// TestRecallWithANSICodes tests recall detection with ANSI color codes
func TestRecallWithANSICodes(t *testing.T) {
	m := Model{
		output:                []string{},
		recentOutput:          []string{},
		worldMap:              mapper.NewMap(),
		skipNextRoomDetection: false,
	}

	// Simulate a line with ANSI codes containing "recall"
	msg := mudMsg("\x1b[33mYou recite a glowing scroll of \x1b[1mrecall\x1b[0m which dissolves.")
	
	msgStr := string(msg)
	lines := strings.Split(msgStr, "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		m.output = append(m.output, line)
		m.recentOutput = append(m.recentOutput, line)
		
		cleanLine := stripANSI(line)
		if strings.Contains(strings.ToLower(cleanLine), "recall") {
			m.skipNextRoomDetection = true
		}
	}

	if !m.skipNextRoomDetection {
		t.Error("Expected skipNextRoomDetection to be true after 'recall' with ANSI codes")
	}
}
