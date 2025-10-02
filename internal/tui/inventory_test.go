package tui

import (
	"testing"
	"time"
)

func TestDetectAndUpdateInventory(t *testing.T) {
	tests := []struct {
		name          string
		recentOutput  []string
		expectedItems []string
	}{
		{
			name: "basic inventory detection",
			recentOutput: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
				"You are carrying:",
				"a sharp short sword",
				"a glowing scroll of recall..it glows dimly",
				"a torch [4]",
				"a rusty knife",
				"",
				"86H 109V 7563X 0.00% 79C T:2 Exits:D>",
			},
			expectedItems: []string{
				"a sharp short sword",
				"a glowing scroll of recall..it glows dimly",
				"a torch [4]",
				"a rusty knife",
			},
		},
		{
			name: "empty inventory",
			recentOutput: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
				"You are carrying:",
				"",
				"86H 109V 7563X 0.00% 79C T:2 Exits:D>",
			},
			expectedItems: []string{},
		},
		{
			name: "no inventory in output",
			recentOutput: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> look",
				"The Temple Square",
				"You are standing in the temple square.",
				"86H 109V 7563X 0.00% 79C T:2 Exits:NESW>",
			},
			expectedItems: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				recentOutput: tt.recentOutput,
			}

			// Store the initial time
			initialTime := m.inventoryTime

			m.detectAndUpdateInventory()

			if tt.expectedItems == nil {
				// Inventory should not be updated
				if m.inventoryTime != initialTime && !m.inventoryTime.IsZero() {
					t.Errorf("Inventory time was updated when it shouldn't have been")
				}
				return
			}

			// Check that inventory was updated
			if len(m.inventory) != len(tt.expectedItems) {
				t.Errorf("Expected %d items, got %d", len(tt.expectedItems), len(m.inventory))
			}

			for i, expected := range tt.expectedItems {
				if i >= len(m.inventory) {
					break
				}
				if m.inventory[i] != expected {
					t.Errorf("Item %d: expected %q, got %q", i, expected, m.inventory[i])
				}
			}

			// Check that timestamp was updated (should be recent)
			if m.inventoryTime.IsZero() {
				t.Error("Inventory time was not set")
			} else if time.Since(m.inventoryTime) > time.Second {
				t.Error("Inventory time is not recent")
			}
		})
	}
}
