package mapper

import (
	"testing"
)

// TestVisualMapRender tests the map rendering with a realistic map layout
// This test is meant to be run with -v flag to see the visual output
func TestVisualMapRender(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual test in short mode")
	}

	// Create a test map (don't rely on external files)
	m := createTestMap()

	t.Log("=== Test Map Rendering ===")
	t.Logf("Total rooms: %d", len(m.Rooms))

	currentRoom := m.GetCurrentRoom()
	if currentRoom == nil {
		t.Fatal("No current room set")
	}

	t.Logf("Current room: %s\n", currentRoom.Title)

	// Render the map at different sizes
	sizes := []struct {
		width  int
		height int
		label  string
	}{
		{30, 10, "Small (30x10)"},
		{40, 15, "Medium (40x15)"},
		{50, 20, "Large (50x20)"},
	}

	for _, size := range sizes {
		t.Logf("\n=== %s ===", size.label)
		content := m.FormatMapPanel(size.width, size.height)
		t.Logf("\n%s", content)
	}

	// Also test with north temple as current room (has up exit)
	m.CurrentRoomID = "north temple|the northern temple hall.|south,up"
	currentRoom = m.GetCurrentRoom()
	if currentRoom != nil {
		t.Logf("\n=== With '%s' as current room (has UP exit) ===", currentRoom.Title)
		content := m.FormatMapPanel(40, 15)
		t.Logf("\n%s", content)
	}
}

func createTestMap() *Map {
	m := NewMap()

	// Create a temple complex
	square := NewRoom("Temple Square", "You are standing in a large temple square. The ancient stones speak of a glorious past.", []string{"north", "south", "east", "west"})
	north := NewRoom("North Temple", "The northern temple hall. Stone pillars reach toward the ceiling.", []string{"south", "up"})
	tower := NewRoom("Bell Tower", "A tall bell tower. You can see far into the distance.", []string{"down"})
	south := NewRoom("South Garden", "A peaceful southern garden. Flowers bloom in every season.", []string{"north"})
	east := NewRoom("East Market", "A busy eastern market. Merchants sell their wares.", []string{"west"})
	west := NewRoom("West Temple", "An ancient western temple. The air is thick with incense.", []string{"east"})

	// Link rooms
	square.UpdateExit("north", north.ID)
	square.UpdateExit("south", south.ID)
	square.UpdateExit("east", east.ID)
	square.UpdateExit("west", west.ID)

	north.UpdateExit("south", square.ID)
	north.UpdateExit("up", tower.ID)

	tower.UpdateExit("down", north.ID)

	south.UpdateExit("north", square.ID)
	east.UpdateExit("west", square.ID)
	west.UpdateExit("east", square.ID)

	// Add all rooms
	m.AddOrUpdateRoom(square)
	m.LinkRooms()
	m.AddOrUpdateRoom(north)
	m.LinkRooms()
	m.AddOrUpdateRoom(tower)
	m.LinkRooms()
	m.AddOrUpdateRoom(south)
	m.LinkRooms()
	m.AddOrUpdateRoom(east)
	m.LinkRooms()
	m.AddOrUpdateRoom(west)
	m.LinkRooms()

	// Set square as current
	m.CurrentRoomID = square.ID

	return m
}

// BenchmarkRenderMap benchmarks the map rendering performance
func BenchmarkRenderMap(b *testing.B) {
	m := createTestMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FormatMapPanel(40, 15)
	}
}
