# Tick Timer Feature

The tick timer feature allows you to automate commands based on the MUD's tick counter. When your MUD displays a tick timer (e.g., `T:24` in the prompt), the client will track it and execute commands at specific tick times.

## Overview

Many DikuMUDs display a tick counter in the prompt showing seconds until the next "tick" (a game event that occurs at regular intervals, typically every 60 or 75 seconds). The tick timer feature:

1. **Automatically detects** tick times from prompts containing `T:NN` patterns
2. **Persists the tick interval** for each server (no manual configuration needed)
3. **Projects the current tick time** even when no new prompt appears
4. **Executes commands** at specific tick times based on your configured triggers

## Usage

### Adding Tick Triggers

Add a tick trigger to execute commands at a specific tick time:

```
/ticktrigger <tick_time> "commands"
```

**Examples:**

```bash
# Cast heal spell at T:5
/ticktrigger 5 "cast 'heal'"

# Execute multiple commands at T:4 (separated by semicolons)
/ticktrigger 4 "cast 'bless';say Ready for battle!"

# Cast a buff at T:10
/ticktrigger 10 "cast 'armor'"
```

### Listing Tick Triggers

View all active tick triggers:

```
/ticktriggers list
```

Output example:
```
=== Active Tick Triggers (Interval: 75s) ===
  1. T:5 -> "cast 'heal'"
  2. T:4 -> "cast 'bless';say Ready!"
  3. T:10 -> "cast 'armor'"
(Current estimated tick time: T:24)
```

### Removing Tick Triggers

Remove a tick trigger by its number:

```
/ticktriggers remove <number>
```

Example:
```
/ticktriggers remove 1
```

## How It Works

### Tick Detection

The client automatically detects tick times from MUD output. It looks for the pattern `T:NN` in prompts. For example:

```
101H 132V 54710X 49.60% 570C [Hero:Good] [goblin:Bad] T:24 Exits:NS>
```

From this prompt, the client extracts `T:24` and knows there are 24 seconds until the next tick.

### Tick Interval

The first time a tick is detected, the client sets the tick interval to 75 seconds (a common default for DikuMUDs). This interval is automatically saved and will be reused for future connections to the same server.

If your MUD uses a different interval (e.g., 60 seconds), the tick timer will adjust over time as it receives more prompts.

### Tick Projection

Once the client has seen a tick time, it maintains an internal timer that projects the current tick time. This means:

- You don't need to constantly receive prompts for the timer to work
- Commands will fire at the correct time even if you're idle
- The timer automatically adjusts when new prompts arrive

### Command Execution

When the projected tick time matches a trigger's tick time:

1. The commands are queued for execution
2. Multiple commands (separated by `;`) are executed sequentially
3. There's a 1-second delay between each command
4. The queue can be stopped with `/stop`

## Configuration Files

Tick timer settings are saved per-server in:
```
~/.config/dikuclient/ticktimer_<host>_<port>.json
```

This file contains:
- Tick interval (e.g., 75 seconds)
- Last seen tick time and timestamp
- All configured tick triggers

You can manually edit this file if needed, but it's generally not necessary.

## Tips and Best Practices

### Timing Your Actions

- **Healing**: Set healing triggers a few seconds before tick (e.g., `T:5`) to ensure you're healthy when the tick hits
- **Buffs**: Apply buffs just after tick (e.g., `T:74`) to maximize their duration
- **Fighting**: Some MUDs have combat advantages at certain tick times

### Multiple Triggers at Same Time

You can have multiple triggers at the same tick time:

```
/ticktrigger 5 "cast 'heal'"
/ticktrigger 5 "drink potion"
/ticktrigger 5 "say Healing up!"
```

All commands will be queued and executed in order.

### Using with Regular Triggers

Tick triggers work well with regular text-based triggers. For example:

```
# Regular trigger: Auto-attack when enemy arrives
/trigger "<enemy> has arrived" "kill <enemy>"

# Tick trigger: Heal before tick
/ticktrigger 5 "cast 'heal'"
```

### Checking Tick Timer Status

Use `/ticktriggers list` to see:
- Your current tick interval
- All active tick triggers
- Current estimated tick time

## Examples

### PvP Setup

```bash
# Refresh armor at T:72 (right after tick)
/ticktrigger 72 "cast 'armor'"

# Refresh shield at T:71
/ticktrigger 71 "cast 'shield'"

# Heal at T:5 (before next tick)
/ticktrigger 5 "cast 'heal'"
```

### Efficient Questing

```bash
# Use recall command at T:2 (safe time to leave)
/ticktrigger 2 "recall"

# Rebuff after returning
/ticktrigger 70 "cast 'armor';cast 'shield';cast 'bless'"
```

### Idle Regeneration

```bash
# Remind yourself to stay active
/ticktrigger 1 "say Still here!"

# Automatic healing when idle
/ticktrigger 5 "cast 'heal'"
```

## Troubleshooting

### Tick Not Detected

If the client doesn't detect ticks:
1. Check that your MUD includes `T:NN` in prompts
2. Try typing a command to get a fresh prompt
3. Use `/ticktriggers list` to see if interval is set

### Wrong Tick Interval

If the interval seems incorrect:
1. The client will learn the correct interval over time
2. You can manually edit the config file if needed
3. Delete the config file to force re-detection

### Commands Not Firing

If commands don't execute at the expected time:
1. Check that you're connected to the MUD
2. Use `/ticktriggers list` to verify triggers are configured
3. Ensure your current tick time is close to the trigger time
4. Check the tick interval is correct

### Multiple Commands Not Working

If multi-command triggers don't work:
1. Ensure commands are separated by semicolons (`;`)
2. Check that each command is valid
3. Remember there's a 1-second delay between commands

## Getting Help

For detailed help on tick triggers:
```
/help ticktrigger
```

For general help:
```
/help
```

## Technical Details

- Timer precision: Updates every second
- Supported prompt formats: Any prompt containing `T:NN` (where NN is 1-99)
- ANSI support: Works with colored prompts
- Persistence: Per-server configuration files
- Thread-safe: Uses Go's bubbletea message passing

## See Also

- `/help trigger` - Text-based triggers
- `/help alias` - Command aliases
- `/help stop` - Stop command execution
