# DikuMUD Client

A modern, efficient DikuMUD client written in Go with a beautiful Text User Interface (TUI) built using Bubble Tea.

## Features

- **Clean TUI Interface**: Modern terminal UI with panels for game output, stats, inventory, and map
- **MUD Connection**: Connect to any MUD server via telnet protocol
- **Command Input/Output**: Interactive command line for sending commands to the MUD
- **Account Management**: Save and manage multiple MUD accounts with auto-login support
- **Auto-Login**: Automatically login with saved username and password
- **Empty Panels**: Placeholder panels for future features (character stats, inventory, map)

## Installation

### Build from source

```bash
git clone https://github.com/anicolao/dikuclient.git
cd dikuclient
go build -o dikuclient ./cmd/dikuclient
```

## Usage

### Quick Start

Run without arguments to see the interactive account selection menu:

```bash
./dikuclient
```

### Connect to a MUD server

```bash
# Connect to a MUD server directly
./dikuclient --host mud.server.com --port 4000

# Example with a public MUD
./dikuclient --host aardmud.org --port 23
```

### Account Management

```bash
# Save account while connecting
./dikuclient --host mud.server.com --port 4000 --save-account

# List saved accounts
./dikuclient --list-accounts

# Use a saved account
./dikuclient --account myaccount

# Delete a saved account
./dikuclient --delete-account myaccount
```

### Auto-Login

When you save an account with username and password, the client will automatically:
1. Detect login prompts (name, login, account, character)
2. Send your username
3. Detect password prompts
4. Send your password

This allows seamless automatic login to your favorite MUDs.

### Logging

```bash
# Enable logging of MUD output and TUI content
./dikuclient --host mud.server.com --port 4000 --log-all
```

### Controls

- **Type commands** in the input area at the bottom and press `Enter` to send
- **Ctrl+C** or **Esc** to quit the application
- **Arrow keys** to navigate through command history (left/right for cursor positioning)

## Project Structure

```
dikuclient/
├── cmd/
│   └── dikuclient/         # Main entry point
├── internal/
│   ├── client/             # MUD connection logic
│   └── tui/                # TUI application
├── DESIGN.md               # Design documentation
└── README.md               # This file
```

## Development Status

This is a barebones implementation (Phase 1) that includes:
- ✅ Basic TUI framework setup with Bubble Tea
- ✅ MUD connection handling
- ✅ Command input/output
- ✅ Empty placeholder panels for future features

### Future Enhancements (Planned)

- Phase 2: Multi-pane layout, syntax highlighting, configuration system
- Phase 3: Web mode with WebSocket support
- Phase 4: Plugin system, mapping, performance optimizations

## License

See LICENSE file for details.
