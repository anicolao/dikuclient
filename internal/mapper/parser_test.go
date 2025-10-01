package mapper

import (
	"testing"
)

func TestParseExitsLine(t *testing.T) {
	tests := []struct {
		line     string
		expected []string
	}{
		{"Exits: north, south, east", []string{"north", "south", "east"}},
		{"[ Exits: n s e w ]", []string{"north", "south", "east", "west"}},
		{"Obvious exits: north and south", []string{"north", "south"}},
		{"exits: up down", []string{"up", "down"}},
		{"Exits: ne, sw", []string{"northeast", "southwest"}},
		{"Not an exit line", nil},
		// New compact format tests
		{"Exits:EW>", []string{"east", "west"}},
		{"Exits:NESW>", []string{"north", "east", "south", "west"}},
		{"119H 131V 4923X 0.00% 60C T:60 Exits:EW> ", []string{"east", "west"}},
		{"Exits:N>", []string{"north"}},
	}

	for _, test := range tests {
		result := parseExitsLine(test.line)
		if len(result) != len(test.expected) {
			t.Errorf("parseExitsLine(%q) returned %d exits, want %d", test.line, len(result), len(test.expected))
			continue
		}
		for i, exit := range result {
			if exit != test.expected[i] {
				t.Errorf("parseExitsLine(%q)[%d] = %q, want %q", test.line, i, exit, test.expected[i])
			}
		}
	}
}

func TestParseRoomInfo(t *testing.T) {
	lines := []string{
		"Temple Square",
		"You are standing in a large temple square. The ancient stones",
		"speak of a glorious past.",
		"Exits: north, south, east",
	}

	info := ParseRoomInfo(lines)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	if info.Title != "Temple Square" {
		t.Errorf("Title = %q, want %q", info.Title, "Temple Square")
	}

	if len(info.Exits) != 3 {
		t.Errorf("Got %d exits, want 3", len(info.Exits))
	}

	expectedExits := map[string]bool{"north": true, "south": true, "east": true}
	for _, exit := range info.Exits {
		if !expectedExits[exit] {
			t.Errorf("Unexpected exit: %q", exit)
		}
	}
}

func TestDetectMovement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"north", "north"},
		{"n", "north"},
		{"south", "south"},
		{"s", "south"},
		{"  east  ", "east"},
		{"look", ""},
		{"inventory", ""},
	}

	for _, test := range tests {
		result := DetectMovement(test.input)
		if result != test.expected {
			t.Errorf("DetectMovement(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}
