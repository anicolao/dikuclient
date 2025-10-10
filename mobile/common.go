package mobile

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/anicolao/dikuclient/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// ClientInstance represents a running dikuclient instance
type ClientInstance struct {
	program *tea.Program
	model   *tui.Model
	ptyMaster *os.File
	ptyName string
	mu      sync.Mutex
	running bool
}

var (
	currentInstance *ClientInstance
	instanceMu      sync.Mutex
)

// createPTY creates a pseudo-terminal for the TUI
func createPTY() (*os.File, string, error) {
	// This is a placeholder - actual PTY creation is platform-specific
	// On iOS/Android, the native side creates the PTY and we'll get file descriptors
	return nil, "", fmt.Errorf("PTY creation must be done by platform-specific code")
}

// StartClientWithPTY starts the dikuclient with a given PTY file descriptor
func StartClientWithPTY(host string, port int, ptyFd int) error {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	if currentInstance != nil && currentInstance.running {
		return fmt.Errorf("client already running")
	}

	// Convert file descriptor to *os.File
	ptyFile := os.NewFile(uintptr(ptyFd), "pty")
	if ptyFile == nil {
		return fmt.Errorf("invalid PTY file descriptor")
	}

	// Create the TUI model
	model := tui.NewModelWithAuth(host, port, "", "", nil, nil, nil, false)

	// Create the Bubble Tea program with custom I/O
	program := tea.NewProgram(
		&model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(ptyFile),
		tea.WithOutput(ptyFile),
	)

	currentInstance = &ClientInstance{
		program:   program,
		model:     &model,
		ptyMaster: ptyFile,
		running:   true,
	}

	// Run the program in a goroutine
	go func() {
		if _, err := program.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		}
		instanceMu.Lock()
		if currentInstance != nil {
			currentInstance.running = false
		}
		instanceMu.Unlock()
	}()

	return nil
}

// SendInput sends keyboard input to the running client
func SendInput(input string) error {
	instanceMu.Lock()
	instance := currentInstance
	instanceMu.Unlock()

	if instance == nil || !instance.running {
		return fmt.Errorf("no client running")
	}

	// Write input to PTY
	if instance.ptyMaster != nil {
		_, err := instance.ptyMaster.Write([]byte(input))
		return err
	}

	return fmt.Errorf("PTY not available")
}

// StopClient stops the running client
func StopClient() error {
	instanceMu.Lock()
	instance := currentInstance
	currentInstance = nil
	instanceMu.Unlock()

	if instance == nil {
		return fmt.Errorf("no client running")
	}

	if instance.program != nil {
		instance.program.Quit()
	}

	if instance.ptyMaster != nil {
		instance.ptyMaster.Close()
	}

	instance.running = false
	return nil
}

// IsRunning checks if a client instance is currently running
func IsRunning() bool {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return currentInstance != nil && currentInstance.running
}

// ReadOutput reads output from the PTY (for testing/debugging)
func ReadOutput(buf []byte) (int, error) {
	instanceMu.Lock()
	instance := currentInstance
	instanceMu.Unlock()

	if instance == nil || !instance.running {
		return 0, fmt.Errorf("no client running")
	}

	if instance.ptyMaster != nil {
		return instance.ptyMaster.Read(buf)
	}

	return 0, io.EOF
}
