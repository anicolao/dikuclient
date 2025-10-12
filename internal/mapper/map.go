package mapper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Map represents the entire MUD world map
type Map struct {
	Rooms          map[string]*Room `json:"rooms"`            // roomID -> Room
	CurrentRoomID  string           `json:"current_room_id"`  // ID of current room
	PreviousRoomID string           `json:"previous_room_id"` // ID of previous room (for linking)
	LastDirection  string           `json:"last_direction"`   // Last movement direction
	RoomNumbering  []string         `json:"room_numbering"`   // Ordered list of room IDs for durable numbering
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

	// Migrate: populate RoomNumbering if it's empty (for existing map files)
	migrated := false
	if len(m.RoomNumbering) == 0 && len(m.Rooms) > 0 {
		// Add all existing rooms to numbering in a deterministic order
		roomIDs := make([]string, 0, len(m.Rooms))
		for id := range m.Rooms {
			roomIDs = append(roomIDs, id)
		}
		// Sort by room ID for consistency
		sort.Strings(roomIDs)
		m.RoomNumbering = roomIDs
		migrated = true
	}

	// Save the map if we migrated to persist the room numbering
	if migrated {
		if err := m.Save(); err != nil {
			// Log error but don't fail - migration still worked in memory
			fmt.Fprintf(os.Stderr, "Warning: failed to save migrated map: %v\n", err)
		}
	}

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
	// Check if we're following a known exit from the current room
	// If so, we might be revisiting an existing room
	var knownDestID string
	if m.CurrentRoomID != "" && m.LastDirection != "" {
		if currentRoom, exists := m.Rooms[m.CurrentRoomID]; exists {
			if destID, hasExit := currentRoom.Exits[m.LastDirection]; hasExit && destID != "" {
				// We have a known destination for this exit
				if _, destExists := m.Rooms[destID]; destExists {
					// Verify the destination room matches the characteristics of the room we're entering
					// Extract exits for comparison
					exits := make([]string, 0, len(room.Exits))
					for direction := range room.Exits {
						exits = append(exits, direction)
					}
					baseID := GenerateRoomID(room.Title, room.Description, exits)
					
					// Extract base ID from destination room
					destBaseID := destID
					lastPipe := strings.LastIndex(destID, "|")
					if lastPipe > 0 && lastPipe < len(destID)-1 {
						suffix := destID[lastPipe+1:]
						if _, err := fmt.Sscanf(suffix, "%d", new(int)); err == nil {
							destBaseID = destID[:lastPipe]
						}
					}
					
					// If characteristics match, we're revisiting this known room
					if baseID == destBaseID {
						knownDestID = destID
					}
				}
			}
		}
	}
	
	if knownDestID != "" {
		// We're revisiting a known room via a known exit
		existing := m.Rooms[knownDestID]
		existing.VisitCount++

		// Merge exits (keep existing mappings, add new ones)
		for direction, destID := range room.Exits {
			if _, hasExit := existing.Exits[direction]; !hasExit {
				existing.Exits[direction] = destID
			}
		}
		
		// Update the room object's ID to match the existing room
		room.ID = knownDestID
	} else {
		// Either a new room or revisiting via an unknown path
		// Extract exits from the room for ID generation
		exits := make([]string, 0, len(room.Exits))
		for direction := range room.Exits {
			exits = append(exits, direction)
		}
		
		// Calculate distance to room 0 for this room's position
		distance := m.calculateDistanceToRoom0()
		
		// If we have a current room and can calculate distance, add 1 for the move to the new room
		if m.CurrentRoomID != "" && distance >= 0 {
			distance++
		}
		
		// For the very first room (no current room), distance is 0 as it becomes room 0
		if m.CurrentRoomID == "" && len(m.RoomNumbering) == 0 {
			distance = 0
		}
		
		// Generate room ID with distance - this is the unique identifier
		newID := GenerateRoomIDWithDistance(room.Title, room.Description, exits, distance)
		
		// Check if room with this exact ID (including distance) already exists
		if existing, exists := m.Rooms[newID]; exists {
			// Room already exists at this distance, increment visit count
			existing.VisitCount++

			// Merge exits (keep existing mappings, add new ones)
			for direction, destID := range room.Exits {
				if _, hasExit := existing.Exits[direction]; !hasExit {
					existing.Exits[direction] = destID
				}
			}
			
			// Update the room object's ID to match the existing room
			room.ID = newID
		} else {
			// New room - add it to the map
			room.ID = newID
			m.Rooms[room.ID] = room
			
			// Add to room numbering if not already present
			m.addToRoomNumbering(room.ID)
		}
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

// addToRoomNumbering adds a room ID to the numbering list if not already present
func (m *Map) addToRoomNumbering(roomID string) {
	// Check if already in the list
	for _, id := range m.RoomNumbering {
		if id == roomID {
			return
		}
	}
	// Add to the end
	m.RoomNumbering = append(m.RoomNumbering, roomID)
}

// calculateDistanceToRoom0 calculates the BFS distance from the current room to room 0 (first room in numbering)
// Returns -1 if room 0 doesn't exist or is unreachable
func (m *Map) calculateDistanceToRoom0() int {
	// If no rooms in numbering, there is no room 0
	if len(m.RoomNumbering) == 0 {
		return -1
	}

	// If no current room, we can't calculate distance
	if m.CurrentRoomID == "" {
		return -1
	}

	room0ID := m.RoomNumbering[0]

	// If current room is room 0, distance is 0
	if m.CurrentRoomID == room0ID {
		return 0
	}

	// BFS to find shortest path from current room to room 0
	type queueItem struct {
		roomID   string
		distance int
	}

	visited := make(map[string]bool)
	queue := []queueItem{{roomID: m.CurrentRoomID, distance: 0}}
	visited[m.CurrentRoomID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		room := m.Rooms[current.roomID]
		if room == nil {
			continue
		}

		// Check all exits
		for _, destID := range room.Exits {
			if destID == "" {
				continue // Unknown destination
			}

			if destID == room0ID {
				// Found room 0!
				return current.distance + 1
			}

			if !visited[destID] {
				visited[destID] = true
				queue = append(queue, queueItem{roomID: destID, distance: current.distance + 1})
			}
		}
	}

	// Room 0 is not reachable from current room
	return -1
}

// GetRoomNumber returns the durable room number for a given room ID
// Returns 0 if the room is not in the numbering
func (m *Map) GetRoomNumber(roomID string) int {
	for i, id := range m.RoomNumbering {
		if id == roomID {
			return i + 1 // 1-indexed
		}
	}
	return 0
}

// GetRoomByNumber returns the room for a given durable room number
// Returns nil if the number is out of range
func (m *Map) GetRoomByNumber(number int) *Room {
	if number < 1 || number > len(m.RoomNumbering) {
		return nil
	}
	roomID := m.RoomNumbering[number-1]
	return m.Rooms[roomID]
}
