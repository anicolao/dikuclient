# Feature Summary: Account Management and Auto-Login

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
