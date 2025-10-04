package triggers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Trigger represents a pattern-action pair
type Trigger struct {
	ID      string         `json:"id"`      // Unique identifier
	Pattern string         `json:"pattern"` // Pattern to match (may contain <variable> placeholders)
	Action  string         `json:"action"`  // Action to execute (may contain <variable> placeholders)
	regex   *regexp.Regexp // Compiled regex (not serialized)
}

// Manager manages all triggers
type Manager struct {
	Triggers []*Trigger `json:"triggers"`
	filePath string     // Path to triggers.json (not serialized)
}

// NewManager creates a new trigger manager
func NewManager() *Manager {
	return &Manager{
		Triggers: make([]*Trigger, 0),
	}
}

// GetTriggersPath returns the path to the triggers file
func GetTriggersPath() (string, error) {
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

	return filepath.Join(configDir, "triggers.json"), nil
}

// Load loads triggers from disk
func Load() (*Manager, error) {
	triggersPath, err := GetTriggersPath()
	if err != nil {
		return nil, err
	}

	return LoadFromPath(triggersPath)
}

// LoadFromPath loads triggers from a specific path (useful for testing)
func LoadFromPath(triggersPath string) (*Manager, error) {
	data, err := os.ReadFile(triggersPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty manager if file doesn't exist
			m := NewManager()
			m.filePath = triggersPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read triggers file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse triggers file: %w", err)
	}
	m.filePath = triggersPath

	// Compile regex patterns
	for _, trigger := range m.Triggers {
		if err := trigger.compilePattern(); err != nil {
			return nil, fmt.Errorf("failed to compile pattern for trigger %s: %w", trigger.ID, err)
		}
	}

	return &m, nil
}

// Save saves triggers to disk
func (m *Manager) Save() error {
	triggersPath := m.filePath
	if triggersPath == "" {
		var err error
		triggersPath, err = GetTriggersPath()
		if err != nil {
			return err
		}
		m.filePath = triggersPath
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal triggers: %w", err)
	}

	if err := os.WriteFile(triggersPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write triggers file: %w", err)
	}

	return nil
}

// Add adds a new trigger
func (m *Manager) Add(pattern, action string) (*Trigger, error) {
	// Generate a unique ID
	id := fmt.Sprintf("trigger_%d", len(m.Triggers)+1)
	for m.getTriggerByID(id) != nil {
		id = fmt.Sprintf("trigger_%d_%d", len(m.Triggers)+1, len(m.Triggers))
	}

	trigger := &Trigger{
		ID:      id,
		Pattern: pattern,
		Action:  action,
	}

	if err := trigger.compilePattern(); err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %w", err)
	}

	m.Triggers = append(m.Triggers, trigger)
	return trigger, nil
}

// Remove removes a trigger by index (0-based)
func (m *Manager) Remove(index int) error {
	if index < 0 || index >= len(m.Triggers) {
		return fmt.Errorf("invalid trigger index: %d", index)
	}

	m.Triggers = append(m.Triggers[:index], m.Triggers[index+1:]...)
	return nil
}

// getTriggerByID finds a trigger by its ID
func (m *Manager) getTriggerByID(id string) *Trigger {
	for _, trigger := range m.Triggers {
		if trigger.ID == id {
			return trigger
		}
	}
	return nil
}

// Match checks if a line matches any trigger and returns the action to execute
func (m *Manager) Match(line string) []string {
	actions := make([]string, 0)

	for _, trigger := range m.Triggers {
		if action := trigger.match(line); action != "" {
			actions = append(actions, action)
		}
	}

	return actions
}

// compilePattern compiles the pattern into a regex
// Converts <variable> placeholders to regex capture groups
func (t *Trigger) compilePattern() error {
	// Escape special regex characters except for our placeholders
	pattern := t.Pattern

	// Find all <variable> placeholders
	placeholderRegex := regexp.MustCompile(`<(\w+)>`)

	// Escape the pattern first, but preserve our placeholders
	// Replace <variable> with a temporary marker
	tempPattern := placeholderRegex.ReplaceAllString(pattern, "§§§PLACEHOLDER§§§")

	// Escape regex special characters
	tempPattern = regexp.QuoteMeta(tempPattern)

	// Replace temporary markers with regex capture groups
	// Use (.+?) for non-greedy matching of any characters
	tempPattern = strings.ReplaceAll(tempPattern, "§§§PLACEHOLDER§§§", "(.+?)")

	// Compile the regex
	regex, err := regexp.Compile(tempPattern)
	if err != nil {
		return err
	}

	t.regex = regex
	return nil
}

// match checks if a line matches this trigger and returns the action with substitutions
func (t *Trigger) match(line string) string {
	if t.regex == nil {
		return ""
	}

	matches := t.regex.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}

	// matches[0] is the full match, matches[1:] are the capture groups
	capturedValues := matches[1:]

	// Find variable names in the pattern
	placeholderRegex := regexp.MustCompile(`<(\w+)>`)
	varNames := placeholderRegex.FindAllStringSubmatch(t.Pattern, -1)

	if len(varNames) != len(capturedValues) {
		return ""
	}

	// Build a map of variable name to captured value
	varMap := make(map[string]string)
	for i, varName := range varNames {
		if i < len(capturedValues) {
			// Replace spaces with dots in the captured value
			value := strings.ReplaceAll(capturedValues[i], " ", ".")
			varMap[varName[1]] = value // varName[1] is the variable name without <>
		}
	}

	// Substitute variables in the action
	action := t.Action
	for varName, value := range varMap {
		placeholder := fmt.Sprintf("<%s>", varName)
		action = strings.ReplaceAll(action, placeholder, value)
	}

	return action
}
