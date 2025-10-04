package xpstats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// XPStat represents XP per second statistics for a creature with EMA tracking
type XPStat struct {
	CreatureName string  `json:"creature_name"`
	XPPerSecond  float64 `json:"xp_per_second"` // Exponential moving average of XP/s
	SampleCount  int     `json:"sample_count"`  // Number of samples used
}

// Manager manages XP statistics with persistence
type Manager struct {
	Stats    map[string]*XPStat `json:"stats"`
	filePath string             // Path to xps.json (not serialized)
}

// NewManager creates a new XP stats manager
func NewManager() *Manager {
	return &Manager{
		Stats: make(map[string]*XPStat),
	}
}

// GetXPSPath returns the path to the XP stats file
func GetXPSPath() (string, error) {
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

	return filepath.Join(configDir, "xps.json"), nil
}

// Load loads XP stats from disk
func Load() (*Manager, error) {
	xpsPath, err := GetXPSPath()
	if err != nil {
		return nil, err
	}

	return LoadFromPath(xpsPath)
}

// LoadFromPath loads XP stats from a specific path (useful for testing)
func LoadFromPath(xpsPath string) (*Manager, error) {
	data, err := os.ReadFile(xpsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty manager if file doesn't exist
			m := NewManager()
			m.filePath = xpsPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read XP stats file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse XP stats file: %w", err)
	}

	m.filePath = xpsPath

	// Ensure map is initialized
	if m.Stats == nil {
		m.Stats = make(map[string]*XPStat)
	}

	return &m, nil
}

// Save saves XP stats to disk
func (m *Manager) Save() error {
	if m.filePath == "" {
		return fmt.Errorf("no file path set for XP stats manager")
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XP stats: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write XP stats file: %w", err)
	}

	return nil
}

// UpdateStat updates or creates an XP stat using exponential moving average
// The alpha parameter controls the weight of new samples (higher = more weight to recent samples)
// For focusing on last 5-10 samples, we use alpha = 2/(N+1) where N is around 7-8
func (m *Manager) UpdateStat(creatureName string, newXPPerSecond float64) {
	stat, exists := m.Stats[creatureName]
	
	if !exists {
		// First sample - just store it
		m.Stats[creatureName] = &XPStat{
			CreatureName: creatureName,
			XPPerSecond:  newXPPerSecond,
			SampleCount:  1,
		}
		return
	}

	// Calculate alpha based on desired window size
	// For 5-10 samples, we'll use N=7, giving alpha = 2/(7+1) = 0.25
	const alpha = 0.25
	
	// Exponential moving average: EMA = alpha * new_value + (1 - alpha) * old_EMA
	stat.XPPerSecond = alpha*newXPPerSecond + (1-alpha)*stat.XPPerSecond
	stat.SampleCount++
}

// GetStat returns the XP stat for a creature
func (m *Manager) GetStat(creatureName string) (*XPStat, bool) {
	stat, exists := m.Stats[creatureName]
	return stat, exists
}

// GetAllStats returns all XP stats
func (m *Manager) GetAllStats() map[string]*XPStat {
	return m.Stats
}
