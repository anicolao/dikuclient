# DikuMUD Client

A modern, efficient DikuMUD client written in Go with a beautiful Text User Interface (TUI) built using Bubble Tea.

## Features

- **Clean TUI Interface**: Modern terminal UI with panels for game output, stats, inventory, and map
- **Automatic Map Building**: Explores and builds a persistent map as you move through rooms
- **Navigation Commands**: Find your way with `/point` and `/wayfind` commands
- **Aliases**: Create command shortcuts with parameter substitution (e.g., `/alias "gat" "give all <target>"`)
- **Triggers**: Automated responses to MUD output patterns
- **Web Mode with Terminal Emulation**: Run the full TUI in a browser with identical experience to terminal mode
- **MUD Connection**: Connect to any MUD server via telnet protocol
- **Command Input/Output**: Interactive command line for sending commands to the MUD
- **Account Management**: Save and manage multiple MUD accounts with auto-login support
- **Auto-Login**: Automatically login with saved username and password
- **Empty Panels**: Placeholder panels for future features (character stats, inventory)

## Installation

### Run one-off without installing

You can run the client directly without building or installing:

```bash
go run github.com/anicolao/dikuclient/cmd/dikuclient@latest --host mud.server.com --port 4000
```

Or from a cloned repository:

```bash
git clone https://github.com/anicolao/dikuclient.git
cd dikuclient
go run ./cmd/dikuclient --host mud.server.com --port 4000
```

### Install with go install

Install the binary to your `$GOPATH/bin` directory:

```bash
go install github.com/anicolao/dikuclient/cmd/dikuclient@latest
```

Then run it directly (ensure `$GOPATH/bin` is in your `$PATH`):

```bash
dikuclient --host mud.server.com --port 4000
```

### Build from source

```bash
git clone https://github.com/anicolao/dikuclient.git
cd dikuclient
go build -o dikuclient ./cmd/dikuclient
```

## Usage

### Quick Start - Terminal Mode

Run without arguments to see the interactive account selection menu:

```bash
./dikuclient
```

### Quick Start - Web Mode

Start the web server and connect via browser. The web mode runs the full TUI in a terminal emulator:

```bash
# Start web server on port 8080
./dikuclient --web

# Start on custom port
./dikuclient --web --web-port 3000
```

Then open your browser to `http://localhost:8080` (or your custom port). Enter the MUD server host and port, then click Connect. You'll see the complete TUI interface rendered in the browser with all panels and formatting.

**Session Sharing**: In web mode, you can use the `/share` command to get a shareable URL. Anyone who opens this URL in their browser will see and control the same underlying TUI session. This allows you to seamlessly share your MUD session with others for cooperative play or assistance.

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

### Mapping and Navigation

The client automatically builds a map as you explore:
- Detects rooms from MUD output (title, description, exits)
- Links rooms together based on your movement
- Persists map between sessions
- Provides navigation commands to find your way

**Client Commands** (start with `/`):
- `/point <room>` - Show next direction to reach a room
- `/wayfind <room>` - Show full path to reach a room
- `/map` - Show map information
- `/alias "name" "template"` - Create command aliases with parameter substitution
- `/aliases list` - List all defined aliases
- `/trigger "pattern" "action"` - Add triggers that fire on MUD output
- `/triggers list` - List all defined triggers
- `/stop` - Stop any active command queue or auto-walking
- `/share` - Get shareable URL (web mode only)
- `/help` - Show available commands

**Note:** Aliases and triggers support multiple commands separated by semicolons (`;`). Each command is sent sequentially with a 1-second delay.

**Navigation Examples:**
```
> /point temple
To reach 'Temple Square', go: north

> /wayfind market
Path to 'Market Street' (3 steps):
  north -> east -> south
```

**Alias Examples:**
```
> /alias "gat" "give all <target>"
Alias added: "gat" -> "give all <target>"

> gat mary
(sends: give all mary)

> /alias "k" "kill <target>"
Alias added: "k" -> "kill <target>"

> k goblin
(sends: kill goblin)

> /alias "prep" "get all from corpse;sacrifice corpse"
Alias added: "prep" -> "get all from corpse;sacrifice corpse"

> prep
(sends: get all from corpse)
[1 second later]
(sends: sacrifice corpse)
```

**Trigger Example with Multiple Commands:**
```
> /trigger "You are hungry" "eat bread;drink water"
Trigger added: "You are hungry" -> "eat bread;drink water"

[When MUD outputs: "You are hungry"]
[Queue: eat bread]
[Queue: drink water]
```

See [MAPPER.md](MAPPER.md) for detailed mapping documentation and [ALIASES.md](ALIASES.md) for comprehensive alias usage.

### Logging

```bash
# Enable logging of MUD output and TUI content
./dikuclient --host mud.server.com --port 4000 --log-all
```

### Controls

**Terminal Mode:**
- **Type commands** in the input area at the bottom and press `Enter` to send
- **Client commands** start with `/` (e.g., `/point temple`, `/wayfind market`, `/help`)
- **MUD commands** are sent directly (e.g., `north`, `look`, `inventory`)
- **Ctrl+C** or **Esc** to quit the application
- **Arrow keys** to navigate through command history (left/right for cursor positioning)

**Web Mode:**
- Enter MUD host and port in the connection controls
- Click **Connect** to start TUI session in browser
- **Full TUI interface** displays with status bar, sidebars, and all panels
- **Type directly** in the terminal - all keyboard input is forwarded to the TUI
- Click **Disconnect** to close connection and terminate TUI session

## Project Structure

```
dikuclient/
├── cmd/
│   └── dikuclient/         # Main entry point
├── internal/
│   ├── client/             # MUD connection logic
│   ├── config/             # Configuration and account management
│   ├── mapper/             # Automatic mapping and pathfinding
│   ├── tui/                # TUI application
│   └── web/                # Web server and WebSocket handler
├── web/
│   └── static/             # Web interface files (HTML/CSS/JS)
├── DESIGN.md               # Design documentation
├── MAPPER.md               # Mapping system documentation
└── README.md               # This file
```

## Development Status

This implementation includes:
- ✅ Basic TUI framework setup with Bubble Tea
- ✅ MUD connection handling
- ✅ Command input/output
- ✅ Web mode with WebSocket support (Phase 3)
- ✅ Account management with auto-login
- ✅ Automatic map building and room tracking
- ✅ Navigation commands (`/point`, `/wayfind`)
- ✅ Map persistence between sessions
- 🔲 Empty placeholder panels for stats and inventory

### Future Enhancements (Planned)

- Phase 2: Syntax highlighting, enhanced configuration system
- Phase 4: Plugin system, visual map display, performance optimizations
- Additional mapper features: custom movement aliases, special room marking, map sharing

## License

See LICENSE file for details.
