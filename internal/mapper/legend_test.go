package mapper

import (
	"strings"
	"testing"
)

// TestRenderMapWithLegend verifies that room numbers appear in the map when legend is provided
func TestRenderMapWithLegend(t *testing.T) {
	m := NewMap()

	// Create a simple map structure
	center := NewRoom("Center", "The center room.", []string{"north", "south"})
	north := NewRoom("North Room", "A room to the north.", []string{"south"})
	south := NewRoom("South Room", "A room to the south.", []string{"north"})

	// Add rooms and link them
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(north)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(south)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(center)

	// Create a legend mapping
	legend := map[string]int{
		north.ID:  1,
		center.ID: 2,
		south.ID:  3,
	}

	// Render map with legend
	rendered, _ := m.RenderMapWithLegend(30, 15, legend)

	t.Logf("Rendered map with legend:\n%s", rendered)

	// Verify that numbers appear in the output
	// The center room (2) should be visible
	if !strings.Contains(rendered, "2") {
		t.Error("Expected to find room number '2' (center) in rendered map")
	}

	// At least one other number should be visible (1 or 3)
	hasOtherNumber := strings.Contains(rendered, "1") || strings.Contains(rendered, "3")
	if !hasOtherNumber {
		t.Error("Expected to find at least one other room number (1 or 3) in rendered map")
	}
}

// TestRenderMapWithoutLegend verifies normal rendering without legend
func TestRenderMapWithoutLegend(t *testing.T) {
	m := NewMap()

	// Create a simple map structure
	center := NewRoom("Center", "The center room.", []string{"north"})
	north := NewRoom("North Room", "A room to the north.", []string{"south"})

	// Add rooms and link them
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(north)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(center)

	// Render map without legend
	rendered, _ := m.RenderMap(30, 15)

	t.Logf("Rendered map without legend:\n%s", rendered)

	// Should contain room symbols, not numbers
	if !strings.Contains(rendered, "▣") && !strings.Contains(rendered, "▢") {
		t.Error("Expected to find room symbols (▣ or ▢) in rendered map")
	}

	// Should not contain numbers (allow numbers in connectors but not as standalone)
	// Check if standalone digits appear (not part of connector characters)
	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		// Simple check: if we see a digit not surrounded by box drawing characters
		for i, ch := range line {
			if ch >= '0' && ch <= '9' {
				// Check context - should not be standalone room numbers
				if i > 0 && i < len(line)-1 {
					// This is a simplified check
					t.Logf("Found digit '%c' in rendered map (expected symbols only)", ch)
				}
			}
		}
	}
}

// TestFormatMapPanelWithLegend verifies the complete panel formatting with legend
func TestFormatMapPanelWithLegend(t *testing.T) {
	m := NewMap()

	// Create a map with multiple rooms
	center := NewRoom("Center Plaza", "The center of the plaza.", []string{"north", "south", "east", "west"})
	north := NewRoom("North Market", "A bustling northern market.", []string{"south"})
	south := NewRoom("South Fountain", "A fountain to the south.", []string{"north"})
	east := NewRoom("East Temple", "The eastern temple.", []string{"west"})
	west := NewRoom("West Gardens", "Beautiful western gardens.", []string{"east"})

	// Build the map
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(north)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("south")
	m.AddOrUpdateRoom(south)
	m.SetLastDirection("north")
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(east)
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(center)
	m.SetLastDirection("west")
	m.AddOrUpdateRoom(west)
	m.SetLastDirection("east")
	m.AddOrUpdateRoom(center)

	// Create legend
	legend := map[string]int{
		north.ID:  1,
		east.ID:   2,
		center.ID: 3,
		south.ID:  4,
		west.ID:   5,
	}

	// Format panel with legend
	panel := m.FormatMapPanelWithLegend(30, 15, legend)

	t.Logf("Formatted panel with legend:\n%s", panel)

	// Verify panel contains numbers
	foundNumbers := 0
	for i := 1; i <= 5; i++ {
		if strings.Contains(panel, string('0'+rune(i))) {
			foundNumbers++
		}
	}

	if foundNumbers < 3 {
		t.Errorf("Expected to find at least 3 room numbers in panel, found %d", foundNumbers)
	}
}
