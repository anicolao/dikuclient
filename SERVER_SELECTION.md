# Two-Step Server and Character Selection

This document describes the new two-step server and character selection process in the DikuMUD client.

## Overview

The client now separates server selection from character selection, making it easier to manage multiple characters across different servers.

## Terminal Mode

When you run `./dikuclient` without any parameters, you'll see a two-step selection menu:

### Step 1: Server Selection

The first menu shows:
1. **Add a new server** - Create a new server entry
2. **Servers** - List of all saved servers
3. **Characters** - List of all saved characters with their servers
4. **Exit** - Exit the application

Example:
```
DikuMUD Client - Server Selection
==================================
  1. Add a new server

Servers:
  2. AardMUD (aardmud.org:23)
  3. MyTestServer (localhost:4000)

Characters:
  4. Warrior on aardmud.org:23
  5. Mage on aardmud.org:23
  6. TestChar on localhost:4000
  7. Exit

Select option:
```

### Step 2: Character Selection

After selecting a server, you'll see the character selection menu for that server:

1. **Create a new character** - Create a new character on this server
2. **Existing characters** - List of characters for the selected server
3. **Back to server selection** - Return to the server menu

Example:
```
Character Selection for AardMUD (aardmud.org:23)
====================================
  1. Create a new character

Existing characters:
  2. Warrior (mywarrior)
  3. Mage (mymage)
  4. Back to server selection

Select option:
```

## Adding Servers

When you select "Add a new server", you'll be prompted for:
- **Server name**: A friendly name for the server (e.g., "AardMUD")
- **Hostname**: The server hostname (e.g., "aardmud.org")
- **Port**: The server port (default: 4000)

The server is automatically saved to your configuration.

## Creating Characters

When you select "Create a new character", you'll be prompted for:
- **Character name** (optional): A friendly name for this character
- **Username** (optional): The login username for this character
- **Password** (optional): The password for this character

**Important**: If you provide a character name, the character will be automatically saved. If you don't provide a name (leave it blank), you'll connect anonymously without saving the character.

## Web Mode with URL Parameters

You can now specify server and port via URL parameters in web mode:

```
http://localhost:8080/?server=aardmud.org&port=23
```

When these parameters are provided:
- The client jumps directly to the character selection menu for that server
- You don't need to go through the server selection step
- The server information is automatically used for the connection

This is useful for:
- Bookmarking direct connections to specific servers
- Sharing URLs with friends for quick access
- Creating server-specific shortcuts

### Example URLs

```
# Connect to AardMUD
http://localhost:8080/?server=aardmud.org&port=23

# Connect to a local test server
http://localhost:8080/?server=localhost&port=4000

# With a specific session ID
http://localhost:8080/?id=my-session&server=aardmud.org&port=23
```

## Data Storage

The configuration is stored in `~/.config/dikuclient/accounts.json`:

```json
{
  "servers": [
    {
      "name": "AardMUD",
      "host": "aardmud.org",
      "port": 23
    }
  ],
  "characters": [
    {
      "name": "Warrior",
      "host": "aardmud.org",
      "port": 23,
      "username": "mywarrior"
    }
  ]
}
```

Passwords are stored separately in `~/.config/dikuclient/.passwords` for security.

## Backward Compatibility

The old "accounts" structure is still supported for backward compatibility. Any legacy accounts will appear in the server selection menu under "Legacy accounts".

## Command Line Options

You can still use the original command line options:

```bash
# Connect directly to a server (bypasses menu)
./dikuclient --host aardmud.org --port 23

# Use a saved account (legacy)
./dikuclient --account MyAccount

# Save a new account during connection
./dikuclient --host aardmud.org --port 23 --save-account
```

## Benefits

1. **Clearer organization**: Servers and characters are separate concepts
2. **Easier navigation**: Two-step process is more intuitive
3. **Better for multiple characters**: Easy to see all characters for a server
4. **Flexible connections**: Can connect anonymously or save characters
5. **Web mode shortcuts**: Direct URLs for quick server access
