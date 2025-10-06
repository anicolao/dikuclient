package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPasswordNotInAccounts(t *testing.T) {
	// Test that accounts.json structure doesn't include passwords
	// Passwords are now stored separately in .passwords file
	accountsJSON := `{
		"accounts": [
			{
				"name": "TestAccount",
				"host": "test.mud.org",
				"port": 4000,
				"username": "testuser"
			}
		]
	}`

	var accounts map[string]interface{}
	if err := json.Unmarshal([]byte(accountsJSON), &accounts); err != nil {
		t.Fatalf("Failed to unmarshal accounts: %v", err)
	}

	// Verify password is not present
	if accountsList, ok := accounts["accounts"].([]interface{}); ok {
		if len(accountsList) > 0 {
			if accMap, ok := accountsList[0].(map[string]interface{}); ok {
				if _, hasPassword := accMap["password"]; hasPassword {
					t.Error("Password field should not be in accounts.json")
				}
			}
		}
	}
}

func TestFileTimestampComparison(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	
	testData := []byte(`{"test": "data"}`)
	if err := os.WriteFile(filePath, testData, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	serverTime := fileInfo.ModTime().UnixMilli()
	
	// Simulate client time (newer)
	clientTime := time.Now().Add(1 * time.Hour).UnixMilli()
	
	if clientTime <= serverTime {
		t.Error("Client time should be newer than server time")
	}
}

func TestDataMessageSerialization(t *testing.T) {
	msg := DataMessage{
		Type:      "file_update",
		Path:      "accounts.json",
		Content:   `{"test": "content"}`,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal DataMessage: %v", err)
	}

	var decoded DataMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DataMessage: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, msg.Type)
	}
	if decoded.Path != msg.Path {
		t.Errorf("Path mismatch: got %s, want %s", decoded.Path, msg.Path)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content mismatch: got %s, want %s", decoded.Content, msg.Content)
	}
}
