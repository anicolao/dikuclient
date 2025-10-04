package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAndGetAccount(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an account
	account := Account{
		Name:     "test-mud",
		Host:     "mud.test.com",
		Port:     4000,
		Username: "testuser",
		Password: "testpass",
	}

	err = cfg.AddAccount(account)
	if err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// Load config and verify
	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(cfg2.Accounts))
	}

	retrieved, err := cfg2.GetAccount("test-mud")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if retrieved.Name != account.Name || retrieved.Host != account.Host ||
		retrieved.Port != account.Port || retrieved.Username != account.Username ||
		retrieved.Password != account.Password {
		t.Errorf("Retrieved account doesn't match: %+v", retrieved)
	}
}

func TestDeleteAccount(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	account := Account{
		Name: "test-mud",
		Host: "mud.test.com",
		Port: 4000,
	}

	err = cfg.AddAccount(account)
	if err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	err = cfg.DeleteAccount("test-mud")
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Accounts) != 0 {
		t.Fatalf("Expected 0 accounts after deletion, got %d", len(cfg2.Accounts))
	}
}

func TestUpdateAccount(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	account := Account{
		Name:     "test-mud",
		Host:     "mud.test.com",
		Port:     4000,
		Username: "user1",
	}

	err = cfg.AddAccount(account)
	if err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// Update the account
	account.Username = "user2"
	account.Password = "newpass"

	err = cfg.AddAccount(account)
	if err != nil {
		t.Fatalf("Failed to update account: %v", err)
	}

	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Accounts) != 1 {
		t.Fatalf("Expected 1 account after update, got %d", len(cfg2.Accounts))
	}

	retrieved, err := cfg2.GetAccount("test-mud")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if retrieved.Username != "user2" || retrieved.Password != "newpass" {
		t.Errorf("Account not updated correctly: %+v", retrieved)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got: %v", err)
	}

	if len(cfg.Accounts) != 0 {
		t.Fatalf("Expected empty config for non-existent file, got %d accounts", len(cfg.Accounts))
	}
}

func TestGetConfigPathWithEnvVar(t *testing.T) {
	// Save original environment
	origEnv := os.Getenv("DIKUCLIENT_CONFIG_DIR")
	defer func() {
		if origEnv != "" {
			os.Setenv("DIKUCLIENT_CONFIG_DIR", origEnv)
		} else {
			os.Unsetenv("DIKUCLIENT_CONFIG_DIR")
		}
	}()

	// Test with environment variable set
	tmpDir := t.TempDir()
	customConfigDir := filepath.Join(tmpDir, "custom-config")
	os.Setenv("DIKUCLIENT_CONFIG_DIR", customConfigDir)

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	expectedPath := filepath.Join(customConfigDir, "accounts.json")
	if configPath != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, configPath)
	}

	// Verify directory was created
	if _, err := os.Stat(customConfigDir); os.IsNotExist(err) {
		t.Errorf("Expected config directory to be created at %s", customConfigDir)
	}

	// Test that config can be loaded from custom path
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add an account and verify it's saved to the custom location
	account := Account{
		Name: "test-custom",
		Host: "custom.mud.com",
		Port: 4000,
	}

	err = cfg.AddAccount(account)
	if err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// Verify file exists at custom location
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected accounts.json to be created at %s", expectedPath)
	}

	// Load again and verify
	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	retrieved, err := cfg2.GetAccount("test-custom")
	if err != nil {
		t.Fatalf("Failed to get account from custom config: %v", err)
	}

	if retrieved.Name != account.Name {
		t.Errorf("Account from custom config doesn't match")
	}
}
