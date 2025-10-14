# Barsoom Room Parsing Mode

## Overview

This feature implements a "Barsoom mode" for room parsing. Once the client detects the `--<` marker (which indicates a Barsoom MUD room format), it permanently switches to Barsoom-only parsing and no longer uses heuristic room detection.

## Key Changes

### 1. Barsoom Mode Flag

Added a `barsoomMode` flag to the `Model` struct in `internal/tui/app.go`:

```go
barsoomMode bool // True if we've ever seen --< marker (switch to Barsoom parsing only)
```

Once this flag is set to `true`, the client will only attempt to parse rooms using the Barsoom format and will not fall back to heuristic parsing.

### 2. Mode Switching Logic

In `detectAndUpdateRoom()` (internal/tui/app.go):

- When a Barsoom room is detected, the `barsoomMode` flag is set to `true`
- Once in Barsoom mode, if no Barsoom room is found in the recent output, the function returns early without attempting heuristic parsing
- This ensures that only Barsoom-formatted rooms are parsed after the first `--<` marker is seen

### 3. Full Description for Room Disambiguation

For Barsoom rooms, the entire description is used to disambiguate rooms instead of just the first sentence.

New functions in `internal/mapper/room.go`:

- `GenerateBarsoomRoomID(title, description string, exits []string) string`: Generates a room ID using the full description
- `NewBarsoomRoom(title, description string, exits []string) *Room`: Creates a new room with an ID based on the full description

Example:

**Regular Room ID (first sentence only):**
```
temple square|you are standing in a large temple square.|east,north,south
```

**Barsoom Room ID (full description):**
```
temple square|you are standing in a large temple square. the ancient stones speak of a glorious past.|east,north,south
```

This allows for better disambiguation of rooms with similar titles but different descriptions.

## Behavior

### Before Barsoom Mode

- The client uses heuristics to detect rooms from MUD output
- Rooms are identified by title and first sentence of description

### After Seeing `--<` Marker

- The `barsoomMode` flag is set to `true`
- Only Barsoom-formatted rooms (with `--<` and `>--` markers) are detected
- Heuristic parsing is completely disabled
- Rooms are identified by title and **full description**

### Regular Output in Barsoom Mode

When regular (non-Barsoom) output is received in Barsoom mode:
- The description split is cleared (no room description shown in the sticky viewport)
- No room is added to the map (because it's not in Barsoom format)
- The user must use Barsoom-formatted room descriptions for mapping

## Testing

Added comprehensive tests in `internal/tui/barsoom_test.go`:

- `TestBarsoomModeSwitching`: Verifies that the mode switches permanently after seeing `--<`
- Tests confirm that regular output is ignored in Barsoom mode

Added tests in `internal/mapper/room_test.go`:

- `TestGenerateBarsoomRoomID`: Verifies full description is used for Barsoom room IDs
- `TestNewBarsoomRoom`: Verifies `NewBarsoomRoom` creates rooms with full description IDs

## Example Usage

When connected to a Barsoom MUD:

```
119H 110V 3674X 0.00% 77C T:56 Exits:EW>
--<
Temple Square
    You are standing in a large temple square. The ancient stones
speak of a glorious past.
>-- Exits:NSE
```

After seeing this output:
1. The client switches to Barsoom mode
2. The room is parsed with full description for the ID
3. Future rooms must use Barsoom format to be detected
4. The description split viewport shows the room description at the top

## Benefits

1. **Definitive Room Detection**: No ambiguity about room boundaries when `--<` and `>--` markers are present
2. **Better Disambiguation**: Using full description instead of first sentence provides more accurate room identification
3. **Clean Separation**: Once in Barsoom mode, no heuristic guessing is needed
4. **Backward Compatible**: Regular rooms still work before Barsoom mode is activated
