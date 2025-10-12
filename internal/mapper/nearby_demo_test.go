package mapper

import (
"fmt"
"testing"
)

// TestNearbyDemo demonstrates the FindNearbyRooms functionality
func TestNearbyDemo(t *testing.T) {
// Create a simple map structure:
//           Room4
//             |
//           Room2
//             |
// Room5 - Room1 - Room3
//             |
//           Room6

m := NewMap()

room1 := NewRoom("Center Plaza", "The center of the plaza.", []string{"north", "south", "east", "west"})
room2 := NewRoom("North Market", "A bustling northern market.", []string{"south", "north"})
room3 := NewRoom("East Temple", "The eastern temple.", []string{"west"})
room4 := NewRoom("Far North Gate", "The northern gate of the city.", []string{"south"})
room5 := NewRoom("West Gardens", "Beautiful western gardens.", []string{"east"})
room6 := NewRoom("South Fountain", "A fountain to the south.", []string{"north"})

// Build the map starting from Center Plaza
m.AddOrUpdateRoom(room1)
	m.LinkRooms()

// Go north to Room2
m.SetLastDirection("north")
m.AddOrUpdateRoom(room2)
	m.LinkRooms()

// Go north to Room4
m.SetLastDirection("north")
m.AddOrUpdateRoom(room4)
	m.LinkRooms()

// Return to Room2
m.SetLastDirection("south")
m.AddOrUpdateRoom(room2)
	m.LinkRooms()

// Return to Center Plaza
m.SetLastDirection("south")
m.AddOrUpdateRoom(room1)
	m.LinkRooms()

// Go east to Room3
m.SetLastDirection("east")
m.AddOrUpdateRoom(room3)
	m.LinkRooms()

// Return to Center Plaza
m.SetLastDirection("west")
m.AddOrUpdateRoom(room1)
	m.LinkRooms()

// Go west to Room5
m.SetLastDirection("west")
m.AddOrUpdateRoom(room5)
	m.LinkRooms()

// Return to Center Plaza
m.SetLastDirection("east")
m.AddOrUpdateRoom(room1)
	m.LinkRooms()

// Go south to Room6
m.SetLastDirection("south")
m.AddOrUpdateRoom(room6)
	m.LinkRooms()

// Return to Center Plaza (our final position)
m.SetLastDirection("north")
m.AddOrUpdateRoom(room1)
	m.LinkRooms()

// Now test FindNearbyRooms from Center Plaza
t.Log("=== Testing FindNearbyRooms ===")
t.Logf("Current room: %s\n", m.GetCurrentRoom().Title)

nearby := m.FindNearbyRooms(5)

if len(nearby) == 0 {
t.Fatal("No nearby rooms found!")
}

t.Logf("Found %d rooms within 5 steps:\n", len(nearby))

currentDistance := -1
for i, nr := range nearby {
// Print distance header when it changes
if nr.Distance != currentDistance {
currentDistance = nr.Distance
stepLabel := "step"
if currentDistance > 1 {
stepLabel = "steps"
}
t.Logf("%d %s away:", currentDistance, stepLabel)
}

t.Logf("  %d. %s", i+1, nr.Room.Title)
}

// Verify expected results
// From Center Plaza, we should find all 5 other rooms
if len(nearby) != 5 {
t.Errorf("Expected 5 nearby rooms, got %d", len(nearby))
}

// Rooms at distance 1: North Market, East Temple, West Gardens, South Fountain (4 rooms)
distance1Count := 0
for _, nr := range nearby {
if nr.Distance == 1 {
distance1Count++
}
}
if distance1Count != 4 {
t.Errorf("Expected 4 rooms at distance 1, got %d", distance1Count)
}

// Rooms at distance 2: Far North Gate (1 room)
distance2Count := 0
for _, nr := range nearby {
if nr.Distance == 2 {
distance2Count++
}
}
if distance2Count != 1 {
t.Errorf("Expected 1 room at distance 2, got %d", distance2Count)
}

t.Log("\n=== Test successful! ===")
}

// ExampleMap_FindNearbyRooms demonstrates usage of FindNearbyRooms
func ExampleMap_FindNearbyRooms() {
m := NewMap()

// Create a simple map
center := NewRoom("Center", "The center room.", []string{"north", "south"})
north := NewRoom("North", "The north room.", []string{"south"})
south := NewRoom("South", "The south room.", []string{"north"})

m.AddOrUpdateRoom(center)
	m.LinkRooms()
m.SetLastDirection("north")
m.AddOrUpdateRoom(north)
	m.LinkRooms()
m.SetLastDirection("south")
m.AddOrUpdateRoom(center)
	m.LinkRooms()
m.SetLastDirection("south")
m.AddOrUpdateRoom(south)
	m.LinkRooms()
m.SetLastDirection("north")
m.AddOrUpdateRoom(center)
	m.LinkRooms()

// Find nearby rooms
nearby := m.FindNearbyRooms(5)

fmt.Printf("Found %d nearby rooms\n", len(nearby))
for _, nr := range nearby {
fmt.Printf("%s at distance %d\n", nr.Room.Title, nr.Distance)
}

// Output:
// Found 2 nearby rooms
// North at distance 1
// South at distance 1
}
