package mapper

import (
	"testing"
)

// TestDistanceFeatureDemo demonstrates the distance-based room ID disambiguation
func TestDistanceFeatureDemo(t *testing.T) {
	m := NewMap()

	// Simulate exploring a Barsoom-style world with identical-looking rooms
	t.Log("=== Exploring Barsoom ===")
	t.Log("")

	// Start at a unique room (becomes room 0)
	start := NewRoom("Temple Entrance", "A grand temple entrance.", []string{"north"})
	m.AddOrUpdateRoom(start)
	m.LinkRooms()
	t.Logf("1. Entered: %s", start.Title)
	t.Logf("   Room ID: %s", start.ID)
	t.Log("")

	// Move north to a corridor
	corridor1 := NewRoom("Stone Corridor", "A long stone corridor.", []string{"north", "south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(corridor1)
	m.LinkRooms()
	t.Logf("2. Moved north to: %s", corridor1.Title)
	t.Logf("   Room ID: %s", corridor1.ID)
	t.Log("")

	// Move north again - another identical corridor!
	corridor2 := NewRoom("Stone Corridor", "A long stone corridor.", []string{"north", "south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(corridor2)
	m.LinkRooms()
	t.Logf("3. Moved north to another: %s", corridor2.Title)
	t.Logf("   Room ID: %s", corridor2.ID)
	t.Log("")

	// Check if the two corridors have different IDs
	if corridor1.ID != corridor2.ID {
		t.Log("✓ SUCCESS: Identical rooms at different distances have DIFFERENT IDs!")
		t.Logf("  Corridor 1 is at distance 1: %s", corridor1.ID)
		t.Logf("  Corridor 2 is at distance 2: %s", corridor2.ID)
	} else {
		t.Error("✗ FAIL: Identical rooms should have different IDs")
	}
	t.Log("")

	// Move north to yet another identical corridor
	corridor3 := NewRoom("Stone Corridor", "A long stone corridor.", []string{"north", "south"})
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(corridor3)
	m.LinkRooms()
	t.Logf("4. Moved north to yet another: %s", corridor3.Title)
	t.Logf("   Room ID: %s", corridor3.ID)
	t.Log("")

	// Now go back south - should recognize corridor2
	m.SetLastDirection("south")
	corridor2_revisit := NewRoom("Stone Corridor", "A long stone corridor.", []string{"north", "south"})
	m.AddOrUpdateRoom(corridor2_revisit)
	m.LinkRooms()
	t.Logf("5. Moved south back to: %s", corridor2_revisit.Title)
	t.Logf("   Room ID: %s", corridor2_revisit.ID)
	t.Log("")

	if corridor2_revisit.ID == corridor2.ID {
		t.Log("✓ SUCCESS: Revisiting via known exit preserves room identity!")
		t.Logf("  Same room ID: %s", corridor2.ID)
	} else {
		t.Error("✗ FAIL: Revisiting a room should yield the same ID")
	}
	t.Log("")

	// Verify we have the expected number of rooms
	expectedRooms := 4 // start + 3 corridors
	if len(m.GetAllRooms()) != expectedRooms {
		t.Errorf("Expected %d unique rooms, got %d", expectedRooms, len(m.GetAllRooms()))
	}

	// Summary
	t.Log("=== Summary ===")
	t.Logf("Total unique rooms in map: %d", len(m.GetAllRooms()))
	t.Log("")
	t.Log("All rooms:")
	for id, room := range m.GetAllRooms() {
		t.Logf("  - %s (visits: %d)", room.Title, room.VisitCount)
		t.Logf("    ID: %s", id)
	}
}
