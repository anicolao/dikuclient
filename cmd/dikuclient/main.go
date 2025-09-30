package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anicolao/dikuclient/internal/config"
	"github.com/anicolao/dikuclient/internal/tui"
	"github.com/anicolao/dikuclient/internal/web"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	host          = flag.String("host", "", "MUD server hostname")
	port          = flag.Int("port", 4000, "MUD server port")
	logAll        = flag.Bool("log-all", false, "Enable logging of MUD output and TUI content")
	accountName   = flag.String("account", "", "Use saved account")
	saveAccount   = flag.Bool("save-account", false, "Save account credentials")
	listAccounts  = flag.Bool("list-accounts", false, "List saved accounts")
	deleteAccount = flag.String("delete-account", "", "Delete saved account")
	webMode       = flag.Bool("web", false, "Start in web mode (HTTP server with WebSocket)")
	webPort       = flag.Int("web-port", 8080, "Web server port")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Handle account management commands
	if *listAccounts {
		handleListAccounts(cfg)
		return
	}

	if *deleteAccount != "" {
		handleDeleteAccount(cfg, *deleteAccount)
		return
	}

	// Handle web mode
	if *webMode {
		fmt.Printf("Starting web server on port %d...\n", *webPort)
		fmt.Printf("Open http://localhost:%d in your browser\n", *webPort)
		if *logAll {
			fmt.Printf("Logging enabled for spawned TUI instances (--log-all)\n")
		}
		if err := web.StartWithLogging(*webPort, *logAll); err != nil {
			fmt.Printf("Error starting web server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Determine connection parameters (terminal mode)
	var finalHost string
	var finalPort int
	var username, password string

	if *accountName != "" {
		// Use saved account
		account, err := cfg.GetAccount(*accountName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		finalHost = account.Host
		finalPort = account.Port
		username = account.Username
		password = account.Password
		fmt.Printf("Using saved account: %s\n", *accountName)
	} else if *host != "" {
		// Use command line parameters
		finalHost = *host
		finalPort = *port

		// If save-account is set, prompt for account name and credentials
		if *saveAccount {
			account, err := promptForAccountDetails(finalHost, finalPort)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			username = account.Username
			password = account.Password

			if err := cfg.AddAccount(*account); err != nil {
				fmt.Printf("Error saving account: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Account '%s' saved successfully.\n", account.Name)
		}
	} else {
		// No host or account specified - show interactive menu
		account, err := selectOrCreateAccount(cfg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if account == nil {
			// User cancelled
			return
		}
		finalHost = account.Host
		finalPort = account.Port
		username = account.Username
		password = account.Password
	}

	var mudLogFile, tuiLogFile, telnetDebugLog *os.File

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

		telnetDebugLog, err = os.Create(fmt.Sprintf("telnet-debug-%s.log", timestamp))
		if err != nil {
			fmt.Printf("Error creating telnet debug log file: %v\n", err)
			os.Exit(1)
		}
		defer telnetDebugLog.Close()

		fmt.Printf("Logging enabled:\n")
		fmt.Printf("  MUD output: mud-output-%s.log\n", timestamp)
		fmt.Printf("  TUI content: tui-content-%s.log\n", timestamp)
		fmt.Printf("  Telnet/UTF-8 debug: telnet-debug-%s.log\n", timestamp)
		fmt.Println()
	}

	// Create the TUI model with auto-login credentials
	model := tui.NewModelWithAuth(finalHost, finalPort, username, password, mudLogFile, tuiLogFile, telnetDebugLog)

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

func handleListAccounts(cfg *config.Config) {
	accounts := cfg.ListAccounts()
	if len(accounts) == 0 {
		fmt.Println("No saved accounts.")
		return
	}

	fmt.Println("Saved accounts:")
	for i, account := range accounts {
		fmt.Printf("  %d. %s (%s:%d)\n", i+1, account.Name, account.Host, account.Port)
		if account.Username != "" {
			fmt.Printf("     Username: %s\n", account.Username)
		}
	}
}

func handleDeleteAccount(cfg *config.Config, name string) {
	if err := cfg.DeleteAccount(name); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Account '%s' deleted successfully.\n", name)
}

func promptForAccountDetails(host string, port int) (*config.Account, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter account name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)

	fmt.Print("Enter username (optional): ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	username = strings.TrimSpace(username)

	fmt.Print("Enter password (optional): ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	password = strings.TrimSpace(password)

	return &config.Account{
		Name:     name,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}, nil
}

func selectOrCreateAccount(cfg *config.Config) (*config.Account, error) {
	accounts := cfg.ListAccounts()

	fmt.Println("\nDikuMUD Client - Account Selection")
	fmt.Println("===================================")

	if len(accounts) > 0 {
		fmt.Println("\nSaved accounts:")
		for i, account := range accounts {
			fmt.Printf("  %d. %s (%s:%d)\n", i+1, account.Name, account.Host, account.Port)
		}
		fmt.Printf("  %d. Connect to new server\n", len(accounts)+1)
		fmt.Printf("  %d. Exit\n", len(accounts)+2)
	} else {
		fmt.Println("\nNo saved accounts found.")
		fmt.Println("  1. Connect to new server")
		fmt.Println("  2. Exit")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nSelect option: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid choice")
	}

	if len(accounts) > 0 {
		if choice >= 1 && choice <= len(accounts) {
			// Use existing account
			return &accounts[choice-1], nil
		} else if choice == len(accounts)+1 {
			// Create new connection
			return createNewAccount(cfg, reader)
		} else if choice == len(accounts)+2 {
			// Exit
			return nil, nil
		}
	} else {
		if choice == 1 {
			// Create new connection
			return createNewAccount(cfg, reader)
		} else if choice == 2 {
			// Exit
			return nil, nil
		}
	}

	return nil, fmt.Errorf("invalid choice")
}

func createNewAccount(cfg *config.Config, reader *bufio.Reader) (*config.Account, error) {
	fmt.Print("\nEnter hostname: ")
	host, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	host = strings.TrimSpace(host)

	fmt.Print("Enter port (default 4000): ")
	portStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	portStr = strings.TrimSpace(portStr)
	port := 4000
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
	}

	fmt.Print("Save this account? (y/n): ")
	save, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	save = strings.TrimSpace(strings.ToLower(save))

	var account config.Account
	if save == "y" || save == "yes" {
		fmt.Print("Enter account name: ")
		name, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)

		fmt.Print("Enter username (optional): ")
		username, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		username = strings.TrimSpace(username)

		fmt.Print("Enter password (optional): ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(password)

		account = config.Account{
			Name:     name,
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
		}

		if err := cfg.AddAccount(account); err != nil {
			return nil, fmt.Errorf("failed to save account: %w", err)
		}
		fmt.Printf("Account '%s' saved.\n", name)
	} else {
		account = config.Account{
			Host: host,
			Port: port,
		}
	}

	return &account, nil
}
