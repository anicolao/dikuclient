# Mapper Feature Example

This document shows a practical example of using the mapping and navigation features.

## Example Session

```
$ ./dikuclient --host localhost --port 4000

Connected to localhost:4000

Welcome to the MUD!
Name: player

Temple Square
You are standing in a large temple square. The ancient stones
speak of a glorious past.
Exits: north, south, east

> /map
=== Map Information ===
Total rooms explored: 1
Current room: Temple Square
Exits: north, south, east

> north

Temple Hall
A grand hall inside the temple. Torches line the walls.
Exits: north, south, west

> /map
=== Map Information ===
Total rooms explored: 2
Current room: Temple Hall
Exits: north, south, west

> north

Inner Sanctum
The innermost chamber of the temple. Very sacred.
Exits: south

> west
You can't go that way.

> /wayfind temple square
Path to 'Temple Square' (2 steps):
  south -> south

> south

Temple Hall
A grand hall inside the temple. Torches line the walls.
Exits: north, south, west

> south

Temple Square
You are standing in a large temple square. The ancient stones
speak of a glorious past.
Exits: north, south, east

> east

Market Street
A busy market street filled with merchants.
Exits: west, south, north

> /map
=== Map Information ===
Total rooms explored: 4
Current room: Market Street
Exits: west, south, north

> /point temple hall
To reach 'Temple Hall', go: west

> /wayfind sanctum
Path to 'Inner Sanctum' (3 steps):
  west -> north -> north

> /point temple
Found 3 rooms matching 'temple':
  1. Temple Hall
  2. Inner Sanctum
  3. Temple Square
Please be more specific.

> /point temple inner
To reach 'Inner Sanctum', go: west

> /help
=== Client Commands ===
  /point <room>    - Show next direction to reach a room
  /wayfind <room> - Show full path to reach a room
  /map            - Show map information
  /help           - Show this help message

Room search matches all terms in room title, description, or exits
```

## Map File Example

After this session, the map is saved to `~/.config/dikuclient/map.json`:

```json
{
  "rooms": {
    "abc123def456": {
      "id": "abc123def456",
      "title": "Temple Square",
      "description": "You are standing in a large temple square. The ancient stones speak of a glorious past.",
      "first_sentence": "You are standing in a large temple square.",
      "exits": {
        "north": "def456ghi789",
        "south": "",
        "east": "xyz789abc123"
      },
      "visit_count": 2
    },
    "def456ghi789": {
      "id": "def456ghi789",
      "title": "Temple Hall",
      "description": "A grand hall inside the temple. Torches line the walls.",
      "first_sentence": "A grand hall inside the temple.",
      "exits": {
        "north": "ghi789jkl012",
        "south": "abc123def456",
        "west": ""
      },
      "visit_count": 2
    },
    "ghi789jkl012": {
      "id": "ghi789jkl012",
      "title": "Inner Sanctum",
      "description": "The innermost chamber of the temple. Very sacred.",
      "first_sentence": "The innermost chamber of the temple.",
      "exits": {
        "south": "def456ghi789"
      },
      "visit_count": 1
    },
    "xyz789abc123": {
      "id": "xyz789abc123",
      "title": "Market Street",
      "description": "A busy market street filled with merchants.",
      "first_sentence": "A busy market street filled with merchants.",
      "exits": {
        "west": "abc123def456",
        "south": "",
        "north": ""
      },
      "visit_count": 1
    }
  },
  "current_room_id": "xyz789abc123",
  "previous_room_id": "abc123def456",
  "last_direction": "east"
}
```

## Key Features Demonstrated

1. **Automatic Detection**: Rooms are detected from MUD output without any configuration
2. **Bidirectional Linking**: Moving from A to B automatically creates both forward and backward links
3. **Pathfinding**: The `/wayfind` command calculates the shortest path using BFS
4. **Smart Search**: The `/point` command finds rooms by matching keywords
5. **Disambiguation**: When multiple rooms match, the system lists them for clarification
6. **Persistence**: The map persists across sessions, so you never lose progress
7. **Visit Tracking**: The system tracks how many times you've visited each room
8. **Unknown Exits**: Exits that haven't been explored yet show as empty strings

## Room ID Generation

Each room gets a unique ID based on:
- Room title: "Temple Square"
- First sentence: "You are standing in a large temple square."
- Exits: "north, south, east" (sorted alphabetically)

These are combined and hashed to create a consistent, unique identifier. This means:
- The same room always gets the same ID
- Different rooms (even with similar names) get different IDs
- Exit changes create a new room ID (treating it as a different location)

## Pathfinding Algorithm

The system uses Breadth-First Search (BFS) which:
- Always finds the shortest path (minimum number of steps)
- Explores the map level by level from your current location
- Handles complex topologies including loops and one-way passages
- Returns `nil` if no path exists (disconnected areas)

## Usage Tips

1. **Explore first**: The more you explore, the more useful the map becomes
2. **Use specific terms**: If too many matches, add more keywords
3. **Check your location**: Use `/map` to see where you are
4. **Plan routes**: Use `/wayfind` before long journeys
5. **Trust the map**: The pathfinding always gives optimal routes
