# Barsoom MUD Room Description Feature

## Overview

This feature adds special handling for Barsoom MUD room descriptions that use `--<` and `>--` markers to bracket room information.

## Room Format

Barsoom MUD uses a special format for room descriptions:

```
119H 110V 3674X 0.00% 77C T:56 Exits:EW>
--<
Temple Square
    You are standing in a large temple square. The ancient stones
speak of a glorious past.
>--
Exits: north, south, east
```

### Components:

1. **Start Marker**: `--<` (alone on a line)
2. **Room Title**: First non-empty line after the start marker
3. **Room Description**: One or more paragraphs describing the room
4. **End Marker**: `>--` (alone on a line)
5. **Exits**: Listed after the end marker

## Features

### 1. Marker Suppression

The `--<` and `>--` markers are automatically suppressed from the display output, providing a cleaner reading experience. They are still processed internally for room detection and mapping.

**Input:**
```
--<
Temple Square
    You are standing in a large temple square.
>--
```

**Display:**
```
Temple Square
    You are standing in a large temple square.
```

### 2. Description Split Viewport

When a Barsoom room is detected, the room description is displayed in a sticky viewport at the top of the screen. This ensures the room description is always visible while exploring.

**Normal View:**
```
┌── Temple Square [north, south, east] ──┐
│                                         │
│  Temple Square                          │
│                                         │
│    You are standing in a large temple  │
│  square. The ancient stones speak of a │
│  glorious past.                         │
│                                         │
├─────────────────────────────────────────┤
│                                         │
│  > look                                 │
│  You see a room description here.       │
│  > inventory                            │
│  You are carrying: a sword, a torch     │
│  [Live output continues below...]       │
│                                         │
└─────────────────────────────────────────┘
```

### 3. Three-Way Split (When Scrolling)

If the user scrolls up while viewing a Barsoom room, the layout becomes a three-way split:

```
┌── Temple Square [north, south, east] ──┐
│  [Description - Stuck to Top]           │
│  Temple Square                          │
│    You are standing in a large...      │
├─────────────────────────────────────────┤
│  [Scrollable Content - Middle]          │
│  > look                                 │
│  You see a room description here.       │
│  > inventory                            │
│  ...                                    │
├─────────────────────────────────────────┤
│  [Live Output - Stuck to Bottom]        │
│  > kill orc                             │
│  You attack the orc!                    │
│  The orc hits you!                      │
│  [Always shows latest output]           │
└─────────────────────────────────────────┘
```

### 4. Automatic Mapping

Barsoom rooms are automatically detected and added to the world map with proper room title, description, and exits. This ensures accurate pathfinding and navigation.

### 5. Regular Rooms Continue to Work

Non-Barsoom rooms (without the markers) continue to work normally, without the description split. The system automatically detects the room format and adapts accordingly.

## Implementation Details

### Parser Changes

- **File**: `internal/mapper/parser.go`
- **Function**: `parseBarsoomRoom()` - Detects and parses Barsoom format rooms
- **Fields Added**: `IsBarsoomRoom`, `BarsoomStartIdx`, `BarsoomEndIdx` in `RoomInfo` struct

### TUI Changes

- **File**: `internal/tui/app.go`
- **Fields Added**: `descriptionViewport`, `currentRoomDescription`, `hasDescriptionSplit`
- **Marker Suppression**: In `mudMsg` handler, filters `--<` and `>--` from display output
- **Description Split**: In `detectAndUpdateRoom()`, sets up description viewport content
- **Layout**: In `renderMainContent()`, supports three-way split when both description and scroll splits are active

## Testing

Comprehensive tests have been added to verify:

1. **Parser Tests** (`internal/mapper/parser_test.go`):
   - Basic Barsoom format parsing
   - Multi-paragraph descriptions
   - Marker index tracking

2. **Demo Test** (`internal/mapper/barsoom_demo_test.go`):
   - Visual demonstration of all scenarios
   - Comparison with regular room format

3. **Integration Tests** (`internal/tui/barsoom_test.go`):
   - Marker suppression verification
   - Description split activation
   - Regular room behavior

## Usage

No configuration is required. The feature automatically activates when connecting to a Barsoom MUD or any MUD that uses the `--<` and `>--` room format.

To see the feature in action:
1. Connect to a Barsoom MUD
2. Move to a room - the description will appear in the sticky top viewport
3. Scroll up with Page Up or mouse wheel - the description stays at the top
4. Scroll down to exit split mode - the layout returns to normal

## Benefits

- **Cleaner Output**: Markers are hidden from view
- **Better Navigation**: Room descriptions always visible at the top
- **Accurate Mapping**: Proper room detection ensures reliable auto-mapping
- **Seamless Experience**: Works alongside existing split viewport feature
- **Backward Compatible**: Regular rooms continue to work as before
