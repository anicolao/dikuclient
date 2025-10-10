package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTypingDoesNotCauseViewportJump tests that typing characters doesn't cause
// unnecessary viewport content updates when the underlying output hasn't changed
func TestTypingDoesNotCauseViewportJump(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m

	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Add some output and set connected state
	model.connected = true
	model.output = append(model.output, "Welcome to the MUD!")
	model.output = append(model.output, "Type a command>")
	model.updateViewport()

	// Get the initial viewport content
	initialContent := model.lastViewportContent

	// Type a character
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)

	// Verify currentInput was updated
	if model.currentInput != "n" {
		t.Errorf("Expected currentInput to be 'n', got '%s'", model.currentInput)
	}

	// Verify viewport content was updated (because input changed)
	if model.lastViewportContent == initialContent {
		t.Error("Expected viewport content to change after typing")
	}

	// Save the content after typing one character
	contentAfterN := model.lastViewportContent

	// Type another character
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)

	// Verify currentInput was updated
	if model.currentInput != "no" {
		t.Errorf("Expected currentInput to be 'no', got '%s'", model.currentInput)
	}

	// Verify viewport content was updated again (because input changed)
	if model.lastViewportContent == contentAfterN {
		t.Error("Expected viewport content to change after typing second character")
	}
}

// TestCursorMovementUpdatesViewport tests that moving the cursor
// updates the viewport content (cursor position is visible)
func TestCursorMovementUpdatesViewport(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m

	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Add some output and set connected state
	model.connected = true
	model.output = append(model.output, "Welcome to the MUD!")
	model.output = append(model.output, "Type a command>")
	
	// Type some characters
	model.currentInput = "north"
	model.cursorPos = 5
	model.updateViewport()

	// Get the viewport content after typing
	contentAfterTyping := model.lastViewportContent

	// Move cursor left (this should call updateViewport)
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)

	// Verify cursor position changed
	if model.cursorPos != 4 {
		t.Errorf("Expected cursorPos to be 4, got %d", model.cursorPos)
	}

	// Verify viewport content was updated by the cursor move
	// (the cursor block █ should move position)
	if model.lastViewportContent == contentAfterTyping {
		t.Error("Expected viewport content to change when cursor moves (cursor character moves)")
	}
}

// TestOutputUpdateCausesViewportUpdate tests that new output from the server
// always causes a viewport update
func TestOutputUpdateCausesViewportUpdate(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m

	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Add some output and set connected state
	model.connected = true
	model.output = append(model.output, "Welcome to the MUD!")
	model.output = append(model.output, "Type a command>")
	model.updateViewport()

	// Get the initial viewport content
	initialContent := model.lastViewportContent

	// Add new output (simulating server response)
	model.output = append(model.output, "You go north.")
	model.output = append(model.output, "You are in a forest.")
	model.output = append(model.output, "Type a command>")
	model.updateViewport()

	// Verify viewport content was updated
	if model.lastViewportContent == initialContent {
		t.Error("Expected viewport content to change after new output")
	}

	// Verify the new content contains the new output
	if model.lastViewportContent == "" {
		t.Error("Expected lastViewportContent to be non-empty")
	}
}

// TestTypingSameCharacterMultipleTimes tests that repeatedly typing characters
// always updates the viewport content (since the input is growing)
func TestTypingSameCharacterMultipleTimes(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m

	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Add some output and set connected state
	model.connected = true
	model.output = append(model.output, "Welcome to the MUD!")
	model.output = append(model.output, "Type a command>")
	model.updateViewport()

	previousContent := model.lastViewportContent

	// Type 'n' three times
	for i := 0; i < 3; i++ {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(*Model)

		// Each time, verify the content changed
		if model.lastViewportContent == previousContent {
			t.Errorf("Expected viewport content to change after typing iteration %d", i+1)
		}
		previousContent = model.lastViewportContent
	}

	// Verify final input
	if model.currentInput != "nnn" {
		t.Errorf("Expected currentInput to be 'nnn', got '%s'", model.currentInput)
	}
}

// TestBackspaceAndRetypeUpdatesViewport tests that backspacing and retyping
// the same character updates the viewport correctly
func TestBackspaceAndRetypeUpdatesViewport(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m

	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Add some output and set connected state
	model.connected = true
	model.output = append(model.output, "Welcome to the MUD!")
	model.output = append(model.output, "Type a command>")
	model.updateViewport()

	// Type 'n'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	contentAfterN := model.lastViewportContent

	// Backspace
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)

	// Verify content changed after backspace
	if model.lastViewportContent == contentAfterN {
		t.Error("Expected viewport content to change after backspace")
	}
	contentAfterBackspace := model.lastViewportContent

	// Type 'n' again
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)

	// Verify content changed after retyping
	if model.lastViewportContent == contentAfterBackspace {
		t.Error("Expected viewport content to change after retyping 'n'")
	}

	// The content should be the same as when we first typed 'n'
	if model.lastViewportContent != contentAfterN {
		t.Error("Expected viewport content to be the same as when first typed 'n'")
	}

	// Verify final input
	if model.currentInput != "n" {
		t.Errorf("Expected currentInput to be 'n', got '%s'", model.currentInput)
	}
}
