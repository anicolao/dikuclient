package mobile

import (
	"fmt"
)

// Mobile provides a simplified API compatible with gomobile bind
// This exposes functions that can be called from Swift (iOS) or Kotlin/Java (Android)

// StartClient starts the dikuclient and connects to the specified MUD server
// ptyFd is the file descriptor for the pseudo-terminal created by the native side
// Returns error message as string (empty string means success)
func StartClient(host string, port int, ptyFd int) string {
	err := StartClientWithPTY(host, port, ptyFd)
	if err != nil {
		return err.Error()
	}
	return ""
}

// SendText sends text input to the running client
// Returns error message as string (empty string means success)
func SendText(text string) string {
	err := SendInput(text)
	if err != nil {
		return err.Error()
	}
	return ""
}

// Stop stops the running client
// Returns error message as string (empty string means success)
func Stop() string {
	err := StopClient()
	if err != nil {
		return err.Error()
	}
	return ""
}

// CheckRunning checks if a client is currently running
func CheckRunning() bool {
	return IsRunning()
}

// Version returns the client version
func Version() string {
	return "0.1.0-mobile"
}

// GetDefaultPort returns the default MUD port
func GetDefaultPort() int {
	return 4000
}

// ValidateConnection validates host and port parameters
// Returns error message as string (empty string means valid)
func ValidateConnection(host string, port int) string {
	if host == "" {
		return "Host cannot be empty"
	}
	if port < 1 || port > 65535 {
		return fmt.Sprintf("Invalid port: %d (must be 1-65535)", port)
	}
	return ""
}
