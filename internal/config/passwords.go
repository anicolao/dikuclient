package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PasswordStore manages password storage separate from accounts.json
type PasswordStore struct {
	passwords map[string]string // key: "host:port:username", value: password
	filePath  string
	readOnly  bool // true in web mode to prevent writing
}

// NewPasswordStore creates a new password store
func NewPasswordStore(readOnly bool) *PasswordStore {
	return &PasswordStore{
		passwords: make(map[string]string),
		readOnly:  readOnly,
	}
}

// GetPasswordPath returns the path to the .passwords file
func GetPasswordPath() (string, error) {
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

	return filepath.Join(configDir, ".passwords"), nil
}

// Load loads passwords from disk
func (ps *PasswordStore) Load() error {
	passwordPath, err := GetPasswordPath()
	if err != nil {
		return err
	}
	ps.filePath = passwordPath

	data, err := os.ReadFile(passwordPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's ok
			return nil
		}
		return fmt.Errorf("failed to read password file: %w", err)
	}

	// Parse file format: account|password per line
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			ps.passwords[parts[0]] = parts[1]
		}
	}

	return scanner.Err()
}

// Save saves passwords to disk
func (ps *PasswordStore) Save() error {
	if ps.readOnly {
		return fmt.Errorf("password store is read-only (web mode)")
	}

	if ps.filePath == "" {
		var err error
		ps.filePath, err = GetPasswordPath()
		if err != nil {
			return err
		}
	}

	// Build file content
	var lines []string
	for account, password := range ps.passwords {
		lines = append(lines, fmt.Sprintf("%s|%s", account, password))
	}

	data := []byte(strings.Join(lines, "\n"))
	if len(data) > 0 {
		data = append(data, '\n')
	}

	if err := os.WriteFile(ps.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write password file: %w", err)
	}

	return nil
}

// MakeAccountKey creates a key for the password map from account details
func MakeAccountKey(host string, port int, username string) string {
	return fmt.Sprintf("%s:%d:%s", host, port, username)
}

// SetPassword sets a password for an account
func (ps *PasswordStore) SetPassword(host string, port int, username string, password string) {
	key := MakeAccountKey(host, port, username)
	if password == "" {
		delete(ps.passwords, key)
	} else {
		ps.passwords[key] = password
	}
}

// GetPassword retrieves a password for an account
func (ps *PasswordStore) GetPassword(host string, port int, username string) string {
	key := MakeAccountKey(host, port, username)
	return ps.passwords[key]
}

// DeletePassword removes a password for an account
func (ps *PasswordStore) DeletePassword(host string, port int, username string) {
	key := MakeAccountKey(host, port, username)
	delete(ps.passwords, key)
}

// GetAllPasswords returns all stored passwords as a map
func (ps *PasswordStore) GetAllPasswords() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range ps.passwords {
		result[k] = v
	}
	return result
}

// LoadFromMap loads passwords from a map (useful for client-side storage)
func (ps *PasswordStore) LoadFromMap(passwords map[string]string) {
	ps.passwords = make(map[string]string)
	for k, v := range passwords {
		ps.passwords[k] = v
	}
}

// IsReadOnly returns true if the password store is in read-only mode (web mode)
func (ps *PasswordStore) IsReadOnly() bool {
	return ps.readOnly
}
