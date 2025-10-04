package mapper

import (
	"fmt"
	"testing"
)

// TestRenderMapDemo demonstrates the map rendering with various scenarios
func TestRenderMapDemo(t *testing.T) {
	t.Log("=== Map Rendering Demonstration ===\n")

	// Demo 1: Simple cross layout
	t.Log("Demo 1: Simple Cross Layout")
	t.Log("----------------------------")
	demoSimpleCross(t)

	// Demo 2: Temple complex with vertical exits
	t.Log("\nDemo 2: Temple Complex with Vertical Exits")
	t.Log("-------------------------------------------")
	demoTempleComplex(t)

	// Demo 3: Linear path
	t.Log("\nDemo 3: Linear Path")
	t.Log("-------------------")
	demoLinearPath(t)

	// Demo 4: Large area
	t.Log("\nDemo 4: Large Area Map")
	t.Log("----------------------")
	demoLargeArea(t)
}

func demoSimpleCross(t *testing.T) {
	m := NewMap()

	// Create cross: N, S, E, W around center
	center := NewRoom("Town Square", "The bustling town square.", []string{"north", "south", "east", "west"})
	north := NewRoom("North Gate", "The northern gate of the city.", []string{"south"})
	south := NewRoom("South Gate", "The southern gate of the city.", []string{"north"})
	east := NewRoom("East Market", "A busy eastern market.", []string{"west"})
	west := NewRoom("West Temple", "An ancient temple to the west.", []string{"east"})

	// Link all rooms
	center.UpdateExit("north", north.ID)
	center.UpdateExit("south", south.ID)
	center.UpdateExit("east", east.ID)
	center.UpdateExit("west", west.ID)
	north.UpdateExit("south", center.ID)
	south.UpdateExit("north", center.ID)
	east.UpdateExit("west", center.ID)
	west.UpdateExit("east", center.ID)

	// Add to map
	m.AddOrUpdateRoom(center)
	m.AddOrUpdateRoom(north)
	m.AddOrUpdateRoom(south)
	m.AddOrUpdateRoom(east)
	m.AddOrUpdateRoom(west)
	m.CurrentRoomID = center.ID

	// Render
	content := m.FormatMapPanel(30, 10)
	t.Logf("Current room: %s\n", center.Title)
	t.Logf("Map:\n%s\n", content)
}

func demoTempleComplex(t *testing.T) {
	m := NewMap()

	// Create temple with vertical exits
	entrance := NewRoom("Temple Entrance", "The grand entrance to the temple.", []string{"north"})
	hall := NewRoom("Temple Hall", "A vast ceremonial hall.", []string{"south", "north", "up"})
	sanctum := NewRoom("Inner Sanctum", "The innermost chamber.", []string{"south", "down"})
	tower := NewRoom("Bell Tower", "A tall tower overlooking the temple.", []string{"down"})
	crypt := NewRoom("Temple Crypt", "Dark catacombs beneath the temple.", []string{"up"})

	// Link rooms
	entrance.UpdateExit("north", hall.ID)
	hall.UpdateExit("south", entrance.ID)
	hall.UpdateExit("north", sanctum.ID)
	hall.UpdateExit("up", tower.ID)
	sanctum.UpdateExit("south", hall.ID)
	sanctum.UpdateExit("down", crypt.ID)
	tower.UpdateExit("down", hall.ID)
	crypt.UpdateExit("up", sanctum.ID)

	// Add to map
	m.AddOrUpdateRoom(entrance)
	m.AddOrUpdateRoom(hall)
	m.AddOrUpdateRoom(sanctum)
	m.AddOrUpdateRoom(tower)
	m.AddOrUpdateRoom(crypt)
	m.CurrentRoomID = hall.ID

	// Render
	content := m.FormatMapPanel(30, 12)
	t.Logf("Current room: %s\n", hall.Title)
	t.Logf("Map (with vertical exits ⇱ showing up):\n%s\n", content)
}

func demoLinearPath(t *testing.T) {
	m := NewMap()

	// Create a long corridor
	rooms := make([]*Room, 7)
	for i := 0; i < 7; i++ {
		title := fmt.Sprintf("Corridor %d", i+1)
		desc := fmt.Sprintf("Part %d of a long corridor.", i+1)
		exits := []string{}
		if i > 0 {
			exits = append(exits, "west")
		}
		if i < 6 {
			exits = append(exits, "east")
		}
		rooms[i] = NewRoom(title, desc, exits)
	}

	// Link rooms
	for i := 0; i < 6; i++ {
		rooms[i].UpdateExit("east", rooms[i+1].ID)
		rooms[i+1].UpdateExit("west", rooms[i].ID)
	}

	// Add to map
	for _, room := range rooms {
		m.AddOrUpdateRoom(room)
	}
	m.CurrentRoomID = rooms[3].ID // Middle room

	// Render
	content := m.FormatMapPanel(40, 8)
	t.Logf("Current room: %s\n", rooms[3].Title)
	t.Logf("Map (showing linear path):\n%s\n", content)
}

func demoLargeArea(t *testing.T) {
	m := NewMap()

	// Create a grid of rooms (3x3)
	grid := make([][]*Room, 3)
	for y := 0; y < 3; y++ {
		grid[y] = make([]*Room, 3)
		for x := 0; x < 3; x++ {
			title := fmt.Sprintf("Room (%d,%d)", x, y)
			desc := fmt.Sprintf("A room at position %d,%d.", x, y)
			exits := []string{}

			if x > 0 {
				exits = append(exits, "west")
			}
			if x < 2 {
				exits = append(exits, "east")
			}
			if y > 0 {
				exits = append(exits, "north")
			}
			if y < 2 {
				exits = append(exits, "south")
			}

			grid[y][x] = NewRoom(title, desc, exits)
		}
	}

	// Link rooms horizontally and vertically
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if x > 0 {
				grid[y][x].UpdateExit("west", grid[y][x-1].ID)
			}
			if x < 2 {
				grid[y][x].UpdateExit("east", grid[y][x+1].ID)
			}
			if y > 0 {
				grid[y][x].UpdateExit("north", grid[y-1][x].ID)
			}
			if y < 2 {
				grid[y][x].UpdateExit("south", grid[y+1][x].ID)
			}
		}
	}

	// Add all rooms to map
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			m.AddOrUpdateRoom(grid[y][x])
		}
	}
	m.CurrentRoomID = grid[1][1].ID // Center room

	// Render
	content := m.FormatMapPanel(30, 12)
	t.Logf("Current room: %s\n", grid[1][1].Title)
	t.Logf("Map (3x3 grid):\n%s\n", content)
}

// TestRenderMapWithVerticalExitsDemo shows how vertical exits are displayed
func TestRenderMapWithVerticalExitsDemo(t *testing.T) {
	t.Log("=== Vertical Exits Display Demo ===\n")

	// Demo with up exit only
	t.Log("Demo: Room with UP exit only")
	m1 := NewMap()
	room1 := NewRoom("Ground Floor", "The ground floor.", []string{"up"})
	m1.AddOrUpdateRoom(room1)
	content1 := m1.FormatMapPanel(30, 8)
	t.Logf("Map:\n%s\n", content1)

	// Demo with down exit only
	t.Log("Demo: Room with DOWN exit only")
	m2 := NewMap()
	room2 := NewRoom("Upper Floor", "The upper floor.", []string{"down"})
	m2.AddOrUpdateRoom(room2)
	content2 := m2.FormatMapPanel(30, 8)
	t.Logf("Map:\n%s\n", content2)

	// Demo with both exits
	t.Log("Demo: Room with BOTH UP and DOWN exits")
	m3 := NewMap()
	room3 := NewRoom("Middle Floor", "The middle floor.", []string{"up", "down"})
	m3.AddOrUpdateRoom(room3)
	content3 := m3.FormatMapPanel(30, 8)
	t.Logf("Map:\n%s\n", content3)
}

// TestAdjacentNotConnectedDemo demonstrates the problem with adjacent but unconnected rooms
func TestAdjacentNotConnectedDemo(t *testing.T) {
	m := NewMap()

	// Create a scenario where rooms are adjacent but not connected
	// This is a "T" shape layout:
	//     N
	//   W C
	//     S
	// The issue: West room (W) appears next to North room (N) but they're not connected
	
	center := NewRoom("Center", "Center room", []string{"north", "south", "west"})
	north := NewRoom("North", "North room", []string{"south"})
	south := NewRoom("South", "South room", []string{"north"})
	west := NewRoom("West", "West room", []string{"east"})

	// Link rooms - NOTE: North and West are NOT connected
	center.UpdateExit("north", north.ID)
	center.UpdateExit("south", south.ID)
	center.UpdateExit("west", west.ID)
	north.UpdateExit("south", center.ID)
	south.UpdateExit("north", center.ID)
	west.UpdateExit("east", center.ID)

	// Add rooms
	m.AddOrUpdateRoom(center)
	m.AddOrUpdateRoom(north)
	m.AddOrUpdateRoom(south)
	m.AddOrUpdateRoom(west)
	m.CurrentRoomID = center.ID

	// Render
	content := m.FormatMapPanel(30, 10)
	t.Logf("Current layout (W and N are adjacent but not connected):\n%s\n", content)
}
