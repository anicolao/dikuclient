package tui

import (
	"os"
	"strings"
	"testing"
)

// TestShareCommandInWebMode tests that the /share command works in web mode
func TestShareCommandInWebMode(t *testing.T) {
	// Set up environment variables to simulate web mode
	os.Setenv("DIKUCLIENT_WEB_SESSION_ID", "test-session-123")
	os.Setenv("DIKUCLIENT_WEB_SERVER_URL", "http://localhost:8080")
	defer os.Unsetenv("DIKUCLIENT_WEB_SESSION_ID")
	defer os.Unsetenv("DIKUCLIENT_WEB_SERVER_URL")

	// Create a model
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Check that environment variables were read
	if model.webSessionID != "test-session-123" {
		t.Errorf("Expected webSessionID to be 'test-session-123', got '%s'", model.webSessionID)
	}
	if model.webServerURL != "http://localhost:8080" {
		t.Errorf("Expected webServerURL to be 'http://localhost:8080', got '%s'", model.webServerURL)
	}

	// Execute the /share command
	model.handleShareCommand()

	// Check that the output contains the expected URL
	found := false
	expectedURL := "http://localhost:8080/?id=test-session-123"
	for _, line := range model.output {
		// Strip ANSI codes for easier checking
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, expectedURL) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected output to contain '%s', got: %v", expectedURL, model.output)
	}

	// Check that the output contains the expected header
	foundHeader := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "Share This Session") {
			foundHeader = true
			break
		}
	}

	if !foundHeader {
		t.Error("Expected output to contain 'Share This Session' header")
	}
}

// TestShareCommandNotInWebMode tests that the /share command shows an error when not in web mode
func TestShareCommandNotInWebMode(t *testing.T) {
	// Make sure environment variables are not set
	os.Unsetenv("DIKUCLIENT_WEB_SESSION_ID")
	os.Unsetenv("DIKUCLIENT_WEB_SERVER_URL")

	// Create a model
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Check that environment variables are empty
	if model.webSessionID != "" {
		t.Errorf("Expected webSessionID to be empty, got '%s'", model.webSessionID)
	}
	if model.webServerURL != "" {
		t.Errorf("Expected webServerURL to be empty, got '%s'", model.webServerURL)
	}

	// Execute the /share command
	model.handleShareCommand()

	// Check that the output contains an error message
	found := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "only available in web mode") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected output to contain error message, got: %v", model.output)
	}
}

// TestHelpCommandIncludesShare tests that /help includes the /share command
func TestHelpCommandIncludesShare(t *testing.T) {
	model := NewModel("localhost", 4000, nil, nil)
	model.width = 80
	model.height = 24

	// Execute the /help command
	model.handleHelpCommand()

	// Check that the output contains the /share command
	found := false
	for _, line := range model.output {
		cleanLine := stripANSI(line)
		if strings.Contains(cleanLine, "/share") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected /help output to include /share command")
	}
}
