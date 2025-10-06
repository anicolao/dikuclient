# Alias Feature Documentation

## Overview

The alias feature allows you to create shortcuts for commonly used commands with parameter substitution. Similar to triggers, aliases can use variable placeholders like `<target>` that get replaced with actual values when you use the alias.

## Basic Usage

### Creating an Alias

Use the `/alias` command with the alias name and template:

```
/alias "name" "template"
```

**Example:**
```
/alias "gat" "give all <target>"
```

Now when you type `gat mary`, it will be expanded to `give all mary` and sent to the server.

### Listing Aliases

To see all your defined aliases:

```
/aliases list
```

or simply:

```
/aliases
```

### Removing an Alias

To remove an alias by its number:

```
/aliases remove <number>
```

**Example:**
```
/aliases remove 1
```

## Parameter Substitution

Aliases support flexible parameter substitution based on the number of parameters in the template and the number of arguments provided.

### Single Placeholder: `<target>`

When the template has only one placeholder, all arguments are joined with spaces and substituted:

**Example:**
```
/alias "k" "kill <target>"
k goblin              → kill goblin
k big scary goblin    → kill big scary goblin
```

### Two Placeholders: `<object> <target>`

With two placeholders, the first argument goes to the first placeholder, the second to the second:

**Example:**
```
/alias "gt" "give <object> <target>"
gt sword mary    → give sword mary
gt bread bob     → give bread bob
```

### Variable Arguments with `<args>`

Use `<args>` as a special placeholder to capture all remaining arguments:

#### Pattern: `<target> <args>`

**Example:**
```
/alias "tell" "tell <target> <args>"
tell mary hello there friend    → tell mary hello there friend
```

#### Pattern: `<arg1> <arg2> <args>`

**Example:**
```
/alias "action" "perform <arg1> <arg2> <args>"
action one two three four    → perform one two three four
```

#### Pattern: `<target> <arg1> <arg2> <args>`

**Example:**
```
/alias "complex" "cmd <target> <arg1> <arg2> <args>"
complex bob first second the rest    → cmd bob first second the rest
```

## Complete Examples

### Common MUD Aliases

```
# Quick kill
/alias "k" "kill <target>"

# Give all items to a player
/alias "gat" "give all <target>"

# Give specific item to a player
/alias "gt" "give <object> <target>"

# Quick casting
/alias "fb" "cast fireball <target>"

# Tell with automatic formatting
/alias "t" "tell <target> <args>"

# Quick movement aliases
/alias "n" "north"
/alias "s" "south"
/alias "e" "east"
/alias "w" "west"
```

### Advanced Usage: Multiple Commands with Semicolons

Aliases can contain multiple commands separated by semicolons (`;`). Each command will be sent sequentially with a 1-second delay between them, allowing you to automate multi-step sequences.

```
# Multi-step command alias - sends 2 commands, one per second
/alias "prep" "get all from corpse;sacrifice corpse"

# Combat macro - bash followed by consider
/alias "combo" "bash <target>;consider <target>"

# Social interaction - bow then speak
/alias "greet" "bow <target>;say Hello, <target>!"

# Complex sequence - up to 5 actions
/alias "fullprep" "kill <target>;get all from corpse;sacrifice corpse;sit;rest"
```

**Important Notes:**
- Commands are queued and sent one per second automatically
- Use `/stop` command to cancel a queued sequence
- Additional aliases or triggers will add to the end of the queue
- This works the same way as the `/go` auto-walking feature

## Persistence

Aliases are automatically saved to `~/.config/dikuclient/aliases.json` (or `$DIKUCLIENT_CONFIG_DIR/aliases.json` if set) when you add or remove them. They will be loaded automatically the next time you start the client.

## Comparison with Triggers

| Feature | Aliases | Triggers |
|---------|---------|----------|
| **Activation** | User types the alias name | Server output matches pattern |
| **Purpose** | Command shortcuts | Automated responses to game events |
| **Parameters** | Flexible argument substitution | Variable capture from pattern |
| **Example** | `/alias "k" "kill <target>"` | `/trigger "hungry" "eat bread"` |

## Tips

1. **Keep alias names short** - The whole point is to save typing!
2. **Use consistent naming** - Prefix related aliases (e.g., `ga` for "give all", `gt` for "give to")
3. **Document complex aliases** - Use `/aliases list` to remind yourself what you've created
4. **Test your aliases** - Try them with different arguments to ensure they work as expected

## Troubleshooting

### Alias doesn't work

- Check that you created it correctly with `/aliases list`
- Verify the alias name matches exactly (aliases are case-sensitive)
- Ensure you're not using special characters in the alias name (only alphanumeric characters allowed)

### Wrong expansion

- Review the parameter substitution rules above
- The number of placeholders determines how arguments are distributed
- Use `<args>` as a placeholder to capture multiple remaining arguments

### Can't create alias

- Alias names must be alphanumeric (no spaces, hyphens, or special characters)
- You cannot create an alias with the same name as an existing alias
- Use `/aliases remove` to delete an existing alias first if you want to replace it
