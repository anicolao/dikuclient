package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Account represents a saved MUD account
type Account struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config represents the application configuration
type Config struct {
	Accounts       []Account `json:"accounts"`
	DefaultAccount string    `json:"default_account,omitempty"`
	configPath     string    // Path to the config file (for testing)
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dikuclient")
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
