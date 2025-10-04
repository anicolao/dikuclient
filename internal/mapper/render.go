package mapper

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

		// Process each exit
		for direction, destID := range room.Exits {
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
	charsPerRoom := 2 // room + space
	maxRoomsPerLine := width / charsPerRoom

	// Calculate viewport bounds to center on (0,0)
	viewHalfWidth := maxRoomsPerLine / 2
	viewHalfHeight := height / 2

	viewMinX := -viewHalfWidth
	viewMaxX := viewHalfWidth
	viewMinY := -viewHalfHeight
	viewMaxY := viewHalfHeight

	// Build the display line by line
	var lines []string
	for y := viewMinY; y <= viewMaxY; y++ {
		var line strings.Builder
		for x := viewMinX; x <= viewMaxX; x++ {
			coord := Coordinate{X: x, Y: y}
			room := grid[coord]

			if room != nil {
				// Check if this is the current room (at 0,0)
				if x == 0 && y == 0 {
					line.WriteString(currentRoomStyle.Render("▣")) // Current room - filled square
				} else {
					line.WriteString(visitedRoomStyle.Render("▢")) // Visited room - hollow square
				}
			} else {
				line.WriteString(" ") // Empty space
			}

			// Add space between rooms (except last column)
			if x < viewMaxX {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
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

// FormatMapPanel formats the complete map panel with header and vertical exits indicator
func (m *Map) FormatMapPanel(width, height int) string {
	mapContent, roomTitle := m.RenderMap(width, height-3) // Reserve space for header and vertical exits

	if roomTitle == "" {
		roomTitle = "Map"
	}

	// Get vertical exits
	hasUp, hasDown := m.GetVerticalExits()
	verticalSymbol := RenderVerticalExits(hasUp, hasDown)

	// Build the panel content
	var lines []string
	
	// Add the map content
	lines = append(lines, strings.Split(mapContent, "\n")...)

	// Add vertical exits indicator if present (centered below the map)
	if verticalSymbol != "" {
		// Add an empty line first
		lines = append(lines, "")
		// Center the vertical symbol
		padding := strings.Repeat(" ", width/2)
		lines = append(lines, fmt.Sprintf("%s%s", padding, verticalSymbol))
	}

	return strings.Join(lines, "\n")
}
