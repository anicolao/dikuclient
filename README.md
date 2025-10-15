# DikuMUD Client

A modern, efficient DikuMUD client written in Go with a beautiful Text User Interface (TUI) built using Bubble Tea.

## Features

- **Clean TUI Interface**: Modern terminal UI with panels for game output, stats, inventory, and map
- **Automatic Map Building**: Explores and builds a persistent map as you move through rooms
- **Navigation Commands**: Find your way with `/point`, `/wayfind`, and `/go` commands
- **Room Management**: Search, list, and find nearby rooms with `/rooms`, `/nearby`, and `/legend`
- **Aliases**: Create command shortcuts with parameter substitution (e.g., `/alias "gat" "give all <target>"`)
- **Triggers**: Automated responses to MUD output patterns
- **Tick Timer**: Automatically track tick times and execute commands at specific tick values (e.g., cast heal at T:5)
- **Web Mode with Terminal Emulation**: Run the full TUI in a browser with identical experience to terminal mode
- **Session Sharing**: Share your web session with others using the `/share` command
- **MUD Connection**: Connect to any MUD server via telnet protocol
- **Command Input/Output**: Interactive command line with history and search (Ctrl+R)
- **Account Management**: Save and manage multiple MUD accounts with auto-login support
- **Auto-Login**: Automatically login with saved username and password

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
- `/go <room>` - Auto-walk to a room (one step per second)
- `/stop` - Stop auto-walk or command queue
- `/map` - Show map information
- `/rooms [filter]` - List all known rooms (optionally filtered)
- `/nearby` - List all rooms within 5 steps
- `/legend` - List all rooms currently on the map
- `/alias "name" "template"` - Create command aliases with parameter substitution
- `/aliases list` - List all defined aliases
- `/aliases remove <n>` - Remove alias by number
- `/trigger "pattern" "action"` - Add triggers that fire on MUD output
- `/triggers list` - List all defined triggers
- `/triggers remove <n>` - Remove trigger by number
- `/ticktrigger <time> "commands"` - Add tick-based triggers (e.g., `/ticktrigger 5 "cast 'heal'"`)
- `/ticktriggers list` - List all tick triggers
- `/ticktriggers remove <n>` - Remove tick trigger by number
- `/share` - Get shareable URL (web mode only)
- `/help [command]` - Show available commands or detailed help for a specific command

**Note:** Aliases, triggers, and tick triggers support multiple commands separated by semicolons (`;`). Each command is sent sequentially with a 1-second delay.

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

**Tick Timer Examples:**
```
> /ticktrigger 5 "cast 'heal'"
Tick trigger added: T:5 -> "cast 'heal'"

> /ticktrigger 4 "cast 'bless';say Ready!"
Tick trigger added: T:4 -> "cast 'bless';say Ready!"

[When tick time reaches T:5]
[Queue: cast 'heal']

[When tick time reaches T:4]
[Queue: cast 'bless']
[Queue: say Ready!]
```

See [MAPPER.md](MAPPER.md) for detailed mapping documentation, [ALIASES.md](ALIASES.md) for comprehensive alias usage, and [TICK_TIMER_FEATURE.md](TICK_TIMER_FEATURE.md) for tick timer details.

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
│   ├── ticktimer/          # Tick timer and tick-based triggers
│   ├── tui/                # TUI application
│   └── web/                # Web server and WebSocket handler
├── web/
│   └── static/             # Web interface files (HTML/CSS/JS)
├── DESIGN.md               # Design documentation
├── MAPPER.md               # Mapping system documentation
├── TICK_TIMER_FEATURE.md   # Tick timer feature documentation
└── README.md               # This file
```

## Development Status

This implementation is feature-complete for a modern MUD client:
- ✅ TUI framework with Bubble Tea
- ✅ MUD connection handling via telnet
- ✅ Command input/output with history and search
- ✅ Web mode with WebSocket support and session sharing
- ✅ Account management with auto-login
- ✅ Automatic map building and room tracking
- ✅ Navigation commands (`/point`, `/wayfind`, `/go`)
- ✅ Room search and discovery (`/rooms`, `/nearby`, `/legend`)
- ✅ Map persistence between sessions
- ✅ Aliases with parameter substitution and multi-command support
- ✅ Triggers with pattern matching and variable capture
- ✅ Tick timer with automatic interval detection and tick-based triggers

## License

See LICENSE file for details.
