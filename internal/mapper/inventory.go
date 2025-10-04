package mapper

import (
	"regexp"
	"strings"
)

// InventoryInfo contains parsed inventory information
type InventoryInfo struct {
	Items     []string
	DebugInfo string // Debug information about parsing
}

// inventoryHeaderPattern matches "You are carrying:"
var inventoryHeaderPattern = regexp.MustCompile(`(?i)^you are carrying:\s*$`)

// ParseInventoryInfo attempts to parse inventory information from MUD output
// It looks for "You are carrying:" followed by item lines
func ParseInventoryInfo(lines []string, enableDebug bool) *InventoryInfo {
	if len(lines) == 0 {
		return nil
	}

	// Find the inventory header line by scanning backwards
	headerIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		if inventoryHeaderPattern.MatchString(line) {
			headerIdx = i
			break
		}
	}

	// If no header found, we can't parse the inventory
	if headerIdx == -1 {
		return nil
	}

	// Look for the prompt line after the header to know where inventory ends
	// Prompts typically end with > and contain stats (H, V, X, etc.)
	promptIdx := -1
	for i := headerIdx + 1; i < len(lines); i++ {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		if isPromptLine(line) {
			promptIdx = i
			break
		}
	}

	// If no prompt found after header, inventory may still be incomplete
	if promptIdx == -1 {
		return nil
	}

	// Collect all lines between header and prompt as inventory items
	items := []string{}
	for i := headerIdx + 1; i < promptIdx; i++ {
		line := stripANSI(lines[i])
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		items = append(items, line)
	}

	return &InventoryInfo{
		Items: items,
	}
}
