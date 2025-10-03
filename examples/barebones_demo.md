# Barebones MUD Client Demo

This document demonstrates the barebones functionality of the DikuMUD Client.

## What You Get (Barebones)

The barebones implementation provides a **simple telnet replacement** with a clean TUI:

1. **Connection** to any MUD server
2. **Display** of MUD output with ANSI colors
3. **Input** prompt to send commands
4. **Empty panes** ready for future features

## Running the Barebones Client

### Step 1: Build the Client

```bash
cd /path/to/dikuclient
go build -o dikuclient ./cmd/dikuclient
```

### Step 2: Connect to a MUD

```bash
# Basic connection (barebones mode)
./dikuclient --host aardmud.org --port 23

# Or any other MUD server
./dikuclient --host your.mud.server --port 4000
```

## What You'll See

```
┌──────────────────────────────────────────────────────────────────┐
│ Connected to aardmud.org:23                                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────┐  ┌─────────────────────────────────────┐│
│  │  Game Output       │  │  ╭─────────────────────────────────╮││
│  │  ╭─────────────────╯  │  │ Character Stats                 │││
│  │  │                    │  │                                   │││
│  │  │ Welcome to        │  │ (not implemented)                 │││
│  │  │ Aardwolf MUD!     │  │                                   │││
│  │  │                    │  ├─────────────────────────────────────┤│
│  │  │ By what name do   │  │ Inventory                         │││
│  │  │ you wish to be    │  │                                   │││
│  │  │ known?            │  │ (not populated)                   │││
│  │  │                    │  │                                   │││
│  │  │ >                  │  ├─────────────────────────────────────┤│
│  │  │                    │  │ Map                               │││
│  │  │                    │  │                                   │││
│  │  │                    │  │ (not implemented)                 │││
│  │  │                    │  │                                   │││
│  │  ╰─────────────────╮  │  ╰─────────────────────────────────╯││
│  └────────────────────┘  └─────────────────────────────────────┘│
│                                                                  │
│ > type_your_commands_here_                                       │
└──────────────────────────────────────────────────────────────────┘
```

## Using the Barebones Client

### Connecting

The client automatically connects when you start it with `--host` and `--port` flags.

### Sending Commands

1. Type your command at the bottom prompt
2. Press `Enter` to send
3. The MUD server responds in the main output area

### Example Session

```
> north
You walk north.

A Dark Forest
You are in a dark, foreboding forest. Trees loom overhead,
blocking out most of the sunlight. Paths lead in all directions.
Exits: north, south, east, west

> look
A Dark Forest
You are in a dark, foreboding forest. Trees loom overhead,
blocking out most of the sunlight. Paths lead in all directions.
Exits: north, south, east, west

> quit
Goodbye!
```

### Quitting

Press `Ctrl+C` or `Esc` to exit the client.

## Empty Panes Explained

The sidebar shows three panels that are currently empty placeholders:

### Character Stats Panel
```
┌─────────────────────┐
│ Character Stats     │
│                     │
│ (not implemented)   │
└─────────────────────┘
```

**Purpose**: Will eventually show:
- HP/Mana/Movement
- Level/Experience
- Combat status

**Current State**: Empty placeholder

### Inventory Panel
```
┌─────────────────────┐
│ Inventory           │
│                     │
│ (not populated)     │
└─────────────────────┘
```

**Purpose**: Will eventually show:
- Items in your inventory
- Equipment worn
- Weight/capacity

**Current State**: Empty placeholder (shows "not populated")

### Map Panel
```
┌─────────────────────┐
│ Map                 │
│                     │
│ (not implemented)   │
└─────────────────────┘
```

**Purpose**: Will eventually show:
- ASCII map of the area
- Your current location
- Recently visited rooms

**Current State**: Empty placeholder

## Technical Details

### Architecture

The barebones client uses:
- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (charmbracelet)
- **Protocol**: Raw TCP with telnet IAC handling
- **Concurrency**: Goroutines for I/O

### Key Components

1. **Connection Layer** (`internal/client/connection.go`)
   - TCP socket connection
   - Telnet protocol negotiation
   - Buffered I/O

2. **TUI Layer** (`internal/tui/app.go`)
   - Bubble Tea model
   - Layout management
   - Event handling

3. **Main Entry** (`cmd/dikuclient/main.go`)
   - CLI argument parsing
   - Program initialization

### Data Flow

```
User Input → TUI → Connection → MUD Server
                                     ↓
MUD Output ← TUI ← Connection ← Server Response
```

## Comparison to Raw Telnet

### Traditional Telnet
```bash
$ telnet aardmud.org 23
Trying 199.193.253.35...
Connected to aardmud.org.
Escape character is '^]'.

Welcome to Aardwolf MUD!
[plain text output scrolls by]
[no visual structure]
[no panel organization]
```

### DikuMUD Client (Barebones)
```bash
$ ./dikuclient --host aardmud.org --port 23
[Clean TUI with organized panels]
[Status bar showing connection]
[Separated input/output areas]
[Empty panes ready for features]
[ANSI color support]
```

## Advantages Over Raw Telnet

1. **Visual Organization**: Clear separation of concerns
2. **Status Display**: Always know your connection state
3. **Input Area**: Dedicated command line
4. **Future-Ready**: Panels ready for stats/inventory/maps
5. **ANSI Support**: Proper color rendering
6. **Keyboard Handling**: Better input control
7. **Extensible**: Easy to add features

## What's NOT Included (Barebones)

To keep the barebones version simple, these features are NOT active:

- ❌ Automatic mapping
- ❌ Trigger system
- ❌ Account management
- ❌ Auto-login
- ❌ Command history
- ❌ Web mode
- ❌ Session logging (unless explicitly enabled)

If you want these features, they're available in the full version. See the main README.md.

## Summary

The barebones implementation is **exactly what it claims to be**: a simple, clean telnet replacement with a nice TUI. It provides:

✓ Connection to MUD servers  
✓ Display of game output  
✓ Command input  
✓ Clean visual layout  
✓ Empty panes for future features  

Nothing more, nothing less. Perfect for users who want a better-than-telnet experience without complexity.

## Next Steps

Once you're comfortable with the barebones version, you can explore:

1. **Account Management**: Save credentials with `--save-account`
2. **Session Logging**: Use `--log-all` to log sessions
3. **Web Mode**: Run in browser with `--web`
4. **Advanced Features**: See README.md for mapper, triggers, etc.

But for now, enjoy the simplicity of a clean, functional MUD client!
