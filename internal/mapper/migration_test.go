package mapper

import (
"encoding/json"
"os"
"path/filepath"
"testing"
)

func TestRoomNumberingMigration(t *testing.T) {
// Create a temporary map file without room_numbering
tmpDir := t.TempDir()
mapPath := filepath.Join(tmpDir, "map.json")

// Create map data without room_numbering field (simulating old map file)
room1 := NewRoom("Room A", "First room", []string{"north"})
room2 := NewRoom("Room B", "Second room", []string{"south"})
room3 := NewRoom("Room C", "Third room", []string{"east"})

oldMap := map[string]interface{}{
"rooms": map[string]*Room{
room1.ID: room1,
room2.ID: room2,
room3.ID: room3,
},
"current_room_id":  room1.ID,
"previous_room_id": "",
"last_direction":   "",
}

data, err := json.MarshalIndent(oldMap, "", "  ")
if err != nil {
t.Fatalf("Failed to marshal test data: %v", err)
}

if err := os.WriteFile(mapPath, data, 0600); err != nil {
t.Fatalf("Failed to write test map file: %v", err)
}

// Load the map
loadedMap, err := LoadFromPath(mapPath)
if err != nil {
t.Fatalf("Failed to load map: %v", err)
}

// Verify RoomNumbering was populated
if len(loadedMap.RoomNumbering) != 3 {
t.Errorf("Expected RoomNumbering to have 3 entries, got %d", len(loadedMap.RoomNumbering))
}

// Verify all rooms are in the numbering
foundRooms := make(map[string]bool)
for _, id := range loadedMap.RoomNumbering {
foundRooms[id] = true
}

if !foundRooms[room1.ID] {
t.Error("Room1 not found in RoomNumbering")
}
if !foundRooms[room2.ID] {
t.Error("Room2 not found in RoomNumbering")
}
if !foundRooms[room3.ID] {
t.Error("Room3 not found in RoomNumbering")
}

t.Logf("RoomNumbering after migration: %v", loadedMap.RoomNumbering)
}
