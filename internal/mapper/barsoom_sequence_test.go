package mapper

import (
	"strings"
	"testing"
)

// TestBarsoomSequenceMapping tests the mapping behavior with a real sequence of Barsoom rooms
// This test is based on actual MUD output to verify correct room detection and linking
func TestBarsoomSequenceMapping(t *testing.T) {
	m := NewMap()

	// Simulate the sequence of movements through Barsoom
	// Starting in "The Temple of the Jeddak"
	
	// Initial room - Temple of the Jeddak (detected when we first see it)
	templeOutput := []string{
		"26H 100F 78V 91C T:75 Exits:NSD> l",
		"--<",
		"The Temple of the Jeddak",
		"   You are in the southern end of the temple hall of Lesser Helium's great temple.",
		"The temple has been constructed from giant marble blocks, scarlet and yellow in",
		"color, and most of the walls are covered by ancient murals depicting the history",
		"of the red Martian race and the glory of Helium.",
		"   Large steps lead down through the grand temple gate, descending from the elevated",
		"platform upon which the temple is built to the temple plaza below.",
		">-- Exits:NSD",
		"A tarnished copper lamp lies here, looking bright.",
		"",
		"26H 100F 78V 91C T:72 Exits:NSD>",
	}

	templeRoom := ParseBarsoomRoomOnly(templeOutput, false)
	if templeRoom == nil || templeRoom.Title == "" {
		t.Fatal("Failed to parse Temple of the Jeddak room")
	}

	t.Logf("Temple room parsed: %s", templeRoom.Title)
	t.Logf("Temple exits: %v", templeRoom.Exits)

	// Add the temple room as the starting room (room 0)
	temple := NewRoom(templeRoom.Title, templeRoom.Description, templeRoom.Exits)
	m.AddOrUpdateRoom(temple)
	m.LinkRooms()

	t.Logf("Temple room ID: %s", temple.ID)
	if !strings.HasSuffix(temple.ID, "|0") {
		t.Errorf("Temple should be room 0, got ID: %s", temple.ID)
	}

	// Now user types 's' to go south
	// The output shows Temple Plaza
	templePlazaOutput := []string{
		"26H 100F 78V 91C T:72 Exits:NSD> s",
		"--<",
		"The Temple Plaza",
		"   You are standing on the temple plaza of Lesser Helium.  Huge marble steps lead",
		"up to the temple gate.  The entrance to the Noble's Guild is to the west, and the",
		"Traveler's Rest Inn is to the east.  Just south of here you see the market plaza,",
		"the heart of Lesser Helium.",
		">-- Exits:NESW",
		"",
		"26H 100F 77V 91C T:69 Exits:NESW>",
	}

	templePlazaRoom := ParseBarsoomRoomOnly(templePlazaOutput, false)
	if templePlazaRoom == nil || templePlazaRoom.Title == "" {
		t.Fatal("Failed to parse Temple Plaza room")
	}

	t.Logf("Temple Plaza room parsed: %s", templePlazaRoom.Title)
	t.Logf("Temple Plaza exits: %v", templePlazaRoom.Exits)

	// User moved south from temple
	m.SetLastDirection("south")
	plaza := NewRoom(templePlazaRoom.Title, templePlazaRoom.Description, templePlazaRoom.Exits)
	m.AddOrUpdateRoom(plaza)
	m.LinkRooms()

	t.Logf("Temple Plaza room ID: %s", plaza.ID)
	if !strings.HasSuffix(plaza.ID, "|1") {
		t.Errorf("Temple Plaza should be at distance 1, got ID: %s", plaza.ID)
	}

	// Verify the link from temple to plaza exists
	templeInMap := m.Rooms[temple.ID]
	if templeInMap.Exits["south"] != plaza.ID {
		t.Errorf("Temple south exit should point to plaza. Got: %s, Want: %s", templeInMap.Exits["south"], plaza.ID)
	}

	// Verify reverse link (tentative) from plaza to temple exists
	plazaInMap := m.Rooms[plaza.ID]
	if plazaInMap.Exits["north"] != temple.ID {
		t.Errorf("Plaza north exit should point to temple. Got: %s, Want: %s", plazaInMap.Exits["north"], temple.ID)
	}

	// Now user types 's' to go south to Market Plaza
	marketPlazaOutput := []string{
		"26H 100F 77V 91C T:69 Exits:NESW> s",
		"--<",
		"Market Plaza",
		"   You are standing on the market plaza, the heart of Lesser Helium.",
		"A magnificent statue commemorating the unification of the Twin Cities stands in the",
		"middle of the plaza.  Roads lead in every direction, north to the temple plaza, south",
		"to the common plaza, east and westbound is the main concourse.",
		">-- Exits:NESW",
		"A water fountain provides fresh water here.",
		"A Red Martian janitor cleans the area quietly.",
		"",
		"26H 100F 75V 91C T:67 Exits:NESW>",
	}

	marketPlazaRoom := ParseBarsoomRoomOnly(marketPlazaOutput, false)
	if marketPlazaRoom == nil || marketPlazaRoom.Title == "" {
		t.Fatal("Failed to parse Market Plaza room")
	}

	t.Logf("Market Plaza room parsed: %s", marketPlazaRoom.Title)

	m.SetLastDirection("south")
	market := NewRoom(marketPlazaRoom.Title, marketPlazaRoom.Description, marketPlazaRoom.Exits)
	m.AddOrUpdateRoom(market)
	m.LinkRooms()

	t.Logf("Market Plaza room ID: %s", market.ID)
	if !strings.HasSuffix(market.ID, "|2") {
		t.Errorf("Market Plaza should be at distance 2, got ID: %s", market.ID)
	}

	// Now user goes west to Main Concourse
	concourse1Output := []string{
		"26H 100F 75V 91C T:66 Exits:NESW> w",
		"--<",
		"Main Concourse",
		"   You are on the main concourse of Lesser Helium.  South of here is the entrance",
		"to the Armory, and the provisions merchant is to the north.  East of here is the",
		"market plaza.",
		">-- Exits:NESW",
		"A scavenging calot prowls here, searching for scraps.",
		"",
		"26H 100F 73V 91C T:65 Exits:NESW>",
	}

	concourse1Room := ParseBarsoomRoomOnly(concourse1Output, false)
	if concourse1Room == nil || concourse1Room.Title == "" {
		t.Fatal("Failed to parse Main Concourse room (first)")
	}

	t.Logf("Main Concourse (1) room parsed: %s", concourse1Room.Title)

	m.SetLastDirection("west")
	concourse1 := NewRoom(concourse1Room.Title, concourse1Room.Description, concourse1Room.Exits)
	m.AddOrUpdateRoom(concourse1)
	m.LinkRooms()

	t.Logf("Main Concourse (1) room ID: %s", concourse1.ID)

	// Now user goes west again to another Main Concourse (identical title!)
	concourse2Output := []string{
		"26H 100F 73V 91C T:65 Exits:NESW> w",
		"--<",
		"Main Concourse",
		"   You are at the western end of Lesser Helium's main concourse.  South of here is",
		"the entrance to the Scientific Academy.  The concourse continues east towards the",
		"market plaza.  The scientific supply shop is to the north and to the west is the",
		"city's west gate.",
		">-- Exits:NESW",
		"",
		"26H 100F 71V 91C T:63 Exits:NESW>",
	}

	concourse2Room := ParseBarsoomRoomOnly(concourse2Output, false)
	if concourse2Room == nil || concourse2Room.Title == "" {
		t.Fatal("Failed to parse Main Concourse room (second)")
	}

	t.Logf("Main Concourse (2) room parsed: %s", concourse2Room.Title)

	m.SetLastDirection("west")
	concourse2 := NewRoom(concourse2Room.Title, concourse2Room.Description, concourse2Room.Exits)
	m.AddOrUpdateRoom(concourse2)
	m.LinkRooms()

	t.Logf("Main Concourse (2) room ID: %s", concourse2.ID)

	// Verify the two Main Concourse rooms have different IDs (due to distance)
	if concourse1.ID == concourse2.ID {
		t.Error("Two Main Concourse rooms should have different IDs due to different descriptions and distances")
	}

	// Log the final map state
	t.Logf("\n=== Final Map State ===")
	t.Logf("Total rooms: %d", len(m.Rooms))
	for id, room := range m.Rooms {
		t.Logf("  %s: %s", room.Title, id)
		for dir, dest := range room.Exits {
			if dest != "" {
				destRoom := m.Rooms[dest]
				if destRoom != nil {
					t.Logf("    %s -> %s", dir, destRoom.Title)
				} else {
					t.Logf("    %s -> %s (unknown)", dir, dest)
				}
			}
		}
	}

	// Verify we can find a path back from the last room to the first
	path := m.FindPath(temple.ID)
	t.Logf("\nPath back to temple: %v", path)
	if path == nil {
		t.Error("Should be able to find a path back to the temple")
	}

	// Expected path: east (to concourse1), east (to market), north (to plaza), north (to temple)
	expectedPath := []string{"east", "east", "north", "north"}
	if len(path) != len(expectedPath) {
		t.Errorf("Path length mismatch. Got %d steps, expected %d. Path: %v", len(path), len(expectedPath), path)
	}
	
	// Verify the path is correct
	for i, dir := range expectedPath {
		if i >= len(path) {
			break
		}
		if path[i] != dir {
			t.Errorf("Step %d: expected %s, got %s", i+1, dir, path[i])
		}
	}
}
