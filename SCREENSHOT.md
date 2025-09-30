# DikuMUD Client - User Interface

## Application Layout

The DikuMUD client features a modern TUI (Text User Interface) built with Bubble Tea:

```
┌─ Status Bar ──────────────────────────────────────────────────────────────┐
│ Connected to localhost:4000 ──────────────────────────────────────────────│
├───────────────────────────────────────────────────────────────────────────┤
│                                                │                           │
│  ┌─ Game Output ─────────────────────────┐    │  ┌─ Character Stats ───┐ │
│  │                                        │    │  │                      │ │
│  │ Connected to localhost:4000            │    │  │  (not implemented)   │ │
│  │ Welcome to Test MUD!                   │    │  │                      │ │
│  │ ===================                    │    │  └──────────────────────┘ │
│  │                                        │    │                           │
│  │ Available commands: look, inventory,   │    │  ┌─ Inventory ──────────┐ │
│  │   stats, help, quit                    │    │  │                      │ │
│  │                                        │    │  │  (not implemented)   │ │
│  │ > look                                 │    │  │                      │ │
│  │ You are in a test room. There are      │    │  └──────────────────────┘ │
│  │   exits to the north and south.        │    │                           │
│  │                                        │    │  ┌─ Map ────────────────┐ │
│  │ > inventory                            │    │  │                      │ │
│  │ You are carrying:                      │    │  │  (not implemented)   │ │
│  │   - A rusty sword                      │    │  │                      │ │
│  │   - 3 health potions                   │    │  └──────────────────────┘ │
│  │                                        │    │                           │
│  │ > stats                                │    │                           │
│  │ Your stats:                            │    │                           │
│  │   HP: 100/100                          │    │                           │
│  │   MP: 75/100                           │    │                           │
│  │   Level: 5                             │    │                           │
│  │                                        │    │                           │
│  └────────────────────────────────────────┘    │                           │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─ Input Area ──────────────────────────┐                                │
│  │┃ Type your command here...             │                                │
│  │┃                                       │                                │
│  │┃                                       │                                │
│  └────────────────────────────────────────┘                                │
└───────────────────────────────────────────────────────────────────────────┘
```

## Features Demonstrated

### 1. **Status Bar**
- Shows connection status
- Displays host and port information

### 2. **Game Output Panel** (Main Area)
- Displays all MUD server messages
- Shows command history with "> " prefix
- Auto-scrolls to show latest content
- Bordered with rounded corners

### 3. **Sidebar Panels** (Right Side)
Three stacked panels for future features:
- **Character Stats**: Placeholder for HP, MP, Level, etc.
- **Inventory**: Placeholder for items and equipment
- **Map**: Placeholder for room mapping

All sidebar panels currently show "(not implemented)" text in italics.

### 4. **Input Area**
- Multi-line textarea for command input
- Placeholder text: "Type your command here..."
- Press Enter to send commands
- Commands are echoed in the game output panel

## Controls

- **Type** commands and press **Enter** to send
- **Esc** or **Ctrl+C** to quit the application
- The interface is fully keyboard-driven

## Technical Details

- Built with Go and Bubble Tea framework
- Uses Lipgloss for styling and layout
- Bubbles for UI components (viewport, textarea)
- Responsive layout that adapts to terminal size
- Concurrent MUD connection handling
- Clean separation of concerns (connection, UI, application logic)
