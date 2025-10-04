package tui

import (
	"strings"
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestLegendCommand verifies the /legend command functionality
func TestLegendCommand(t *testing.T) {
	// Create a model with a simple map
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Create a test map with several rooms
	center := mapper.NewRoom("Center", "The center room.", []string{"north", "south", "east"})
	north := mapper.NewRoom("North Room", "A room to the north.", []string{"south"})
	south := mapper.NewRoom("South Room", "A room to the south.", []string{"north"})
	east := mapper.NewRoom("East Room", "A room to the east.", []string{"west"})

	// Add all rooms to the map
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("north")
	m.worldMap.AddOrUpdateRoom(north)
	m.worldMap.SetLastDirection("south")
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("south")
	m.worldMap.AddOrUpdateRoom(south)
	m.worldMap.SetLastDirection("north")
	m.worldMap.AddOrUpdateRoom(center)
	m.worldMap.SetLastDirection("east")
	m.worldMap.AddOrUpdateRoom(east)
	m.worldMap.SetLastDirection("west")
	m.worldMap.AddOrUpdateRoom(center)

	// Execute the /legend command
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/legend\x1b[0m"
	m.handleClientCommand("/legend")
	m.output = append(m.output, "")
	m.output = append(m.output, "")
	m.output = append(m.output, savedPrompt)

	// Verify the output contains all rooms
	hasLegendHeader := false
	foundCenter := false
	foundNorth := false
	foundSouth := false
	foundEast := false

	for _, line := range m.output {
		cleanLine := strings.ReplaceAll(line, "\x1b[93m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[0m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[92m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[96m", "")
		cleanLine = strings.ReplaceAll(cleanLine, "\x1b[90m", "")

		if strings.Contains(cleanLine, "All Mapped Rooms") {
			hasLegendHeader = true
		}
		if strings.Contains(cleanLine, "Center") {
			foundCenter = true
		}
		if strings.Contains(cleanLine, "North Room") {
			foundNorth = true
		}
		if strings.Contains(cleanLine, "South Room") {
			foundSouth = true
		}
		if strings.Contains(cleanLine, "East Room") {
			foundEast = true
		}
	}

	if !hasLegendHeader {
		t.Error("Expected to find 'All Mapped Rooms' header in output")
	}
	if !foundCenter {
		t.Error("Expected to find Center in legend")
	}
	if !foundNorth {
		t.Error("Expected to find North Room in legend")
	}
	if !foundSouth {
		t.Error("Expected to find South Room in legend")
	}
	if !foundEast {
		t.Error("Expected to find East Room in legend")
	}

	// Verify mapLegend is populated
	if m.mapLegend == nil {
		t.Error("Expected mapLegend to be populated")
	}
	if len(m.mapLegend) != 4 {
		t.Errorf("Expected mapLegend to have 4 entries, got %d", len(m.mapLegend))
	}
}

// TestLegendCommandNoRooms verifies behavior when no rooms are mapped
func TestLegendCommandNoRooms(t *testing.T) {
	m := Model{
		output:    []string{"> "},
		connected: true,
		worldMap:  mapper.NewMap(),
	}

	// Execute the /legend command without any rooms
	savedPrompt := m.output[len(m.output)-1]
	m.output[len(m.output)-1] = savedPrompt + "\x1b[93m/legend\x1b[0m"
	m.handleClientCommand("/legend")

	// Should have "no rooms" message
	hasMessage := false
	for _, line := range m.output {
		if strings.Contains(line, "No rooms have been explored") {
			hasMessage = true
			break
		}
	}
	if !hasMessage {
		t.Error("Expected 'No rooms have been explored' message")
	}
}
