package mapper

import (
	"testing"
)

// TestLinkingSequence tests the linking behavior when rooms are detected immediately
func TestLinkingSequence(t *testing.T) {
	m := NewMap()

	// Scenario: start at room 0 and go s then s then west
	
	t.Log("\n=== STEP 1: Room 0 detected ===")
	room0 := NewRoom("Room 0", "Starting room.", []string{"south"})
	t.Logf("Before AddOrUpdateRoom: Current=%s, Previous=%s, LastDir=%s", 
		m.CurrentRoomID, m.PreviousRoomID, m.LastDirection)
	m.AddOrUpdateRoom(room0)
	t.Logf("After AddOrUpdateRoom: Current=%s, Previous=%s", 
		m.CurrentRoomID, m.PreviousRoomID)
	m.LinkRooms()
	t.Logf("Room 0 ID: %s", room0.ID)
	
	t.Log("\n=== STEP 2: User types 's', Room 1 detected ===")
	room1 := NewRoom("Room 1", "First room.", []string{"north", "south"})
	m.SetLastDirection("south")
	t.Logf("Before AddOrUpdateRoom: Current=%s, Previous=%s, LastDir=%s", 
		m.CurrentRoomID, m.PreviousRoomID, m.LastDirection)
	m.AddOrUpdateRoom(room1)
	t.Logf("After AddOrUpdateRoom: Current=%s, Previous=%s", 
		m.CurrentRoomID, m.PreviousRoomID)
	m.LinkRooms()
	t.Logf("Room 1 ID: %s", room1.ID)
	t.Logf("Room 0 south exit: %s", m.Rooms[room0.ID].Exits["south"])
	
	// Verify room0 south exit points to room1
	if m.Rooms[room0.ID].Exits["south"] != room1.ID {
		t.Errorf("Room 0 south exit should point to room 1. Got: %s, Want: %s", 
			m.Rooms[room0.ID].Exits["south"], room1.ID)
	}
	
	t.Log("\n=== STEP 3: User types 's', Room 2 detected ===")
	room2 := NewRoom("Room 2", "Second room.", []string{"north", "west"})
	m.SetLastDirection("south")
	t.Logf("Before AddOrUpdateRoom: Current=%s, Previous=%s, LastDir=%s", 
		m.CurrentRoomID, m.PreviousRoomID, m.LastDirection)
	m.AddOrUpdateRoom(room2)
	t.Logf("After AddOrUpdateRoom: Current=%s, Previous=%s", 
		m.CurrentRoomID, m.PreviousRoomID)
	m.LinkRooms()
	t.Logf("Room 2 ID: %s", room2.ID)
	t.Logf("Room 1 south exit: %s", m.Rooms[room1.ID].Exits["south"])
	
	// Verify room1 south exit points to room2
	if m.Rooms[room1.ID].Exits["south"] != room2.ID {
		t.Errorf("Room 1 south exit should point to room 2. Got: %s, Want: %s", 
			m.Rooms[room1.ID].Exits["south"], room2.ID)
	}
	
	t.Log("\n=== STEP 4: User types 'west', Room 3 detected ===")
	room3 := NewRoom("Room 3", "Third room.", []string{"east"})
	m.SetLastDirection("west")
	t.Logf("Before AddOrUpdateRoom: Current=%s, Previous=%s, LastDir=%s", 
		m.CurrentRoomID, m.PreviousRoomID, m.LastDirection)
	m.AddOrUpdateRoom(room3)
	t.Logf("After AddOrUpdateRoom: Current=%s, Previous=%s", 
		m.CurrentRoomID, m.PreviousRoomID)
	m.LinkRooms()
	t.Logf("Room 3 ID: %s", room3.ID)
	t.Logf("Room 2 west exit: %s", m.Rooms[room2.ID].Exits["west"])
	t.Logf("Room 2 south exit: %s", m.Rooms[room2.ID].Exits["south"])
	
	// Verify room2 west exit points to room3
	if m.Rooms[room2.ID].Exits["west"] != room3.ID {
		t.Errorf("Room 2 west exit should point to room 3. Got: %s, Want: %s", 
			m.Rooms[room2.ID].Exits["west"], room3.ID)
	}
	
	// Verify room2 south exit is NOT set to room3 (the bug the user is reporting)
	if m.Rooms[room2.ID].Exits["south"] == room3.ID {
		t.Errorf("Room 2 south exit should NOT point to room 3. Got: %s", 
			m.Rooms[room2.ID].Exits["south"])
	}
	
	// Log all room links for debugging
	t.Log("\n=== Final Room Links ===")
	for id, room := range m.Rooms {
		t.Logf("Room: %s", id)
		for dir, dest := range room.Exits {
			t.Logf("  %s -> %s", dir, dest)
		}
	}
}
