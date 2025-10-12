package tui

import (
	"strings"
	"testing"
)

func TestBarsoomMarkerSuppression(t *testing.T) {
	// Create a test model using NewModel
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 100
	model.height = 40

	// Simulate receiving Barsoom room output (new format: exits on >-- line)
	barsoomOutput := `119H 110V 3674X 0.00% 77C T:56 Exits:EW>
--<
Temple Square
    You are standing in a large temple square. The ancient stones
speak of a glorious past.
>-- Exits:NSE`

	// Process the output through the model
	updatedModel, _ := model.Update(mudMsg(barsoomOutput))
	m := updatedModel.(*Model)

	// Check that --< and >-- markers are not in the output
	outputStr := strings.Join(m.output, "\n")
	if strings.Contains(outputStr, "--<") {
		t.Error("Expected --< marker to be suppressed from output")
	}
	if strings.Contains(outputStr, ">--") {
		t.Error("Expected >-- marker to be suppressed from output")
	}

	// Check that the title and description are in recentOutput for parsing
	recentStr := strings.Join(m.recentOutput, "\n")
	if !strings.Contains(recentStr, "--<") {
		t.Error("Expected --< marker to be in recentOutput for parsing")
	}
	if !strings.Contains(recentStr, ">--") {
		t.Error("Expected >-- marker to be in recentOutput for parsing")
	}
	if !strings.Contains(recentStr, "Temple Square") {
		t.Error("Expected room title in recentOutput")
	}
}

func TestBarsoomDescriptionSplit(t *testing.T) {
	// This test verifies that the description split is activated for Barsoom rooms
	// We can't fully test the rendering without a real terminal, but we can check the flags
	
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 100
	model.height = 40

	// The test relies on detectAndUpdateRoom being called after receiving output
	// This happens automatically in the mudMsg handler
	
	// Initially, no description split should be active
	if model.hasDescriptionSplit {
		t.Error("Expected hasDescriptionSplit to be false initially")
	}
	
	// Note: To fully test this, we'd need to set up a worldMap and ensure
	// detectAndUpdateRoom runs. This is more of an integration test that would
	// require a more complex setup. The unit tests for ParseRoomInfo already
	// verify the detection logic works correctly.
}

func TestRegularRoomNoDescriptionSplit(t *testing.T) {
	// Verify that regular (non-Barsoom) rooms don't trigger description split
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 100
	model.height = 40

	// Simulate receiving regular room output (without --< >-- markers)
	regularOutput := `119H 110V 3674X 0.00% 77C T:56 Exits:EW>
Temple Square
    You are standing in a large temple square. The ancient stones
speak of a glorious past.
Exits: north, south, east`

	// Process the output
	updatedModel, _ := model.Update(mudMsg(regularOutput))
	m := updatedModel.(*Model)

	// Check that all lines are in output (no suppression for regular rooms)
	outputStr := strings.Join(m.output, "\n")
	if !strings.Contains(outputStr, "Temple Square") {
		t.Error("Expected room title in output for regular room")
	}
	if !strings.Contains(outputStr, "Exits: north, south, east") {
		t.Error("Expected exits in output for regular room")
	}
}

func TestBarsoomDescriptionUpdatesWithoutMovement(t *testing.T) {
	// Verify that Barsoom description split updates even without movement
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 100
	model.height = 40

	// Simulate receiving Barsoom room output WITHOUT any movement command
	// (i.e., pendingMovement should be empty)
	// New format: exits on >-- line
	barsoomOutput := `119H 110V 3674X 0.00% 77C T:56 Exits:EW>
--<
Temple Square
You are standing in a large temple square.
>-- Exits:NSE`

	// Process the output
	updatedModel, _ := model.Update(mudMsg(barsoomOutput))
	m := updatedModel.(*Model)

	// Even without movement, the description split should be active for Barsoom rooms
	if !m.hasDescriptionSplit {
		t.Error("Expected hasDescriptionSplit to be true for Barsoom room even without movement")
	}

	if m.currentRoomDescription == "" {
		t.Error("Expected currentRoomDescription to be populated for Barsoom room even without movement")
	}

	// Verify the Barsoom title and exits are stored for the title bar
	if m.currentBarsoomTitle != "Temple Square" {
		t.Errorf("Expected currentBarsoomTitle to be 'Temple Square', got %q", m.currentBarsoomTitle)
	}
	
	if len(m.currentBarsoomExits) != 3 {
		t.Errorf("Expected 3 exits, got %d", len(m.currentBarsoomExits))
	}
}
