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

	var title string
	var descriptionLines []string
	var exits []string
	var debugInfo strings.Builder

	debugInfo.WriteString("[MAPPER DEBUG] Attempting to parse room from lines:\n")
	for i, line := range lines {
		if i < 5 { // Only show first 5 lines in debug
			debugInfo.WriteString(fmt.Sprintf("  Line %d: %q\n", i, line))
		}
	}

	// First non-empty line is usually the title
	foundTitle := false
	for i, line := range lines {
		line = stripANSI(line)
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Check if this is an exits line
		if parsedExits := parseExitsLine(line); len(parsedExits) > 0 {
			exits = parsedExits
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found exits line: %q -> %v\n", line, parsedExits))
			// If we haven't found a title yet and have description lines, 
			// the first description line might be the title
			if !foundTitle && len(descriptionLines) > 0 {
				title = descriptionLines[0]
				descriptionLines = descriptionLines[1:]
			}
			break
		}

		if !foundTitle {
			// First non-exit line is the title
			title = line
			foundTitle = true
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found title: %q\n", title))
		} else {
			// Subsequent lines are description (until we hit exits)
			descriptionLines = append(descriptionLines, line)
			// Stop collecting after a reasonable number of lines
			if i > 20 {
				break
			}
		}
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
