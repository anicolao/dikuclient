package mapper

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMapPersistence(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	mapPath := filepath.Join(tmpDir, "test_map.json")

	// Create a map and add some rooms
	m := NewMap()
	m.mapPath = mapPath

	room1 := NewRoom("Room 1", "This is room 1.", []string{"north", "south"})
	room2 := NewRoom("Room 2", "This is room 2.", []string{"south"})

	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()

	// Save the map
	if err := m.Save(); err != nil {
		t.Fatalf("Failed to save map: %v", err)
	}

	// Load the map
	loaded, err := LoadFromPath(mapPath)
	if err != nil {
		t.Fatalf("Failed to load map: %v", err)
	}

	// Verify loaded map has the same rooms
	if len(loaded.Rooms) != 2 {
		t.Errorf("Loaded map has %d rooms, want 2", len(loaded.Rooms))
	}

	if loaded.CurrentRoomID != room2.ID {
		t.Errorf("CurrentRoomID = %q, want %q", loaded.CurrentRoomID, room2.ID)
	}
}

func TestMapFindRooms(t *testing.T) {
	m := NewMap()

	room1 := NewRoom("Temple Square", "A large temple square.", []string{"north"})
	room2 := NewRoom("Market Street", "A busy market street.", []string{"south"})
	room3 := NewRoom("Temple Entrance", "The entrance to the temple.", []string{"east"})

	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()

	// Search for "temple"
	results := m.FindRooms("temple")
	if len(results) != 2 {
		t.Errorf("Found %d rooms with 'temple', want 2", len(results))
	}

	// Search for "temple square"
	results = m.FindRooms("temple square")
	if len(results) != 1 {
		t.Errorf("Found %d rooms with 'temple square', want 1", len(results))
	}

	// Search for non-existent room
	results = m.FindRooms("castle")
	if len(results) != 0 {
		t.Errorf("Found %d rooms with 'castle', want 0", len(results))
	}
}

func TestMapPathfinding(t *testing.T) {
	m := NewMap()

	// Create a simple linear path: room1 -> room2 -> room3
	room1 := NewRoom("Room 1", "First room.", []string{"north"})
	room2 := NewRoom("Room 2", "Second room.", []string{"north", "south"})
	room3 := NewRoom("Room 3", "Third room.", []string{"south"})

	t.Logf("Room1 ID: %s", room1.ID)
	t.Logf("Room2 ID: %s", room2.ID)
	t.Logf("Room3 ID: %s", room3.ID)

	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	m.LinkRooms()
	t.Logf("After room1: Current=%s, Previous=%s", m.CurrentRoomID, m.PreviousRoomID)

	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	m.LinkRooms()
	t.Logf("After room2: Current=%s, Previous=%s", m.CurrentRoomID, m.PreviousRoomID)
	t.Logf("Room1 exits: %v", m.Rooms[room1.ID].Exits)
	t.Logf("Room2 exits: %v", m.Rooms[room2.ID].Exits)

	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	m.LinkRooms()
	t.Logf("After room3: Current=%s, Previous=%s", m.CurrentRoomID, m.PreviousRoomID)
	t.Logf("Room2 exits: %v", m.Rooms[room2.ID].Exits)
	t.Logf("Room3 exits: %v", m.Rooms[room3.ID].Exits)

	// Verify the links were created
	if m.Rooms[room1.ID].Exits["north"] != room2.ID {
		t.Errorf("room1 north exit = %q, want %q", m.Rooms[room1.ID].Exits["north"], room2.ID)
	}
	if m.Rooms[room2.ID].Exits["south"] != room1.ID {
		t.Errorf("room2 south exit = %q, want %q", m.Rooms[room2.ID].Exits["south"], room1.ID)
	}
	if m.Rooms[room2.ID].Exits["north"] != room3.ID {
		t.Errorf("room2 north exit = %q, want %q", m.Rooms[room2.ID].Exits["north"], room3.ID)
	}

	// Now we're at room3, try to find path back to room1
	path := m.FindPath(room1.ID)
	if path == nil {
		t.Fatal("FindPath returned nil")
	}

	t.Logf("Path: %v", path)

	// Should be: south, south
	if len(path) != 2 {
		t.Errorf("Path length = %d, want 2", len(path))
	}

	if len(path) >= 1 && path[0] != "south" {
		t.Errorf("First step = %q, want 'south'", path[0])
	}
}

func TestReverseDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"north", "south"},
		{"south", "north"},
		{"east", "west"},
		{"west", "east"},
		{"up", "down"},
		{"down", "up"},
		{"ne", "sw"},
		{"unknown", ""},
	}

	for _, test := range tests {
		result := getReverseDirection(test.input)
		if result != test.expected {
			t.Errorf("getReverseDirection(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestLoadNonExistentMap(t *testing.T) {
	tmpDir := t.TempDir()
	mapPath := filepath.Join(tmpDir, "nonexistent.json")

	m, err := LoadFromPath(mapPath)
	if err != nil {
		t.Fatalf("LoadFromPath should not error on non-existent file: %v", err)
	}

	if len(m.Rooms) != 0 {
		t.Errorf("New map should have 0 rooms, got %d", len(m.Rooms))
	}
}

func TestFindNearbyRooms(t *testing.T) {
	m := NewMap()

	// Create a linear path of rooms: Room1 -> Room2 -> Room3 -> Room4 -> Room5 -> Room6
	room1 := NewRoom("Room 1", "First room.", []string{"north"})
	room2 := NewRoom("Room 2", "Second room.", []string{"south", "north"})
	room3 := NewRoom("Room 3", "Third room.", []string{"south", "north"})
	room4 := NewRoom("Room 4", "Fourth room.", []string{"south", "north"})
	room5 := NewRoom("Room 5", "Fifth room.", []string{"south", "north"})
	room6 := NewRoom("Room 6", "Sixth room.", []string{"south"})

	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room4)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room5)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room6)
	m.LinkRooms()

	// Now current room is Room6, go back to Room3
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room5)
	m.LinkRooms()
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room4)
	m.LinkRooms()
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()

	// Test from Room3 - should find Room1, Room2, Room4, Room5 within 5 steps
	// Room6 is also within 3 steps
	nearby := m.FindNearbyRooms(5)

	if len(nearby) == 0 {
		t.Fatal("Expected to find nearby rooms, got none")
	}

	// Check that rooms are sorted by distance
	for i := 1; i < len(nearby); i++ {
		if nearby[i-1].Distance > nearby[i].Distance {
			t.Errorf("Rooms not sorted by distance: room %d has distance %d, room %d has distance %d",
				i-1, nearby[i-1].Distance, i, nearby[i].Distance)
		}
	}

	// Room2 and Room4 should be at distance 1 (adjacent to Room3)
	adjacentRooms := 0
	for _, nr := range nearby {
		if nr.Distance == 1 {
			adjacentRooms++
			if nr.Room.Title != "Room 2" && nr.Room.Title != "Room 4" {
				t.Errorf("Unexpected adjacent room: %s", nr.Room.Title)
			}
		}
	}
	if adjacentRooms != 2 {
		t.Errorf("Expected 2 adjacent rooms (distance 1), got %d", adjacentRooms)
	}

	// Current room (Room3) should not be in the list
	for _, nr := range nearby {
		if nr.Room.ID == room3.ID {
			t.Error("Current room should not be in nearby rooms list")
		}
	}
}

func TestFindNearbyRoomsNoCurrentRoom(t *testing.T) {
	m := NewMap()

	// Add a room but don't set it as current
	room1 := NewRoom("Room 1", "First room.", []string{"north"})
	m.Rooms[room1.ID] = room1

	nearby := m.FindNearbyRooms(5)

	if nearby != nil {
		t.Errorf("Expected nil when no current room, got %d rooms", len(nearby))
	}
}

func TestRoomIDWithDistance(t *testing.T) {
	m := NewMap()

	// Create first room with certain characteristics
	room1 := NewRoom("Identical Room", "This room looks the same.", []string{"north", "south"})
	m.AddOrUpdateRoom(room1)
	m.LinkRooms()

	// Move to another room
	intermediate := NewRoom("Intermediate", "A connecting room.", []string{"north", "south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(intermediate)
	m.LinkRooms()

	// Go back south to room1 - this should recognize it via the exit mapping
	revisitRoom1 := NewRoom("Identical Room", "This room looks the same.", []string{"north", "south"})
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(revisitRoom1)
	m.LinkRooms()

	// The revisited room should have the same ID as room1 (it IS room1)
	if revisitRoom1.ID != room1.ID {
		t.Errorf("Revisiting a room via known exit should yield the same ID. Room1: %s, Revisit: %s", room1.ID, revisitRoom1.ID)
	}

	// Verify distance is in the ID
	if !strings.Contains(room1.ID, "|0") {
		t.Errorf("Room1 should have distance 0 in ID, got: %s", room1.ID)
	}

	// Only 2 distinct rooms should exist (room1 and intermediate)
	if len(m.Rooms) != 2 {
		t.Errorf("Expected 2 rooms in map, got %d", len(m.Rooms))
	}

	t.Logf("Room1 ID: %s", room1.ID)
	t.Logf("Revisit ID: %s", revisitRoom1.ID)
	t.Logf("Intermediate ID: %s", intermediate.ID)
}

func TestRoomIDDistanceCalculation(t *testing.T) {
	m := NewMap()

	// Build a simple path to test distance calculation
	room0 := NewRoom("Start", "The starting room.", []string{"north"})
	m.AddOrUpdateRoom(room0)
	m.LinkRooms()

	room1 := NewRoom("Room 1", "One step away.", []string{"south", "north"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room1)
	m.LinkRooms()

	room2 := NewRoom("Room 2", "Two steps away.", []string{"south", "north"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()

	room3 := NewRoom("Room 3", "Three steps away.", []string{"south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()

	// Verify distances in IDs
	if !strings.HasSuffix(room0.ID, "|0") {
		t.Errorf("Room0 should end with |0, got: %s", room0.ID)
	}
	if !strings.HasSuffix(room1.ID, "|1") {
		t.Errorf("Room1 should end with |1, got: %s", room1.ID)
	}
	if !strings.HasSuffix(room2.ID, "|2") {
		t.Errorf("Room2 should end with |2, got: %s", room2.ID)
	}
	if !strings.HasSuffix(room3.ID, "|3") {
		t.Errorf("Room3 should end with |3, got: %s", room3.ID)
	}
}

func TestFindNearbyRoomsMaxDistance(t *testing.T) {
	m := NewMap()

	// Create a branching structure
	//       Room1
	//       /   \
	//   Room2   Room3
	//     |       |
	//   Room4   Room5
	//     |
	//   Room6
	
	room1 := NewRoom("Room 1", "First room.", []string{"north", "east"})
	room2 := NewRoom("Room 2", "Second room.", []string{"south", "north"})
	room3 := NewRoom("Room 3", "Third room.", []string{"west", "north"})
	room4 := NewRoom("Room 4", "Fourth room.", []string{"south", "north"})
	room5 := NewRoom("Room 5", "Fifth room.", []string{"south"})
	room6 := NewRoom("Room 6", "Sixth room.", []string{"south"})

	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	
	// Build north branch
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room4)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room6)
	m.LinkRooms()
	
	// Go back to room1
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room4)
	m.LinkRooms()
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room2)
	m.LinkRooms()
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room1)
	m.LinkRooms()
	
	// Build east branch
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room5)
	m.LinkRooms()
	
	// Go back to room1
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room3)
	m.LinkRooms()
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(room1)
	m.LinkRooms()

	// Test with maxDistance = 2 from Room1
	nearby := m.FindNearbyRooms(2)

	// Should find: Room2 (1 step), Room3 (1 step), Room4 (2 steps), Room5 (2 steps)
	// Should NOT find: Room6 (3 steps)
	if len(nearby) != 4 {
		t.Errorf("Expected 4 rooms within 2 steps, got %d", len(nearby))
	}

	for _, nr := range nearby {
		if nr.Room.Title == "Room 6" {
			t.Error("Room 6 should not be within 2 steps")
		}
		if nr.Distance > 2 {
			t.Errorf("Room %s has distance %d, expected <= 2", nr.Room.Title, nr.Distance)
		}
	}
}
