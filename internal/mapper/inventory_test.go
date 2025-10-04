package mapper

import (
	"reflect"
	"testing"
)

func TestParseInventoryInfo(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		expectedItems []string
	}{
		{
			name: "basic inventory from problem statement",
			lines: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
				"You are carrying:",
				"a sharp short sword",
				"a glowing scroll of recall..it glows dimly",
				"a torch [4]",
				"a rusty knife",
				"a bowl of Otik's spiced potatoes [2]",
				"an entire loaf of bread [4]",
				"",
				"86H 109V 7563X 0.00% 79C T:2 Exits:D>",
			},
			expectedItems: []string{
				"a sharp short sword",
				"a glowing scroll of recall..it glows dimly",
				"a torch [4]",
				"a rusty knife",
				"a bowl of Otik's spiced potatoes [2]",
				"an entire loaf of bread [4]",
			},
		},
		{
			name: "empty inventory",
			lines: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
				"You are carrying:",
				"",
				"86H 109V 7563X 0.00% 79C T:2 Exits:D>",
			},
			expectedItems: []string{},
		},
		{
			name: "inventory with ANSI codes",
			lines: []string{
				"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m109V \x1b[0;37m7563X 0.00% 79C \x1b[0;37mT:3 Exits:D>",
				"You are carrying:",
				"\x1b[33m\x1b[1ma sharp short sword",
				"a torch [4]",
				"",
				"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m109V \x1b[0;37m7563X 0.00% 79C \x1b[0;37mT:2 Exits:D>",
			},
			expectedItems: []string{
				"a sharp short sword",
				"a torch [4]",
			},
		},
		{
			name: "no inventory header",
			lines: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> look",
				"The Temple Square",
				"You are standing in the temple square.",
				"86H 109V 7563X 0.00% 79C T:2 Exits:NESW>",
			},
			expectedItems: nil,
		},
		{
			name: "incomplete inventory (no closing prompt)",
			lines: []string{
				"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
				"You are carrying:",
				"a sharp short sword",
			},
			expectedItems: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInventoryInfo(tt.lines, false)

			if tt.expectedItems == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Expected result, got nil")
			}

			if !reflect.DeepEqual(result.Items, tt.expectedItems) {
				t.Errorf("Items mismatch.\nExpected: %v\nGot: %v", tt.expectedItems, result.Items)
			}
		})
	}
}
