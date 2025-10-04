# Map Building and Navigation Guide

DikuClient now includes an automatic mapping system that tracks rooms as you explore and provides navigation commands to help you find your way around the MUD.

## Features

### Automatic Map Building

As you move through the MUD, the client automatically:
- Detects room information (title, description, exits)
- Creates a unique room ID based on room characteristics
- Links rooms together based on your movement
- Persists the map to disk for future sessions

The map is saved to `~/.config/dikuclient/map.json` and loads automatically when you connect.

### Room Detection

The mapper recognizes common MUD room formats:
- Room titles (first non-empty line)
- Room descriptions (text before exits line)
- Exit lines in various formats:
  - `Exits: north, south, east`
  - `[ Exits: n s e w ]`
  - `Obvious exits: north and south`

The system handles ANSI color codes and various formatting styles.

### Navigation Commands

All client commands start with a forward slash (`/`) to distinguish them from MUD commands.

#### `/point <room>`

Shows the next direction to take to reach a destination.

**Usage:**
```
/point temple square
/point market
/point fountain north
/point 1              # Select first room from previous search
/point 2 temple       # Select second room matching 'temple'
```

The command searches for rooms matching all the provided terms (case-insensitive). If multiple rooms match, it will list them and ask you to be more specific. You can then use `/point <number>` to select a specific room from the list, or `/point <number> <search terms>` to search and select in one command.

**Example:**
```
> /point temple square
To reach 'Temple Square', go: south

> /point temple
Found 3 rooms matching 'temple':
  1. Inner Sanctum
  2. Temple Entrance
  3. Temple Square
Please be more specific, or use /point <number> to select a room.

> /point 2
To reach 'Temple Entrance', go: north
```

#### `/wayfind <room>`

Shows the complete path from your current location to the destination.

**Usage:**
```
/wayfind temple square
/wayfind market
/wayfind 1              # Select first room from previous search
/wayfind 2 temple       # Select second room matching 'temple'
```

**Example:**
```
> /wayfind temple square
Path to 'Temple Square' (3 steps):
  1. south -> Market Square
  2. west -> Main Street
  3. south -> Temple Square

> /wayfind temple
Found 3 rooms matching 'temple':
  1. Inner Sanctum
  2. Temple Entrance
  3. Temple Square
Please be more specific, or use /wayfind <number> to select a room.

> /wayfind 3
Path to 'Temple Square' (3 steps):
  south -> west -> south
```

#### `/go <room>`

Auto-walks to the specified room, executing one movement command per second.

**Usage:**
```
/go temple square
/go market
/go                   # Cancel current auto-walk
/go 1                 # Select first room from previous search
/go 2 temple          # Select second room matching 'temple'
```

The auto-walk can be cancelled at any time by typing `/go` with no arguments.

When multiple rooms match your search, you can use `/go <number>` to select a specific room from the list, or `/go <number> <search terms>` to search and select in one command.

**Example:**
```
> /go temple square
Auto-walking to 'Temple Square' (3 steps). Type /go to cancel.
[Auto-walk: south (1/3)]
[Auto-walk: west (2/3)]
[Auto-walk: south (3/3)]
[Auto-walk complete!]

> /go temple
Found 3 rooms matching 'temple':
  1. Inner Sanctum
  2. Temple Entrance
  3. Temple Square
Please be more specific, or use /go <number> to select a room.

> /go 1
Auto-walking to 'Inner Sanctum' (5 steps). Type /go to cancel.
[...]

> /go
Auto-walk cancelled.
```

#### `/rooms [filter]`

Lists all explored rooms. Optionally filter by search terms to find specific rooms. The room numbers shown can be used with `/go <number>`, `/point <number>`, or `/wayfind <number>`.

**Usage:**
```
/rooms                  # List all rooms
/rooms temple           # List only rooms matching 'temple'
/rooms market street    # List rooms matching both 'market' and 'street'
```

**Example:**
```
> /rooms
=== Known Rooms (42) ===
  1. Dark Alley [north, south]
  2. Market Square [north, south, east, west]
  3. Temple Entrance [north, east]
  ...

> /rooms temple
=== Rooms matching 'temple' (3) ===
  1. Inner Sanctum [south]
  2. Temple Entrance [north, east]
  3. Temple Square [north, south, east]

> /go 2
Auto-walking to 'Temple Entrance' (2 steps). Type /go to cancel.
```

This is useful for checking which rooms match your search terms before using `/point`, `/wayfind`, or `/go`. You can then select a specific room by number.

#### `/nearby`

Lists all rooms within 5 steps of your current location, organized by distance. This helps you understand what's around you and plan your exploration.

**Usage:**
```
/nearby
```

**Example:**
```
> /nearby
=== Nearby Rooms (12 within 5 steps) ===
1 step away:
  1. Market Square [north, south, east, west]
  2. Temple Entrance [north, east]
  3. Dark Alley [south]
2 steps away:
  4. Town Square [north, south, east]
  5. Guild Hall [west, east]
3 steps away:
  6. Fountain Plaza [north, south]
  7. City Gates [south, west]
  ...
```

The rooms are displayed in order by distance (closest first), making it easy to see what's immediately around you versus what's further away. This is particularly useful when deciding where to explore next or when looking for a specific type of location nearby.

When you run `/nearby`, the rooms are numbered and these numbers temporarily appear on the visual map display (in the sidebar), replacing the usual room symbols. This makes it easy to correlate the list with the visual map.

#### `/legend`

Lists all rooms currently on the map, sorted alphabetically by room title.

**Usage:**
```
/legend
```

**Example:**
```
> /legend
=== All Mapped Rooms (42) ===
  1. Armory [north, south]
  2. Barracks [west, east]
  3. City Gates [south, east]
  4. Dark Alley [north, south]
  5. East Market [west, north]
  6. Fountain Plaza [north, south, east]
  ...
```

Like `/nearby`, when you run `/legend`, the rooms are numbered and these numbers appear on the visual map display, replacing the usual room symbols. This provides a complete overview of all explored areas and their locations on the map.

#### `/map`

Shows information about the current map state.

**Example:**
```
> /map
=== Map Information ===
Total rooms explored: 42
Current room: Inner Sanctum
Exits: south
```

#### `/help`

Shows a list of available client commands.

## How It Works

### Room Identification

Each room is identified by a unique ID in a human-readable format:

**Format:** `title|first_sentence|exits`

**Example:** `temple square|a large temple square with ancient stones.|east,north`

The ID is generated from:
1. Room title (lowercase)
2. First sentence of the description (lowercase)
3. Available exits (sorted, comma-separated)

This format ensures:
- Rooms are uniquely identified even if they have similar names
- The same room is always recognized as the same location
- The map JSON file is human-readable for debugging and manual inspection
- You can easily see room connections when viewing the map file

### Room Linking

As you move between rooms:
1. The client detects your movement command (n, north, s, south, etc.)
2. When a new room is detected, it creates a link from the previous room
3. Reverse links are automatically created (e.g., if you go north, a south exit is added to the new room)

### Pathfinding

The pathfinding algorithm uses Breadth-First Search (BFS) to find the shortest path between any two rooms in the explored map. This ensures you always get the most efficient route.

### Room Search

The search system matches all query terms against:
- Room title
- First sentence of description  
- Exit names

All matching is case-insensitive, and all terms must be present for a match.

## Tips

1. **Explore thoroughly**: The more you explore, the more useful the map becomes
2. **Use specific search terms**: If `/point temple` matches too many rooms, try `/point temple north`
3. **Check the map**: Use `/map` to see how many rooms you've discovered
4. **Movement detection**: The mapper automatically detects cardinal directions (n, s, e, w, ne, nw, se, sw, up, down)

## Limitations

- Only standard movement commands are detected (not custom commands like "enter", "climb", etc.)
- Rooms must follow common MUD formatting to be detected automatically
- The map doesn't handle area resets or dynamic room changes
- Maximum of 30 recent lines are checked for room detection

## Future Enhancements

Potential improvements for the mapping system:
- Visual ASCII map display in sidebar
- Mark special rooms (shops, banks, trainers)
- Support for custom movement aliases
- Import/export maps
- Share maps between users
- Mark dangerous areas
- Track mob spawns

## Technical Details

The mapping system is implemented in the `internal/mapper` package and includes:
- `room.go`: Room data structure and ID generation
- `map.go`: Map management, pathfinding, and persistence
- `parser.go`: MUD output parsing and room detection

The map data is stored as JSON in the config directory with proper file permissions (0600) to ensure privacy.
