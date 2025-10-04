package mapper

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// sortDirections sorts directions in a consistent order for deterministic rendering
func sortDirections(dirs []string) {
	// Define priority order for common directions
	priority := map[string]int{
		"north": 0, "n": 0,
		"south": 1, "s": 1,
		"east": 2, "e": 2,
		"west": 3, "w": 3,
		"up": 4, "u": 4,
		"down": 5, "d": 5,
	}

	sort.Slice(dirs, func(i, j int) bool {
		pi, oki := priority[dirs[i]]
		pj, okj := priority[dirs[j]]

		// If both have priority, sort by priority
		if oki && okj {
			return pi < pj
		}
		// If only one has priority, it comes first
		if oki {
			return true
		}
		if okj {
			return false
		}
		// Otherwise sort alphabetically
		return dirs[i] < dirs[j]
	})
}

// Coordinate represents a position in the 2D grid
type Coordinate struct {
	X, Y int
}

// RenderMap generates a visual representation of the map for the given dimensions
// Returns the rendered map as a string and the current room title
func (m *Map) RenderMap(width, height int) (string, string) {
	currentRoom := m.GetCurrentRoom()
	if currentRoom == nil {
		return "(exploring...)", ""
	}

	// Build the room grid centered on current room
	roomGrid := m.buildRoomGrid(currentRoom, width, height)

	// Render the grid to string
	rendered := renderGrid(roomGrid, width, height, nil)

	return rendered, currentRoom.Title
}

// RenderMapWithLegend generates a visual representation of the map with room numbers
// Returns the rendered map as a string and the current room title
func (m *Map) RenderMapWithLegend(width, height int, legend map[string]int) (string, string) {
	currentRoom := m.GetCurrentRoom()
	if currentRoom == nil {
		return "(exploring...)", ""
	}

	// Build the room grid centered on current room
	roomGrid := m.buildRoomGrid(currentRoom, width, height)

	// Render the grid to string with legend
	rendered := renderGrid(roomGrid, width, height, legend)

	return rendered, currentRoom.Title
}

// RoomMarker represents a room or unexplored area in the grid
type RoomMarker struct {
	Room       *Room
	IsUnknown  bool // True if this is an unexplored exit
}

// buildRoomGrid creates a 2D grid of rooms centered on the current room
func (m *Map) buildRoomGrid(currentRoom *Room, width, height int) map[Coordinate]*RoomMarker {
	grid := make(map[Coordinate]*RoomMarker)

	// Place current room at center (0, 0)
	center := Coordinate{0, 0}
	grid[center] = &RoomMarker{Room: currentRoom, IsUnknown: false}

	// BFS to place connected rooms relative to current room
	visited := make(map[string]bool)
	visited[currentRoom.ID] = true
	exploredCoords := make(map[Coordinate]bool)
	exploredCoords[center] = true

	type queueItem struct {
		roomID string
		coord  Coordinate
	}

	queue := []queueItem{{roomID: currentRoom.ID, coord: center}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		room := m.Rooms[current.roomID]
		if room == nil {
			continue
		}

		// Process each exit in a deterministic order
		// Sort directions to ensure consistent rendering
		var directions []string
		for direction := range room.Exits {
			directions = append(directions, direction)
		}
		// Sort in a specific order: north, south, east, west, up, down, then alphabetically
		sortDirections(directions)

		for _, direction := range directions {
			destID := room.Exits[direction]
			
			// Calculate new coordinate based on direction
			newCoord := current.coord
			switch direction {
			case "north", "n":
				newCoord.Y--
			case "south", "s":
				newCoord.Y++
			case "east", "e":
				newCoord.X++
			case "west", "w":
				newCoord.X--
			case "up", "u", "down", "d":
				// Vertical exits don't add new grid positions
				continue
			default:
				// Unknown direction, skip
				continue
			}

			// Skip if we've already processed this coordinate
			if exploredCoords[newCoord] {
				continue
			}
			exploredCoords[newCoord] = true

			// Empty destID means there's an exit but it hasn't been explored yet
			if destID == "" {
				// This is an unexplored exit - mark it as unknown
				grid[newCoord] = &RoomMarker{Room: nil, IsUnknown: true}
			} else {
				destRoom := m.Rooms[destID]
				if destRoom == nil {
					// This is an unexplored exit - mark it as unknown
					grid[newCoord] = &RoomMarker{Room: nil, IsUnknown: true}
				} else {
					// This is an explored room
					grid[newCoord] = &RoomMarker{Room: destRoom, IsUnknown: false}
					if !visited[destID] {
						visited[destID] = true
						queue = append(queue, queueItem{roomID: destID, coord: newCoord})
					}
				}
			}
		}
	}

	return grid
}

// GetVisibleRoomIDs returns the IDs of rooms that are visible on the current map display
func (m *Map) GetVisibleRoomIDs(width, height int) []string {
	currentRoom := m.GetCurrentRoom()
	if currentRoom == nil {
		return nil
	}

	// Build the room grid to see what's visible
	grid := m.buildRoomGrid(currentRoom, width, height)

	// Extract room IDs from the grid
	roomIDs := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, marker := range grid {
		if marker != nil && !marker.IsUnknown && marker.Room != nil {
			if !seen[marker.Room.ID] {
				seen[marker.Room.ID] = true
				roomIDs = append(roomIDs, marker.Room.ID)
			}
		}
	}

	return roomIDs
}

// renderGrid converts the room grid to a visual string representation
// If legend is provided, rooms in the legend will be shown with their number instead of symbol
func renderGrid(grid map[Coordinate]*RoomMarker, width, height int, legend map[string]int) string {
	// Define styles for different room types
	currentRoomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/gold
	visitedRoomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // White
	unexploredRoomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dark gray
	connectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dark gray for connections

	// Calculate how many characters we can fit
	// When using legend, we need more space per room for numbers
	charsPerRoom := 3 // room + double connector space
	linesPerRoom := 2 // room line + connector line
	
	// If legend is active, adjust spacing for multi-digit numbers
	if legend != nil && len(legend) > 0 {
		// Find max number to determine spacing
		maxNum := 0
		for _, num := range legend {
			if num > maxNum {
				maxNum = num
			}
		}
		// Adjust chars per room based on max number width
		if maxNum >= 100 {
			charsPerRoom = 5 // 3 digits + 2 connector
		} else if maxNum >= 10 {
			charsPerRoom = 4 // 2 digits + 2 connector
		}
	}

	maxRoomsPerLine := width / charsPerRoom
	maxRoomsPerHeight := height / linesPerRoom

	// Calculate viewport bounds to center on (0,0)
	viewHalfWidth := maxRoomsPerLine / 2
	viewHalfHeight := maxRoomsPerHeight / 2

	viewMinX := -viewHalfWidth
	viewMaxX := viewHalfWidth
	viewMinY := -viewHalfHeight
	viewMaxY := viewHalfHeight

	// Build the display line by line, alternating between room lines and connector lines
	var lines []string
	for y := viewMinY; y <= viewMaxY; y++ {
		// Room line
		var roomLine strings.Builder
		// Connector line (vertical connections below this row)
		var connLine strings.Builder

		for x := viewMinX; x <= viewMaxX; x++ {
			coord := Coordinate{X: x, Y: y}
			marker := grid[coord]

			// Render the room symbol
			if marker != nil {
				if marker.IsUnknown {
					// Unexplored room - always show as gray ▦
					roomLine.WriteString(unexploredRoomStyle.Render("▦"))
				} else {
					room := marker.Room
					isCurrentRoom := (x == 0 && y == 0)
					
					// Check if this room is in the legend
					if legend != nil {
						if roomNum, inLegend := legend[room.ID]; inLegend {
							// Show room number from legend
							symbol := fmt.Sprintf("%d", roomNum)
							if isCurrentRoom {
								roomLine.WriteString(currentRoomStyle.Render(symbol))
							} else {
								roomLine.WriteString(visitedRoomStyle.Render(symbol))
							}
						} else {
							// Not in legend, use regular symbol
							if isCurrentRoom {
								roomLine.WriteString(currentRoomStyle.Render("▣"))
							} else {
								roomLine.WriteString(visitedRoomStyle.Render("▢"))
							}
						}
					} else {
						// No legend, use symbols based on vertical exits
						hasUp := false
						hasDown := false
						for direction := range room.Exits {
							switch direction {
							case "up", "u":
								hasUp = true
							case "down", "d":
								hasDown = true
							}
						}

						// Determine the symbol based on vertical exits
						var symbol string
						
						if hasUp && hasDown {
							symbol = "⇅" // Both up and down
						} else if hasUp {
							symbol = "⇱" // Up only
						} else if hasDown {
							symbol = "⇲" // Down only
						} else {
							// No vertical exits - use regular room symbols
							if isCurrentRoom {
								symbol = "▣" // Current room - filled square
							} else {
								symbol = "▢" // Visited room - hollow square
							}
						}
						
						// Apply color - current room is always yellow, others are white
						if isCurrentRoom {
							roomLine.WriteString(currentRoomStyle.Render(symbol))
						} else {
							roomLine.WriteString(visitedRoomStyle.Render(symbol))
						}
					}
				}
			} else {
				roomLine.WriteString(" ") // Empty space
			}

			// Render horizontal connector (to the right of this room)
			if x < viewMaxX {
				// Check if there's a connection to the east
				hasEastConnection := false
				eastCoord := Coordinate{X: x + 1, Y: y}
				eastMarker := grid[eastCoord]
				
				if marker != nil && eastMarker != nil {
					// Check from current room to east
					if !marker.IsUnknown && marker.Room != nil {
						// Check if current room has east exit
						for dir, destID := range marker.Room.Exits {
							if (dir == "east" || dir == "e") {
								// Connection exists if:
								// 1. East room is unexplored (destID is empty or room doesn't exist)
								// 2. East room is known and IDs match
								if eastMarker.IsUnknown || 
								   (eastMarker.Room != nil && destID == eastMarker.Room.ID) {
									hasEastConnection = true
									break
								}
							}
						}
					}
					// Also check from east room to current (for unexplored rooms pointing back)
					if !hasEastConnection && !eastMarker.IsUnknown && eastMarker.Room != nil {
						// Check if east room has west exit pointing to current
						for dir, destID := range eastMarker.Room.Exits {
							if (dir == "west" || dir == "w") {
								if marker.IsUnknown ||
								   (marker.Room != nil && destID == marker.Room.ID) {
									hasEastConnection = true
									break
								}
							}
						}
					}
				}
				
				if hasEastConnection {
					roomLine.WriteString(connectionStyle.Render("──"))
				} else {
					roomLine.WriteString("  ")
				}
			}

			// Render vertical connector (below this room)
			if y < viewMaxY {
				// Check if there's a connection to the south
				hasSouthConnection := false
				southCoord := Coordinate{X: x, Y: y + 1}
				southMarker := grid[southCoord]
				
				if marker != nil && southMarker != nil {
					// Check from current room to south
					if !marker.IsUnknown && marker.Room != nil {
						// Check if current room has south exit
						for dir, destID := range marker.Room.Exits {
							if (dir == "south" || dir == "s") {
								// Connection exists if:
								// 1. South room is unexplored (destID is empty or room doesn't exist)
								// 2. South room is known and IDs match
								if southMarker.IsUnknown ||
								   (southMarker.Room != nil && destID == southMarker.Room.ID) {
									hasSouthConnection = true
									break
								}
							}
						}
					}
					// Also check from south room to current (for unexplored rooms pointing back)
					if !hasSouthConnection && !southMarker.IsUnknown && southMarker.Room != nil {
						// Check if south room has north exit pointing to current
						for dir, destID := range southMarker.Room.Exits {
							if (dir == "north" || dir == "n") {
								if marker.IsUnknown ||
								   (marker.Room != nil && destID == marker.Room.ID) {
									hasSouthConnection = true
									break
								}
							}
						}
					}
				}
				
				if hasSouthConnection {
					connLine.WriteString(connectionStyle.Render("│"))
				} else {
					connLine.WriteString(" ")
				}
			}

			// Add spacing for connector column in connector line (2 spaces for double connector)
			if x < viewMaxX && y < viewMaxY {
				connLine.WriteString("  ")
			}
		}

		lines = append(lines, roomLine.String())
		if y < viewMaxY {
			lines = append(lines, connLine.String())
		}
	}

	return strings.Join(lines, "\n")
}

// GetVerticalExits returns symbols for up/down exits from current room
// Returns: hasUp, hasDown
func (m *Map) GetVerticalExits() (bool, bool) {
	currentRoom := m.GetCurrentRoom()
	if currentRoom == nil {
		return false, false
	}

	hasUp := false
	hasDown := false

	for direction := range currentRoom.Exits {
		switch direction {
		case "up", "u":
			hasUp = true
		case "down", "d":
			hasDown = true
		}
	}

	return hasUp, hasDown
}

// RenderVerticalExits returns the symbol for vertical exits
func RenderVerticalExits(hasUp, hasDown bool) string {
	if hasUp && hasDown {
		return "⇅" // Both up and down
	} else if hasUp {
		return "⇱" // Up only
	} else if hasDown {
		return "⇲" // Down only
	}
	return ""
}

// FormatMapPanel formats the complete map panel
func (m *Map) FormatMapPanel(width, height int) string {
	// Render the map using the full available height
	mapContent, _ := m.RenderMap(width, height)
	return mapContent
}

// FormatMapPanelWithLegend formats the complete map panel with optional room number legend
func (m *Map) FormatMapPanelWithLegend(width, height int, legend map[string]int) string {
	// Render the map with legend using the full available height
	mapContent, _ := m.RenderMapWithLegend(width, height, legend)
	return mapContent
}
