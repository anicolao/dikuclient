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

	info := ParseRoomInfo(lines, false)
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

func TestParseRoomInfo_RealMUDOutput(t *testing.T) {
	// Test with actual MUD output format from Arctic MUD
	lines := []string{
		"tunnel. The passageway is clear of significant debris and has likely been",
		"cleared by recent rainfall.",
		"\x1b[33m\x1b[1mThe corpse of the slimy earthworm is lying here.",
		"\x1b[0;37m\x1b[31m\x1b[1m\x1b[0;37m",
		"\x1b[32m\x1b[0;37m\x1b[32m119H \x1b[0;37m\x1b[32m110V \x1b[0;37m3674X 0.00% 77C \x1b[0;37mT:56 Exits:EW>",
		"You nearly retch as the powerful odor reaches your nose.",
		"The Solace Dump",
		"    This is the Solace Dump. It is basically a huge sinkhole which is",
		"filled with trash and other odds and ends. The smell is quite horrible. The",
		"thick foliage around the edge of the dump is effective in keeping the smell",
		"and sight of the dump isolated from the rest of Solace. A small path leads",
		"north. A large sewer grate, which is perpetually open, allows access to the",
		"sewers below.",
		"A long-tailed ground dove sits here on a low branch.",
		"A small child plays with some colorful objects.",
		"A small child points a thin wooden stick at you.",
		"You feel righteous.",
		"",
		"119H 108V 3674X 0.00% 77C T:34 Exits:ND>",
	}

	info := ParseRoomInfo(lines, true) // Enable debug for this test
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	// Print debug info
	if info.DebugInfo != "" {
		t.Logf("Debug Info:\n%s", info.DebugInfo)
	}

	// Should parse "The Solace Dump" as the title
	if info.Title != "The Solace Dump" {
		t.Errorf("Title = %q, want %q", info.Title, "The Solace Dump")
	}

	// Should have parsed exits
	if len(info.Exits) != 2 {
		t.Errorf("Got %d exits, want 2", len(info.Exits))
	}

	// Should have north and down exits
	expectedExits := map[string]bool{"north": true, "down": true}
	for _, exit := range info.Exits {
		if !expectedExits[exit] {
			t.Errorf("Unexpected exit: %q", exit)
		}
	}

	// Description should not be empty
	if info.Description == "" {
		t.Error("Description is empty")
	}

	t.Logf("Parsed room: %q", info.Title)
	t.Logf("Description: %q", info.Description)
	t.Logf("Exits: %v", info.Exits)
}
