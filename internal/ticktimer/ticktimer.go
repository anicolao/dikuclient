package ticktimer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TickTrigger represents a trigger that fires at a specific tick time
type TickTrigger struct {
	TickTime int    `json:"tick_time"` // Time in seconds (e.g., 4 means T:4)
	Commands string `json:"commands"`  // Commands to execute (can be multiple separated by ;)
}

// Manager manages tick timing and tick-based triggers
type Manager struct {
	TickInterval   int           `json:"tick_interval"`    // Tick interval in seconds (e.g., 60 or 75)
	LastTickTime   time.Time     `json:"last_tick_time"`   // When the last tick occurred (or was estimated)
	LastSeenValue  int           `json:"last_seen_value"`  // Last T:NN value seen in prompt
	LastUpdateTime time.Time     `json:"last_update_time"` // When LastSeenValue was captured
	TickTriggers   []TickTrigger `json:"tick_triggers"`    // Triggers to fire at specific tick times
	filePath       string        // Path to tick timer config file (not serialized)
}

// NewManager creates a new tick timer manager
func NewManager(tickInterval int) *Manager {
	return &Manager{
		TickInterval:   tickInterval,
		LastTickTime:   time.Time{},
		LastSeenValue:  0,
		LastUpdateTime: time.Time{},
		TickTriggers:   make([]TickTrigger, 0),
	}
}

// GetTickTimerPath returns the path to the tick timer config file for a specific server
func GetTickTimerPath(host string, port int) (string, error) {
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

	filename := fmt.Sprintf("ticktimer_%s_%d.json", host, port)
	return filepath.Join(configDir, filename), nil
}

// Load loads tick timer config from disk for a specific server
func Load(host string, port int, tickInterval int) (*Manager, error) {
	tickTimerPath, err := GetTickTimerPath(host, port)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tickTimerPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return new manager if file doesn't exist
			m := NewManager(tickInterval)
			m.filePath = tickTimerPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read tick timer file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse tick timer file: %w", err)
	}
	m.filePath = tickTimerPath

	// Update tick interval if it has changed
	if tickInterval > 0 && m.TickInterval != tickInterval {
		m.TickInterval = tickInterval
	}

	return &m, nil
}

// Save saves tick timer config to disk
func (m *Manager) Save() error {
	if m.filePath == "" {
		return fmt.Errorf("tick timer path not set")
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tick timer: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write tick timer file: %w", err)
	}

	return nil
}

// UpdateFromPrompt updates the tick timer based on a T:NN value seen in the prompt
func (m *Manager) UpdateFromPrompt(tickValue int) {
	now := time.Now()
	m.LastSeenValue = tickValue
	m.LastUpdateTime = now

	// If we have a tick interval, estimate when the last tick occurred
	if m.TickInterval > 0 {
		// Calculate when the tick should have occurred
		// If T:24, and interval is 75, then tick was 75-24=51 seconds ago
		secondsSinceLastTick := m.TickInterval - tickValue
		m.LastTickTime = now.Add(-time.Duration(secondsSinceLastTick) * time.Second)
	}
}

// GetCurrentTickTime returns the estimated current tick time (seconds until next tick)
func (m *Manager) GetCurrentTickTime() int {
	if m.TickInterval <= 0 || m.LastUpdateTime.IsZero() {
		return 0
	}

	now := time.Now()
	elapsedSinceUpdate := now.Sub(m.LastUpdateTime).Seconds()
	
	// Calculate current tick time
	currentTickTime := m.LastSeenValue - int(elapsedSinceUpdate)
	
	// Handle wrap-around (if we passed a tick)
	if currentTickTime <= 0 {
		currentTickTime = m.TickInterval + currentTickTime
		// Ensure it's positive
		for currentTickTime <= 0 {
			currentTickTime += m.TickInterval
		}
	}

	return currentTickTime
}

// GetSecondsSinceLastTick returns how many seconds have passed since the last tick
func (m *Manager) GetSecondsSinceLastTick() float64 {
	if m.LastTickTime.IsZero() {
		return 0
	}
	return time.Since(m.LastTickTime).Seconds()
}

// GetSecondsUntilNextTick returns how many seconds until the next tick
func (m *Manager) GetSecondsUntilNextTick() float64 {
	if m.TickInterval <= 0 || m.LastTickTime.IsZero() {
		return 0
	}
	
	secondsSinceTick := m.GetSecondsSinceLastTick()
	secondsUntilTick := float64(m.TickInterval) - secondsSinceTick
	
	// Handle case where we've passed the tick
	if secondsUntilTick <= 0 {
		secondsUntilTick = float64(m.TickInterval) - (secondsSinceTick - float64(m.TickInterval))
	}
	
	return secondsUntilTick
}

// AddTrigger adds a new tick trigger
func (m *Manager) AddTrigger(tickTime int, commands string) error {
	if tickTime < 0 || tickTime > m.TickInterval {
		return fmt.Errorf("tick time must be between 0 and %d", m.TickInterval)
	}

	trigger := TickTrigger{
		TickTime: tickTime,
		Commands: commands,
	}

	m.TickTriggers = append(m.TickTriggers, trigger)
	return nil
}

// RemoveTrigger removes a tick trigger by index (0-based)
func (m *Manager) RemoveTrigger(index int) error {
	if index < 0 || index >= len(m.TickTriggers) {
		return fmt.Errorf("invalid trigger index: %d", index)
	}

	m.TickTriggers = append(m.TickTriggers[:index], m.TickTriggers[index+1:]...)
	return nil
}

// GetTriggersToFire returns all triggers that should fire at the current estimated tick time
// lastFiredTickTime is used to avoid firing the same trigger multiple times
func (m *Manager) GetTriggersToFire(lastFiredTickTime int) []string {
	currentTickTime := m.GetCurrentTickTime()
	
	if currentTickTime == 0 || currentTickTime == lastFiredTickTime {
		return nil
	}

	var commands []string
	for _, trigger := range m.TickTriggers {
		if trigger.TickTime == currentTickTime {
			commands = append(commands, trigger.Commands)
		}
	}

	return commands
}
