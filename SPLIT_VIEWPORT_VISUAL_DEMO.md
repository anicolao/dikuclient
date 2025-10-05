# Split Viewport Visual Demonstration

## Feature Behavior

This document shows the visual behavior of the split viewport feature with ASCII art examples.

### State 1: Normal Mode (Before Scroll)
User is at the bottom, viewing live output. **No split active.**

```
┌────────────────────────────────────────────────────────────────┐
│ Connected to mud.server.com:4000                               │
├────────────────────────────────────────────────────────────────┤
│ ┌── Temple Square [n,s,e,w] ─────┐  ┌── Tells ──────────────┐ │
│ │                                 │  │                       │ │
│ │ Line 90: Previous output...     │  │ Gandalf: Hi there!    │ │
│ │ Line 91: More output...         │  │ Frodo: Hello back!    │ │
│ │ Line 92: Even more output...    │  │                       │ │
│ │ Line 93: Keep going...          │  ├───────────────────────┤ │
│ │ Line 94: Almost there...        │  │                       │ │
│ │ Line 95: Getting closer...      │  │  [Inventory Panel]    │ │
│ │ Line 96: Nearly at end...       │  │                       │ │
│ │ Line 97: Just a bit more...     │  ├───────────────────────┤ │
│ │ Line 98: Almost done...         │  │                       │ │
│ │ Line 99: Last line!             │  │  [Map Panel]          │ │
│ │ Line 100: Current output ▂      │  │                       │ │
│ │                                 │  └───────────────────────┘ │
│ └─────────────────────────────────┘                            │
└────────────────────────────────────────────────────────────────┘
```

### State 2: User Presses Page Up
User scrolls back to see earlier output. **Split mode activates!**

```
┌────────────────────────────────────────────────────────────────┐
│ Connected to mud.server.com:4000                               │
├────────────────────────────────────────────────────────────────┤
│ ┌── Temple Square [n,s,e,w] ─────┐  ┌── Tells ──────────────┐ │
│ │ 📖 READING HISTORY (2/3)        │  │                       │ │
│ │                                 │  │ Gandalf: Hi there!    │ │
│ │ Line 50: Older output here...   │  │ Frodo: Hello back!    │ │
│ │ Line 51: Previous combat...     │  │                       │ │
│ │ Line 52: You killed goblin!     │  ├───────────────────────┤ │
│ │ Line 53: XP gained: 150         │  │                       │ │
│ │ Line 54: More history...        │  │  [Inventory Panel]    │ │
│ │ Line 55: Reading back...        │  │                       │ │
│ │ Line 56: User scrolled here ▲   │  ├───────────────────────┤ │
│ ├─────────────────────────────────┤  │                       │ │
│ │ 🔴 LIVE OUTPUT (1/3)            │  │  [Map Panel]          │ │
│ │                                 │  │                       │ │
│ │ Line 98: Current output...      │  └───────────────────────┘ │
│ │ Line 99: Still happening...     │                            │
│ │ Line 100: Latest message! ▂     │                            │
│ └─────────────────────────────────┘                            │
└────────────────────────────────────────────────────────────────┘
         ↑                                ↑
    User's scroll               Live tracking continues
    position preserved          (always at bottom)
```

### State 3: More Output Arrives While Split
New lines appear ONLY in bottom pane. **Split remains active.**

```
┌────────────────────────────────────────────────────────────────┐
│ Connected to mud.server.com:4000                               │
├────────────────────────────────────────────────────────────────┤
│ ┌── Temple Square [n,s,e,w] ─────┐  ┌── Tells ──────────────┐ │
│ │ 📖 READING HISTORY (2/3)        │  │                       │ │
│ │                                 │  │ Gandalf: Hi there!    │ │
│ │ Line 50: Older output here...   │  │ Frodo: Hello back!    │ │
│ │ Line 51: Previous combat...     │  │ Legolas: Greetings!   │ │
│ │ Line 52: You killed goblin!     │  │                       │ │
│ │ Line 53: XP gained: 150         │  ├───────────────────────┤ │
│ │ Line 54: More history...        │  │                       │ │
│ │ Line 55: Reading back...        │  │  [Inventory Panel]    │ │
│ │ Line 56: User scrolled here ▲   │  │                       │ │
│ ├─────────────────────────────────┤  ├───────────────────────┤ │
│ │ 🔴 LIVE OUTPUT (1/3)            │  │                       │ │
│ │                                 │  │  [Map Panel]          │ │
│ │ Line 101: NEW! Orc attacks!     │  │                       │ │
│ │ Line 102: NEW! You defend!      │  └───────────────────────┘ │
│ │ Line 103: NEW! Combat! ▂        │                            │
│ └─────────────────────────────────┘                            │
└────────────────────────────────────────────────────────────────┘
         ↑                                ↑
    Still at Line 56              New lines only appear here!
    (unchanged)                   User can see updates
```

### State 4: User Scrolls Back Down (Page Down)
User returns to bottom. **Split disappears automatically.**

```
┌────────────────────────────────────────────────────────────────┐
│ Connected to mud.server.com:4000                               │
├────────────────────────────────────────────────────────────────┤
│ ┌── Temple Square [n,s,e,w] ─────┐  ┌── Tells ──────────────┐ │
│ │                                 │  │                       │ │
│ │ Line 97: Previous output...     │  │ Gandalf: Hi there!    │ │
│ │ Line 98: More output...         │  │ Frodo: Hello back!    │ │
│ │ Line 99: Even more output...    │  │ Legolas: Greetings!   │ │
│ │ Line 100: Latest message...     │  │                       │ │
│ │ Line 101: Orc attacks!          │  ├───────────────────────┤ │
│ │ Line 102: You defend!           │  │                       │ │
│ │ Line 103: Combat continues!     │  │  [Inventory Panel]    │ │
│ │ Line 104: You win!              │  │                       │ │
│ │ Line 105: XP gained!            │  ├───────────────────────┤ │
│ │ Line 106: Back to normal ▂      │  │                       │ │
│ │                                 │  │  [Map Panel]          │ │
│ │                                 │  │                       │ │
│ └─────────────────────────────────┘  └───────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
                    ↑
            Back to single pane mode
            (no split, at bottom)
```

## Key Features Demonstrated

1. **Automatic Split**: No manual command needed - just scroll up!
2. **Dual View**: See both history AND live output simultaneously
3. **Smart Tracking**: Bottom pane always shows latest output
4. **Preserved Position**: Top pane stays where you scrolled to
5. **Automatic Unsplit**: Returns to normal when you scroll to bottom
6. **Seamless Updates**: New content appears in bottom pane while split

## User Experience Benefits

### Problem Solved
**Before**: Scrolling up meant missing new output  
**After**: Can read history while tracking live updates

### Use Cases
1. **Combat Review**: Check previous hits while fight continues
2. **Tell History**: Read old messages while receiving new ones
3. **Room Description**: Review past rooms while exploring
4. **XP Analysis**: Study XP gains while still fighting

### Interaction
- **Page Up / Mouse Wheel Up**: Enter split mode
- **Page Down / Mouse Wheel Down**: Exit split mode (when at bottom)
- **No Manual Toggle**: Everything automatic based on scroll position

## Implementation Notes

### Technical Details
- Split ratio: 2/3 (history) : 1/3 (live)
- Both viewports share same content
- Minimal overhead (only when split)
- Tested with 6 comprehensive unit tests

### Border Rendering
- Top pane: Normal top border, no bottom border
- Bottom pane: Horizontal line separator, normal bottom border
- Seamless visual integration with sidebar panels

### Performance
- No content duplication
- Efficient viewport updates
- Immediate response to scroll events
