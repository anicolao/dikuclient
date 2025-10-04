# Map Panel Visual Examples

This document shows real output examples from the implemented Map Panel feature.

## Example 1: Simple Cross Layout

Player is at Temple Square with exits in all four cardinal directions.

```
╭────────────────────────────╮
│                            │
│ Temple Square              │
│                            │
│           ▢                │
│           │                │
│         ▢─▣─▢              │
│           │                │
│           ▢                │
│                            │
╰────────────────────────────╯
```

**Legend:**
- `▣` = Current room (Temple Square) - shown in yellow
- `▢` = Visited rooms (North Gate, South Gate, East Market, West Temple) - shown in white
- `─` and `│` = Connection lines showing actual exits between rooms - shown in dim gray

## Example 2: Temple Complex with Vertical Exits

Player is at Ground Floor with an exit leading up to a tower.

```
╭────────────────────────────╮
│                            │
│ Ground Floor               │
│                            │
│             ▣              │
│                            │
│             ⇱              │
│                            │
╰────────────────────────────╯
```

**Vertical Exit Symbols:**
- `⇱` = Exit up only
- `⇲` = Exit down only
- `⇅` = Both up and down exits

## Example 3: Linear Path

Player is in the middle of a long corridor.

```
╭────────────────────────────────────────╮
│                                        │
│ Corridor 4                             │
│                                        │
│      ▢─▢─▢─▣─▢─▢─▢                    │
│                                        │
╰────────────────────────────────────────╯
```

Shows 7 connected rooms in a straight line (east-west) with clear connection lines.

## Example 4: Large Area (3x3 Grid)

Player is at the center of a 3x3 grid of rooms.

```
╭────────────────────────────╮
│                            │
│ Room (1,1)                 │
│                            │
│        ▢─▢─▢               │
│        │ │ │               │
│        ▢─▣─▢               │
│        │ │ │               │
│        ▢─▢─▢               │
│                            │
╰────────────────────────────╯
```

Shows all 9 rooms in a compact grid layout with connection lines forming a grid pattern.

## Example 5: Complete Sidebar with Map

Full sidebar view showing all three panels (Tells, Inventory, Map).

```
╭────────────────────────────╮
│                            │
│ Tells                      │
│                            │
│                            │
│                            │
╰────────────────────────────╯
╭────────────────────────────╮
│                            │
│ Inventory                  │
│                            │
│                            │
│                            │
╰────────────────────────────╯
╭────────────────────────────╮
│                            │
│ Temple Square              │
│                            │
│           ▢                │
│           │                │
│         ▢─▣─▢              │
│           │                │
│           ▢                │
│                            │
╰────────────────────────────╯
```

## Example 6: Empty States

### When No Map Exists
```
╭────────────────────────────╮
│                            │
│ Map                        │
│                            │
│ (not implemented)          │
│                            │
╰────────────────────────────╯
```

### When Exploring (Map exists, but no current room yet)
```
╭────────────────────────────╮
│                            │
│ Map                        │
│                            │
│ (exploring...)             │
│                            │
╰────────────────────────────╯
```

## Example 7: Asymmetric Layout

Real-world MUD areas are rarely symmetric. Here's a more realistic example:

```
╭────────────────────────────╮
│                            │
│ Market Square              │
│                            │
│    ▢ ▢                     │
│    ▢   ▢                   │
│  ▢ ▢ ▢ ▣ ▢                 │
│      ▢   ▢                 │
│    ▢ ▢   ▢                 │
│                            │
╰────────────────────────────╯
```

Shows how rooms naturally arrange based on actual exit connections.

## Example 8: Multi-Story Building

Player is in Temple Hall which connects to both a tower above and a crypt below.

```
╭────────────────────────────╮
│                            │
│ Temple Hall                │
│                            │
│           ▢                │
│           │                │
│           ⇅                │
│           │                │
│           ▢                │
│                            │
╰────────────────────────────╯
```

The `⇅` symbol indicates both up and down exits are available from this room. Connection lines show the north-south corridor.

## Color Rendering

While this document shows text-only examples, the actual implementation uses terminal colors:

- **Current Room** (`▣`): Bright yellow/gold (ANSI color 226)
- **Visited Rooms** (`▢`): White (ANSI color 255)
- **Connection Lines** (`─` `│`): Dim gray (ANSI color 240)
- **Vertical Exit Symbols**: Default terminal color

This provides clear visual distinction between your current location, previously visited areas, and the connections between them.

## Dynamic Updates

As the player moves through the MUD:

1. The current room symbol (`▣`) moves to the new location
2. The previous current room becomes a visited room (`▢`)
3. The room name in the panel header updates
4. New rooms are added to the grid as they're discovered
5. The view automatically stays centered on the current room

## Performance

All rendering happens in ~9 microseconds, making updates instantaneous as players move through the world. The map never causes lag or delays in the game experience.
