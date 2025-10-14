package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Server represents a MUD server
type Server struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Character represents a character on a specific server
type Character struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

// Account represents a saved MUD account (legacy - kept for backward compatibility)
// Note: Password is NOT stored in accounts.json, it's stored separately in .passwords file
type Account struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"` // Never serialize to JSON
}

// Config represents the application configuration
type Config struct {
	Servers        []Server    `json:"servers,omitempty"`
	Characters     []Character `json:"characters,omitempty"`
	Accounts       []Account   `json:"accounts"` // Legacy field for backward compatibility
	DefaultAccount string      `json:"default_account,omitempty"`
	configPath     string      // Path to the config file (for testing)
}

// GetConfigPath returns the path to the configuration file
// If DIKUCLIENT_CONFIG_DIR environment variable is set, it will be used as the config directory
// Otherwise, it defaults to ~/.config/dikuclient
func GetConfigPath() (string, error) {
	var configDir string

	// Check for environment variable override
	if envConfigDir := os.Getenv("DIKUCLIENT_CONFIG_DIR"); envConfigDir != "" {
		configDir = envConfigDir
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config", "dikuclient")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "accounts.json"), nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	return LoadConfigFromPath(configPath)
}

// LoadConfigFromPath loads configuration from a specific path (useful for testing)
func LoadConfigFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{Accounts: []Account{}, configPath: configPath}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	config.configPath = configPath

	return &config, nil
}

// SaveConfig saves the configuration to disk
func (c *Config) SaveConfig() error {
	configPath := c.configPath
	if configPath == "" {
		var err error
		configPath, err = GetConfigPath()
		if err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddAccount adds a new account to the configuration
func (c *Config) AddAccount(account Account) error {
	// Check if account with same name already exists
	for i, existing := range c.Accounts {
		if existing.Name == account.Name {
			// Update existing account
			c.Accounts[i] = account
			return c.SaveConfig()
		}
	}

	// Add new account
	c.Accounts = append(c.Accounts, account)
	return c.SaveConfig()
}

// GetAccount retrieves an account by name
func (c *Config) GetAccount(name string) (*Account, error) {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account '%s' not found", name)
}

// DeleteAccount removes an account from the configuration
func (c *Config) DeleteAccount(name string) error {
	for i, account := range c.Accounts {
		if account.Name == name {
			c.Accounts = append(c.Accounts[:i], c.Accounts[i+1:]...)
			return c.SaveConfig()
		}
	}
	return fmt.Errorf("account '%s' not found", name)
}

// ListAccounts returns all saved accounts
func (c *Config) ListAccounts() []Account {
	return c.Accounts
}

// AddServer adds a new server to the configuration
func (c *Config) AddServer(server Server) error {
	// Check if server with same name already exists
	for i, existing := range c.Servers {
		if existing.Name == server.Name {
			// Update existing server
			c.Servers[i] = server
			return c.SaveConfig()
		}
	}

	// Add new server
	c.Servers = append(c.Servers, server)
	return c.SaveConfig()
}

// GetServer retrieves a server by name
func (c *Config) GetServer(name string) (*Server, error) {
	for _, server := range c.Servers {
		if server.Name == name {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server '%s' not found", name)
}

// ListServers returns all saved servers
func (c *Config) ListServers() []Server {
	return c.Servers
}

// DeleteServer removes a server from the configuration
func (c *Config) DeleteServer(name string) error {
	for i, server := range c.Servers {
		if server.Name == name {
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			// Also delete all characters for this server
			c.Characters = filterCharactersByServer(c.Characters, server.Host, server.Port)
			return c.SaveConfig()
		}
	}
	return fmt.Errorf("server '%s' not found", name)
}

// AddCharacter adds a new character to the configuration
func (c *Config) AddCharacter(character Character) error {
	// Check if character with same username and server already exists
	for i, existing := range c.Characters {
		if existing.Username == character.Username && existing.Host == character.Host && existing.Port == character.Port {
			// Update existing character
			c.Characters[i] = character
			return c.SaveConfig()
		}
	}

	// Add new character
	c.Characters = append(c.Characters, character)
	return c.SaveConfig()
}

// GetCharacter retrieves a character by username and server
func (c *Config) GetCharacter(username string, host string, port int) (*Character, error) {
	for _, character := range c.Characters {
		if character.Username == username && character.Host == host && character.Port == port {
			return &character, nil
		}
	}
	return nil, fmt.Errorf("character '%s' not found on %s:%d", username, host, port)
}

// ListCharacters returns all saved characters
func (c *Config) ListCharacters() []Character {
	return c.Characters
}

// ListCharactersForServer returns all characters for a specific server
func (c *Config) ListCharactersForServer(host string, port int) []Character {
	var characters []Character
	for _, char := range c.Characters {
		if char.Host == host && char.Port == port {
			characters = append(characters, char)
		}
	}
	return characters
}

// DeleteCharacter removes a character from the configuration
func (c *Config) DeleteCharacter(username string, host string, port int) error {
	for i, character := range c.Characters {
		if character.Username == username && character.Host == host && character.Port == port {
			c.Characters = append(c.Characters[:i], c.Characters[i+1:]...)
			return c.SaveConfig()
		}
	}
	return fmt.Errorf("character '%s' not found on %s:%d", username, host, port)
}

// Helper function to filter out characters for a specific server
func filterCharactersByServer(characters []Character, host string, port int) []Character {
	var filtered []Character
	for _, char := range characters {
		if char.Host != host || char.Port != port {
			filtered = append(filtered, char)
		}
	}
	return filtered
}
