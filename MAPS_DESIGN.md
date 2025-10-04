# Map Panel UI Design Specification

## Overview

The Map Panel is a visual component in the DikuMUD client's sidebar that displays a representation of the current location and surrounding areas. This document specifies the user interface design for the Map Panel, including its visual layout, display modes, interactions, and states.

The Map Panel works in conjunction with the automatic mapping system (documented in MAPPER.md) to provide players with spatial awareness as they explore the MUD world.

## Purpose

The Map Panel serves several key purposes:
- **Spatial Orientation**: Shows the player's current position relative to explored areas
- **Exit Visualization**: Displays available exits from the current room
- **Quick Reference**: Provides at-a-glance information about nearby rooms
- **Navigation Aid**: Helps players understand their surroundings without memorizing text

## Visual Layout

### Panel Position and Size

The Map Panel is located in the right sidebar of the TUI, positioned as the third (bottom) panel:

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
- **Height**: Equal share with Tells and Inventory panels (approximately 1/3 of sidebar height)
- **Minimum Size**: Should display at least a 3x3 room grid when space permits
- **Responsive**: Adapts to terminal resize events

### Panel Header

The panel header should display:
- **Title**: "Map" in bold text
- **Optional Subtitle**: Room count or current area name in a smaller, lighter font

Example:
```
┌────────────────────────────┐
│ Map                        │
│ (42 rooms explored)        │
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

## Display Modes

### Mode 1: Local Area View (Primary Mode)

The default and most common display mode showing the immediate area around the player.

#### Visual Representation

Rooms are displayed using ASCII characters in a grid layout:

```
┌────────────────────────────┐
│ Map                        │
│ (42 rooms)                 │
│                            │
│     [?]─[?]─[?]            │
│      │   │   │             │
│     [?]─[@]─[?]            │
│      │   │   │             │
│     [?]─[?]─[?]            │
│                            │
└────────────────────────────┘
```

**Legend:**
- `[@]` = Current room (player's location)
- `[?]` = Explored room (visited before)
- `[!]` = Unexplored exit (room exists but not visited)
- `[ ]` = Empty space (no known room)
- `─` = Horizontal connection (east/west)
- `│` = Vertical connection (north/south)
- `┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼` = Corner/junction connectors

#### Room Representation Details

**Current Room (@):**
- Displayed in the center of the view when possible
- Highlighted with distinct styling (bright color, possibly inverted)
- Always visible unless player moves to edge of explored area

**Visited Rooms (?):**
- Shows rooms the player has previously entered
- Standard styling (normal brightness)
- May use different symbols to indicate special rooms (see Special Room Markers below)

**Exits to Unexplored Rooms (!):**
- Indicates there is an exit in that direction
- Room details unknown until visited
- Helps players identify unexplored areas

#### Directional Connections

**Cardinal Directions:**
- North/South: Vertical line `│`
- East/West: Horizontal line `─`
- Displayed between adjacent room markers

**Diagonal Connections:**
- Northeast: `╱` or `/`
- Northwest: `╲` or `\`
- Southeast: `╲` or `\`
- Southwest: `╱` or `/`
- May be displayed at an angle if space permits

**Up/Down Connections:**
- Indicated by `^` (up) or `v` (down) symbol adjacent to room marker
- Example: `[?]^` means there's an exit up from this room
- Example: `[?]v` means there's an exit down from this room

#### Grid Layout Specifications

**Standard Grid Size:**
- Minimum: 3x3 rooms (current room plus one in each direction)
- Preferred: 5x5 rooms for better context
- Maximum: As large as panel height permits

**Spacing:**
- Rooms separated by connection symbols (`─` or `│`)
- Minimum one character between room markers
- Consistent spacing maintained across entire grid

**Centering:**
- Current room centered in display when possible
- View shifts when player moves to edge of explored area
- Smooth transitions as player navigates

### Mode 2: Compact View

When panel height is limited (too small for standard grid), display a text-based alternative:

```
┌────────────────────────────┐
│ Map (compact)              │
│                            │
│ Current: Temple Square     │
│                            │
│ Exits:                     │
│   N: Market District       │
│   S: Guard Post            │
│   E: Training Ground       │
│   W: (unexplored)          │
│                            │
└────────────────────────────┘
```

This mode shows:
- Current room name
- List of exits with destinations
- Known room names or "(unexplored)" for unknown exits
- No graphical representation

### Mode 3: Room Info View (Alternative)

An alternative focused on detailed room information:

```
┌────────────────────────────┐
│ Map                        │
│                            │
│ Temple Square              │
│ Visited: 3 times           │
│                            │
│ Exits: N, S, E, W          │
│                            │
│ Nearby (1 step):           │
│ • Market District (N)      │
│ • Guard Post (S)           │
│                            │
└────────────────────────────┘
```

## Display Features

### Color Coding

The map panel should use color to convey information:

**Room States:**
- Current Room (`[@]`): Bright cyan or yellow (highly visible)
- Visited Rooms (`[?]`): Normal white or gray
- Unexplored Exits (`[!]`): Dim white or dark gray
- Special Rooms: Varies by type (see Special Room Markers)

**Connection Lines:**
- Normal connections: Gray or dim white
- Recently traveled path: Brighter or colored temporarily

**Background:**
- Panel background: Terminal default or subtle dark color
- Matches other sidebar panels for visual consistency

### Special Room Markers

Different room types can be indicated with different symbols or colors:

**Symbol Variations:**
- `[@]` = Current location
- `[H]` = Home/recall point
- `[$]` = Shop or merchant
- `[#]` = Bank or storage
- `[T]` = Trainer or guild
- `[!]` = Dangerous area (marked by player)
- `[*]` = Points of interest (marked by player)
- `[?]` = Standard explored room

**Color Variations (Alternative or Combined):**
- Default rooms: White/gray
- Shops: Yellow/gold
- Banks: Green
- Trainers: Blue
- Dangerous areas: Red
- Points of interest: Magenta
- Current room: Bright cyan (overrides other colors)

### Room Labels

When space permits, rooms can display abbreviated names:

```
┌────────────────────────────┐
│ Map                        │
│                            │
│  Market─Temple─Training    │
│     │      │       │       │
│  Guild──[@]────Plaza       │
│     │      │       │       │
│  Alley──Bank──Fountain     │
│                            │
└────────────────────────────┘
```

- Names truncated to fit available space (e.g., "Temple Square" → "Temple")
- Only displayed when panel width allows
- Can be toggled on/off by user preference
- Shown in smaller or lighter font than room markers

### Dynamic Elements

**Update Indicators:**
- Brief highlight or animation when map updates
- Visual feedback when new room is discovered
- Flash or pulse on current room when player moves

**Path Preview:**
- When using `/point` or `/wayfind` commands, highlight the suggested path
- Different color/style for navigation route
- Arrow indicators showing direction of travel

**Auto-Walk Indicator:**
- During `/go` auto-walk, show progress along path
- Highlight next destination
- Mark completed vs. remaining steps

## Interaction Specifications

### Scrolling Behavior

**Automatic Scrolling:**
- Map auto-centers on player's current room
- View shifts smoothly as player moves
- No manual scrolling needed in normal usage

**Manual Scrolling (Optional Future Enhancement):**
- Arrow keys or mouse wheel to pan the map
- Temporary pan mode that returns to center after timeout
- Shows larger explored area beyond immediate vicinity

### Viewport Focus

**Center on Current Room:**
- Default behavior: player always centered (when possible)
- Map scrolls to keep player visible when they move

**Edge Cases:**
- At edges of explored area, map may not be centered
- Player position may shift toward edge if no rooms beyond

### Refresh Rate

**Update Triggers:**
- New room detected: Immediate update
- Player movement: Immediate update
- Map data changes: Immediate update
- Terminal resize: Immediate re-render

**Performance:**
- Updates should be lightweight (no visible lag)
- Only redraw map panel, not entire screen
- Efficient rendering for large maps

## Panel States

### State 1: Not Implemented

Before implementation is complete:
- Shows "(not implemented)" message
- Panel is visible but non-functional
- Placeholder for future implementation

### State 2: Exploring (No Data)

Mapper is active but hasn't detected first room yet:
- Shows "(exploring...)" message
- Indicates system is working but waiting for room data
- Changes to Active state when first room detected

### State 3: Active (Normal Display)

Mapper has data and is displaying the map:
- Shows current room and surrounding area
- Updates automatically as player moves
- Primary operational state

### State 4: Compact Mode

Panel is too small for graphical display:
- Switches to text-based compact view
- Shows current room name and exits
- Automatic fallback when height insufficient

### State 5: Empty/Loading

Temporary state during transitions:
- Brief blank or loading indicator
- Occurs during map file loading
- Should be very brief (< 100ms)

### State 6: Error State (Optional)

If map data is corrupted or unavailable:
- Shows "(map unavailable)" message
- Optionally includes error reason
- Suggests user action if applicable

## Responsive Behavior

### Terminal Resize Handling

**Width Adjustments:**
- Narrow width (< 30 cols): Switch to compact mode or hide room labels
- Medium width (30-40 cols): Standard display with abbreviated labels
- Wide width (> 40 cols): Full display with complete room labels

**Height Adjustments:**
- Short height (< 8 rows): Compact mode (text list)
- Medium height (8-12 rows): 3x3 grid
- Tall height (> 12 rows): 5x5 or larger grid

### Automatic Mode Switching

- Panel automatically selects best display mode based on available space
- Smooth transitions between modes
- Maintains data continuity across mode changes
- No user intervention required

### Panel Collapse/Expand (Optional)

- Allow users to collapse panels to give more space to others
- Keyboard shortcut to toggle map panel visibility
- Saves screen real estate when map not needed

## Integration with Mapping Commands

The Map Panel should reflect the state and output of mapping commands:

### Command: `/map`

When user types `/map`:
- Command output appears in main window
- Map panel highlights or updates to draw attention
- Visual sync between text info and graphical display

### Command: `/point <destination>`

When finding a path:
- Map panel highlights the suggested path
- Next direction indicator more prominent
- Path shown in different color (e.g., green or yellow)
- Remains visible until player moves or cancels

### Command: `/wayfind <destination>`

When showing full path:
- Entire path highlighted on map
- Each step numbered or colored progressively
- Shows complete route visually
- Helps player understand spatial relationship

### Command: `/go <destination>`

During auto-walk:
- Map shows complete path
- Current destination highlighted
- Progress indicator (e.g., rooms turn green as passed)
- Next room in path emphasized
- Updates in real-time as player moves

### Command: `/rooms [filter]`

When listing rooms:
- Could highlight matching rooms on map (optional)
- Visual correlation between text list and map
- Helps player locate rooms spatially

## Visual Design Principles

### Clarity

- Use clear, distinct symbols for different room types
- Maintain adequate contrast between elements
- Avoid visual clutter or overlapping information
- Prioritize readability over decoration

### Consistency

- Match styling of other sidebar panels (Tells, Inventory)
- Use same color scheme and border styles
- Maintain consistent spacing and alignment
- Follow established UI patterns in the client

### Information Density

- Balance between detail and simplicity
- Show enough information to be useful
- Avoid overwhelming the player
- Use progressive disclosure (more detail on demand)

### Visual Hierarchy

- Current room most prominent
- Adjacent rooms clearly visible
- Distant rooms less emphasized
- Connections subtle but clear

## Accessibility Considerations

### Symbol Alternatives

- Provide ASCII-only mode (no box-drawing characters)
- Alternative symbols that work in limited terminals
- Configurable character set for different terminal types

### Color Blindness

- Don't rely solely on color for information
- Use symbols + color combination
- Provide monochrome-friendly mode
- Test with color blindness simulators

### Screen Readers

- Provide text alternatives for visual elements
- Map state should be describable in text
- Integration with accessibility tools
- Text-based navigation option

## Configuration Options (Future)

### Display Preferences

Users should eventually be able to configure:
- **Display mode preference**: Grid, Compact, or Info view
- **Symbol set**: Standard, ASCII-only, or Custom
- **Color scheme**: Full color, Limited, or Monochrome
- **Room labels**: On, Off, or Auto
- **Grid size**: 3x3, 5x5, 7x7, or Auto
- **Special markers**: Enable/disable specific room types

### Toggle Options

- Show/hide unexplored exits
- Show/hide connection lines
- Show/hide room labels
- Show/hide special markers
- Auto-center on/off

## Technical Requirements

### Performance Requirements

- **Render time**: Map panel updates should complete in < 10ms
- **Memory**: Map display should not significantly impact memory usage
- **CPU**: Rendering should use minimal CPU (< 1% on modern systems)
- **Responsiveness**: No input lag caused by map rendering

### Terminal Compatibility

- **Character sets**: Support both UTF-8 box-drawing and ASCII fallback
- **Color support**: Gracefully degrade from 256-color to 16-color to monochrome
- **Size ranges**: Handle terminals from 80x24 to 200x60+
- **Terminal types**: Work in common terminals (xterm, tmux, screen, etc.)

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

## Error Handling

### Graceful Degradation

- If map data unavailable: Show empty state message
- If rendering fails: Fall back to compact mode
- If terminal too small: Show minimal info
- If invalid room data: Display what's valid, omit invalid

### Error Messages

Error messages should be:
- **Concise**: Fit within panel space
- **Actionable**: Suggest what user can do
- **Non-intrusive**: Don't block other functionality
- **Informative**: Explain what went wrong

Example error states:
```
(map unavailable)
(no current room)
(panel too small)
(map data error)
```

## Future Enhancements

These are not part of the initial specification but should be considered in the design:

### Interactive Features

- Click/select rooms to see details
- Right-click for room menu
- Hover for room tooltips
- Drag to pan the view

### Advanced Visualization

- Multi-level display (show up/down levels simultaneously)
- 3D visualization mode
- Area boundaries and zones
- Terrain types or biomes

### Customization

- User-defined room markers
- Custom color schemes
- Adjustable zoom levels
- Overlay layers (show/hide different information)

### Social Features

- Show other party members' positions
- Track mob locations
- Mark quest objectives
- Share map annotations

### Path Features

- Show multiple path options
- Avoid dangerous areas
- Prefer certain route types
- Save and replay routes

## Summary

The Map Panel UI should provide players with clear, at-a-glance spatial awareness within the MUD world. The design prioritizes:

1. **Usability**: Easy to understand at a glance
2. **Integration**: Works seamlessly with existing mapper commands
3. **Flexibility**: Adapts to different terminal sizes and user needs
4. **Performance**: Lightweight and responsive
5. **Accessibility**: Works for all users regardless of terminal or abilities

The implementation should start with the Local Area View (Mode 1) as the primary display mode, with Compact View (Mode 2) as an automatic fallback for space constraints. Additional features and modes can be added incrementally based on user feedback and needs.
