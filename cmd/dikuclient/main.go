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
	mapDebug      = flag.Bool("map-debug", false, "Enable mapper debug output")
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

	// Load password store (read-only in web mode to prevent writing)
	webSessionID := os.Getenv("DIKUCLIENT_WEB_SESSION_ID")
	isWebMode := webSessionID != ""
	passwordStore := config.NewPasswordStore(isWebMode)
	
	if err := passwordStore.Load(); err != nil {
		fmt.Printf("Error loading passwords: %v\n", err)
		// Continue anyway - passwords file might not exist yet
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

	// Check if web mode has specified a server for character selection
	webServer := os.Getenv("DIKUCLIENT_WEB_SERVER")
	webPort := os.Getenv("DIKUCLIENT_WEB_PORT")
	
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
		password = passwordStore.GetPassword(account.Host, account.Port, account.Username)
		fmt.Printf("Using saved account: %s\n", *accountName)
	} else if *host != "" {
		// Use command line parameters
		finalHost = *host
		finalPort = *port

		// If save-account is set, prompt for account name and credentials
		if *saveAccount {
			account, err := promptForAccountDetails(finalHost, finalPort, passwordStore)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			username = account.Username
			password = account.Password

			// Save account (without password)
			if err := cfg.AddAccount(*account); err != nil {
				fmt.Printf("Error saving account: %v\n", err)
				os.Exit(1)
			}
			
			// Save password separately (only in non-web mode)
			if password != "" && !passwordStore.IsReadOnly() {
				passwordStore.SetPassword(account.Host, account.Port, account.Username, password)
				if err := passwordStore.Save(); err != nil {
					fmt.Printf("Error saving password: %v\n", err)
					os.Exit(1)
				}
			}
			
			if passwordStore.IsReadOnly() {
				fmt.Printf("Account '%s' saved. Password will be captured automatically during login.\n", account.Name)
			} else {
				fmt.Printf("Account '%s' saved successfully.\n", account.Name)
			}

			// Flush output before TUI initialization
			// This prevents escape codes from being displayed literally
			os.Stdout.Sync()
		}
	} else if webServer != "" && webPort != "" {
		// Web mode with server specified - show character selection for that server
		portNum, err := strconv.Atoi(webPort)
		if err != nil {
			fmt.Printf("Error: invalid web port: %v\n", err)
			os.Exit(1)
		}
		
		server := &config.Server{
			Name: fmt.Sprintf("%s:%s", webServer, webPort),
			Host: webServer,
			Port: portNum,
		}
		
		reader := bufio.NewReader(os.Stdin)
		account, err := selectOrCreateCharacter(cfg, passwordStore, server, reader)
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
		password = passwordStore.GetPassword(account.Host, account.Port, account.Username)

		// Flush output before TUI initialization
		// This prevents escape codes from being displayed literally
		os.Stdout.Sync()
	} else {
		// No host or account specified - show interactive menu
		account, err := selectOrCreateAccount(cfg, passwordStore)
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
		password = passwordStore.GetPassword(account.Host, account.Port, account.Username)

		// Flush output before TUI initialization
		// This prevents escape codes from being displayed literally
		os.Stdout.Sync()
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
	model := tui.NewModelWithAuth(finalHost, finalPort, username, password, mudLogFile, tuiLogFile, telnetDebugLog, *mapDebug)

	// Create the Bubble Tea program
	// Explicitly specify input/output to ensure proper terminal handling
	p := tea.NewProgram(
		&model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
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

func promptForAccountDetails(host string, port int, passwordStore *config.PasswordStore) (*config.Account, error) {
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

	var password string
	// Only prompt for password in non-web mode
	// In web mode, password will be captured during login automatically
	if !passwordStore.IsReadOnly() {
		fmt.Print("Enter password (optional): ")
		passwordInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(passwordInput)
	}

	return &config.Account{
		Name:     name,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}, nil
}

func selectOrCreateAccount(cfg *config.Config, passwordStore *config.PasswordStore) (*config.Account, error) {
	reader := bufio.NewReader(os.Stdin)
	
	// Step 1: Select server or character
	selection, err := selectServerOrCharacter(cfg, passwordStore, reader)
	if err != nil {
		return nil, err
	}
	if selection == nil {
		// User cancelled
		return nil, nil
	}
	
	// If a character was directly selected, return it
	if selection.account != nil {
		return selection.account, nil
	}
	
	// Otherwise, go to character selection for the server
	return selectOrCreateCharacter(cfg, passwordStore, selection.server, reader)
}

// serverOrCharacterSelection represents either a server or a direct character selection
type serverOrCharacterSelection struct {
	server  *config.Server
	account *config.Account
}

func selectServerOrCharacter(cfg *config.Config, passwordStore *config.PasswordStore, reader *bufio.Reader) (*serverOrCharacterSelection, error) {
	servers := cfg.ListServers()
	characters := cfg.ListCharacters()
	accounts := cfg.ListAccounts() // Legacy accounts

	fmt.Println("\nDikuMUD Client - Server Selection")
	fmt.Println("==================================")

	optionNum := 1
	
	// Option 1: Add a new server
	fmt.Printf("  %d. Add a new server\n", optionNum)
	optionNum++
	
	// List all servers
	serverStartIdx := optionNum
	if len(servers) > 0 {
		fmt.Println("\nServers:")
		for _, server := range servers {
			fmt.Printf("  %d. %s (%s:%d)\n", optionNum, server.Name, server.Host, server.Port)
			optionNum++
		}
	}
	serverEndIdx := optionNum
	
	// List all characters with their servers
	charStartIdx := optionNum
	if len(characters) > 0 {
		fmt.Println("\nCharacters:")
		for _, char := range characters {
			displayName := char.Username
			if displayName == "" {
				displayName = "(unnamed)"
			}
			fmt.Printf("  %d. %s on %s:%d\n", optionNum, displayName, char.Host, char.Port)
			optionNum++
		}
	}
	charEndIdx := optionNum
	
	// List legacy accounts (for backward compatibility)
	accountStartIdx := optionNum
	if len(accounts) > 0 {
		fmt.Println("\nLegacy accounts:")
		for _, account := range accounts {
			displayName := account.Name
			if account.Username != "" {
				displayName = fmt.Sprintf("%s (%s)", account.Name, account.Username)
			}
			fmt.Printf("  %d. %s (%s:%d)\n", optionNum, displayName, account.Host, account.Port)
			optionNum++
		}
	}
	accountEndIdx := optionNum
	
	// Exit option
	fmt.Printf("  %d. Exit\n", optionNum)
	exitOption := optionNum

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

	// Add a new server
	if choice == 1 {
		server, err := createNewServer(cfg, reader)
		if err != nil {
			return nil, err
		}
		return &serverOrCharacterSelection{server: server}, nil
	}
	
	// Select an existing server
	if choice >= serverStartIdx && choice < serverEndIdx {
		idx := choice - serverStartIdx
		return &serverOrCharacterSelection{server: &servers[idx]}, nil
	}
	
	// Select a character directly - return account for immediate connection
	if choice >= charStartIdx && choice < charEndIdx {
		idx := choice - charStartIdx
		char := characters[idx]
		password := passwordStore.GetPassword(char.Host, char.Port, char.Username)
		account := &config.Account{
			Name:     char.Username,
			Host:     char.Host,
			Port:     char.Port,
			Username: char.Username,
			Password: password,
		}
		return &serverOrCharacterSelection{account: account}, nil
	}
	
	// Select a legacy account - return account for immediate connection
	if choice >= accountStartIdx && choice < accountEndIdx {
		idx := choice - accountStartIdx
		acc := accounts[idx]
		password := passwordStore.GetPassword(acc.Host, acc.Port, acc.Username)
		acc.Password = password
		return &serverOrCharacterSelection{account: &acc}, nil
	}
	
	// Exit
	if choice == exitOption {
		return nil, nil
	}

	return nil, fmt.Errorf("invalid choice")
}

func selectOrCreateCharacter(cfg *config.Config, passwordStore *config.PasswordStore, server *config.Server, reader *bufio.Reader) (*config.Account, error) {
	characters := cfg.ListCharactersForServer(server.Host, server.Port)
	
	fmt.Printf("\nCharacter Selection for %s (%s:%d)\n", server.Name, server.Host, server.Port)
	fmt.Println("====================================")
	
	optionNum := 1
	
	// Option 1: Create a new character
	fmt.Printf("  %d. Create a new character\n", optionNum)
	optionNum++
	
	// List existing characters
	charStartIdx := optionNum
	if len(characters) > 0 {
		fmt.Println("\nExisting characters:")
		for _, char := range characters {
			displayName := char.Username
			if displayName == "" {
				displayName = "(unnamed)"
			}
			fmt.Printf("  %d. %s\n", optionNum, displayName)
			optionNum++
		}
	}
	charEndIdx := optionNum
	
	// Back option
	fmt.Printf("  %d. Back to server selection\n", optionNum)
	backOption := optionNum
	
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
	
	// Create a new character
	if choice == 1 {
		return createNewCharacter(cfg, passwordStore, server, reader)
	}
	
	// Select an existing character
	if choice >= charStartIdx && choice < charEndIdx {
		idx := choice - charStartIdx
		char := characters[idx]
		password := passwordStore.GetPassword(char.Host, char.Port, char.Username)
		return &config.Account{
			Name:     char.Username,
			Host:     char.Host,
			Port:     char.Port,
			Username: char.Username,
			Password: password,
		}, nil
	}
	
	// Back to server selection
	if choice == backOption {
		return selectOrCreateAccount(cfg, passwordStore)
	}
	
	return nil, fmt.Errorf("invalid choice")
}

func createNewServer(cfg *config.Config, reader *bufio.Reader) (*config.Server, error) {
	fmt.Print("\nEnter server name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	
	fmt.Print("Enter hostname: ")
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
	
	server := config.Server{
		Name: name,
		Host: host,
		Port: port,
	}
	
	// Save the server
	if err := cfg.AddServer(server); err != nil {
		return nil, fmt.Errorf("failed to save server: %w", err)
	}
	
	fmt.Printf("Server '%s' saved.\n", name)
	return &server, nil
}

func createNewCharacter(cfg *config.Config, passwordStore *config.PasswordStore, server *config.Server, reader *bufio.Reader) (*config.Account, error) {
	fmt.Print("\nEnter username (optional): ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	username = strings.TrimSpace(username)
	
	var password string
	// Only prompt for password in non-web mode
	if !passwordStore.IsReadOnly() {
		fmt.Print("Enter password (optional): ")
		passwordInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(passwordInput)
	}
	
	// If character has a username, save it by default
	if username != "" {
		character := config.Character{
			Host:     server.Host,
			Port:     server.Port,
			Username: username,
		}
		
		if err := cfg.AddCharacter(character); err != nil {
			return nil, fmt.Errorf("failed to save character: %w", err)
		}
		
		// Save password separately (only in non-web mode)
		if password != "" && !passwordStore.IsReadOnly() {
			passwordStore.SetPassword(server.Host, server.Port, username, password)
			if err := passwordStore.Save(); err != nil {
				return nil, fmt.Errorf("failed to save password: %w", err)
			}
		}
		
		if passwordStore.IsReadOnly() {
			fmt.Printf("Character '%s' saved. Password will be captured automatically during login.\n", username)
		} else {
			fmt.Printf("Character '%s' saved.\n", username)
		}
	}
	
	return &config.Account{
		Name:     username,
		Host:     server.Host,
		Port:     server.Port,
		Username: username,
		Password: password,
	}, nil
}

func createNewAccount(cfg *config.Config, passwordStore *config.PasswordStore, reader *bufio.Reader) (*config.Account, error) {
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

		var password string
		// Only prompt for password in non-web mode
		// In web mode, password will be captured during login automatically
		if !passwordStore.IsReadOnly() {
			fmt.Print("Enter password (optional): ")
			passwordInput, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			password = strings.TrimSpace(passwordInput)
		}

		account = config.Account{
			Name:     name,
			Host:     host,
			Port:     port,
			Username: username,
			Password: password, // Will be saved separately (not in JSON)
		}

		// Save account (without password in JSON)
		if err := cfg.AddAccount(account); err != nil {
			return nil, fmt.Errorf("failed to save account: %w", err)
		}
		
		// Save password separately (only in non-web mode)
		if password != "" && !passwordStore.IsReadOnly() {
			passwordStore.SetPassword(host, port, username, password)
			if err := passwordStore.Save(); err != nil {
				return nil, fmt.Errorf("failed to save password: %w", err)
			}
		}
		
		if passwordStore.IsReadOnly() {
			fmt.Printf("Account '%s' saved. Password will be captured automatically during login.\n", name)
		} else {
			fmt.Printf("Account '%s' saved.\n", name)
		}
	} else {
		account = config.Account{
			Host: host,
			Port: port,
		}
	}

	return &account, nil
}
