package mapper

import (
	"strings"
	"testing"
)

// TestRenderMapBasic tests basic map rendering
func TestRenderMapBasic(t *testing.T) {
	m := NewMap()

	// Create a simple cross-shaped map
	//     N
	//   W C E
	//     S
	center := NewRoom("Center Room", "You are at the center.", []string{"north", "south", "east", "west"})
	north := NewRoom("North Room", "You are in the north room.", []string{"south"})
	south := NewRoom("South Room", "You are in the south room.", []string{"north"})
	east := NewRoom("East Room", "You are in the east room.", []string{"west"})
	west := NewRoom("West Room", "You are in the west room.", []string{"east"})

	// Link rooms
	center.UpdateExit("north", north.ID)
	center.UpdateExit("south", south.ID)
	center.UpdateExit("east", east.ID)
	center.UpdateExit("west", west.ID)

	north.UpdateExit("south", center.ID)
	south.UpdateExit("north", center.ID)
	east.UpdateExit("west", center.ID)
	west.UpdateExit("east", center.ID)

	// Add rooms to map
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	m.AddOrUpdateRoom(north)
	m.LinkRooms()
	m.AddOrUpdateRoom(south)
	m.LinkRooms()
	m.AddOrUpdateRoom(east)
	m.LinkRooms()
	m.AddOrUpdateRoom(west)
	m.LinkRooms()

	// Set center as current room
	m.CurrentRoomID = center.ID

	// Render the map
	rendered, title := m.RenderMap(30, 10)

	t.Logf("Rendered map:\n%s", rendered)
	t.Logf("Current room title: %s", title)

	// Verify current room title
	if title != "Center Room" {
		t.Errorf("Expected title 'Center Room', got '%s'", title)
	}

	// Verify the current room symbol appears in output
	if !strings.Contains(rendered, "▣") {
		t.Error("Expected current room symbol ▣ to appear in rendered map")
	}

	// Verify visited room symbols appear
	if !strings.Contains(rendered, "▢") {
		t.Error("Expected visited room symbols ▢ to appear in rendered map")
	}
}

// TestRenderMapEmpty tests rendering with no current room
func TestRenderMapEmpty(t *testing.T) {
	m := NewMap()

	rendered, title := m.RenderMap(30, 10)

	t.Logf("Rendered map (empty): %s", rendered)
	t.Logf("Title (empty): %s", title)

	if rendered != "(exploring...)" {
		t.Errorf("Expected '(exploring...)' for empty map, got '%s'", rendered)
	}

	if title != "" {
		t.Errorf("Expected empty title for empty map, got '%s'", title)
	}
}

// TestGetVerticalExits tests vertical exit detection
func TestGetVerticalExits(t *testing.T) {
	m := NewMap()

	// Test with no room
	hasUp, hasDown := m.GetVerticalExits()
	if hasUp || hasDown {
		t.Error("Expected no vertical exits when no current room")
	}

	// Test with up exit
	room := NewRoom("Test Room", "A test room.", []string{"north", "up"})
	m.AddOrUpdateRoom(room)
	m.LinkRooms()
	hasUp, hasDown = m.GetVerticalExits()
	if !hasUp || hasDown {
		t.Errorf("Expected only up exit, got up=%v down=%v", hasUp, hasDown)
	}

	// Test with down exit
	room2 := NewRoom("Test Room 2", "Another test room.", []string{"south", "down"})
	m.CurrentRoomID = room2.ID
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	hasUp, hasDown = m.GetVerticalExits()
	if hasUp || !hasDown {
		t.Errorf("Expected only down exit, got up=%v down=%v", hasUp, hasDown)
	}

	// Test with both exits
	room3 := NewRoom("Test Room 3", "A third test room.", []string{"up", "down"})
	m.CurrentRoomID = room3.ID
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	hasUp, hasDown = m.GetVerticalExits()
	if !hasUp || !hasDown {
		t.Errorf("Expected both vertical exits, got up=%v down=%v", hasUp, hasDown)
	}
}

// TestRenderVerticalExits tests vertical exit symbol rendering
func TestRenderVerticalExits(t *testing.T) {
	tests := []struct {
		hasUp    bool
		hasDown  bool
		expected string
	}{
		{false, false, ""},
		{true, false, "⇱"},
		{false, true, "⇲"},
		{true, true, "⇅"},
	}

	for _, tc := range tests {
		result := RenderVerticalExits(tc.hasUp, tc.hasDown)
		if result != tc.expected {
			t.Errorf("RenderVerticalExits(%v, %v) = %q, expected %q",
				tc.hasUp, tc.hasDown, result, tc.expected)
		}
	}
}

// TestFormatMapPanel tests the complete map panel formatting
func TestFormatMapPanel(t *testing.T) {
	m := NewMap()

	// Create a simple map with vertical exits
	center := NewRoom("Temple Square", "A sacred place.", []string{"north", "south", "up", "down"})
	north := NewRoom("North Temple", "The northern temple.", []string{"south"})

	// Link rooms
	center.UpdateExit("north", north.ID)
	north.UpdateExit("south", center.ID)

	// Add rooms
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	m.AddOrUpdateRoom(north)
	m.LinkRooms()

	// Set center as current
	m.CurrentRoomID = center.ID

	// Check vertical exits before formatting
	hasUp, hasDown := m.GetVerticalExits()
	t.Logf("Vertical exits: up=%v, down=%v", hasUp, hasDown)

	// Format the panel with enough height for content
	panel := m.FormatMapPanel(30, 15)

	t.Logf("Formatted panel:\n%s", panel)
	t.Logf("Panel length: %d", len(panel))

	// Current room has both up and down exits, so it should show ⇅ instead of ▣
	if !strings.Contains(panel, "⇅") {
		t.Error("Expected vertical exit symbol ⇅ in panel for current room with both exits")
	}
}

// TestRenderMapWithUnexplored tests rendering with unexplored exits
func TestRenderMapWithUnexplored(t *testing.T) {
	m := NewMap()

	// Create and add center room
	center := NewRoom("Town Square", "The bustling town square.", []string{"north", "south", "east", "west"})
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	
	// Move north and add north gate
	north := NewRoom("North Gate", "The northern gate.", []string{"south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(north)
	m.LinkRooms()
	
	// Move back south to center
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	
	// Manually add unexplored exits to center
	centerRoom := m.Rooms[center.ID]
	centerRoom.UpdateExit("south", "")
	centerRoom.UpdateExit("east", "")
	centerRoom.UpdateExit("west", "")
	
	// Render the map
	rendered, title := m.RenderMap(40, 15)
	
	t.Logf("Map with unexplored exits:\n%s", rendered)
	t.Logf("Current room title: %s", title)
	
	// Verify current room title
	if title != "Town Square" {
		t.Errorf("Expected title 'Town Square', got '%s'", title)
	}
	
	// Verify unexplored room symbol appears
	if !strings.Contains(rendered, "▦") {
		t.Error("Expected unexplored room symbol ▦ to appear in rendered map")
	}
	
	// Verify the current room symbol appears
	if !strings.Contains(rendered, "▣") {
		t.Error("Expected current room symbol ▣ to appear in rendered map")
	}
	
	// Verify visited room symbol appears (north gate)
	if !strings.Contains(rendered, "▢") {
		t.Error("Expected visited room symbol ▢ to appear in rendered map")
	}
}

// TestRenderMapWithEmptyStringExits tests rendering with empty string exits
func TestRenderMapWithEmptyStringExits(t *testing.T) {
	m := NewMap()

	// Create and add center room
	center := NewRoom("Center", "The center room.", []string{"north", "south", "east", "west"})
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	
	// Move north and add north room
	north := NewRoom("North", "The north room.", []string{"south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(north)
	m.LinkRooms()
	
	// Move back south to center
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(center)
	m.LinkRooms()
	
	// Manually add unexplored exits with empty strings (as seen in real map data)
	centerRoom := m.Rooms[center.ID]
	centerRoom.UpdateExit("south", "")
	centerRoom.UpdateExit("east", "")
	centerRoom.UpdateExit("west", "")
	
	// Render the map
	rendered, title := m.RenderMap(40, 15)
	
	t.Logf("Map with empty string exits:\n%s", rendered)
	t.Logf("Current room title: %s", title)
	
	// Verify unexplored rooms are shown
	unexploredCount := strings.Count(rendered, "▦")
	if unexploredCount < 3 {
		t.Errorf("Expected at least 3 unexplored room symbols, got %d", unexploredCount)
	}
	
	// Verify current room and visited room
	if !strings.Contains(rendered, "▣") {
		t.Error("Expected current room symbol ▣")
	}
	if !strings.Contains(rendered, "▢") {
		t.Error("Expected visited room symbol ▢")
	}
}

// TestRenderMapLinear tests rendering of a linear path
func TestRenderMapLinear(t *testing.T) {
	m := NewMap()

	// Create a linear path by simulating movement: Room1 -> Room2 -> Room3
	room1 := NewRoom("Room 1", "First room.", []string{"east"})
	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	
	room2 := NewRoom("Room 2", "Second room.", []string{"west", "east"})
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	
	room3 := NewRoom("Room 3", "Third room.", []string{"west"})
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	
	// Go back to room2
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()

	// Render
	rendered, title := m.RenderMap(40, 10)

	t.Logf("Linear map:\n%s", rendered)
	t.Logf("Title: %s", title)

	// Verify title is correct
	if title != "Room 2" {
		t.Errorf("Expected title 'Room 2', got '%s'", title)
	}

	// Count rooms in output
	currentCount := strings.Count(rendered, "▣")
	visitedCount := strings.Count(rendered, "▢")

	if currentCount != 1 {
		t.Errorf("Expected 1 current room symbol, found %d", currentCount)
	}

	if visitedCount < 2 {
		t.Errorf("Expected at least 2 visited room symbols, found %d", visitedCount)
	}
}
