package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAIAPIKeyStorage(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "dikuclient-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create password store
	ps := NewPasswordStore(false)
	ps.filePath = filepath.Join(tmpDir, ".passwords")
	
	// Load the password store (initializes the map)
	if err := ps.Load(); err != nil {
		t.Fatalf("Failed to load password store: %v", err)
	}

	// Test setting API key
	testKey := "sk-test-key-12345"
	err = SetAIAPIKey(ps, testKey)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	// Test retrieving API key
	retrievedKey := GetAIAPIKey(ps)
	if retrievedKey != testKey {
		t.Errorf("Expected API key %s, got %s", testKey, retrievedKey)
	}

	// Test that API key is persisted
	ps2 := NewPasswordStore(false)
	ps2.filePath = filepath.Join(tmpDir, ".passwords")
	if err := ps2.Load(); err != nil {
		t.Fatalf("Failed to load password store: %v", err)
	}

	retrievedKey2 := GetAIAPIKey(ps2)
	if retrievedKey2 != testKey {
		t.Errorf("API key not persisted correctly. Expected %s, got %s", testKey, retrievedKey2)
	}

	// Test deleting API key
	err = DeleteAIAPIKey(ps2)
	if err != nil {
		t.Fatalf("Failed to delete API key: %v", err)
	}

	retrievedKey3 := GetAIAPIKey(ps2)
	if retrievedKey3 != "" {
		t.Errorf("API key not deleted. Expected empty string, got %s", retrievedKey3)
	}
}

func TestAIAPIKeyNotInConfig(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "dikuclient-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up config path
	configPath := filepath.Join(tmpDir, "accounts.json")

	// Create a config with AI settings
	cfg := &Config{
		AI: AIConfig{
			Type:   "openai",
			URL:    "https://api.openai.com/v1/chat/completions",
			APIKey: "sk-test-key-should-not-be-saved",
		},
		configPath: configPath,
	}

	// Save the config
	err = cfg.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read the config file as text
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configText := string(data)

	// Verify API key is NOT in the file
	if containsString(configText, "sk-test-key") {
		t.Error("API key found in config file! This is a security issue.")
	}
	if containsString(configText, "APIKey") {
		t.Error("APIKey field found in config file! This is a security issue.")
	}

	// Verify Type and URL are in the file
	if !containsString(configText, "openai") {
		t.Error("AI type not found in config file")
	}
	if !containsString(configText, "https://api.openai.com") {
		t.Error("AI URL not found in config file")
	}
}

func TestAIAPIKeyReadOnlyMode(t *testing.T) {
	// Create password store in read-only mode (web mode)
	ps := NewPasswordStore(true)

	// Try to set API key
	err := SetAIAPIKey(ps, "test-key")
	if err == nil {
		t.Error("Expected error when setting API key in read-only mode, got nil")
	}

	// Try to delete API key
	err = DeleteAIAPIKey(ps)
	if err == nil {
		t.Error("Expected error when deleting API key in read-only mode, got nil")
	}
}

func TestAIAPIKeyFilePermissions(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "dikuclient-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create password store and save an API key
	ps := NewPasswordStore(false)
	ps.filePath = filepath.Join(tmpDir, ".passwords")
	
	if err := ps.Load(); err != nil {
		t.Fatalf("Failed to load password store: %v", err)
	}

	err = SetAIAPIKey(ps, "test-key")
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	// Check file permissions
	fileInfo, err := os.Stat(ps.filePath)
	if err != nil {
		t.Fatalf("Failed to stat password file: %v", err)
	}

	// Verify permissions are 0600 (owner read/write only)
	expectedPerms := os.FileMode(0600)
	actualPerms := fileInfo.Mode().Perm()
	if actualPerms != expectedPerms {
		t.Errorf("Incorrect file permissions. Expected %o, got %o", expectedPerms, actualPerms)
	}
}

func containsString(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 && text != substr && (text == substr || len(text) > len(substr) && (text[:len(substr)] == substr || text[len(text)-len(substr):] == substr || containsInMiddle(text, substr)))
}

func containsInMiddle(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
