# Intended Design of the Room Mapper

This document describes the intended functionality of the room mapping feature as inferred from the existing codebase and user feedback. The goal is to establish a common understanding of the design before attempting to fix any bugs.

## 1. Core Principle: Path-Dependent Room Identity

The fundamental principle of this mapper is that a room's identity is defined not just by its content (title, description, exits) but also by the **path taken to reach it**.

-   **Unique ID:** A room's unique identifier is a composite key containing its content and its distance from the starting room (room 0).
-   **Format:** `title|first_sentence|exits|distance`
-   **Example:** If a room titled "A Fork in the Road" is discovered 5 steps from the start, its ID might be `a fork in the road|the path splits here.|north,south|5`. If an identical-looking room is found 12 steps from the start, it is treated as a completely separate entity with the ID `a fork in the road|the path splits here.|north,south|12`.

The map represents the explored portion of the MUD's world graph. Since there is no way to be certain if two rooms with identical descriptions are the same physical room, the mapper treats rooms that are reachable via different length paths from the starting room as distinct entities. This provides an accurate map of the explored areas under this constraint.

## 2. Distance Calculation: Incremental Path Length

The `distance` component of the room ID is calculated incrementally based on the player's movement.

-   **Initial Room:** The first room discovered is considered "room 0" and has a distance of `0`.
-   **Movement:** When the player moves from a room at `distance: N` to a new room, the new room is assigned `distance: N + 1`.
-   **No Global Optimization:** The system does not perform a global recalculation (like a Breadth-First Search from the start) to find the absolute shortest path. The distance recorded is always based on the path of discovery.

## 3. Room Discovery and Updates (`AddOrUpdateRoom`)

The `AddOrUpdateRoom` function is the core of the mapping logic. Its intended behavior is as follows:

1.  **Parse Room Info:** When new MUD output is processed, the client parses it to get the current room's title, description, and exits.
2.  **Handle Revisiting Rooms (Backtracking):** A room can only be revisited by following a "tentative link". When moving from Room A to Room B, a tentative link is created from Room B back to Room A's full ID (`...|distance`).
    -   When the player moves, the system checks if the `LastDirection` from the current room corresponds to a tentative link.
    -   If it does, the client checks if the newly parsed room's content and distance (current distance +/- 1) are consistent with the room ID stored in the tentative link.
    -   If they are consistent, the link is ratified, and the player's `CurrentRoomID` is updated to the existing room's ID. No new room is created.
3.  **Process as a New Room:** If the move does not correspond to a valid tentative link, the system treats it as a new room discovery.
    -   It calculates the new room's distance by taking the current room's distance and adding 1.
    -   It generates a new, unique room ID using this new distance (`title|...|distance`).
    -   It adds this new room to the map's collection of rooms.
    -   It updates `PreviousRoomID` and `CurrentRoomID`.

## 4. Room Linking (`LinkRooms`)

After a movement has been confirmed and the new room has been processed by `AddOrUpdateRoom`, the `LinkRooms` function is called.

-   It creates a one-way link from the `PreviousRoomID` to the `CurrentRoomID` in the direction of `LastDirection`.
-   It also creates a "best guess" reverse link from the `CurrentRoomID` back to the `PreviousRoomID`. This reverse link is tentative and will be confirmed if the player moves back in that direction.

## 5. Summary of Intended State Flow

1.  Player is in Room A (ID: `...|5`).
2.  Player types "north". `SetLastDirection("north")` is called.
3.  MUD sends back the description of Room B.
4.  The client parses Room B's info.
5.  `AddOrUpdateRoom` is called with Room B's info.
    -   It determines this is not backtracking.
    -   It calculates Room B's distance as `5 + 1 = 6`.
    -   It creates a new Room B with ID `...|6` and adds it to the map.
    -   It sets `PreviousRoomID` to `...|5` and `CurrentRoomID` to `...|6`.
6.  `LinkRooms` is called.
    -   It sets the "north" exit of Room A (`...|5`) to point to Room B's ID (`...|6`).
    -   It sets the "south" exit of Room B (`...|6`) to point to Room A's ID (`...|5`).
7.  The map is saved.