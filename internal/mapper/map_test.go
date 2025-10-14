package mapper

import (
	"path/filepath"
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
	m.AddOrUpdateRoom(room2)

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
	m.AddOrUpdateRoom(room2)
	m.AddOrUpdateRoom(room3)

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
	t.Logf("After room1: Current=%s, Previous=%s", m.CurrentRoomID, m.PreviousRoomID)

	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	t.Logf("After room2: Current=%s, Previous=%s", m.CurrentRoomID, m.PreviousRoomID)
	t.Logf("Room1 exits: %v", m.Rooms[room1.ID].Exits)
	t.Logf("Room2 exits: %v", m.Rooms[room2.ID].Exits)

	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room3)
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
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room3)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room4)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room5)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room6)

	// Now current room is Room6, go back to Room3
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room5)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room4)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room3)

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
	
	// Build north branch
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room2)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room4)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room6)
	
	// Go back to room1
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room4)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room2)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room1)
	
	// Build east branch
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(room3)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(room5)
	
	// Go back to room1
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(room3)
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(room1)

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

func TestGetMapPathForServer(t *testing.T) {
	// Test that GetMapPathForServer generates the correct filename
	path, err := GetMapPathForServer("mud.arctic.org", 2700)
	if err != nil {
		t.Fatalf("GetMapPathForServer failed: %v", err)
	}

	// Path should contain map.mud.arctic.org.2700.json
	expectedName := "map.mud.arctic.org.2700.json"
	if filepath.Base(path) != expectedName {
		t.Errorf("Path basename %q does not match %q", filepath.Base(path), expectedName)
	}

	// Test with different server
	path2, err := GetMapPathForServer("localhost", 4000)
	if err != nil {
		t.Fatalf("GetMapPathForServer failed: %v", err)
	}

	expectedName2 := "map.localhost.4000.json"
	if filepath.Base(path2) != expectedName2 {
		t.Errorf("Path basename %q does not match %q", filepath.Base(path2), expectedName2)
	}

	// Paths should be different
	if path == path2 {
		t.Error("Different servers should have different map paths")
	}
}

func TestBarsoomModePersistence(t *testing.T) {
	// Test that BarsoomMode is persisted in the map
	tmpDir := t.TempDir()
	mapPath := filepath.Join(tmpDir, "test_barsoom_map.json")

	// Create a map with BarsoomMode enabled
	m := NewMap()
	m.mapPath = mapPath
	m.BarsoomMode = true

	room1 := NewBarsoomRoom("Temple Square", "You are standing in a large temple square. The ancient stones speak of a glorious past.", []string{"north", "south"})
	m.AddOrUpdateRoom(room1)

	// Save the map
	if err := m.Save(); err != nil {
		t.Fatalf("Failed to save map: %v", err)
	}

	// Load the map
	loaded, err := LoadFromPath(mapPath)
	if err != nil {
		t.Fatalf("Failed to load map: %v", err)
	}

	// Verify BarsoomMode was persisted
	if !loaded.BarsoomMode {
		t.Error("Expected BarsoomMode to be true after loading")
	}

	// Verify the room was loaded
	if len(loaded.Rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(loaded.Rooms))
	}
}
