package mapper

import (
	"testing"
)

// TestWayfindingBug reproduces the bug from the user's log
// Starting from Temple of the Jeddak, going s, s, w, w
// Then wayfinding back should be e, e, n, n (not e, e, n)
func TestWayfindingBug(t *testing.T) {
	m := NewMap()

	// Room 0: The Temple of the Jeddak
	temple := NewRoom("The Temple of the Jeddak", 
		"You are in the southern end of the temple hall of Lesser Helium's great temple.", 
		[]string{"north", "south", "down"})
	m.AddOrUpdateRoom(temple)
	m.LinkRooms()
	t.Logf("Room 0: %s (distance %d)", temple.Title, m.extractDistanceFromRoomID(temple.ID))
	
	// User types "s"
	// Room 1: The Temple Plaza
	templePlaza := NewRoom("The Temple Plaza",
		"You are standing on the temple plaza of Lesser Helium.",
		[]string{"north", "east", "south", "west"})
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(templePlaza)
	m.LinkRooms()
	t.Logf("Room 1: %s (distance %d)", templePlaza.Title, m.extractDistanceFromRoomID(templePlaza.ID))
	
	// User types "s"
	// Room 2: Market Plaza
	marketPlaza := NewRoom("Market Plaza",
		"You are standing on the market plaza, the heart of Lesser Helium.",
		[]string{"north", "east", "south", "west"})
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(marketPlaza)
	m.LinkRooms()
	t.Logf("Room 2: %s (distance %d)", marketPlaza.Title, m.extractDistanceFromRoomID(marketPlaza.ID))
	
	// User types "w"
	// Room 3: Main Concourse (first one)
	mainConcourse1 := NewRoom("Main Concourse",
		"You are on the main concourse of Lesser Helium.",
		[]string{"north", "east", "south", "west"})
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(mainConcourse1)
	m.LinkRooms()
	t.Logf("Room 3: %s (distance %d)", mainConcourse1.Title, m.extractDistanceFromRoomID(mainConcourse1.ID))
	
	// User types "w"
	// Room 4: Main Concourse (western end)
	mainConcourse2 := NewRoom("Main Concourse",
		"You are at the western end of Lesser Helium's main concourse.",
		[]string{"north", "east", "south", "west"})
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(mainConcourse2)
	m.LinkRooms()
	t.Logf("Room 4: %s (distance %d)", mainConcourse2.Title, m.extractDistanceFromRoomID(mainConcourse2.ID))
	
	// Check the links
	t.Log("\n=== Room Links ===")
	t.Logf("Temple → south → %s", m.Rooms[temple.ID].Exits["south"])
	t.Logf("Temple Plaza → south → %s", m.Rooms[templePlaza.ID].Exits["south"])
	t.Logf("Market Plaza → west → %s", m.Rooms[marketPlaza.ID].Exits["west"])
	t.Logf("Main Concourse 1 → west → %s", m.Rooms[mainConcourse1.ID].Exits["west"])
	t.Logf("Main Concourse 2 → east → %s", m.Rooms[mainConcourse2.ID].Exits["east"])
	
	// Now test wayfinding from Main Concourse 2 back to Temple
	path := m.FindPath(temple.ID)
	
	t.Log("\n=== Wayfinding Path ===")
	for i, dir := range path {
		t.Logf("%d. %s", i+1, dir)
	}
	
	// The path should be: east, east, north, north (4 steps)
	expectedPath := []string{"east", "east", "north", "north"}
	if len(path) != len(expectedPath) {
		t.Errorf("Path length mismatch. Got %d, expected %d", len(path), len(expectedPath))
	}
	
	for i, dir := range expectedPath {
		if i >= len(path) {
			t.Errorf("Missing step %d: expected %s", i+1, dir)
			continue
		}
		if path[i] != dir {
			t.Errorf("Step %d mismatch: got %s, expected %s", i+1, path[i], dir)
		}
	}
	
	// Verify that Main Concourse 1 has east → Market Plaza
	if m.Rooms[mainConcourse1.ID].Exits["east"] != marketPlaza.ID {
		t.Errorf("Main Concourse 1 should have east → Market Plaza")
		t.Errorf("  Got: east → %s", m.Rooms[mainConcourse1.ID].Exits["east"])
		t.Errorf("  Expected: east → %s", marketPlaza.ID)
	}
}

// TestWayfindingBugFromActualMap reproduces the ACTUAL bug by simulating what happened
// The user's map shows that Market Plaza.east → Temple Plaza (should be north)
// and Temple Plaza.west → Market Plaza (should be south)
// This suggests rooms were linked with the wrong direction
func TestWayfindingBugFromActualMap(t *testing.T) {
	t.Log("=== Simulating what actually happened based on user's map file ===")
	
	// What if the rooms were detected ONE STEP LATE?
	// E.g., Market Plaza is detected when user types "w" instead of when user types "s"
	
	m := NewMap()
	
	// Room 0: Temple - detected immediately
	temple := NewRoom("The Temple of the Jeddak",
		"You are in the southern end of the temple hall of Lesser Helium's great temple.",
		[]string{"down", "north", "south"})
	m.AddOrUpdateRoom(temple)
	m.LinkRooms()
	t.Logf("Added temple: %s", temple.ID)
	
	// User types "s" but Temple Plaza is NOT detected yet
	// (or maybe pendingMovement gets cleared/overwritten)
	
	// User types "s" again, NOW Temple Plaza is detected
	// But LastDirection is "s" from the SECOND command
	templePlaza := NewRoom("The Temple Plaza",
		"You are standing on the temple plaza of Lesser Helium.",
		[]string{"east", "north", "south", "west"})
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(templePlaza)
	m.LinkRooms()
	t.Logf("Added Temple Plaza with direction 'south': %s", templePlaza.ID)
	
	// User types "w", Market Plaza is detected
	// But we're still at Temple Plaza in the mapper's view!
	marketPlaza := NewRoom("Market Plaza",
		"You are standing on the market plaza, the heart of Lesser Helium.",
		[]string{"east", "north", "south", "west"})
	m.SetLastDirection("west")  // This is WRONG! Should be "south"
	m.AddOrUpdateRoom(marketPlaza)
	m.LinkRooms()
	t.Logf("Added Market Plaza with direction 'west' (WRONG!): %s", marketPlaza.ID)
	
	// User types "w", Main Concourse is detected
	mainConcourse := NewRoom("Main Concourse",
		"You are on the main concourse of Lesser Helium.",
		[]string{"east", "north", "south", "west"})
	m.SetLastDirection("west")  // This is correct
	m.AddOrUpdateRoom(mainConcourse)
	m.LinkRooms()
	t.Logf("Added Main Concourse with direction 'west': %s", mainConcourse.ID)
	
	// Check the links - they should match the user's broken map
	t.Log("\n=== Checking if links match user's broken map ===")
	t.Logf("Temple Plaza → west → %s (expected Market Plaza)", m.Rooms[templePlaza.ID].Exits["west"])
	t.Logf("Market Plaza → east → %s (expected Temple Plaza)", m.Rooms[marketPlaza.ID].Exits["east"])
	
	if m.Rooms[templePlaza.ID].Exits["west"] == marketPlaza.ID {
		t.Log("✓ Bug reproduced! Temple Plaza.west → Market Plaza (should be south)")
	}
	if m.Rooms[marketPlaza.ID].Exits["east"] == templePlaza.ID {
		t.Log("✓ Bug reproduced! Market Plaza.east → Temple Plaza (should be north)")
	}
}
