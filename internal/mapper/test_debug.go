package mapper

import (
"testing"
"fmt"
)

func TestDebugUnexploredConnections(t *testing.T) {
m := NewMap()

// Create a simple setup
center := NewRoom("Center", "Center room", []string{"north"})
center.Exits["north"] = "" // Unexplored exit

m.AddOrUpdateRoom(center)
m.CurrentRoomID = center.ID

// Build grid
grid := m.buildRoomGrid(center, 30, 10)

// Check what's in the grid
for coord, marker := range grid {
if marker.IsUnknown {
t.Logf("Unexplored at (%d, %d)", coord.X, coord.Y)
} else if marker.Room != nil {
t.Logf("Room '%s' at (%d, %d)", marker.Room.Title, coord.X, coord.Y)
}
}

content := m.FormatMapPanel(30, 10)
fmt.Printf("Map:\n%s\n", content)
}
