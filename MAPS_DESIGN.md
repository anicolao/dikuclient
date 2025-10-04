# Map Panel UI Design Specification

## Overview

The Map Panel is a visual component in the DikuMUD client's sidebar that
displays a representation of the current location and surrounding areas. This
document specifies the user interface design for the Map Panel, including its
visual layout, display modes, interactions, and states.

The Map Panel works in conjunction with the automatic mapping system (documented
in MAPPER.md) to provide players with spatial awareness as they explore the MUD
world.

## Purpose

The Map Panel serves several key purposes:

- **Spatial Orientation**: Shows the player's current position relative to
  explored areas
- **Exit Visualization**: Displays available exits from the current room
- **Quick Reference**: Provides at-a-glance information about nearby rooms
- **Navigation Aid**: Helps players understand their surroundings without
  memorizing text

## Visual Layout

### Panel Position and Size

The Map Panel is located in the right sidebar of the TUI, positioned as the
third (bottom) panel:

```
┌──────────────────────────────────────────────────────────────────┐
│ Status Bar                                                       │
├─────────────────────────────────────────┬────────────────────────┤
│                                         │  Tells                 │
│                                         │                        │
│         Main Output Area                ├────────────────────────┤
│                                         │  Inventory             │
│                                         │                        │
│                                         ├────────────────────────┤
│                                         │  Map         ← HERE    │
│                                         │                        │
│                                         │                        │
└─────────────────────────────────────────┴────────────────────────┘
│ Input Area                                                       │
└──────────────────────────────────────────────────────────────────┘
```

- **Width**: Same as other sidebar panels (approximately 1/3 of screen width)
- **Height**: Equal share with Tells and Inventory panels (approximately 1/3 of
  sidebar height)
- **Minimum Size**: Should display at least a 3x3 room grid when space permits
- **Responsive**: Adapts to terminal resize events

### Panel Header

The panel header should display the current room name as the title:

- **Title**: Current room name in bold text (e.g., "Temple Square")

Example:

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
```

### Initial State (Not Implemented)

Before the map panel is implemented, it displays:

```
┌────────────────────────────┐
│ Map                        │
│                            │
│ (not implemented)          │
│                            │
└────────────────────────────┘
```

### Empty State (No Current Room)

When the mapper is active but no room has been detected yet:

```
┌────────────────────────────┐
│ Map                        │
│                            │
│ (exploring...)             │
│                            │
└────────────────────────────┘
```

## Visual Representation

Rooms are displayed using compact Unicode block characters in a pseudo-graphical
layout. The current room is always displayed in the center of the view.

**Basic Example:**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│        ▢                   │
│        │                   │
│      ▢─▣─▢                 │
│        │                   │
│        ▢                   │
│                            │
└────────────────────────────┘
```

**Larger Example with More Rooms:**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│    ▢ ▢         ▢           │
│      ▢   ▢     ▢           │
│  ▢ ▢   ▢ ▢ ▢     ▢         │
│    ▢     ▢       ▢         │
│  ▢ ▢ ▢ ▢ ▣ ▢ ▢ ▢ ▢         │
│  ▢       ▢                 │
│  ▢ ▢     ▢ ▢               │
│  ▢       ▢     ▢           │
│  ▢ ▢ ▢ ▢ ▢ ▢ ▢ ▢ ▢         │
│                            │
└────────────────────────────┘
```

**Example with Up/Down Exits:**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│      ▢─▢─▢                 │
│        │                   │
│        ⇱                   │
│        │                   │
│      ▢─▣─▢                 │
│        │                   │
│        ⇲                   │
│        │                   │
│        ▢                   │
│                            │
└────────────────────────────┘
```

**Example with Room Having Both Up and Down:**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│      ▢─▢─▢                 │
│        │                   │
│        ⇅                   │
│        │                   │
│      ▢─▣─▢                 │
│        │                   │
│        ▢                   │
│                            │
└────────────────────────────┘
```

**Example with Unexplored Areas:**

```
┌────────────────────────────┐
│ Market District            │
│                            │
│      ▢ ▦ ▢                 │
│      │   │                 │
│      ▢─▣─▢                 │
│        │                   │
│      ▢─▢ ▦                 │
│                            │
└────────────────────────────┘
```

**Comprehensive Example (Realistic MUD Area):**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│    ▢ ▢                     │
│    ▢     ▢                 │
│  ▢ ▢   ▢ ▢ ▢     ▢         │
│    ▢     ▢       ▦         │
│  ▢ ▢ ▢ ▢ ▣ ▢ ▢ ▢ ▢         │
│  ▢       ▢                 │
│  ▢ ▢     ⇅ ▢               │
│  ▢       ▢     ▢           │
│  ▢ ▢ ▢ ▢ ▢ ▢ ▢ ▢ ▢         │
│                            │
└────────────────────────────┘
```

This example shows:

- Current room (▣) in center with bright color
- Multiple explored rooms (▢) around it
- Some unexplored areas (▦) shown in gray
- A room with up/down exits (⇅) south of current room
- Asymmetric layout representing actual MUD geography

**Legend:**

- `▣` = Current room (player's location), shown in bright color
- `▢` = Explored rooms (visited before)
- `▦` = Unexplored/unknown rooms (grayed out, not yet visited)
- `⇱` = Exit up only
- `⇲` = Exit down only
- `⇅` = Both up and down exits available

#### Room Representation Details

**Current Room (▣):**

- Always displayed in the center of the view
- Highlighted with distinct styling (bright cyan or yellow color)
- Represented by filled square block character
- Room name displayed in panel header
- The map view always keeps the current room centered

**Visited Rooms (▢):**

- Shows rooms the player has previously entered
- Represented by hollow square block character
- Standard styling (normal brightness)
- May use different symbols to indicate special rooms (see Special Room Markers
  below)

**Unexplored Rooms (▦):**

- Indicates there is a room in that direction but not yet visited
- Represented by grayed-out block character (light gray color)
- Shows up when mapper knows a room exists but player hasn't entered it
- Helps players identify unexplored areas

#### Directional Connections

The mapper supports six directions: North, South, East, West, Up, and Down.

**Cardinal Directions (N, S, E, W):**

- Rooms are placed adjacent to each other on the grid
- North: Room directly above
- South: Room directly below
- East: Room to the right
- West: Room to the left
- Connection lines (─ and │) show which rooms are actually connected
- Adjacent rooms without connection lines are not connected

**Up/Down Connections:** Up and down connections are indicated with special
arrow symbols shown near the current room:

- `⇱` = Exit up only (room above on different level)
- `⇲` = Exit down only (room below on different level)
- `⇅` = Both up and down exits available

**Spacing and Layout:**

- Each room occupies a single character position
- Rooms are displayed with connection lines showing connectivity
- Box drawing characters (─ │) connect rooms that are linked
- The grid shows spatial relationships through position and connections
- Unexplored exits shown as grayed blocks (▦)

**Simple Example with Vertical Exits:**

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│      ▢─▢─▢                 │
│        │                   │
│        ⇱                   │
│        │                   │
│      ▢─▣─▢                 │
│        │                   │
│        ⇲                   │
│        │                   │
│      ▢─▢─▢                 │
│                            │
└────────────────────────────┘
```

**Note on Multi-Level Areas:** The design focuses on showing the current level
with indicators for vertical exits. The display does not attempt to show
multiple levels simultaneously, as this would complicate the compact view.
Instead:

- Arrow indicators (⇱⇲⇅) clearly show which vertical exits are available
- Players understand that going up or down changes the displayed level
- The mapper tracks all levels but displays one level at a time centered on the
  current room

#### Grid Layout Specifications

The map display uses the entire available space in the panel, always centering
the current room:

**Spacing:**

- Each room occupies a single character position
- Connection lines between rooms show actual exits
- Consistent spacing maintained across entire grid
- Example: `▢─▢─▢` shows three connected rooms
- Example: `▢ ▢ ▢` shows three adjacent but unconnected rooms

**Centering:**

- Current room is always centered in display, even if there is nothing to see on
  one side
- The map view scrolls/shifts to keep the current room centered as player
  navigates
- This provides consistent spatial reference and orientation

## Display Features

### Color Coding

The map panel should use color to convey information:

**Room States:**

- Current Room (`▣`): Yellow (highly visible)
- Visited Rooms (`▢`): White
- Unexplored Rooms (`▦`): Dim gray (significantly darker/muted)
- Special Rooms: Varies by type (see Special Room Markers)

**Spatial Indicators:**

- Connection lines (─ │) in dim gray show cardinal direction exits
- Spatial adjacency without lines indicates no connection
- Arrow symbols (⇱⇲⇅) show vertical connections

**Background:**

- Panel background: Terminal default or subtle dark color
- Matches other sidebar panels for visual consistency

### Special Room Markers

Different room types can be indicated with alternative Unicode block characters
or colors:

**Symbol Variations:**

- `▣` = Current location (always in center, bright color)
- `▢` = Standard explored room
- `▦` = Unexplored room (grayed out)
- `◈` = Home/recall point
- `◆` = Shop or merchant
- `◇` = Bank or storage
- `◎` = Trainer or guild

Example with special rooms:

```
┌────────────────────────────┐
│ Temple Square              │
│                            │
│      ◆ ◇                   │
│    ▢ ▣ ▢                   │
│      ◈ ◎                   │
└────────────────────────────┘
```

In this example: Shop (◆) to north-west, Bank (◇) to north-east, Home (◈) to
south-west, Trainer (◎) to south-east

**Color Variations (Alternative or Combined):**

- Default rooms: White/gray
- Shops: Yellow/gold
- Banks: Green
- Trainers: Blue
- Current room: Yellow (overrides other colors)

### Room Labels

Due to the compact nature of the block-based display, individual room labels are
not shown within the map grid. Instead:

**Current Room Name:**

- Always displayed in the panel header
- Provides clear context for the player's location
- Updates automatically as player moves

The compact grid prioritizes spatial overview over individual room
identification.

### Data Requirements

- Map panel should work with existing map data format
- No modification to underlying map data structures
- Read-only access to map data
- No dependencies on specific MUD server output

### Integration Requirements

- Fit within existing TUI framework (Bubble Tea)
- Use existing styling system (lipgloss)
- Coordinate with other sidebar panels
- Respect global application state

## Summary

The Map Panel UI should provide players with clear, at-a-glance spatial
awareness within the MUD world. The design prioritizes:

1. **Usability**: Easy to understand at a glance
2. **Performance**: Lightweight and responsive
3. **Simplicity**: Single display mode with current room always centered

The implementation uses compact Unicode block characters to display the map,
with the current room always centered in the available panel space.
