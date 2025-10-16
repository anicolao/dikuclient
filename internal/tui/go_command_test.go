package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestGoCommandCancellation tests that /go with no args cancels auto-walking
func TestGoCommandCancellation(t *testing.T) {
	m := Model{
		output:        []string{},
		connected:     true,
		worldMap:      mapper.NewMap(),
		autoWalking:   true,
		autoWalkPath:  []string{"north", "east"},
		autoWalkIndex: 0,
	}

	// Call /go with no arguments - should cancel
	m.handleGoCommand([]string{})

	if m.autoWalking {
		t.Error("Expected auto-walking to be cancelled")
	}
	if m.autoWalkPath != nil {
		t.Error("Expected autoWalkPath to be nil")
	}
	if m.autoWalkIndex != 0 {
		t.Error("Expected autoWalkIndex to be reset to 0")
	}

	// Check output message
	foundCancelMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "cancelled") {
			foundCancelMessage = true
			break
		}
	}
	if !foundCancelMessage {
		t.Error("Expected cancellation message in output")
	}
}

// TestGoCommandNoArgsNotWalking tests that /go with no args shows usage when not auto-walking
func TestGoCommandNoArgsNotWalking(t *testing.T) {
	m := Model{
		output:      []string{},
		connected:   true,
		worldMap:    mapper.NewMap(),
		autoWalking: false,
	}

	// Call /go with no arguments when not walking - should show usage
	m.handleGoCommand([]string{})

	if m.autoWalking {
		t.Error("Should not start auto-walking")
	}

	// Check for usage message
	foundUsage := false
	for _, line := range m.output {
		if strings.Contains(line, "Usage") {
			foundUsage = true
			break
		}
	}
	if !foundUsage {
		t.Error("Expected usage message in output")
	}
}

// TestGoCommandNumericSelection tests selecting a room by number
func TestGoCommandNumericSelection(t *testing.T) {
	worldMap := mapper.NewMap()

	// Create and add test rooms
	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Temple Entrance", "The entrance to the temple.", []string{"south"})
	room3 := mapper.NewRoom("Market Square", "A busy market.", []string{"east"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.CurrentRoomID = room3.ID // Start at market

	// Link the rooms so there's a path
	room3.Exits["north"] = room1.ID
	room1.Exits["south"] = room3.ID
	room3.Exits["east"] = room2.ID
	room2.Exits["west"] = room3.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// First, search for "temple" to populate lastRoomSearch
	m.handleGoCommand([]string{"temple"})

	// Should have multiple matches
	if len(m.lastRoomSearch) != 2 {
		t.Fatalf("Expected 2 rooms to match 'temple', got %d", len(m.lastRoomSearch))
	}

	// Clear output
	m.output = []string{}

	// Now use numeric selection to pick the first temple
	cmd := m.handleGoCommand([]string{"1", "temple"})

	// Should start auto-walking
	if !m.autoWalking {
		t.Error("Expected to start auto-walking after numeric selection")
	}

	// Check that we got a valid command back (tea.Cmd for the tick)
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned for auto-walk start")
	}
}

// TestGoCommandNumericSelectionFromPreviousSearch tests /go <number> without search terms
func TestGoCommandNumericSelectionFromPreviousSearch(t *testing.T) {
	worldMap := mapper.NewMap()

	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Temple Entrance", "The entrance to the temple.", []string{"south"})
	room3 := mapper.NewRoom("Market Square", "A busy market.", []string{"east"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.CurrentRoomID = room3.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// First, use /rooms to list all rooms
	m.handleRoomsCommand([]string{})

	// Should populate lastRoomSearch
	if len(m.lastRoomSearch) != 3 {
		t.Fatalf("Expected 3 rooms in lastRoomSearch, got %d", len(m.lastRoomSearch))
	}

	// Clear output
	m.output = []string{}

	// Now use /go 1 to select the first room from previous search
	cmd := m.handleGoCommand([]string{"1"})

	// Should start auto-walking (or show "already at location" if it's the current room)
	// With durable room numbering, room 1 is the first room added (Temple Square)
	// Since rooms are not linked in this test, we expect "No path found"
	hasExpectedMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "Auto-walking") || strings.Contains(line, "already at") || strings.Contains(line, "No path found") {
			hasExpectedMessage = true
			break
		}
	}
	if !hasExpectedMessage {
		t.Error("Expected to either start auto-walking, be already at location, or show no path found")
	}

	// cmd might be nil if we're already at the location
	if m.autoWalking && cmd == nil {
		t.Error("Expected a tea.Cmd when auto-walking starts")
	}
}

// TestGoCommandInvalidNumericSelection tests invalid room numbers
func TestGoCommandInvalidNumericSelection(t *testing.T) {
	worldMap := mapper.NewMap()

	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Market Square", "A busy market.", []string{"south"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.CurrentRoomID = room2.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// Search for "temple" to populate lastRoomSearch
	m.handleGoCommand([]string{"temple"})

	// Clear output
	m.output = []string{}

	// Try to select room number 5 when only 1 exists
	m.handleGoCommand([]string{"5", "temple"})

	// Should have error message
	foundError := false
	for _, line := range m.output {
		if strings.Contains(line, "Invalid room number") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected error message for invalid room number")
	}

	// Should not start auto-walking
	if m.autoWalking {
		t.Error("Should not start auto-walking with invalid room number")
	}
}

// TestGoCommandNumericSelectionNoPreviousSearch tests /go <number> with no previous search
func TestGoCommandNumericSelectionNoPreviousSearch(t *testing.T) {
	worldMap := mapper.NewMap()

	m := Model{
		output:         []string{},
		connected:      true,
		worldMap:       worldMap,
		lastRoomSearch: nil,
	}

	// Try to use /go 1 with no previous search
	m.handleGoCommand([]string{"1"})

	// Should have error message
	foundError := false
	for _, line := range m.output {
		if strings.Contains(line, "No previous room search") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected error message when no previous search exists")
	}

	// Should not start auto-walking
	if m.autoWalking {
		t.Error("Should not start auto-walking with no previous search")
	}
}

// TestPointCommandNumericSelection tests that /point also supports numeric selection
func TestPointCommandNumericSelection(t *testing.T) {
	worldMap := mapper.NewMap()

	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Temple Entrance", "The entrance to the temple.", []string{"south"})
	room3 := mapper.NewRoom("Market Square", "A busy market.", []string{"east"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.CurrentRoomID = room3.ID

	// Link the rooms
	room3.Exits["north"] = room1.ID
	room1.Exits["south"] = room3.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// First, search for "temple"
	m.handlePointCommand([]string{"temple"})

	// Should have multiple matches
	if len(m.lastRoomSearch) != 2 {
		t.Fatalf("Expected 2 rooms to match 'temple', got %d", len(m.lastRoomSearch))
	}

	// Clear output
	m.output = []string{}

	// Use numeric selection
	m.handlePointCommand([]string{"1", "temple"})

	// Should show direction
	foundDirection := false
	for _, line := range m.output {
		if strings.Contains(line, "go:") {
			foundDirection = true
			break
		}
	}
	if !foundDirection {
		t.Error("Expected direction message after numeric selection")
	}
}

// TestWayfindCommandNumericSelection tests that /wayfind also supports numeric selection
func TestWayfindCommandNumericSelection(t *testing.T) {
	worldMap := mapper.NewMap()

	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Temple Entrance", "The entrance to the temple.", []string{"south"})
	room3 := mapper.NewRoom("Market Square", "A busy market.", []string{"east"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.CurrentRoomID = room3.ID

	// Link the rooms
	room3.Exits["north"] = room1.ID
	room1.Exits["south"] = room3.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// First, search for "temple"
	m.handleWayfindCommand([]string{"temple"})

	// Should have multiple matches
	if len(m.lastRoomSearch) != 2 {
		t.Fatalf("Expected 2 rooms to match 'temple', got %d", len(m.lastRoomSearch))
	}

	// Clear output
	m.output = []string{}

	// Use numeric selection
	m.handleWayfindCommand([]string{"1", "temple"})

	// Should show path
	foundPath := false
	for _, line := range m.output {
		if strings.Contains(line, "Path to") {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Error("Expected path message after numeric selection")
	}
}

// TestWayfindOutputFormat tests that /wayfind shows step numbers and room names
func TestWayfindOutputFormat(t *testing.T) {
	worldMap := mapper.NewMap()

	room1 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := mapper.NewRoom("Temple Hall", "The grand hall.", []string{"south", "north"})
	room3 := mapper.NewRoom("Inner Sanctum", "The innermost chamber.", []string{"south"})

	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.CurrentRoomID = room1.ID

	// Link the rooms
	room1.Exits["north"] = room2.ID
	room2.Exits["south"] = room1.ID
	room2.Exits["north"] = room3.ID
	room3.Exits["south"] = room2.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// Search for the Inner Sanctum
	m.handleWayfindCommand([]string{"inner", "sanctum"})

	// Check that we got the right output format
	foundPath := false
	foundStepFormat := false
	for _, line := range m.output {
		if strings.Contains(line, "Path to") && strings.Contains(line, "2 steps") {
			foundPath = true
		}
		// Check for step number format: "1. direction -> Room Name"
		if strings.Contains(line, "1. north -> Temple Hall") {
			foundStepFormat = true
		}
	}

	if !foundPath {
		t.Error("Expected path message with step count")
		t.Logf("Output: %v", m.output)
	}

	if !foundStepFormat {
		t.Error("Expected step format: '1. direction -> Room Name'")
		t.Logf("Output: %v", m.output)
	}

	// Log the output to show the new format
	t.Log("Wayfind output format:")
	for _, line := range m.output {
		t.Log("  " + line)
	}
}

// TestGoCommandAfterWayfindDisambiguation tests that /go can use numbers from /wayfind disambiguation lists
func TestGoCommandAfterWayfindDisambiguation(t *testing.T) {
	worldMap := mapper.NewMap()

	// Create rooms where durable room numbers DON'T match the disambiguation list positions
	// Add rooms in a specific order to control their durable numbers
	room1 := mapper.NewRoom("Market Square", "A busy market.", []string{"east"})
	room2 := mapper.NewRoom("Inn", "A cozy inn.", []string{"west"})
	room3 := mapper.NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room4 := mapper.NewRoom("Temple Entrance", "The entrance to the temple.", []string{"south"})
	room5 := mapper.NewRoom("Tavern", "A noisy tavern.", []string{"south"})

	// Add them in order so they get durable numbers 1, 2, 3, 4, 5
	worldMap.AddOrUpdateRoom(room1)
	worldMap.AddOrUpdateRoom(room2)
	worldMap.AddOrUpdateRoom(room3)
	worldMap.AddOrUpdateRoom(room4)
	worldMap.AddOrUpdateRoom(room5)
	
	worldMap.CurrentRoomID = room1.ID

	// Link the rooms
	room1.Exits["north"] = room3.ID
	room3.Exits["south"] = room1.ID

	m := Model{
		output:    []string{},
		connected: true,
		worldMap:  worldMap,
	}

	// First, use /wayfind to search for "temple" - should show disambiguation list
	m.handleWayfindCommand([]string{"temple"})

	// Should have multiple matches
	if len(m.lastRoomSearch) != 2 {
		t.Fatalf("Expected 2 rooms to match 'temple', got %d", len(m.lastRoomSearch))
	}

	// Verify disambiguation list was shown
	foundDisambiguation := false
	for _, line := range m.output {
		if strings.Contains(line, "Found 2 rooms") {
			foundDisambiguation = true
			break
		}
	}
	if !foundDisambiguation {
		t.Error("Expected disambiguation list to be shown")
	}

	// Clear output
	m.output = []string{}

	// Now use /go with the number from the disambiguation list
	t.Logf("Before /go: lastRoomSearch has %d rooms", len(m.lastRoomSearch))
	t.Logf("Before /go: mapLegendRooms has %d rooms", len(m.mapLegendRooms))
	for i, room := range m.lastRoomSearch {
		t.Logf("  lastRoomSearch[%d]: %s (room number %d)", i, room.Title, m.worldMap.GetRoomNumber(room.ID))
	}
	
	// /wayfind shows: 1. Temple Square, 2. Temple Entrance
	// But their durable room numbers are 3 and 4
	// So /go 1 should select "Temple Square" (the first in the list)
	m.handleGoCommand([]string{"1"})

	t.Logf("After /go, output has %d lines:", len(m.output))
	for _, line := range m.output {
		t.Logf("  %s", line)
	}

	// Should start auto-walking to Temple Square (the first in the disambiguation list)
	if !m.autoWalking {
		t.Error("Expected auto-walking to start")
		for _, line := range m.output {
			t.Logf("  Output: %s", line)
		}
	}

	// Check that we're going to Temple Square (first in list), not Market Square (room #1)
	foundTempleSquare := false
	for _, line := range m.output {
		if strings.Contains(line, "Temple Square") {
			foundTempleSquare = true
			break
		}
	}
	if !foundTempleSquare {
		t.Error("Expected to navigate to 'Temple Square' (first item in disambiguation list)")
		for _, line := range m.output {
			t.Logf("  Output: %s", line)
		}
	}
}
