# Inventory Feature

## Overview
The DikuMUD client now automatically detects inventory output from the MUD server and displays it in the inventory sidebar panel with a timestamp.

## How It Works

### Automatic Detection
When you type `i` or `inventory` in the MUD, the client automatically detects the output and populates the inventory panel.

**Example MUD Output:**
```
86H 109V 7563X 0.00% 79C T:3 Exits:D> i
You are carrying:
a sharp short sword
a glowing scroll of recall..it glows dimly
a torch [4]
a rusty knife
a bowl of Otik's spiced potatoes [2]
an entire loaf of bread [4]

86H 109V 7563X 0.00% 79C T:2 Exits:D>
```

### User Interface

The inventory appears in two places:
1. **Main Output Window** - Shows the full inventory output from the MUD
2. **Inventory Sidebar Panel** - Displays items in a compact format with timestamp

#### Before Inventory Command
```
┌─────────────────────────────────────────────┐
│                Main Output                  │
├─────────────────────────────────────────────┤
│ You enter the room...                       │
│                                             │
│ 86H 109V 7563X 0.00% 79C T:3 Exits:D>      │
│                                             │
└─────────────────────────────────────────────┘

         Sidebar:
    ┌────────────────┐
    │ Inventory      │
    │                │
    │ (not populated)│
    └────────────────┘
```

#### After Inventory Command
```
┌─────────────────────────────────────────────┐
│                Main Output                  │
├─────────────────────────────────────────────┤
│ 86H 109V 7563X 0.00% 79C T:3 Exits:D> i    │
│ You are carrying:                           │
│ a sharp short sword                         │
│ a glowing scroll of recall..it glows dimly  │
│ a torch [4]                                 │
│ a rusty knife                               │
│ a bowl of Otik's spiced potatoes [2]        │
│ an entire loaf of bread [4]                 │
│                                             │
│ 86H 109V 7563X 0.00% 79C T:2 Exits:D>      │
└─────────────────────────────────────────────┘

         Sidebar:
    ┌────────────────┐
    │ Inventory      │
    │ (15:04:23)     │ ← Timestamp
    │                │
    │ a sharp sho... │
    │ a glowing s... │
    │ a torch [4]    │
    │ a rusty knife  │
    │ a bowl of O... │
    │ an entire l... │
    └────────────────┘
```

## Features

### ✓ Automatic Detection
- No configuration needed
- Works with standard "You are carrying:" format
- Detects inventory immediately when output appears

### ✓ Timestamp Label
- Shows when inventory was last updated
- Format: HH:MM:SS (e.g., "15:04:23")
- Displayed in gray text below "Inventory" header

### ✓ Dual Display
- **Main Output**: Full inventory text preserved for reference
- **Sidebar Panel**: Compact, always-visible inventory list

### ✓ Smart Formatting
- Long item names truncated with "..." to fit panel
- Empty lines skipped
- ANSI color codes properly handled
- Overflow indicator ("...") when too many items

### ✓ Persistent Display
- Inventory stays visible until next update
- Doesn't disappear when moving rooms
- Updates only when new inventory command is issued

## Technical Details

### Parser (`internal/mapper/inventory.go`)
- Searches for "You are carrying:" header
- Collects all lines until next MUD prompt
- Strips ANSI codes for clean display
- Handles empty inventory (no items)

### Detection (`internal/tui/app.go`)
- Called automatically when MUD output arrives
- Uses recent output buffer (last 30 lines)
- Updates both inventory items and timestamp
- No performance impact on game play

### Rendering (`internal/tui/app.go`)
- Displays in sidebar panel (right side of screen)
- Adapts to panel height (shows "..." if overflow)
- Truncates items to fit panel width
- Shows "(not populated)" when no inventory detected

## Testing

Comprehensive test coverage includes:
- Parser unit tests
- Detection tests
- Integration tests
- Rendering tests

All tests pass successfully:
```bash
go test ./...
?   	github.com/anicolao/dikuclient/cmd/dikuclient	[no test files]
ok  	github.com/anicolao/dikuclient/internal/client
ok  	github.com/anicolao/dikuclient/internal/mapper
ok  	github.com/anicolao/dikuclient/internal/tui
ok  	github.com/anicolao/dikuclient/internal/web
```

## Example Usage

1. Connect to your MUD server
2. Type `i` or `inventory` command
3. Watch the inventory panel populate automatically
4. The timestamp shows when it was last updated

No special commands or configuration required!
