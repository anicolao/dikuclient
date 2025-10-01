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
```

The command searches for rooms matching all the provided terms (case-insensitive). If multiple rooms match, it will list them and ask you to be more specific.

**Example:**
```
> /point temple square
To reach 'Temple Square', go: south
```

#### `/wayfind <room>`

Shows the complete path from your current location to the destination.

**Usage:**
```
/wayfind temple square
/wayfind market
```

**Example:**
```
> /wayfind temple square
Path to 'Temple Square' (3 steps):
  south -> west -> south
```

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

Each room is identified by a unique ID generated from:
1. Room title
2. First sentence of the description
3. Available exits (sorted)

This ensures that rooms are uniquely identified even if they have similar names, and the same room is always recognized as the same location.

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
