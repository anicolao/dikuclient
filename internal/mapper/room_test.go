package mapper

import (
	"testing"
)

func TestGenerateRoomID(t *testing.T) {
	title := "Temple Square"
	description := "You are standing in a large square. The temple looms before you."
	exits := []string{"north", "south", "east"}

	id1 := GenerateRoomID(title, description, exits)
	id2 := GenerateRoomID(title, description, exits)

	if id1 != id2 {
		t.Errorf("Same room should generate same ID")
	}

	// Different exit order should still generate same ID
	exits2 := []string{"east", "north", "south"}
	id3 := GenerateRoomID(title, description, exits2)
	if id1 != id3 {
		t.Errorf("Exit order shouldn't affect ID")
	}

	// Different room should generate different ID
	id4 := GenerateRoomID("Different Room", description, exits)
	if id1 == id4 {
		t.Errorf("Different rooms should have different IDs")
	}
}

func TestExtractFirstSentence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"This is a sentence. This is another.", "This is a sentence."},
		{"Question? Answer.", "Question?"},
		{"Exclamation! More text.", "Exclamation!"},
		{"Single line without terminator", "Single line without terminator"},
		{"First line\nSecond line", "First line"},
	}

	for _, test := range tests {
		result := extractFirstSentence(test.input)
		if result != test.expected {
			t.Errorf("extractFirstSentence(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestRoomMatchesSearch(t *testing.T) {
	room := NewRoom("Temple Square", "You are in a large temple square.", []string{"north", "south"})

	tests := []struct {
		query    []string
		expected bool
	}{
		{[]string{"temple"}, true},
		{[]string{"square"}, true},
		{[]string{"temple", "square"}, true},
		{[]string{"north"}, true},
		{[]string{"temple", "north"}, true},
		{[]string{"missing"}, false},
		{[]string{"temple", "missing"}, false},
	}

	for _, test := range tests {
		result := room.MatchesSearch(test.query)
		if result != test.expected {
			t.Errorf("MatchesSearch(%v) = %v, want %v", test.query, result, test.expected)
		}
	}
}
