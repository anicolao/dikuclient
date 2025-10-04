package mapper

import (
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
	rendered := renderGrid(roomGrid, width, height)

	return rendered, currentRoom.Title
}

// buildRoomGrid creates a 2D grid of rooms centered on the current room
func (m *Map) buildRoomGrid(currentRoom *Room, width, height int) map[Coordinate]*Room {
	grid := make(map[Coordinate]*Room)

	// Place current room at center (0, 0)
	center := Coordinate{0, 0}
	grid[center] = currentRoom

	// BFS to place connected rooms relative to current room
	visited := make(map[string]bool)
	visited[currentRoom.ID] = true

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
			if destID == "" || visited[destID] {
				continue
			}

			destRoom := m.Rooms[destID]
			if destRoom == nil {
				continue
			}

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

			// Add to grid
			grid[newCoord] = destRoom
			visited[destID] = true
			queue = append(queue, queueItem{roomID: destID, coord: newCoord})
		}
	}

	return grid
}

// renderGrid converts the room grid to a visual string representation
func renderGrid(grid map[Coordinate]*Room, width, height int) string {
	// Define styles for different room types
	currentRoomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow/gold
	visitedRoomStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // White
	verticalExitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // White for vertical exits

	// Calculate the grid bounds
	minX, maxX := 0, 0
	minY, maxY := 0, 0
	for coord := range grid {
		if coord.X < minX {
			minX = coord.X
		}
		if coord.X > maxX {
			maxX = coord.X
		}
		if coord.Y < minY {
			minY = coord.Y
		}
		if coord.Y > maxY {
			maxY = coord.Y
		}
	}

	// Calculate how many characters we can fit
	// Each room is 1 character, with 1 space between rooms
	// So we need 2 characters per room horizontally
	// We need 2 rows per room row (one for rooms, one for vertical exits)
	charsPerRoom := 2 // room + space
	maxRoomsPerLine := width / charsPerRoom
	
	// We need 2 lines per row (room line + vertical exit line)
	maxRoomRows := height / 2

	// Calculate viewport bounds to center on (0,0)
	viewHalfWidth := maxRoomsPerLine / 2
	viewHalfHeight := maxRoomRows / 2

	viewMinX := -viewHalfWidth
	viewMaxX := viewHalfWidth
	viewMinY := -viewHalfHeight
	viewMaxY := viewHalfHeight

	// Build the display line by line
	var lines []string
	for y := viewMinY; y <= viewMaxY; y++ {
		// First line: room symbols
		var roomLine strings.Builder
		for x := viewMinX; x <= viewMaxX; x++ {
			coord := Coordinate{X: x, Y: y}
			room := grid[coord]

			if room != nil {
				// Check if this is the current room (at 0,0)
				if x == 0 && y == 0 {
					roomLine.WriteString(currentRoomStyle.Render("▣")) // Current room - filled square
				} else {
					roomLine.WriteString(visitedRoomStyle.Render("▢")) // Visited room - hollow square
				}
			} else {
				roomLine.WriteString(" ") // Empty space
			}

			// Add space between rooms (except last column)
			if x < viewMaxX {
				roomLine.WriteString(" ")
			}
		}
		lines = append(lines, roomLine.String())
		
		// Second line: vertical exit indicators (between this row and next)
		var verticalLine strings.Builder
		for x := viewMinX; x <= viewMaxX; x++ {
			coord := Coordinate{X: x, Y: y}
			room := grid[coord]

			if room != nil {
				// Check for vertical exits
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
				
				var symbol string
				if hasUp && hasDown {
					symbol = verticalExitStyle.Render("⇅") // Both up and down
				} else if hasUp {
					symbol = verticalExitStyle.Render("⇱") // Up only
				} else if hasDown {
					symbol = verticalExitStyle.Render("⇲") // Down only
				} else {
					symbol = " " // No vertical exit
				}
				
				verticalLine.WriteString(symbol)
			} else {
				verticalLine.WriteString(" ") // Empty space
			}

			// Add space between columns (except last column)
			if x < viewMaxX {
				verticalLine.WriteString(" ")
			}
		}
		lines = append(lines, verticalLine.String())
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
