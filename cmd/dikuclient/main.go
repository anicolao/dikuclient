package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/anicolao/dikuclient/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	host   = flag.String("host", "", "MUD server hostname")
	port   = flag.Int("port", 4000, "MUD server port")
	logAll = flag.Bool("log-all", false, "Enable logging of MUD output and TUI content")
)

func main() {
	flag.Parse()

	if *host == "" {
		fmt.Println("Usage: dikuclient --host <hostname> [--port <port>] [--log-all]")
		fmt.Println("\nExample:")
		fmt.Println("  dikuclient --host mud.server.com --port 4000")
		fmt.Println("  dikuclient --host mud.server.com --port 4000 --log-all")
		os.Exit(1)
	}

	var mudLogFile, tuiLogFile *os.File
	var err error

	// Create log files if --log-all flag is set
	if *logAll {
		timestamp := time.Now().Format("20060102-150405")
		
		mudLogFile, err = os.Create(fmt.Sprintf("mud-output-%s.log", timestamp))
		if err != nil {
			fmt.Printf("Error creating MUD log file: %v\n", err)
			os.Exit(1)
		}
		defer mudLogFile.Close()
		
		tuiLogFile, err = os.Create(fmt.Sprintf("tui-content-%s.log", timestamp))
		if err != nil {
			fmt.Printf("Error creating TUI log file: %v\n", err)
			os.Exit(1)
		}
		defer tuiLogFile.Close()
		
		fmt.Printf("Logging enabled:\n")
		fmt.Printf("  MUD output: mud-output-%s.log\n", timestamp)
		fmt.Printf("  TUI content: tui-content-%s.log\n", timestamp)
		fmt.Println()
	}

	// Create the TUI model
	model := tui.NewModel(*host, *port, mudLogFile, tuiLogFile)

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
