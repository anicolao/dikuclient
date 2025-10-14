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

	// Note: Password is not stored in accounts.json anymore, it's in separate .passwords file
	if retrieved.Name != account.Name || retrieved.Host != account.Host ||
		retrieved.Port != account.Port || retrieved.Username != account.Username {
		t.Errorf("Retrieved account doesn't match: %+v", retrieved)
	}
	
	// Password should be empty since it's not serialized
	if retrieved.Password != "" {
		t.Errorf("Password should not be stored in accounts.json, got: %s", retrieved.Password)
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

	// Note: Password is not stored in accounts.json anymore
	if retrieved.Username != "user2" {
		t.Errorf("Account not updated correctly: %+v", retrieved)
	}
	
	// Password should be empty since it's not serialized
	if retrieved.Password != "" {
		t.Errorf("Password should not be stored in accounts.json, got: %s", retrieved.Password)
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

func TestAddAndGetServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a server
	server := Server{
		Name: "TestServer",
		Host: "mud.example.com",
		Port: 4000,
	}

	err = cfg.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Load config and verify
	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(cfg2.Servers))
	}

	retrieved, err := cfg2.GetServer("TestServer")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if retrieved.Name != server.Name || retrieved.Host != server.Host || retrieved.Port != server.Port {
		t.Errorf("Server mismatch: expected %+v, got %+v", server, retrieved)
	}
}

func TestUpdateServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server := Server{
		Name: "TestServer",
		Host: "mud1.example.com",
		Port: 4000,
	}

	err = cfg.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Update the server
	server.Host = "mud2.example.com"
	server.Port = 4001

	err = cfg.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to update server: %v", err)
	}

	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Servers) != 1 {
		t.Fatalf("Expected 1 server after update, got %d", len(cfg2.Servers))
	}

	retrieved, err := cfg2.GetServer("TestServer")
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if retrieved.Host != "mud2.example.com" || retrieved.Port != 4001 {
		t.Errorf("Server not updated correctly: %+v", retrieved)
	}
}

func TestAddAndGetCharacter(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a character
	character := Character{
		Host:     "mud.example.com",
		Port:     4000,
		Username: "testuser",
	}

	err = cfg.AddCharacter(character)
	if err != nil {
		t.Fatalf("Failed to add character: %v", err)
	}

	// Load config and verify
	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Characters) != 1 {
		t.Fatalf("Expected 1 character, got %d", len(cfg2.Characters))
	}

	retrieved, err := cfg2.GetCharacter("testuser", "mud.example.com", 4000)
	if err != nil {
		t.Fatalf("Failed to get character: %v", err)
	}

	if retrieved.Host != character.Host ||
		retrieved.Port != character.Port || retrieved.Username != character.Username {
		t.Errorf("Character mismatch: expected %+v, got %+v", character, retrieved)
	}
}

func TestListCharactersForServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add characters for different servers
	char1 := Character{
		Host:     "server1.com",
		Port:     4000,
		Username: "user1",
	}

	char2 := Character{
		Host:     "server1.com",
		Port:     4000,
		Username: "user2",
	}

	char3 := Character{
		Host:     "server2.com",
		Port:     4000,
		Username: "user3",
	}

	cfg.AddCharacter(char1)
	cfg.AddCharacter(char2)
	cfg.AddCharacter(char3)

	// List characters for server1.com
	chars := cfg.ListCharactersForServer("server1.com", 4000)
	if len(chars) != 2 {
		t.Fatalf("Expected 2 characters for server1.com, got %d", len(chars))
	}

	// List characters for server2.com
	chars = cfg.ListCharactersForServer("server2.com", 4000)
	if len(chars) != 1 {
		t.Fatalf("Expected 1 character for server2.com, got %d", len(chars))
	}
}

func TestDeleteServer(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add server and characters
	server := Server{
		Name: "TestServer",
		Host: "mud.example.com",
		Port: 4000,
	}

	char := Character{
		Host:     "mud.example.com",
		Port:     4000,
		Username: "testuser",
	}

	cfg.AddServer(server)
	cfg.AddCharacter(char)

	// Delete server
	err = cfg.DeleteServer("TestServer")
	if err != nil {
		t.Fatalf("Failed to delete server: %v", err)
	}

	// Verify server is deleted
	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Servers) != 0 {
		t.Errorf("Expected 0 servers after delete, got %d", len(cfg2.Servers))
	}

	// Verify characters for that server are also deleted
	if len(cfg2.Characters) != 0 {
		t.Errorf("Expected 0 characters after server delete, got %d", len(cfg2.Characters))
	}
}

func TestDeleteCharacter(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "accounts.json")

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add characters
	char1 := Character{
		Host:     "mud.example.com",
		Port:     4000,
		Username: "user1",
	}

	char2 := Character{
		Host:     "mud.example.com",
		Port:     4000,
		Username: "user2",
	}

	cfg.AddCharacter(char1)
	cfg.AddCharacter(char2)

	// Delete one character
	err = cfg.DeleteCharacter("user1", "mud.example.com", 4000)
	if err != nil {
		t.Fatalf("Failed to delete character: %v", err)
	}

	// Verify character is deleted
	cfg2, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg2.Characters) != 1 {
		t.Errorf("Expected 1 character after delete, got %d", len(cfg2.Characters))
	}

	_, err = cfg2.GetCharacter("user1", "mud.example.com", 4000)
	if err == nil {
		t.Errorf("Expected error getting deleted character")
	}

	_, err = cfg2.GetCharacter("user2", "mud.example.com", 4000)
	if err != nil {
		t.Errorf("Expected to find user2, got error: %v", err)
	}
}
