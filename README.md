# DikuMUD Client

A modern, efficient DikuMUD client written in Go with a beautiful Text User Interface (TUI) built using Bubble Tea.

## Features

- **Clean TUI Interface**: Modern terminal UI with panels for game output, stats, inventory, and map
- **MUD Connection**: Connect to any MUD server via telnet protocol
- **Command Input/Output**: Interactive command line for sending commands to the MUD
- **Empty Panels**: Placeholder panels for future features (character stats, inventory, map)

## Installation

### Build from source

```bash
git clone https://github.com/anicolao/dikuclient.git
cd dikuclient
go build -o dikuclient ./cmd/dikuclient
```

## Usage

```bash
# Connect to a MUD server
./dikuclient --host mud.server.com --port 4000

# Example with a public MUD
./dikuclient --host aardmud.org --port 23
```

### Controls

- **Type commands** in the input area at the bottom and press `Enter` to send
- **Ctrl+C** or **Esc** to quit the application

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
