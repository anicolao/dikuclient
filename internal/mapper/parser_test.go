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
		{"Not an exit line", nil},
		// New compact format tests
		{"Exits:EW>", []string{"east", "west"}},
		{"Exits:NESW>", []string{"north", "east", "south", "west"}},
		{"119H 131V 4923X 0.00% 60C T:60 Exits:EW> ", []string{"east", "west"}},
		{"Exits:N>", []string{"north"}},
		// Test with up and down
		{"Exits:UD>", []string{"up", "down"}},
		{"86H 81V 7886X 0.00% 37C T:40 Exits:UD>", []string{"up", "down"}},
		{"Exits:NESWUD>", []string{"north", "east", "south", "west", "up", "down"}},
		// Test with closed doors (parentheses)
		{"Exits:N(S)E>", []string{"north", "south", "east"}},
		{"Exits:N(SE)W>", []string{"north", "south", "east", "west"}},
		{"Exits:(N)S>", []string{"north", "south"}},
		{"Exits:N(S)(E)W>", []string{"north", "south", "east", "west"}},
		{"120H 100V 5000X 0.00% 50C T:30 Exits:N(S)E>", []string{"north", "south", "east"}},
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
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"Temple Square",
		"    You are standing in a large temple square. The ancient stones",
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

func TestParseRoomInfo_BarsoomFormat(t *testing.T) {
	// Test Barsoom MUD format with --< and >-- markers
	lines := []string{
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"--<",
		"Temple Square",
		"    You are standing in a large temple square. The ancient stones",
		"speak of a glorious past.",
		">--",
		"Exits: north, south, east",
	}

	info := ParseRoomInfo(lines, false)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	if info.Title != "Temple Square" {
		t.Errorf("Title = %q, want %q", info.Title, "Temple Square")
	}

	expectedDesc := "You are standing in a large temple square. The ancient stones speak of a glorious past."
	if info.Description != expectedDesc {
		t.Errorf("Description = %q, want %q", info.Description, expectedDesc)
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

	// Verify that Barsoom-specific fields are set
	if !info.IsBarsoomRoom {
		t.Error("Expected IsBarsoomRoom to be true")
	}

	if info.BarsoomStartIdx != 1 {
		t.Errorf("BarsoomStartIdx = %d, want 1", info.BarsoomStartIdx)
	}

	if info.BarsoomEndIdx != 5 {
		t.Errorf("BarsoomEndIdx = %d, want 5", info.BarsoomEndIdx)
	}
}

func TestParseRoomInfo_BarsoomFormatMultipleParagraphs(t *testing.T) {
	// Test Barsoom format with multiple description paragraphs
	lines := []string{
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"--<",
		"Ancient Library",
		"Towering shelves filled with ancient tomes line the walls of this grand library.",
		"The musty smell of old parchment fills the air.",
		"",
		"A large reading table sits in the center of the room, covered with open books.",
		">--",
		"Exits: west",
	}

	info := ParseRoomInfo(lines, false)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	if info.Title != "Ancient Library" {
		t.Errorf("Title = %q, want %q", info.Title, "Ancient Library")
	}

	expectedDesc := "Towering shelves filled with ancient tomes line the walls of this grand library. The musty smell of old parchment fills the air. A large reading table sits in the center of the room, covered with open books."
	if info.Description != expectedDesc {
		t.Errorf("Description = %q, want %q", info.Description, expectedDesc)
	}

	if !info.IsBarsoomRoom {
		t.Error("Expected IsBarsoomRoom to be true")
	}
}

func TestParseRoomInfo_BarsoomFormatMultipleRooms(t *testing.T) {
	// Test that backward search finds the most recent room when multiple rooms are present
	lines := []string{
		// Old room that should NOT be detected
		"119H 110V 3674X 0.00% 77C T:56 Exits:NS>",
		"--<",
		"Old Temple",
		"An ancient temple.",
		">--",
		"Exits: north, south",
		"",
		// Some movement output
		"You move north.",
		"",
		// New room that SHOULD be detected (most recent)
		"120H 110V 3600X 0.00% 77C T:57 Exits:EW>",
		"--<",
		"Temple Square",
		"You are standing in a large temple square. The ancient stones speak of a glorious past.",
		">--",
		"Exits: east, west",
	}

	info := ParseRoomInfo(lines, false)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	// Should find the most recent room (Temple Square), not the old one
	if info.Title != "Temple Square" {
		t.Errorf("Title = %q, want %q (should find most recent room)", info.Title, "Temple Square")
	}

	expectedDesc := "You are standing in a large temple square. The ancient stones speak of a glorious past."
	if info.Description != expectedDesc {
		t.Errorf("Description = %q, want %q", info.Description, expectedDesc)
	}

	expectedExits := map[string]bool{"east": true, "west": true}
	if len(info.Exits) != len(expectedExits) {
		t.Errorf("Got %d exits, want %d", len(info.Exits), len(expectedExits))
	}
	for _, exit := range info.Exits {
		if !expectedExits[exit] {
			t.Errorf("Unexpected exit: %q", exit)
		}
	}

	if !info.IsBarsoomRoom {
		t.Error("Expected IsBarsoomRoom to be true")
	}
}

func TestParseRoomInfo_BarsoomFormatNoExitsYet(t *testing.T) {
	// Test Barsoom format when exits haven't arrived yet
	// This simulates the case where we receive the room description but the exits line hasn't come through yet
	lines := []string{
		"--<",
		"Temple Square",
		"You are standing in a large temple square.",
		">--",
	}

	info := ParseRoomInfo(lines, false)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil - should still parse room even without exits")
	}

	if info.Title != "Temple Square" {
		t.Errorf("Title = %q, want %q", info.Title, "Temple Square")
	}

	expectedDesc := "You are standing in a large temple square."
	if info.Description != expectedDesc {
		t.Errorf("Description = %q, want %q", info.Description, expectedDesc)
	}

	// Exits may be empty if they haven't arrived yet - that's OK
	if len(info.Exits) > 0 {
		t.Logf("Exits found: %v (may be empty initially)", info.Exits)
	}

	if !info.IsBarsoomRoom {
		t.Error("Expected IsBarsoomRoom to be true")
	}
}

func TestParseRoomInfo_BarsoomFormatExitsAfterMarker(t *testing.T) {
	// Test that exits are correctly found AFTER the >-- marker, not before
	lines := []string{
		// Previous room's exit line (should be ignored)
		"5H 100F 82V 0C T:16 Exits:NSD>",
		"--<",
		"Temple Square",
		"You are standing in a large temple square.",
		">--",
		"A guard stands here.",
		"",
		"5H 100F 82V 0C T:15 Exits:NESW>", // This is the correct exits line
	}

	info := ParseRoomInfo(lines, false)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	if info.Title != "Temple Square" {
		t.Errorf("Title = %q, want %q", info.Title, "Temple Square")
	}

	// Should find exits from AFTER the >-- marker, not before
	expectedExits := map[string]bool{"north": true, "east": true, "south": true, "west": true}
	if len(info.Exits) != len(expectedExits) {
		t.Errorf("Got %d exits, want %d (should find exits after >--, not before)", len(info.Exits), len(expectedExits))
	}
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

func TestParseRoomInfo_ArcticWelcomeScreen(t *testing.T) {
	// Test with welcome screen - should NOT detect a room
	lines := []string{
		"Connected to mud.arctic.org:2700",
		"                               Welcome to",
		"     ___",
		"    (___)     @@@@@ @@@@@@@@@   @@@@@@@@ @@@@@@@@@@@ @@@@@@@ @@@@@@@@",
		"     ==_     @@@@@@ @@@   @@@@ @@@@  @@@@    @@@       @@@  @@@@  @@@@",
		"     =/./   @@@@@@@ @@@    @@@ @@@    @@@    @@@  @@@",
		"",
		"By what name do you wish to be known?",
	}

	info := ParseRoomInfo(lines, true) // Enable debug
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	// Print debug info
	if info.DebugInfo != "" {
		t.Logf("Debug Info:\n%s", info.DebugInfo)
	}

	// Should NOT parse welcome screen as a room (no exits)
	if info.Title != "" {
		t.Errorf("Unexpected title parsed from welcome screen: %q", info.Title)
	}
}

func TestParseRoomInfo_ShrineOfTheHolyPaladin(t *testing.T) {
	// Test with actual room from user's session
	lines := []string{
		"Shrine of the Holy Paladin",
		"    This shrine is sparse except for a single white marble statue depicting",
		"a knight with his sword arm raised in an action of beheading a dragon adorning",
		"the middle of the shrine. Small cushions are placed around the statue to provide",
		"comfort for those meditating. This shrine gives you both a sense of awe and",
		"serenity.",
		"A paladin stands here radiating holiness.",
		"",
		"119H 131V 4923X 0.00% 60C T:60 Exits:E>",
	}

	info := ParseRoomInfo(lines, true) // Enable debug
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	// Print debug info
	if info.DebugInfo != "" {
		t.Logf("Debug Info:\n%s", info.DebugInfo)
	}

	// Should parse "Shrine of the Holy Paladin" as the title
	if info.Title != "Shrine of the Holy Paladin" {
		t.Errorf("Title = %q, want %q", info.Title, "Shrine of the Holy Paladin")
	}

	// Should have parsed exits
	if len(info.Exits) != 1 {
		t.Errorf("Got %d exits, want 1", len(info.Exits))
	}

	// Should have east exit
	if len(info.Exits) > 0 && info.Exits[0] != "east" {
		t.Errorf("Exit = %q, want %q", info.Exits[0], "east")
	}

	// Description should not be empty
	if info.Description == "" {
		t.Error("Description is empty")
	}

	t.Logf("Parsed room: %q", info.Title)
	t.Logf("Description: %q", info.Description)
	t.Logf("Exits: %v", info.Exits)
}

// Test case from user feedback: Reception -> Inn of the Last Home with get/d commands
func TestParseRoomInfo_ReceptionToInn(t *testing.T) {
	// This is the exact output when moving down from Reception to Inn
	// The issue: parser was grabbing "The Reception" from earlier output instead of "The Inn of the Last Home"
	lines := []string{
		"\x1b[36mThe Reception\x1b[0;37m",
		"    The reception contains a small desk. A long hall leads behind the desk",
		"here and you can tell this is where adventurers rest and take time off from",
		"their journeys. A small stairway goes down towards what appears to be a",
		"lively tavern.",
		"\x1b[33m\x1b[1mA long sword has been left here.",
		"A gardening spade made of gold lies here in the dirt.",
		"A sharp stinger lies on the ground here.",
		"\x1b[0;37m\x1b[31m\x1b[1mA red-haired young lady is wiping down a table here.",
		"You see the outline of a life form here, carrying: a glowing scroll of recall; a glowing scroll of recall; a glowing scroll of recall.",
		"A beautiful dragonhorse flies here, ready to carry a passenger.",
		"A girl twirls her hair in the Solace Tourist Information booth.",
		"A man wearing heavy armor is standing here.",
		"A receptionist sits here doing her nails.",
		"\x1b[0;37m",
		"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m82V \x1b[0;37m7886X 0.00% 37C \x1b[0;37mT:46 Exits:D>",
		"",
		"Tika Waylan leaves down.",
		"",
		"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m82V \x1b[0;37m7886X 0.00% 37C \x1b[0;37mT:43 Exits:D>",
		"You get a long sword.",
		"You get a golden spade.",
		"You get the queen bee's stinger.",
		"",
		"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m82V \x1b[0;37m7886X 0.00% 37C \x1b[0;37mT:42 Exits:D>",
		"The wonderful smell of spiced potatoes makes your stomach rumble.",
		"\x1b[36mThe Inn of the Last Home\x1b[0;37m",
		"    The Inn of the Last Home is a cozy tavern. The smells of rich food and",
		"the sounds of merry music overflow your senses. A huge fireplace dominates",
		"one wall and a warm glow emanates from its flames. Small tables fill the",
		"room. A stairway leads up, further into the tree, while another stairway",
		"leads down to the forest below.",
		"\x1b[33m\x1b[1mA loaf of bread lies here. [5]",
		"\x1b[0;37m\x1b[31m\x1b[1mA red-haired young lady is wiping down a table here.",
		"A small child plays with some colorful objects.",
		"Otik watches you calmly, while he skillfully mixes a drink.",
		"\x1b[0;37m",
		"\x1b[32m\x1b[0;37m\x1b[33m\x1b[1m86H \x1b[0;37m\x1b[33m\x1b[1m81V \x1b[0;37m7886X 0.00% 37C \x1b[0;37mT:40 Exits:UD>",
	}

	info := ParseRoomInfo(lines, true)
	if info == nil {
		t.Fatal("ParseRoomInfo returned nil")
	}

	// Print debug info
	if info.DebugInfo != "" {
		t.Logf("Debug Info:\n%s", info.DebugInfo)
	}

	// Should parse "The Inn of the Last Home" as the title, NOT "The Reception"
	if info.Title != "The Inn of the Last Home" {
		t.Errorf("Title = %q, want %q", info.Title, "The Inn of the Last Home")
	}

	// Should have parsed exits (up and down)
	if len(info.Exits) != 2 {
		t.Errorf("Got %d exits, want 2", len(info.Exits))
	}

	// Should have up and down exits
	hasUp := false
	hasDown := false
	for _, exit := range info.Exits {
		if exit == "up" {
			hasUp = true
		}
		if exit == "down" {
			hasDown = true
		}
	}
	if !hasUp || !hasDown {
		t.Errorf("Exits = %v, want [up, down]", info.Exits)
	}

	// Description should contain "cozy tavern"
	if info.Description == "" {
		t.Error("Description is empty")
	}

	t.Logf("Parsed room: %q", info.Title)
	t.Logf("Description: %q", info.Description)
	t.Logf("Exits: %v", info.Exits)
}
