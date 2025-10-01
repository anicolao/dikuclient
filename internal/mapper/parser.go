package mapper

import (
	"fmt"
	"regexp"
	"strings"
)

// RoomInfo contains parsed room information
type RoomInfo struct {
	Title       string
	Description string
	Exits       []string
	DebugInfo   string // Debug information about parsing
}

// exitPatterns are common patterns for exit lines in MUDs
var exitPatterns = []*regexp.Regexp{
	// "Exits: north, south, east"
	regexp.MustCompile(`(?i)^exits?\s*:\s*(.+)$`),
	// "[ Exits: n s e w ]"
	regexp.MustCompile(`(?i)^\[\s*exits?\s*:\s*(.+?)\s*\]$`),
	// "Obvious exits: north, south"
	regexp.MustCompile(`(?i)^obvious\s+exits?\s*:\s*(.+)$`),
	// "Exits:EW>" or "Exits:NESW>" (compact format, no spaces)
	regexp.MustCompile(`(?i)exits?\s*:\s*([neswd]+)\s*>`),
}

// directionAliases maps short direction names to full names
var directionAliases = map[string]string{
	"n":  "north",
	"s":  "south",
	"e":  "east",
	"w":  "west",
	"u":  "up",
	"d":  "down",
	"ne": "northeast",
	"nw": "northwest",
	"se": "southeast",
	"sw": "southwest",
}

// ParseRoomInfo attempts to parse room information from MUD output
// It looks for a title line, description, and exits line
func ParseRoomInfo(lines []string) *RoomInfo {
	if len(lines) == 0 {
		return nil
	}

	var debugInfo strings.Builder
	debugInfo.WriteString("[MAPPER DEBUG] Attempting to parse room from lines:\n")
	for i := len(lines) - 1; i >= 0 && i >= len(lines)-10; i-- {
		debugInfo.WriteString(fmt.Sprintf("  Line %d: %q\n", i, lines[i]))
	}

	// Find the exits line first by scanning backwards
	exitsLineIdx := -1
	var exits []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)
		
		if parsedExits := parseExitsLine(line); len(parsedExits) > 0 {
			exits = parsedExits
			exitsLineIdx = i
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found exits line at index %d: %q -> %v\n", i, line, parsedExits))
			break
		}
	}

	// If no exits line found, we can't parse the room
	if exitsLineIdx == -1 {
		debugInfo.WriteString("[MAPPER DEBUG] No exits line found\n")
		return &RoomInfo{
			DebugInfo: debugInfo.String(),
		}
	}

	// Now look for the contiguous block of room text just before the exits line
	// We'll collect non-empty, non-status lines working backwards from just before exits
	var roomLines []string
	startSearchIdx := exitsLineIdx - 1
	
	// Skip backwards over empty lines and status/mob lines to find the end of room description
	for startSearchIdx >= 0 {
		line := stripANSI(lines[startSearchIdx])
		line = strings.TrimSpace(line)
		
		// If it's empty or a status/mob line, skip it
		if line == "" || isStatusOrCombatLine(line) {
			startSearchIdx--
			continue
		}
		
		// Found a non-status line, this is likely the end of room description
		break
	}
	
	// Now collect the contiguous block of room text going backwards
	emptyLineCount := 0
	for i := startSearchIdx; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)
		
		// If we hit another exits line, stop (we've gone too far back)
		if parsedExits := parseExitsLine(line); len(parsedExits) > 0 {
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Hit previous exits line at index %d, stopping\n", i))
			break
		}
		
		// Count empty lines - if we hit 2+ consecutive empty lines, stop
		if line == "" {
			emptyLineCount++
			if emptyLineCount >= 2 {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Hit multiple empty lines at index %d, stopping\n", i))
				break
			}
			continue
		}
		
		// Reset empty line count when we see content
		emptyLineCount = 0
		
		// Add this line to the beginning of our collection
		roomLines = append([]string{line}, roomLines...)
		
		// Don't go back more than 15 lines from where we started
		if startSearchIdx - i > 15 {
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Reached 15 line limit at index %d\n", i))
			break
		}
	}

	debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Collected %d room lines\n", len(roomLines)))

	// The first line that looks like a room title is the title, the rest is description
	var title string
	var descriptionLines []string
	
	if len(roomLines) > 0 {
		// Find the first line that looks like a room title
		titleFound := false
		for i, line := range roomLines {
			if isRoomTitle(line) {
				title = line
				// Everything after the title is description
				if i+1 < len(roomLines) {
					descriptionLines = roomLines[i+1:]
				}
				titleFound = true
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found title at line %d: %q\n", i, title))
				break
			}
		}
		
		// If no clear title found, use the first line as title
		if !titleFound && len(roomLines) > 0 {
			title = roomLines[0]
			if len(roomLines) > 1 {
				descriptionLines = roomLines[1:]
			}
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Using first line as title: %q\n", title))
		}
		
		debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Description lines: %d\n", len(descriptionLines)))
	}

	// If we found title and exits, return the room info
	if title != "" && len(exits) > 0 {
		description := strings.Join(descriptionLines, " ")
		debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Successfully parsed room: %q with exits %v\n", title, exits))
		return &RoomInfo{
			Title:       title,
			Description: description,
			Exits:       exits,
			DebugInfo:   debugInfo.String(),
		}
	}

	debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Failed to parse room (title=%q, exits=%v)\n", title, exits))
	return &RoomInfo{
		DebugInfo: debugInfo.String(),
	}
}

// isStatusOrCombatLine checks if a line is a status update or combat message
// These lines appear between rooms and should be skipped
func isStatusOrCombatLine(line string) bool {
	line = strings.ToLower(line)
	
	// Skip lines that are clearly status/combat/mob descriptions
	skipPatterns := []string{
		"you feel",
		"you are affected",
		"you nearly",
		"you retch",
		"points a",
		"is lying here",
		"sits here",
		"stands here",
		"plays with",
		"is here",
		"a small",
		"a large",
		"a long",
		"the corpse",
	}
	
	for _, pattern := range skipPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	
	return false
}

// isRoomTitle checks if a line looks like a room title
// Room titles are typically short, capitalize, and don't start with "you" or have verbs
func isRoomTitle(line string) bool {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return false
	}
	
	// Room titles don't usually start with these words
	lowerLine := strings.ToLower(line)
	badStarts := []string{"you ", "the corpse", "a small", "a large", "a long"}
	for _, start := range badStarts {
		if strings.HasPrefix(lowerLine, start) {
			return false
		}
	}
	
	// Room titles are typically 2-8 words
	words := strings.Fields(line)
	if len(words) < 2 || len(words) > 8 {
		return false
	}
	
	// Room titles typically start with a capital letter
	if line[0] < 'A' || line[0] > 'Z' {
		return false
	}
	
	return true
}

// parseExitsLine extracts exit directions from an exits line
func parseExitsLine(line string) []string {
	// Try each pattern
	for _, pattern := range exitPatterns {
		if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
			return parseExitsList(matches[1])
		}
	}
	return nil
}

// parseExitsList parses a comma/space separated list of exits
func parseExitsList(exitText string) []string {
	exitText = strings.TrimSpace(exitText)
	
	// Check if it's compact format (no spaces, just letters like "EW" or "NESW")
	if len(exitText) > 0 && !strings.Contains(exitText, " ") && !strings.Contains(exitText, ",") {
		// Split each character as a direction
		var exits []string
		for _, ch := range exitText {
			dir := strings.ToLower(string(ch))
			if isValidDirection(dir) {
				// Expand alias to full direction name
				if fullDir, ok := directionAliases[dir]; ok {
					exits = append(exits, fullDir)
				} else {
					exits = append(exits, dir)
				}
			}
		}
		return exits
	}
	
	// Replace commas with spaces for uniform splitting
	exitText = strings.ReplaceAll(exitText, ",", " ")
	
	// Split on whitespace
	words := strings.Fields(exitText)
	
	var exits []string
	for _, word := range words {
		word = strings.ToLower(word)
		// Remove common noise words
		if word == "and" || word == "or" || word == "none" {
			continue
		}
		
		// Expand aliases
		if fullDir, ok := directionAliases[word]; ok {
			word = fullDir
		}
		
		// Only keep known directions
		if isValidDirection(word) {
			exits = append(exits, word)
		}
	}
	
	return exits
}

// isValidDirection checks if a string is a valid direction
func isValidDirection(dir string) bool {
	validDirections := map[string]bool{
		"north": true, "south": true, "east": true, "west": true,
		"up": true, "down": true,
		"northeast": true, "northwest": true, "southeast": true, "southwest": true,
		"ne": true, "nw": true, "se": true, "sw": true,
		"n": true, "s": true, "e": true, "w": true, "u": true, "d": true,
	}
	return validDirections[strings.ToLower(dir)]
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(str, "")
}

// DetectMovement checks if a line represents a movement command
func DetectMovement(line string) string {
	line = strings.TrimSpace(strings.ToLower(line))
	
	// Check for full direction names
	if isValidDirection(line) {
		// Expand aliases
		if fullDir, ok := directionAliases[line]; ok {
			return fullDir
		}
		return line
	}
	
	return ""
}
