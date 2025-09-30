package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anicolao/dikuclient/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	host = flag.String("host", "", "MUD server hostname")
	port = flag.Int("port", 4000, "MUD server port")
)

func main() {
	flag.Parse()

	if *host == "" {
		fmt.Println("Usage: dikuclient --host <hostname> [--port <port>]")
		fmt.Println("\nExample:")
		fmt.Println("  dikuclient --host mud.server.com --port 4000")
		os.Exit(1)
	}

	// Create the TUI model
	model := tui.NewModel(*host, *port)

	// Create the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
