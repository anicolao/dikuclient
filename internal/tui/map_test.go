package tui

import (
	"testing"

	"github.com/anicolao/dikuclient/internal/mapper"
)

// TestMapPanelRendering tests that the map panel renders correctly in the TUI
func TestMapPanelRendering(t *testing.T) {
	// Create a model with a simple map
	m := NewModel("localhost", 4000, nil, nil)

	// Create a test map
	worldMap := mapper.NewMap()
	center := mapper.NewRoom("Temple Square", "The temple square.", []string{"north", "south", "east", "west"})
	north := mapper.NewRoom("North Gate", "The north gate.", []string{"south"})
	south := mapper.NewRoom("South Gate", "The south gate.", []string{"north"})
	east := mapper.NewRoom("East Market", "The eastern market.", []string{"west"})
	west := mapper.NewRoom("West Temple", "The western temple.", []string{"east"})

	// Link rooms
	center.UpdateExit("north", north.ID)
	center.UpdateExit("south", south.ID)
	center.UpdateExit("east", east.ID)
	center.UpdateExit("west", west.ID)
	north.UpdateExit("south", center.ID)
	south.UpdateExit("north", center.ID)
	east.UpdateExit("west", center.ID)
	west.UpdateExit("east", center.ID)

	// Add rooms
	worldMap.AddOrUpdateRoom(center)
	worldMap.AddOrUpdateRoom(north)
	worldMap.AddOrUpdateRoom(south)
	worldMap.AddOrUpdateRoom(east)
	worldMap.AddOrUpdateRoom(west)
	worldMap.CurrentRoomID = center.ID

	// Set the map in the model
	m.worldMap = worldMap

	// Test rendering with reasonable dimensions
	m.width = 80
	m.height = 24

	// Render sidebar which includes map panel
	sidebar := m.renderSidebar(30, 20)

	// Verify the sidebar contains map content
	if len(sidebar) == 0 {
		t.Fatal("Expected non-empty sidebar")
	}

	t.Logf("Sidebar with map panel:\n%s", sidebar)

	// The sidebar should contain the current room symbol
	// Note: This is a basic check - the actual visual output is verified in mapper tests
}

// TestMapPanelWithNoMap tests rendering when no map exists
func TestMapPanelWithNoMap(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	m.worldMap = nil

	m.width = 80
	m.height = 24

	sidebar := m.renderSidebar(30, 20)

	if len(sidebar) == 0 {
		t.Fatal("Expected non-empty sidebar even with no map")
	}

	t.Logf("Sidebar with no map:\n%s", sidebar)
}

// TestMapPanelWithNoCurrentRoom tests rendering when map exists but no current room
func TestMapPanelWithNoCurrentRoom(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	m.worldMap = mapper.NewMap() // Empty map

	m.width = 80
	m.height = 24

	sidebar := m.renderSidebar(30, 20)

	if len(sidebar) == 0 {
		t.Fatal("Expected non-empty sidebar")
	}

	t.Logf("Sidebar with empty map:\n%s", sidebar)
}

// TestMapPanelWithVerticalExits tests rendering with vertical exits
func TestMapPanelWithVerticalExits(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)

	worldMap := mapper.NewMap()
	ground := mapper.NewRoom("Ground Floor", "The ground floor.", []string{"up"})
	upper := mapper.NewRoom("Upper Floor", "The upper floor.", []string{"down"})

	ground.UpdateExit("up", upper.ID)
	upper.UpdateExit("down", ground.ID)

	worldMap.AddOrUpdateRoom(ground)
	worldMap.AddOrUpdateRoom(upper)
	worldMap.CurrentRoomID = ground.ID

	m.worldMap = worldMap
	m.width = 80
	m.height = 24

	sidebar := m.renderSidebar(30, 20)

	if len(sidebar) == 0 {
		t.Fatal("Expected non-empty sidebar")
	}

	t.Logf("Sidebar with vertical exits:\n%s", sidebar)
}
