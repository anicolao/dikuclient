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
