# Design Document: Improved Room Matching and Disambiguation

## 1. Executive Summary

This document proposes a new design for the room matching and disambiguation feature in the DikuMUD client. The current implementation, which uses the distance from the starting room to create unique room IDs, is unreliable and leads to a corrupted map over time. The new design will focus on creating stable, content-based room identifiers and using map connectivity and area detection to resolve ambiguities.

## 2. Problems with the Current Implementation

The existing room matching system suffers from several critical flaws:

1.  **Unreliable Distance Calculation:** The distance from the starting room (room 0) is calculated incrementally as the user explores. This distance is not stable and can change as new paths and shortcuts are discovered, making it a poor choice for a unique identifier.
2.  **ID Instability:** By including the distance in the room ID, the identifier for a room changes whenever a shorter path is found. This breaks existing links in the map and corrupts the map data.
3.  **Lack of Area Awareness:** The current system treats the entire MUD as a single, flat map. It cannot distinguish between two rooms that have identical descriptions but are in different game areas (e.g., "A Forest Path" in the "Elven Woods" vs. "A Forest Path" in the "Trollshaws").
4.  **Complex and Buggy Logic:** The code in `internal/mapper/map.go` has become complex and difficult to maintain due to attempts to work around the fundamental flaws in the distance-based ID system.

## 3. Proposed New Design

The new design will be based on the following principles:

### 3.1. Stable, Content-Based Room IDs

-   **Primary Identifier:** The primary, unique identifier for a room will be a SHA256 hash of its intrinsic and immutable content:
    -   Title
    -   Description (the full description, not just the first sentence)
    -   Exits (the list of available exits, sorted alphabetically)
-   **Human-Readable ID:** The human-readable format (`title|first_sentence|exits`) will be kept as a secondary, non-unique identifier for debugging and manual inspection.
-   **Removal of Distance from ID:** The distance from the starting room will be completely removed from the room ID. Distance will be a calculated property, not a part of the room's identity.

### 3.2. Disambiguation through Map Connectivity

When the client encounters a room that has the same content hash as an existing room in the map, it will use the map's connectivity to determine if it is the same room or a new, identical-looking room.

The logic will be as follows:

1.  When a room is parsed, generate its content hash.
2.  Check if a room with this hash already exists in the map.
3.  **If no room with this hash exists:** It's a new room. Add it to the map.
4.  **If a room with this hash exists:**
    a.  Check the previous room the player was in.
    b.  Does the existing room with the matching hash have an exit that leads back to the player's previous room?
    c.  **If yes:** It is highly probable that the player has moved into the known room. The client will treat it as the same room.
    d.  **If no:** It is likely a new, identical-looking room in a different part of the MUD. The client will treat it as a new room and add it to the map with a disambiguation suffix or a different internal ID.

### 3.3. Introduction of Zone/Area Detection (Future Enhancement)

While not part of the immediate implementation, the design will be structured to support the future addition of zone or area detection. This could be implemented by:

-   Allowing the user to manually set the current zone (`/mapper zone "Elven Woods"`).
-   Automatically detecting zone changes by looking for specific MUD output (e.g., "You have entered the Trollshaws.").
-   When zones are implemented, a room's identity would be a combination of its content hash and its zone.

### 3.4. Simplified and Robust Code

The new design will allow for a significant simplification of the code in `internal/mapper/map.go`. The complex logic for handling distance-based IDs and backtracking will be replaced with a more straightforward implementation of the connectivity-based disambiguation.

## 4. Implementation Plan

The implementation will be carried out in the following steps:

1.  **Modify `room.go`:**
    -   Change `GenerateRoomID` to generate a SHA256 hash of the room's content.
    -   Keep the human-readable format as a separate field in the `Room` struct.
    -   Remove `GenerateRoomIDWithDistance`.
2.  **Refactor `map.go`:**
    -   Simplify `AddOrUpdateRoom` to implement the connectivity-based disambiguation logic described above.
    -   Remove all code related to calculating and storing distance in the room ID.
    -   Ensure that the map saving and loading functions are updated to handle the new `Room` struct and ID format.
3.  **Update `parser.go`:**
    -   Ensure the parser provides the full description to the `Room` creation functions.
4.  **Testing:**
    -   Create new unit tests to verify the new room ID generation and disambiguation logic.
    -   Create integration tests that simulate walking through a MUD with identical rooms to ensure the map is built correctly.

## 5. Conclusion

This new design will provide a more robust, reliable, and maintainable room matching system for the DikuMUD client. By moving to stable, content-based identifiers and using map connectivity for disambiguation, the client will be able to build accurate maps even in complex MUDs with repeating room descriptions.