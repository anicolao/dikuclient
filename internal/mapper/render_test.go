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
	m.AddOrUpdateRoom(north)
	m.AddOrUpdateRoom(south)
	m.AddOrUpdateRoom(east)
	m.AddOrUpdateRoom(west)

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
	hasUp, hasDown = m.GetVerticalExits()
	if !hasUp || hasDown {
		t.Errorf("Expected only up exit, got up=%v down=%v", hasUp, hasDown)
	}

	// Test with down exit
	room2 := NewRoom("Test Room 2", "Another test room.", []string{"south", "down"})
	m.CurrentRoomID = room2.ID
	m.AddOrUpdateRoom(room2)
	hasUp, hasDown = m.GetVerticalExits()
	if hasUp || !hasDown {
		t.Errorf("Expected only down exit, got up=%v down=%v", hasUp, hasDown)
	}

	// Test with both exits
	room3 := NewRoom("Test Room 3", "A third test room.", []string{"up", "down"})
	m.CurrentRoomID = room3.ID
	m.AddOrUpdateRoom(room3)
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
	m.AddOrUpdateRoom(north)

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

	// Create a center room with some explored and some unexplored exits
	center := NewRoom("Town Square", "The bustling town square.", []string{"north", "south", "east", "west"})
	north := NewRoom("North Gate", "The northern gate.", []string{"south"})
	
	// Link center to north (explored)
	center.UpdateExit("north", north.ID)
	north.UpdateExit("south", center.ID)
	
	// Create unexplored exits (IDs but no actual rooms in the map yet)
	center.UpdateExit("south", "unexplored-south-id")
	center.UpdateExit("east", "unexplored-east-id")
	center.UpdateExit("west", "unexplored-west-id")
	
	// Add only center and north rooms to the map
	m.AddOrUpdateRoom(center)
	m.AddOrUpdateRoom(north)
	m.CurrentRoomID = center.ID
	
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

// TestRenderMapLinear tests rendering of a linear path
func TestRenderMapLinear(t *testing.T) {
	m := NewMap()

	// Create a linear path: Room1 -> Room2 -> Room3
	room1 := NewRoom("Room 1", "First room.", []string{"east"})
	room2 := NewRoom("Room 2", "Second room.", []string{"west", "east"})
	room3 := NewRoom("Room 3", "Third room.", []string{"west"})

	// Link rooms
	room1.UpdateExit("east", room2.ID)
	room2.UpdateExit("west", room1.ID)
	room2.UpdateExit("east", room3.ID)
	room3.UpdateExit("west", room2.ID)

	// Add to map
	m.AddOrUpdateRoom(room1)
	m.AddOrUpdateRoom(room2)
	m.AddOrUpdateRoom(room3)

	// Set room2 as current
	m.CurrentRoomID = room2.ID

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
