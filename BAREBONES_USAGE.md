# Barebones Usage Guide

This document describes how to use the DikuMUD Client as a simple telnet replacement - the core barebones functionality.

## What is the Barebones Implementation?

The barebones implementation provides:
1. **Connection to MUD servers** via telnet protocol
2. **Display of MUD output** in the main viewport
3. **Command input** to send to the MUD
4. **Empty placeholder panes** for future features (stats, inventory, map)

## Quick Start - Barebones Mode

To use the client as a simple telnet replacement:

```bash
# Connect to a MUD server
./dikuclient --host mud.server.com --port 4000

# Example with a public MUD
./dikuclient --host aardmud.org --port 23
```

## What You'll See

When you connect, you'll see a TUI with:

```
┌─────────────────────────────────────────────────────────────┐
│ Connected to mud.server.com:4000                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────┐  ┌──────────────────────────┐    │
│  │   MUD Output         │  │  Character Stats         │    │
│  │                      │  │  (not implemented)       │    │
│  │  [Game text here]    │  ├──────────────────────────┤    │
│  │  [Server messages]   │  │  Inventory               │    │
│  │  [Prompts]           │  │  (not populated)         │    │
│  │                      │  ├──────────────────────────┤    │
│  │                      │  │  Map                     │    │
│  │                      │  │  (not implemented)       │    │
│  └──────────────────────┘  └──────────────────────────┘    │
│                                                             │
│ > your_command_here                                         │
└─────────────────────────────────────────────────────────────┘
```

## Core Features (Barebones)

### 1. Connect to MUD
The client establishes a TCP connection to the MUD server using the telnet protocol.

### 2. Display Output
All text from the MUD server is displayed in the main viewport with ANSI color support preserved.

### 3. Send Commands
Type commands at the input prompt and press Enter to send them to the MUD.

### 4. Empty Placeholder Panes
The sidebar shows three empty panels that are placeholders for:
- **Character Stats**: Currently shows "(not implemented)"
- **Inventory**: Shows "(not populated)" when empty
- **Map**: Currently shows "(not implemented)"

### 5. Basic Controls
- **Type** to enter commands
- **Enter** to send commands
- **Ctrl+C** or **Esc** to quit

## That's It!

This is the complete barebones functionality - a simple, clean telnet replacement with a nice TUI layout. The empty panes are ready for future enhancements.

## Advanced Features (Optional)

While the barebones functionality is simple, the client also supports optional advanced features:
- Account management (`--save-account`, `--account`)
- Session logging (`--log-all`)
- Web mode (`--web`)

These are not part of the barebones experience but are available if needed.

## Technical Details

### Implementation
- **Language**: Go
- **TUI Framework**: Bubble Tea (charmbracelet)
- **Protocol**: Telnet/TCP
- **Architecture**: Event-driven with goroutines for I/O

### Core Components (Barebones)
- `internal/client/connection.go` - TCP/telnet connection handling
- `internal/tui/app.go` - TUI layout and event handling
- `cmd/dikuclient/main.go` - Entry point and CLI

See [DESIGN.md](DESIGN.md) for complete architecture details.
