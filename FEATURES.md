# Feature Summary

This document summarizes the major features added to DikuMUD Client.

## Web Mode with WebSocket Support (Phase 3)

### Overview
The DikuMUD Client now supports a web-based interface accessible through a browser. This enables remote access, multi-user support, and platform-independent usage.

### Features Implemented

#### 1. HTTP Server
- Serves static web interface files (HTML, CSS, JavaScript)
- Configurable port with `--web-port` flag
- Clean, modern dark theme UI

#### 2. WebSocket Handler
- Real-time bidirectional communication between browser and server
- Manages multiple concurrent client connections
- Protocol: Simple text-based (CONNECT:host:port, ERROR:message, etc.)
- Automatic connection cleanup on disconnect

#### 3. Web Client Interface
- Connection controls (host, port, connect/disconnect buttons)
- Real-time output display area with auto-scroll
- Command input field with Enter-to-send
- Status indicator (Connected/Disconnected)
- Responsive layout

#### 4. MUD Integration
- WebSocket server creates TCP connections to MUD servers on behalf of clients
- Forwards MUD output to browser in real-time
- Forwards user commands from browser to MUD server
- Handles telnet protocol negotiation

### Usage

Start web server:
```bash
./dikuclient --web --web-port 8080
```

Then open `http://localhost:8080` in a browser, enter MUD host/port, and click Connect.

### Technical Implementation

#### Server Side (Go)
- **internal/web/server.go**: HTTP server with file serving and WebSocket endpoint
- **internal/web/websocket.go**: WebSocket connection handler and session manager
- Uses gorilla/websocket library for WebSocket support

#### Client Side (JavaScript)
- **web/static/index.html**: Main HTML interface
- **web/static/app.js**: WebSocket client logic
- **web/static/styles.css**: Dark theme styling

#### Protocol
- `CONNECT:host:port` - Client requests connection to MUD server
- `CONNECTED` - Server confirms connection established
- `ERROR:message` - Server reports error
- Regular text - MUD output or user commands

### Security Considerations
- CORS enabled (allow all origins) - should be configured for production
- No authentication implemented - suitable for local/trusted networks
- Plain text WebSocket (ws://) - should use WSS for production

### Files Added
- `internal/web/server.go` - HTTP server
- `internal/web/websocket.go` - WebSocket handler
- `web/static/index.html` - Web interface
- `web/static/app.js` - WebSocket client
- `web/static/styles.css` - Styling

### Files Modified
- `cmd/dikuclient/main.go` - Added web mode flags
- `go.mod` - Added gorilla/websocket dependency

---

## Account Management and Auto-Login (Phase 1)

This document summarizes the account management and auto-login features added to DikuMUD Client.

## Overview

The DikuMUD Client now supports saving multiple MUD accounts with automatic login functionality. This makes it easy to connect to your favorite MUDs without having to type your credentials every time.

## Features Implemented

### 1. Account Storage
- Accounts are stored in `~/.config/dikuclient/accounts.json`
- Supports multiple accounts for different MUD servers
- Each account can store:
  - Account name (for identification)
  - Host and port
  - Username (optional)
  - Password (optional)

### 2. Account Management Commands
- `--list-accounts`: List all saved accounts
- `--account <name>`: Connect using a saved account
- `--save-account`: Save account while connecting
- `--delete-account <name>`: Delete a saved account

### 3. Interactive Menu
When running without arguments, users get an interactive menu to:
- Select from saved accounts
- Create a new connection
- Save new accounts with credentials

### 4. Auto-Login
When username and password are saved:
- Automatically detects login prompts (case-insensitive):
  - Username prompts: "name:", "login:", "account:", "character:"
  - Password prompts: "password:", "pass:"
- Sends credentials automatically
- Shows visual feedback: `[Auto-login: sending username 'user']`
- Seamless login experience

## Usage Examples

### Save an account
```bash
./dikuclient --host aardmud.org --port 23 --save-account
```

### Connect with saved account
```bash
./dikuclient --account AardMUD
```

### Interactive selection
```bash
./dikuclient
```

### List accounts
```bash
./dikuclient --list-accounts
```

## Security Considerations

- Credentials stored in plain text JSON file
- File permissions set to 0600 (owner read/write only)
- Config directory created with 0700 permissions
- Users should protect their systems accordingly

## Testing

The implementation includes:
- Unit tests for config package (save, load, update, delete)
- Unit tests for auto-login prompt detection
- All tests passing

## Files Added/Modified

### New Files
- `internal/config/account.go` - Account storage and management
- `internal/config/account_test.go` - Unit tests
- `internal/tui/autologin_test.go` - Auto-login tests
- `ACCOUNTS.md` - Comprehensive user guide
- `FEATURES.md` - This file

### Modified Files
- `cmd/dikuclient/main.go` - Account management CLI
- `internal/tui/app.go` - Auto-login detection and handling
- `README.md` - Updated with account management documentation

## Technical Implementation

### Config Package
- JSON-based configuration storage
- Thread-safe operations
- Graceful handling of missing files
- Support for multiple accounts

### TUI Integration
- Auto-login state machine (idle → username sent → password sent)
- Prompt detection using pattern matching
- Visual feedback for user
- Preserves existing TUI functionality

### Command-Line Interface
- Flag-based account operations
- Interactive prompts for account creation
- User-friendly error messages
- Backward compatible with existing usage

## Future Enhancements (Not Implemented)

Potential future additions could include:
- Password encryption
- Multiple character support per account
- Account import/export
- Default account selection
- Connection history
- Configurable auto-login patterns

## Conclusion

The account management and auto-login functionality significantly improves the user experience by:
- Eliminating repetitive credential entry
- Supporting multiple MUD connections
- Providing a simple, intuitive interface
- Maintaining security with file permissions
