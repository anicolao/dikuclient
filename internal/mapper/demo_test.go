package mapper

import (
	"fmt"
	"testing"
)

// TestMapperDemo demonstrates the mapper functionality
func TestMapperDemo(t *testing.T) {
	// Create a new map
	m := NewMap()

	// Simulate visiting rooms
	t.Log("1. Visiting Temple Square...")
	room1 := NewRoom("Temple Square", "A large temple square with ancient stones.", []string{"north", "east"})
	m.AddOrUpdateRoom(room1)
	t.Logf("   Room ID: %s", room1.ID)

	t.Log("2. Moving north to Temple Hall...")
	m.SetLastDirection("north")
	room2 := NewRoom("Temple Hall", "A grand hall inside the temple. Torches line the walls.", []string{"south", "north", "west"})
	m.AddOrUpdateRoom(room2)
	t.Logf("   Temple Square north exit now points to: %s", m.Rooms[room1.ID].Exits["north"])
	t.Logf("   Temple Hall south exit now points to: %s", m.Rooms[room2.ID].Exits["south"])

	t.Log("3. Moving north to Inner Sanctum...")
	m.SetLastDirection("north")
	room3 := NewRoom("Inner Sanctum", "The innermost chamber of the temple. Very sacred.", []string{"south"})
	m.AddOrUpdateRoom(room3)

	t.Log("4. Testing pathfinding from Inner Sanctum back to Temple Square...")
	path := m.FindPath(room1.ID)
	if path == nil {
		t.Fatal("Expected to find path")
	}
	t.Logf("   Path: %v (steps: %d)", path, len(path))
	if len(path) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(path))
	}

	t.Log("5. Testing room search...")
	results := m.FindRooms("temple")
	t.Logf("   Found %d rooms with 'temple'", len(results))
	for _, r := range results {
		t.Logf("     - %s", r.Title)
	}

	results = m.FindRooms("temple square")
	t.Logf("   Found %d rooms with 'temple square'", len(results))
	if len(results) != 1 {
		t.Errorf("Expected 1 room, got %d", len(results))
	}

	t.Log("6. Simulating /point command...")
	if len(results) > 0 {
		targetID := results[0].ID
		path = m.FindPath(targetID)
		if path != nil && len(path) > 0 {
			t.Logf("   To reach '%s', go: %s", results[0].Title, path[0])
		}
	}

	t.Log("7. Simulating /wayfind command...")
	path = m.FindPath(room1.ID)
	if path != nil {
		t.Logf("   Path to '%s' (%d steps): %v", room1.Title, len(path), path)
	}

	t.Log("8. Map statistics...")
	t.Logf("   Total rooms: %d", len(m.Rooms))
	current := m.GetCurrentRoom()
	if current != nil {
		t.Logf("   Current room: %s", current.Title)
	}
}

// TestRoomParsingDemo demonstrates room parsing from MUD output
func TestRoomParsingDemo(t *testing.T) {
	t.Log("Testing room parsing from typical MUD output...")

	// Example 1: Simple room
	lines1 := []string{
		"Temple Square",
		"You are standing in a large temple square. The ancient stones",
		"speak of a glorious past.",
		"Exits: north, south, east",
	}

	info1 := ParseRoomInfo(lines1, false)
	if info1 == nil {
		t.Fatal("Failed to parse room 1")
	}
	t.Logf("Parsed room 1:")
	t.Logf("  Title: %s", info1.Title)
	t.Logf("  Description: %s", info1.Description)
	t.Logf("  Exits: %v", info1.Exits)

	// Example 2: Room with different exit format
	lines2 := []string{
		"Market Street",
		"A busy market street filled with merchants and shoppers.",
		"[ Exits: n s e w ]",
	}

	info2 := ParseRoomInfo(lines2, false)
	if info2 == nil {
		t.Fatal("Failed to parse room 2")
	}
	t.Logf("Parsed room 2:")
	t.Logf("  Title: %s", info2.Title)
	t.Logf("  Description: %s", info2.Description)
	t.Logf("  Exits: %v", info2.Exits)

	// Example 3: Room with ANSI codes
	lines3 := []string{
		"\x1b[1;33mGolden Hall\x1b[0m",
		"\x1b[0;37mA magnificent hall with golden walls.\x1b[0m",
		"\x1b[0;32mObvious exits: north and south\x1b[0m",
	}

	info3 := ParseRoomInfo(lines3, false)
	if info3 == nil {
		t.Fatal("Failed to parse room 3")
	}
	t.Logf("Parsed room 3:")
	t.Logf("  Title: %s", info3.Title)
	t.Logf("  Description: %s", info3.Description)
	t.Logf("  Exits: %v", info3.Exits)
}

// ExampleMap demonstrates usage of the mapper
func ExampleMap() {
	m := NewMap()

	// Add first room
	room1 := NewRoom("Starting Room", "You are in a small room.", []string{"north"})
	m.AddOrUpdateRoom(room1)

	// Move north and add second room
	m.SetLastDirection("north")
	room2 := NewRoom("Northern Room", "You are in a larger room.", []string{"south", "east"})
	m.AddOrUpdateRoom(room2)

	// Find path back to start
	path := m.FindPath(room1.ID)
	fmt.Printf("Path back: %v\n", path)

	// Search for rooms
	results := m.FindRooms("room")
	fmt.Printf("Found %d rooms\n", len(results))

	// Output:
	// Path back: [south]
	// Found 2 rooms
}
