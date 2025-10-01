package mapper

import (
	"sort"
	"strings"
)

// Room represents a single room in the MUD world
type Room struct {
	ID            string            `json:"id"`             // Unique identifier based on content
	Title         string            `json:"title"`          // Room title
	Description   string            `json:"description"`    // Full description
	FirstSentence string            `json:"first_sentence"` // First sentence of description
	Exits         map[string]string `json:"exits"`          // direction -> destination room ID
	VisitCount    int               `json:"visit_count"`    // Number of times visited
}

// GenerateRoomID creates a unique ID from title, first sentence, and exits
// Returns human-readable format: "title|first_sentence|exits"
func GenerateRoomID(title, description string, exits []string) string {
	// Extract first sentence from description
	firstSentence := extractFirstSentence(description)

	// Sort exits for consistent ID generation
	sortedExits := make([]string, len(exits))
	copy(sortedExits, exits)
	sort.Strings(sortedExits)

	// Combine elements in human-readable format
	// Use lowercase for consistency but keep it readable
	combined := strings.ToLower(title) + "|" +
		strings.ToLower(firstSentence) + "|" +
		strings.Join(sortedExits, ",")

	return combined
}

// extractFirstSentence extracts the first sentence from a description
func extractFirstSentence(description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return ""
	}

	// Find the first sentence terminator
	for _, terminator := range []string{". ", "! ", "? "} {
		if idx := strings.Index(description, terminator); idx != -1 {
			return strings.TrimSpace(description[:idx+1])
		}
	}

	// If no terminator found, treat first line as sentence
	if idx := strings.Index(description, "\n"); idx != -1 {
		return strings.TrimSpace(description[:idx])
	}

	// Otherwise return the whole description
	return description
}

// NewRoom creates a new Room with generated ID
func NewRoom(title, description string, exits []string) *Room {
	firstSentence := extractFirstSentence(description)
	id := GenerateRoomID(title, description, exits)

	room := &Room{
		ID:            id,
		Title:         title,
		Description:   description,
		FirstSentence: firstSentence,
		Exits:         make(map[string]string),
		VisitCount:    1,
	}

	// Initialize exits with unknown destinations
	for _, direction := range exits {
		room.Exits[direction] = ""
	}

	return room
}

// GetSearchText returns the text used for searching/matching this room
func (r *Room) GetSearchText() string {
	exitNames := make([]string, 0, len(r.Exits))
	for direction := range r.Exits {
		exitNames = append(exitNames, direction)
	}
	sort.Strings(exitNames)

	return strings.ToLower(r.Title + " " + r.FirstSentence + " " + strings.Join(exitNames, " "))
}

// MatchesSearch checks if all query terms are present in the room's search text
func (r *Room) MatchesSearch(queryTerms []string) bool {
	searchText := r.GetSearchText()

	for _, term := range queryTerms {
		if !strings.Contains(searchText, strings.ToLower(term)) {
			return false
		}
	}

	return true
}

// UpdateExit sets the destination for a given exit direction
func (r *Room) UpdateExit(direction, destinationID string) {
	r.Exits[direction] = destinationID
}
