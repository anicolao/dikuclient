package tui

import (
	"os"
	"testing"
)

// TestBarebonesModelCreation tests that we can create a minimal model
// without any advanced features enabled
func TestBarebonesModelCreation(t *testing.T) {
	// Create a barebones model with just host and port
	model := NewModel("localhost", 4000, nil, nil)

	// Verify basic fields are set
	if model.host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", model.host)
	}

	if model.port != 4000 {
		t.Errorf("Expected port 4000, got %d", model.port)
	}

	// Verify no advanced features are initially active
	if model.connected {
		t.Error("Model should not be connected initially")
	}

	if model.username != "" {
		t.Error("Barebones model should not have username set")
	}

	if model.password != "" {
		t.Error("Barebones model should not have password set")
	}

	// Verify empty panes are initialized
	if len(model.output) != 0 {
		t.Error("Output should be empty initially")
	}

	if len(model.inventory) != 0 {
		t.Error("Inventory should be empty initially")
	}

	// Verify viewport is created
	if model.viewport.Width < 0 || model.viewport.Height < 0 {
		t.Error("Viewport dimensions should be non-negative")
	}
}

// TestBarebonesModelWithAuth tests creating a model with auth credentials
func TestBarebonesModelWithAuth(t *testing.T) {
	model := NewModelWithAuth("localhost", 4000, "testuser", "testpass", nil, nil, nil, false)

	if model.host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", model.host)
	}

	if model.port != 4000 {
		t.Errorf("Expected port 4000, got %d", model.port)
	}

	if model.username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", model.username)
	}

	if model.password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", model.password)
	}

	// Verify it's still not connected
	if model.connected {
		t.Error("Model should not be connected initially")
	}
}

// TestBarebonesRendering tests that the barebones TUI can be rendered
// without crashing, even with no connection
func TestBarebonesRendering(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)

	// Set minimal dimensions for rendering
	model.width = 80
	model.height = 24

	// Initialize viewport with dimensions
	model.viewport.Width = 50
	model.viewport.Height = 20

	// Try to render the view - should not panic
	view := model.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should show "Loading..." when width is 0
	model.width = 0
	view = model.View()
	if view != "Loading..." {
		t.Errorf("Expected 'Loading...' when width is 0, got '%s'", view)
	}
}

// TestBarebonesEmptyPanels verifies that empty panels render correctly
func TestBarebonesEmptyPanels(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 100
	model.height = 30
	model.sidebarWidth = 30

	// Set up viewports
	model.viewport.Width = 60
	model.viewport.Height = 25
	model.inventoryViewport.Width = 25
	model.inventoryViewport.Height = 5

	// Render sidebar with empty panels
	sidebar := model.renderSidebar(30, 25)

	if sidebar == "" {
		t.Error("Sidebar should not be empty")
	}

	// The sidebar should contain placeholder text for empty panels
	// (We can't easily test the exact rendering without lipgloss, but we can verify it doesn't crash)
}

// TestBarebonesInputHandling tests basic input handling
func TestBarebonesInputHandling(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Initially, input should be empty
	if model.currentInput != "" {
		t.Error("Current input should be empty initially")
	}

	if model.cursorPos != 0 {
		t.Error("Cursor position should be 0 initially")
	}
}

// TestBarebonesSimpleConnection tests that we can initialize a connection
// (This test doesn't actually connect, just verifies the initialization)
func TestBarebonesSimpleConnection(t *testing.T) {
	// Create a model
	model := NewModel("localhost", 4000, nil, nil)

	// Verify connection is nil initially
	if model.conn != nil {
		t.Error("Connection should be nil initially")
	}

	// Verify connected flag is false
	if model.connected {
		t.Error("Connected flag should be false initially")
	}

	// Init() should return a command to connect
	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() should return a connection command")
	}
}

// TestBarebonesWithLogging tests model creation with logging enabled
func TestBarebonesWithLogging(t *testing.T) {
	// Create temporary log files
	mudLog, err := os.CreateTemp("", "mud-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp mud log: %v", err)
	}
	defer os.Remove(mudLog.Name())
	defer mudLog.Close()

	tuiLog, err := os.CreateTemp("", "tui-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp tui log: %v", err)
	}
	defer os.Remove(tuiLog.Name())
	defer tuiLog.Close()

	// Create model with logging
	model := NewModel("localhost", 4000, mudLog, tuiLog)

	if model.mudLogFile != mudLog {
		t.Error("MUD log file not set correctly")
	}

	if model.tuiLogFile != tuiLog {
		t.Error("TUI log file not set correctly")
	}
}

// TestBarebonesDefaults tests that the model has sensible defaults
func TestBarebonesDefaults(t *testing.T) {
	model := NewModel("testhost", 1234, nil, nil)

	// Check defaults
	if model.sidebarWidth <= 0 {
		t.Error("Sidebar width should be positive")
	}

	if model.autoLoginState != 0 {
		t.Error("Auto-login state should be 0 (idle) initially")
	}

	if model.autoWalking {
		t.Error("Auto-walking should be false initially")
	}

	if model.echoSuppressed {
		t.Error("Echo suppressed should be false initially")
	}
}
