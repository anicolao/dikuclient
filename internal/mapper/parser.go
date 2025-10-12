package mapper

import (
	"fmt"
	"regexp"
	"strings"
)

// RoomInfo contains parsed room information
type RoomInfo struct {
	Title          string
	Description    string
	Exits          []string
	DebugInfo      string // Debug information about parsing
	IsBarsoomRoom  bool   // Whether this is a Barsoom format room (with --< >-- markers)
	BarsoomStartIdx int   // Index of --< line in original lines (for suppression)
	BarsoomEndIdx   int   // Index of >-- line in original lines (for suppression)
}

// exitPatterns are common patterns for exit lines in MUDs
var exitPatterns = []*regexp.Regexp{
	// "Exits: north, south, east"
	regexp.MustCompile(`(?i)^exits?\s*:\s*(.+)$`),
	// "[ Exits: n s e w ]"
	regexp.MustCompile(`(?i)^\[\s*exits?\s*:\s*(.+?)\s*\]$`),
	// "Obvious exits: north, south"
	regexp.MustCompile(`(?i)^obvious\s+exits?\s*:\s*(.+)$`),
	// "Exits:EW>" or "Exits:NESW>" or "Exits:UD>" or "Exits:N(S)E>" (compact format, closed doors in parentheses)
	regexp.MustCompile(`(?i)exits?\s*:\s*([neswud()]+)\s*>`),
}

// directionAliases maps short direction names to full names
var directionAliases = map[string]string{
	"n": "north",
	"s": "south",
	"e": "east",
	"w": "west",
	"u": "up",
	"d": "down",
}

// parseBarsoomRoom attempts to parse a Barsoom MUD room format
// Barsoom format: --< on a line, title on next line, description paragraphs, then >-- Exits:... on a line
// The exits are now always on the same line as the >-- marker
// Searches backwards from the end to find the most recent complete room
func parseBarsoomRoom(lines []string, enableDebug bool, debugInfo *strings.Builder) *RoomInfo {
	// Search backwards from the end to find the most recent complete room
	// Keep track of the end marker and start marker
	endMarkerIdx := -1
	startMarkerIdx := -1
	var exits []string

	// First, find the end marker >-- (exits are on the same line)
	for i := len(lines) - 1; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		// Find the end marker >-- (may have exits on the same line)
		if strings.HasPrefix(line, ">--") {
			endMarkerIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found Barsoom end marker at index %d: %q\n", i, line))
			}
			
			// Parse exits from the same line (format: ">-- Exits:NSD" or just ">--")
			if len(line) > 3 {
				// Remove the ">--" prefix and parse the rest
				exitsPart := strings.TrimSpace(line[3:])
				if parsedExits := parseExitsLine(exitsPart); len(parsedExits) > 0 {
					exits = parsedExits
					if enableDebug {
						debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found exits on end marker line: %v\n", exits))
					}
				}
			}
			break
		}
	}

	// If no end marker found, not a Barsoom room
	if endMarkerIdx == -1 {
		return nil
	}

	// Now search backwards from the end marker to find the start marker
	for i := endMarkerIdx - 1; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		// Find the start marker --<
		if line == "--<" {
			startMarkerIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found Barsoom start marker at index %d\n", i))
			}
			break // Found a complete room, stop searching
		}
	}

	// Validate we found a complete Barsoom room
	if startMarkerIdx == -1 {
		return nil // Not a complete Barsoom format room
	}

	// Title is the first non-empty line after --<
	var title string
	titleIdx := -1
	for i := startMarkerIdx + 1; i < endMarkerIdx; i++ {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)
		if line != "" {
			title = line
			titleIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found Barsoom title at index %d: %q\n", i, title))
			}
			break
		}
	}

	if title == "" {
		return nil // No title found
	}

	// Collect description lines from after title to end marker
	var descriptionLines []string
	for i := titleIdx + 1; i < endMarkerIdx; i++ {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)
		if line != "" {
			descriptionLines = append(descriptionLines, line)
		}
	}

	description := strings.Join(descriptionLines, " ")
	if enableDebug {
		debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Successfully parsed Barsoom room: %q with exits %v\n", title, exits))
	}

	return &RoomInfo{
		Title:           title,
		Description:     description,
		Exits:           exits,
		DebugInfo:       debugInfo.String(),
		IsBarsoomRoom:   true,
		BarsoomStartIdx: startMarkerIdx,
		BarsoomEndIdx:   endMarkerIdx,
	}
}

// ParseBarsoomRoomOnly attempts to parse only Barsoom MUD room format
// This is called on every output to update the description split
func ParseBarsoomRoomOnly(lines []string, enableDebug bool) *RoomInfo {
	if len(lines) == 0 {
		return nil
	}

	var debugInfo strings.Builder
	if enableDebug {
		debugInfo.WriteString("[MAPPER DEBUG] Checking for Barsoom room format:\n")
	}

	return parseBarsoomRoom(lines, enableDebug, &debugInfo)
}

// ParseRoomInfo attempts to parse room information from MUD output
// It looks for a title line, description, and exits line
// New heuristic: search backwards for previous prompt, then forwards for first indented line
// Also supports Barsoom MUD format with --< and >-- markers
func ParseRoomInfo(lines []string, enableDebug bool) *RoomInfo {
	if len(lines) == 0 {
		return nil
	}

	var debugInfo strings.Builder
	if enableDebug {
		debugInfo.WriteString("[MAPPER DEBUG] Attempting to parse room from lines:\n")
		for i := len(lines) - 1; i >= 0 && i >= len(lines)-10; i-- {
			debugInfo.WriteString(fmt.Sprintf("  Line %d: %q\n", i, lines[i]))
		}
	}

	// Check for Barsoom MUD format (--< ... >--)
	if barsoomInfo := parseBarsoomRoom(lines, enableDebug, &debugInfo); barsoomInfo != nil {
		return barsoomInfo
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
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found exits line at index %d: %q -> %v\n", i, line, parsedExits))
			}
			break
		}
	}

	// If no exits line found, we can't parse the room
	if exitsLineIdx == -1 {
		if enableDebug {
			debugInfo.WriteString("[MAPPER DEBUG] No exits line found\n")
		}
		return &RoomInfo{
			DebugInfo: debugInfo.String(),
		}
	}

	// Search backwards from just before exits line to find previous prompt
	// We look for a prompt line first, as that's the most reliable boundary
	// Only fall back to previous exits line if no prompt is found
	previousPromptIdx := -1
	previousExitsIdx := -1
	for i := exitsLineIdx - 1; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		// A prompt line typically ends with > and contains stats (H, V, X, etc.)
		if isPromptLine(line) {
			previousPromptIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found previous prompt at index %d: %q\n", i, line))
			}
			break
		}

		// Track if we find another exits line (but don't stop immediately)
		if previousExitsIdx == -1 && parseExitsLine(line) != nil {
			previousExitsIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found previous exits line at index %d\n", i))
			}
		}

		// Don't search too far back
		if exitsLineIdx-i > 25 {
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Reached search limit at index %d\n", i))
			}
			break
		}
	}

	// Use the prompt as boundary if found, otherwise use previous exits line
	startSearchIdx := 0
	if previousPromptIdx >= 0 {
		startSearchIdx = previousPromptIdx + 1
		if enableDebug {
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Using prompt at %d as boundary\n", previousPromptIdx))
		}
	} else if previousExitsIdx >= 0 {
		startSearchIdx = previousExitsIdx + 1
		if enableDebug {
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Using previous exits at %d as boundary\n", previousExitsIdx))
		}
	}

	firstIndentedIdx := -1
	for i := startSearchIdx; i < exitsLineIdx; i++ {
		line := lines[i] // Don't strip ANSI yet - we need to check original indentation
		stripped := stripANSI(line)

		// Skip empty lines
		if strings.TrimSpace(stripped) == "" {
			continue
		}

		// Check if line is indented (starts with whitespace)
		if len(stripped) > 0 && (stripped[0] == ' ' || stripped[0] == '\t') {
			firstIndentedIdx = i
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found first indented line at index %d: %q\n", i, stripped))
			}
			break
		}
	}

	// The title is the line before the first indented line
	var title string
	var descriptionStartIdx int

	if firstIndentedIdx > startSearchIdx {
		// Found indented line, so title is the line before it
		for i := firstIndentedIdx - 1; i >= startSearchIdx; i-- {
			line := stripANSI(lines[i])
			line = strings.TrimSpace(line)

			// Skip empty lines
			if line == "" {
				continue
			}

			// This is the title
			title = line
			descriptionStartIdx = firstIndentedIdx
			if enableDebug {
				debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Found title at index %d: %q\n", i, title))
			}
			break
		}
	}

	// If we didn't find a title using indentation, fail
	if title == "" {
		if enableDebug {
			debugInfo.WriteString("[MAPPER DEBUG] No indented line found, cannot parse room\n")
		}
		return &RoomInfo{
			DebugInfo: debugInfo.String(),
		}
	}

	// Collect description from descriptionStartIdx until we hit exits or status/mob lines
	var descriptionLines []string
	for i := descriptionStartIdx; i < exitsLineIdx; i++ {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Stop if we hit status/mob lines (these come after description)
		if isStatusOrCombatLine(line) {
			break
		}

		descriptionLines = append(descriptionLines, line)
	}

	if enableDebug {
		debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Collected %d description lines\n", len(descriptionLines)))
	}

	// If we found title and exits, return the room info
	if title != "" && len(exits) > 0 {
		description := strings.Join(descriptionLines, " ")
		if enableDebug {
			debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Successfully parsed room: %q with exits %v\n", title, exits))
		}
		return &RoomInfo{
			Title:       title,
			Description: description,
			Exits:       exits,
			DebugInfo:   debugInfo.String(),
		}
	}

	if enableDebug {
		debugInfo.WriteString(fmt.Sprintf("[MAPPER DEBUG] Failed to parse room (title=%q, exits=%v)\n", title, exits))
	}
	return &RoomInfo{
		DebugInfo: debugInfo.String(),
	}
}

// isPromptLine checks if a line looks like a MUD prompt
func isPromptLine(line string) bool {
	// Prompts typically end with > and contain stats like "119H 108V"
	if !strings.HasSuffix(line, ">") {
		return false
	}

	// Look for stat indicators (H for health, V for movement, etc.)
	return strings.Contains(line, "H ") && strings.Contains(line, "V ")
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

	// Check if it's compact format (no spaces, just letters like "EW" or "NESW" or "N(S)E")
	if len(exitText) > 0 && !strings.Contains(exitText, " ") && !strings.Contains(exitText, ",") {
		// Split each character as a direction, handling parentheses for closed doors
		var exits []string
		for _, ch := range exitText {
			// Skip parentheses - they indicate closed doors but we still want the exit
			if ch == '(' || ch == ')' {
				continue
			}

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
