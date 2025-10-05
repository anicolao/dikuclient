package aliases

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Alias represents a command alias with parameter substitution
type Alias struct {
	ID       string `json:"id"`       // Unique identifier
	Name     string `json:"name"`     // Alias name (e.g., "gat")
	Template string `json:"template"` // Template with placeholders (e.g., "give all <target>")
}

// Manager manages all aliases
type Manager struct {
	Aliases  []*Alias `json:"aliases"`
	filePath string   // Path to aliases.json (not serialized)
}

// NewManager creates a new alias manager
func NewManager() *Manager {
	return &Manager{
		Aliases: make([]*Alias, 0),
	}
}

// GetAliasesPath returns the path to the aliases file
func GetAliasesPath() (string, error) {
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

	return filepath.Join(configDir, "aliases.json"), nil
}

// Load loads aliases from disk
func Load() (*Manager, error) {
	aliasesPath, err := GetAliasesPath()
	if err != nil {
		return nil, err
	}

	return LoadFromPath(aliasesPath)
}

// LoadFromPath loads aliases from a specific path (useful for testing)
func LoadFromPath(aliasesPath string) (*Manager, error) {
	data, err := os.ReadFile(aliasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty manager if file doesn't exist
			m := NewManager()
			m.filePath = aliasesPath
			return m, nil
		}
		return nil, fmt.Errorf("failed to read aliases file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse aliases file: %w", err)
	}
	m.filePath = aliasesPath

	return &m, nil
}

// Save saves aliases to disk
func (m *Manager) Save() error {
	aliasesPath := m.filePath
	if aliasesPath == "" {
		var err error
		aliasesPath, err = GetAliasesPath()
		if err != nil {
			return err
		}
		m.filePath = aliasesPath
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	if err := os.WriteFile(aliasesPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write aliases file: %w", err)
	}

	return nil
}

// Add adds a new alias
func (m *Manager) Add(name, template string) (*Alias, error) {
	// Validate alias name (must be alphanumeric, no spaces)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(name) {
		return nil, fmt.Errorf("alias name must be alphanumeric")
	}

	// Check if alias already exists
	if m.getAliasByName(name) != nil {
		return nil, fmt.Errorf("alias '%s' already exists", name)
	}

	// Generate a unique ID
	id := fmt.Sprintf("alias_%d", len(m.Aliases)+1)
	for m.getAliasByID(id) != nil {
		id = fmt.Sprintf("alias_%d_%d", len(m.Aliases)+1, len(m.Aliases))
	}

	alias := &Alias{
		ID:       id,
		Name:     name,
		Template: template,
	}

	m.Aliases = append(m.Aliases, alias)
	return alias, nil
}

// Remove removes an alias by index (0-based)
func (m *Manager) Remove(index int) error {
	if index < 0 || index >= len(m.Aliases) {
		return fmt.Errorf("invalid alias index: %d", index)
	}

	m.Aliases = append(m.Aliases[:index], m.Aliases[index+1:]...)
	return nil
}

// getAliasByID finds an alias by its ID
func (m *Manager) getAliasByID(id string) *Alias {
	for _, alias := range m.Aliases {
		if alias.ID == id {
			return alias
		}
	}
	return nil
}

// getAliasByName finds an alias by its name
func (m *Manager) getAliasByName(name string) *Alias {
	for _, alias := range m.Aliases {
		if alias.Name == name {
			return alias
		}
	}
	return nil
}

// Expand expands an alias with the given arguments
// Returns the expanded command and true if the command matches an alias,
// or the original command and false if it doesn't
func (m *Manager) Expand(command string) (string, bool) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return command, false
	}

	// Check if first part is an alias
	alias := m.getAliasByName(parts[0])
	if alias == nil {
		return command, false
	}

	// Parse arguments
	args := parts[1:]
	
	// Expand the template with arguments
	expanded := m.expandTemplate(alias.Template, args)
	return expanded, true
}

// expandTemplate expands a template with the given arguments
// Following the parameter substitution rules from the problem statement
func (m *Manager) expandTemplate(template string, args []string) string {
	// Find all placeholders in the template
	placeholderRegex := regexp.MustCompile(`<(\w+)>`)
	placeholders := placeholderRegex.FindAllStringSubmatch(template, -1)
	
	if len(placeholders) == 0 {
		// No placeholders, return template as-is
		return template
	}
	
	// Build a map of placeholder names to their values based on argument count
	varMap := make(map[string]string)
	
	numArgs := len(args)
	numPlaceholders := len(placeholders)
	
	// Apply parameter substitution rules
	if numPlaceholders == 1 {
		// Single placeholder gets all remaining args joined with spaces
		placeholderName := placeholders[0][1]
		if numArgs > 0 {
			varMap[placeholderName] = strings.Join(args, " ")
		} else {
			varMap[placeholderName] = ""
		}
	} else if numPlaceholders == 2 {
		placeholderName1 := placeholders[0][1]
		placeholderName2 := placeholders[1][1]
		
		// Check for special case: <args> as second placeholder
		if placeholderName2 == "args" {
			// <arg1> <args> pattern
			if numArgs >= 1 {
				varMap[placeholderName1] = args[0]
				if numArgs > 1 {
					varMap[placeholderName2] = strings.Join(args[1:], " ")
				} else {
					varMap[placeholderName2] = ""
				}
			} else {
				varMap[placeholderName1] = ""
				varMap[placeholderName2] = ""
			}
		} else {
			// Regular two-argument pattern: <object> <target>
			if numArgs >= 1 {
				varMap[placeholderName1] = args[0]
			} else {
				varMap[placeholderName1] = ""
			}
			if numArgs >= 2 {
				varMap[placeholderName2] = args[1]
			} else {
				varMap[placeholderName2] = ""
			}
		}
	} else if numPlaceholders == 3 {
		placeholderName1 := placeholders[0][1]
		placeholderName2 := placeholders[1][1]
		placeholderName3 := placeholders[2][1]
		
		// Check for <args> as third placeholder
		if placeholderName3 == "args" {
			// <arg1> <arg2> <args> pattern
			if numArgs >= 1 {
				varMap[placeholderName1] = args[0]
			} else {
				varMap[placeholderName1] = ""
			}
			if numArgs >= 2 {
				varMap[placeholderName2] = args[1]
			} else {
				varMap[placeholderName2] = ""
			}
			if numArgs > 2 {
				varMap[placeholderName3] = strings.Join(args[2:], " ")
			} else {
				varMap[placeholderName3] = ""
			}
		} else {
			// Regular three-argument pattern
			if numArgs >= 1 {
				varMap[placeholderName1] = args[0]
			} else {
				varMap[placeholderName1] = ""
			}
			if numArgs >= 2 {
				varMap[placeholderName2] = args[1]
			} else {
				varMap[placeholderName2] = ""
			}
			if numArgs >= 3 {
				varMap[placeholderName3] = args[2]
			} else {
				varMap[placeholderName3] = ""
			}
		}
	} else if numPlaceholders == 4 {
		// <target> <arg1> <arg2> <args> pattern
		placeholderName1 := placeholders[0][1]
		placeholderName2 := placeholders[1][1]
		placeholderName3 := placeholders[2][1]
		placeholderName4 := placeholders[3][1]
		
		if numArgs >= 1 {
			varMap[placeholderName1] = args[0]
		} else {
			varMap[placeholderName1] = ""
		}
		if numArgs >= 2 {
			varMap[placeholderName2] = args[1]
		} else {
			varMap[placeholderName2] = ""
		}
		if numArgs >= 3 {
			varMap[placeholderName3] = args[2]
		} else {
			varMap[placeholderName3] = ""
		}
		if numArgs > 3 {
			varMap[placeholderName4] = strings.Join(args[3:], " ")
		} else {
			varMap[placeholderName4] = ""
		}
	} else {
		// More than 4 placeholders - assign args positionally
		for i, placeholder := range placeholders {
			placeholderName := placeholder[1]
			if i < numArgs {
				varMap[placeholderName] = args[i]
			} else {
				varMap[placeholderName] = ""
			}
		}
	}
	
	// Substitute placeholders in the template
	result := template
	for varName, value := range varMap {
		placeholder := fmt.Sprintf("<%s>", varName)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}
