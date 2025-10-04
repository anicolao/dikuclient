package mapper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Map represents the entire MUD world map
type Map struct {
	Rooms          map[string]*Room `json:"rooms"`            // roomID -> Room
	CurrentRoomID  string           `json:"current_room_id"`  // ID of current room
	PreviousRoomID string           `json:"previous_room_id"` // ID of previous room (for linking)
	LastDirection  string           `json:"last_direction"`   // Last movement direction
	mapPath        string           // Path to the map file (not serialized)
}

// NewMap creates a new empty map
func NewMap() *Map {
	return &Map{
		Rooms: make(map[string]*Room),
	}
}

// GetMapPath returns the path to the map file
func GetMapPath() (string, error) {
	var configDir string

	// Check for environment variable override
	if envConfigDir := os.Getenv("DIKUCLIENT_CONFIG_DIR"); envConfigDir != "" {
		configDir = envConfigDir
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config", "dikuclient")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "map.json"), nil
}

// Load loads the map from disk
func Load() (*Map, error) {
	mapPath, err := GetMapPath()
	if err != nil {
		return nil, err
	}

	return LoadFromPath(mapPath)
}

// LoadFromPath loads map from a specific path (useful for testing)
func LoadFromPath(mapPath string) (*Map, error) {
	data, err := os.ReadFile(mapPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty map if file doesn't exist
			m := NewMap()
			m.mapPath = mapPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	var m Map
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse map file: %w", err)
	}
	m.mapPath = mapPath

	return &m, nil
}

// Save saves the map to disk
func (m *Map) Save() error {
	mapPath := m.mapPath
	if mapPath == "" {
		var err error
		mapPath, err = GetMapPath()
		if err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := os.WriteFile(mapPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write map file: %w", err)
	}

	return nil
}

// AddOrUpdateRoom adds a new room or updates an existing one
func (m *Map) AddOrUpdateRoom(room *Room) {
	if existing, exists := m.Rooms[room.ID]; exists {
		// Room already exists, increment visit count
		existing.VisitCount++

		// Merge exits (keep existing mappings, add new ones)
		for direction, destID := range room.Exits {
			if _, hasExit := existing.Exits[direction]; !hasExit {
				existing.Exits[direction] = destID
			}
		}
	} else {
		// New room - add it to the map
		m.Rooms[room.ID] = room
	}

	// Get the room from the map (whether new or existing)
	currentRoom := m.Rooms[room.ID]

	// Link from current room (before we move) if we have the information
	if m.CurrentRoomID != "" && m.LastDirection != "" && m.CurrentRoomID != room.ID {
		// Link current room (where we are now) to new room (where we're going)
		if fromRoom, exists := m.Rooms[m.CurrentRoomID]; exists {
			fromRoom.UpdateExit(m.LastDirection, room.ID)
		}

		// Link new room back to current room (reverse direction)
		reverseDir := getReverseDirection(m.LastDirection)
		if reverseDir != "" {
			currentRoom.UpdateExit(reverseDir, m.CurrentRoomID)
		}
	}

	// Update current room tracking
	m.PreviousRoomID = m.CurrentRoomID
	m.CurrentRoomID = room.ID
}

// SetLastDirection records the direction of the last movement
func (m *Map) SetLastDirection(direction string) {
	m.LastDirection = direction
}

// FindRooms searches for rooms matching all query terms
func (m *Map) FindRooms(query string) []*Room {
	queryTerms := strings.Fields(strings.ToLower(query))
	if len(queryTerms) == 0 {
		return nil
	}

	var matches []*Room
	for _, room := range m.Rooms {
		if room.MatchesSearch(queryTerms) {
			matches = append(matches, room)
		}
	}

	return matches
}

// FindPath finds the shortest path from current room to target room
func (m *Map) FindPath(targetRoomID string) []string {
	if m.CurrentRoomID == "" || targetRoomID == "" {
		return nil
	}

	if m.CurrentRoomID == targetRoomID {
		return []string{} // Already at target
	}

	// BFS to find shortest path
	type queueItem struct {
		roomID string
		path   []string
	}

	visited := make(map[string]bool)
	queue := []queueItem{{roomID: m.CurrentRoomID, path: []string{}}}
	visited[m.CurrentRoomID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		room := m.Rooms[current.roomID]
		if room == nil {
			continue
		}

		// Check all exits
		for direction, destID := range room.Exits {
			if destID == "" {
				continue // Unknown destination
			}

			if destID == targetRoomID {
				// Found target!
				return append(current.path, direction)
			}

			if !visited[destID] {
				visited[destID] = true
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = direction
				queue = append(queue, queueItem{roomID: destID, path: newPath})
			}
		}
	}

	return nil // No path found
}

// PathStep represents a single step in a path with direction and destination room
type PathStep struct {
	Direction string
	RoomTitle string
}

// FindPathWithRooms finds the shortest path and returns steps with room information
func (m *Map) FindPathWithRooms(targetRoomID string) []PathStep {
	if m.CurrentRoomID == "" || targetRoomID == "" {
		return nil
	}

	if m.CurrentRoomID == targetRoomID {
		return []PathStep{} // Already at target
	}

	// BFS to find shortest path
	type queueItem struct {
		roomID string
		path   []PathStep
	}

	visited := make(map[string]bool)
	queue := []queueItem{{roomID: m.CurrentRoomID, path: []PathStep{}}}
	visited[m.CurrentRoomID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		room := m.Rooms[current.roomID]
		if room == nil {
			continue
		}

		// Check all exits
		for direction, destID := range room.Exits {
			if destID == "" {
				continue // Unknown destination
			}

			destRoom := m.Rooms[destID]
			if destRoom == nil {
				continue
			}

			step := PathStep{
				Direction: direction,
				RoomTitle: destRoom.Title,
			}

			if destID == targetRoomID {
				// Found target!
				return append(current.path, step)
			}

			if !visited[destID] {
				visited[destID] = true
				newPath := make([]PathStep, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = step
				queue = append(queue, queueItem{roomID: destID, path: newPath})
			}
		}
	}

	return nil // No path found
}

// GetCurrentRoom returns the current room
func (m *Map) GetCurrentRoom() *Room {
	if m.CurrentRoomID == "" {
		return nil
	}
	return m.Rooms[m.CurrentRoomID]
}

// GetAllRooms returns all rooms in the map
func (m *Map) GetAllRooms() map[string]*Room {
	return m.Rooms
}

// NearbyRoom represents a room with its distance from the current location
type NearbyRoom struct {
	Room     *Room
	Distance int
}

// FindNearbyRooms finds all rooms within maxDistance steps of the current room
// Returns rooms sorted by distance (closest first)
func (m *Map) FindNearbyRooms(maxDistance int) []NearbyRoom {
	if m.CurrentRoomID == "" {
		return nil
	}

	// BFS to find all reachable rooms within maxDistance
	type queueItem struct {
		roomID   string
		distance int
	}

	visited := make(map[string]int) // roomID -> distance
	queue := []queueItem{{roomID: m.CurrentRoomID, distance: 0}}
	visited[m.CurrentRoomID] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Skip if we've reached max distance
		if current.distance >= maxDistance {
			continue
		}

		room := m.Rooms[current.roomID]
		if room == nil {
			continue
		}

		// Check all exits
		for _, destID := range room.Exits {
			if destID == "" {
				continue // Unknown destination
			}

			// Only visit if we haven't seen it or found a shorter path
			if prevDist, seen := visited[destID]; !seen || current.distance+1 < prevDist {
				visited[destID] = current.distance + 1
				queue = append(queue, queueItem{roomID: destID, distance: current.distance + 1})
			}
		}
	}

	// Convert to NearbyRoom slice and sort by distance
	var nearby []NearbyRoom
	for roomID, distance := range visited {
		if roomID == m.CurrentRoomID {
			continue // Don't include current room
		}
		if room := m.Rooms[roomID]; room != nil {
			nearby = append(nearby, NearbyRoom{
				Room:     room,
				Distance: distance,
			})
		}
	}

	// Sort by distance, then by title for consistency
	for i := 0; i < len(nearby); i++ {
		for j := i + 1; j < len(nearby); j++ {
			if nearby[i].Distance > nearby[j].Distance ||
				(nearby[i].Distance == nearby[j].Distance && nearby[i].Room.Title > nearby[j].Room.Title) {
				nearby[i], nearby[j] = nearby[j], nearby[i]
			}
		}
	}

	return nearby
}

// getReverseDirection returns the opposite direction
func getReverseDirection(direction string) string {
	reverseMap := map[string]string{
		"north":     "south",
		"south":     "north",
		"east":      "west",
		"west":      "east",
		"up":        "down",
		"down":      "up",
		"ne":        "sw",
		"nw":        "se",
		"se":        "nw",
		"sw":        "ne",
		"northeast": "southwest",
		"northwest": "southeast",
		"southeast": "northwest",
		"southwest": "northeast",
	}

	dir := strings.ToLower(direction)
	if reverse, ok := reverseMap[dir]; ok {
		return reverse
	}
	return ""
}
