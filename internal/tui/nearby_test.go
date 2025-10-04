package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestNearbyCommand verifies the /nearby command functionality
func TestNearbyCommand(t *testing.T) {
	// Create a model with a simple map
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Create a test map with a center room and surrounding rooms
	center := mapper.NewRoom("Center", "The center room.", []string{"north", "south", "east", "west"})
	north := mapper.NewRoom("North Room", "A room to the north.", []string{"south", "north"})
	farNorth := mapper.NewRoom("Far North", "A distant northern room.", []string{"south"})
	south := mapper.NewRoom("South Room", "A room to the south.", []string{"north"})
	east := mapper.NewRoom("East Room", "A room to the east.", []string{"west"})
	west := mapper.NewRoom("West Room", "A room to the west.", []string{"east"})

	// Add all rooms to the map
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("north")
	m.worldMap.AddOrUpdateRoom(north)
	m.worldMap.SetLastDirection("north")
	m.worldMap.AddOrUpdateRoom(farNorth)
	
	// Go back to center
	m.worldMap.SetLastDirection("south")
	m.worldMap.AddOrUpdateRoom(north)
	m.worldMap.SetLastDirection("south")
	m.worldMap.AddOrUpdateRoom(center)
	
	// Add other rooms
	m.worldMap.SetLastDirection("south")
	m.worldMap.AddOrUpdateRoom(south)
	m.worldMap.SetLastDirection("north")
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("east")
	m.worldMap.AddOrUpdateRoom(east)
	m.worldMap.SetLastDirection("west")
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("west")
	m.worldMap.AddOrUpdateRoom(west)
	m.worldMap.SetLastDirection("east")
	m.worldMap.AddOrUpdateRoom(center)

	// Execute the /nearby command
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/nearby\x1b[0m"
	m.handleClientCommand("/nearby")
	m.output = append(m.output, "")
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	// Verify the output contains nearby rooms
	hasNearbyHeader := false
	foundNorthRoom := false
	foundSouthRoom := false
	foundEastRoom := false
	foundWestRoom := false
	foundFarNorth := false

	for _, line := range m.output {
		cleanLine := strings.ReplaceAll(line, "\x1b[93m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[0m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[92m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[96m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[90m", "")

		if strings.Contains(cleanLine, "Nearby Rooms") {
			hasNearbyHeader = true
		}
		if strings.Contains(cleanLine, "North Room") {
			foundNorthRoom = true
		}
		if strings.Contains(cleanLine, "South Room") {
			foundSouthRoom = true
		}
		if strings.Contains(cleanLine, "East Room") {
			foundEastRoom = true
		}
		if strings.Contains(cleanLine, "West Room") {
			foundWestRoom = true
		}
		if strings.Contains(cleanLine, "Far North") {
			foundFarNorth = true
		}
	}

	if !hasNearbyHeader {
		t.Error("Expected to find 'Nearby Rooms' header in output")
	}

	// All adjacent rooms should be found
	if !foundNorthRoom {
		t.Error("Expected to find North Room in nearby list")
	}
	if !foundSouthRoom {
		t.Error("Expected to find South Room in nearby list")
	}
	if !foundEastRoom {
		t.Error("Expected to find East Room in nearby list")
	}
	if !foundWestRoom {
		t.Error("Expected to find West Room in nearby list")
	}

	// Far North should also be found (2 steps away)
	if !foundFarNorth {
		t.Error("Expected to find Far North in nearby list")
	}
}

// TestNearbyCommandNoCurrentRoom verifies behavior when no current room is set
func TestNearbyCommandNoCurrentRoom(t *testing.T) {
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Add a room but don't set it as current
	room := mapper.NewRoom("Test Room", "A test room.", []string{"north"})
	m.worldMap.Rooms[room.ID] = room

	// Execute the /nearby command
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/nearby\x1b[0m"
	m.handleClientCommand("/nearby")

	// Should have error message
	hasError := false
	for _, line := range m.output {
		if strings.Contains(line, "No current room") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error message when no current room is set")
	}
}

// TestNearbyCommandNoNearbyRooms verifies behavior when there are no nearby rooms
func TestNearbyCommandNoNearbyRooms(t *testing.T) {
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Create a single isolated room
	room := mapper.NewRoom("Isolated Room", "A lonely room.", []string{})
	m.worldMap.AddOrUpdateRoom(room)

	// Execute the /nearby command
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/nearby\x1b[0m"
	m.handleClientCommand("/nearby")

	// Should have "no nearby rooms" message
	hasMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "No nearby rooms") {
			hasMessage = true
			break
		}
	}
	if !hasMessage {
		t.Error("Expected 'No nearby rooms' message for isolated room")
	}
}
