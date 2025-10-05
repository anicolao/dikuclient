package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manager manages command history
type Manager struct {
	Commands []string `json:"commands"`
	filePath string   // Path to history.json (not serialized)
}

// NewManager creates a new history manager
func NewManager() *Manager {
	return &Manager{
		Commands: make([]string, 0),
	}
}

// GetHistoryPath returns the path to the history file
func GetHistoryPath() (string, error) {
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

	return filepath.Join(configDir, "history.json"), nil
}

// Load loads command history from disk
func Load() (*Manager, error) {
	historyPath, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	return LoadFromPath(historyPath)
}

// LoadFromPath loads history from a specific path (useful for testing)
func LoadFromPath(historyPath string) (*Manager, error) {
	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty manager if file doesn't exist
			m := NewManager()
			m.filePath = historyPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}
	m.filePath = historyPath

	return &m, nil
}

// Save saves command history to disk
func (m *Manager) Save() error {
	historyPath := m.filePath
	if historyPath == "" {
		var err error
		historyPath, err = GetHistoryPath()
		if err != nil {
			return err
		}
		m.filePath = historyPath
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Add adds a command to history (avoiding consecutive duplicates)
func (m *Manager) Add(command string) {
	if command == "" {
		return
	}

	// Don't add if it's the same as the last command
	if len(m.Commands) > 0 && m.Commands[len(m.Commands)-1] == command {
		return
	}

	m.Commands = append(m.Commands, command)
}

// GetCommands returns a copy of the command history
func (m *Manager) GetCommands() []string {
	commands := make([]string, len(m.Commands))
	copy(commands, m.Commands)
	return commands
}
