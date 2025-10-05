package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSplitViewportPgUp(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add many output lines to ensure viewport has scrollable content
	for i := 0; i < 100; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Initially not split
	if model.isSplit {
		t.Error("Model should not be split initially")
	}
	
	// Verify we're at bottom before scrolling up
	if !model.viewport.AtBottom() {
		t.Error("Viewport should be at bottom initially")
	}
	
	// Press Page Up
	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	// Should be split now
	if !model.isSplit {
		t.Error("Model should be split after Page Up")
	}
}

func TestSplitViewportPgDown(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add many output lines to ensure viewport has scrollable content
	for i := 0; i < 100; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Press Page Up to enter split mode
	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	if !model.isSplit {
		t.Error("Model should be split after Page Up")
	}
	
	// Press Page Down until at bottom
	for i := 0; i < 10; i++ {
		msg = tea.KeyMsg{Type: tea.KeyPgDown}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(*Model)
		
		if model.viewport.AtBottom() {
			break
		}
	}
	
	// Should exit split mode when at bottom
	if model.isSplit {
		t.Error("Model should not be split when at bottom after Page Down")
	}
}

func TestSplitViewportMouseWheel(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add many output lines to ensure viewport has scrollable content
	for i := 0; i < 100; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Initially not split
	if model.isSplit {
		t.Error("Model should not be split initially")
	}
	
	// Scroll up with mouse wheel
	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
	}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	// Should be split now
	if !model.isSplit {
		t.Error("Model should be split after mouse wheel up")
	}
}

func TestSplitViewportNewContentAtBottom(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add many output lines to ensure viewport has scrollable content
	for i := 0; i < 100; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Press Page Up to enter split mode
	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	if !model.isSplit {
		t.Error("Model should be split after Page Up")
	}
	
	// Scroll to bottom
	model.viewport.GotoBottom()
	
	// Add new content
	model.output = append(model.output, "new line")
	model.updateViewport()
	
	// Should exit split mode since viewport was at bottom
	if model.isSplit {
		t.Error("Model should exit split mode when new content arrives and viewport was at bottom")
	}
}

func TestSplitViewportNewContentWhileScrolled(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add many output lines to ensure viewport has scrollable content
	for i := 0; i < 100; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Press Page Up to enter split mode
	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	if !model.isSplit {
		t.Error("Model should be split after Page Up")
	}
	
	// Don't scroll back to bottom, just add new content
	model.output = append(model.output, "new line")
	model.updateViewport()
	
	// Should stay in split mode since viewport was not at bottom
	if !model.isSplit {
		t.Error("Model should stay in split mode when new content arrives while scrolled up")
	}
}

func TestSplitViewportRendering(t *testing.T) {
	m := NewModel("localhost", 4000, nil, nil)
	model := &m
	model.width = 100
	model.height = 40
	model.connected = true
	
	// Resize to set viewport dimensions
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)
	
	// Add some output lines
	for i := 0; i < 50; i++ {
		model.output = append(model.output, "test line")
	}
	model.updateViewport()
	
	// Enter split mode
	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	
	// Render the view
	view := model.View()
	
	// Check that view is not empty
	if view == "" {
		t.Error("Split view should not be empty")
	}
	
	// View should contain content
	if len(view) < 100 {
		t.Errorf("Split view seems too short: %d characters", len(view))
	}
}
