# Split Viewport Feature

## Overview

The DikuMUD client now supports a split viewport feature that automatically activates when the user scrolls back in the main window. This allows users to review past output while still tracking new output from the MUD server in real-time.

## How It Works

### Automatic Split Mode

When you scroll up in the main viewport (using Page Up or mouse wheel), the window automatically splits into two sections:

1. **Top Section (2/3 of screen)**: Shows your scrolled position, preserving where you were reading
2. **Bottom Section (1/3 of screen)**: Continuously tracks new MUD output at the bottom

### Automatic Unsplit

The split automatically disappears when you:
- Scroll back to the bottom (using Page Down or mouse wheel down)
- Were already at the bottom when new content arrives

## User Interface

### Normal Mode (Not Split)
```
┌── Current Room [exits] ──────────────────────┐
│                                              │
│  Game output continues here...               │
│  > look                                      │
│  You see a room description here.            │
│  > kill orc                                  │
│  You attack the orc!                         │
│  The orc hits you!                           │
│  [Live output at bottom]                     │
│                                              │
└──────────────────────────────────────────────┘
```

### Split Mode (After Scrolling Up)
```
┌── Current Room [exits] ──────────────────────┐
│                                              │
│  [Scrolled back view - 2/3 of screen]        │
│  > look                                      │
│  You see a room description here.            │
│  > inventory                                 │
│  You are carrying: a sword, a torch          │
│                                              │
├──────────────────────────────────────────────┤
│                                              │
│  [Live tracking view - 1/3 of screen]        │
│  > kill orc                                  │
│  You attack the orc!                         │
│  The orc hits you!                           │
│  [Always shows latest output]                │
│                                              │
└──────────────────────────────────────────────┘
```

## Controls

### Entering Split Mode
- **Page Up**: Scroll up one page
- **Mouse Wheel Up**: Scroll up by lines

### Exiting Split Mode
- **Page Down**: Scroll down (automatically exits when at bottom)
- **Mouse Wheel Down**: Scroll down by lines (automatically exits when at bottom)

## Use Cases

1. **Reading Previous Combat Logs**: Scroll back to see what happened earlier while still monitoring current combat
2. **Reviewing Tells**: Read old conversation messages while tracking new incoming tells
3. **Checking Past Room Descriptions**: Review earlier room info while moving through areas
4. **Analyzing Combat XP**: Look at XP gain history while still fighting

## Technical Details

### Implementation
- Uses Bubble Tea's viewport component
- Two viewports: main (user-controlled) and split (auto-bottom)
- State tracked with `isSplit` boolean flag
- Automatic state management based on scroll position

### Performance
- Minimal overhead: only active when split
- Both viewports share the same content buffer
- No duplication of output data

### Testing
- Comprehensive unit tests in `internal/tui/split_viewport_test.go`
- Tests cover:
  - Page Up/Down behavior
  - Mouse wheel behavior
  - New content handling (at bottom vs scrolled)
  - Rendering in split mode

## Examples

### Scenario 1: Reading Old Combat While Fighting
```
User scrolls up (Page Up) to read earlier combat messages
├─ Top 2/3: Shows "You killed a goblin" from 30 seconds ago
└─ Bottom 1/3: Shows "The orc attacks!" happening now

User can review the past while monitoring present
```

### Scenario 2: Reviewing Tells While Exploring
```
User scrolls up to read old tell messages
├─ Top 2/3: Shows conversation from 5 minutes ago
└─ Bottom 1/3: Shows "You arrive in Temple Square" from current movement

User can catch up on conversation without missing navigation
```

### Scenario 3: Automatic Return to Normal
```
User scrolls back down to bottom
└─ Split mode automatically exits
    Returns to single viewport showing all current output
```

## Future Enhancements

Potential improvements (not in current implementation):
- Configurable split ratio (e.g., 1/2 1/2 or 3/4 1/4)
- Manual split toggle command
- Independent scrolling in both viewports
- Split position persistence across sessions
