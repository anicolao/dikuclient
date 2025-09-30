# Account Management Guide

DikuMUD Client provides comprehensive account management features to help you organize and quickly connect to multiple MUD servers.

## Overview

Accounts are stored in `~/.config/dikuclient/accounts.json` in JSON format. The file is created automatically when you save your first account. Account credentials are stored locally on your machine.

## Creating Accounts

### Interactive Account Creation

Run the client without any arguments to enter the interactive menu:

```bash
./dikuclient
```

You'll be presented with:
1. A list of saved accounts (if any)
2. Option to connect to a new server
3. Option to exit

When connecting to a new server, you'll be prompted to:
- Enter the hostname
- Enter the port (default: 4000)
- Choose whether to save the account
- If saving, provide:
  - Account name (for easy identification)
  - Username (optional, for auto-login)
  - Password (optional, for auto-login)

### Save While Connecting

Connect to a server and save the account in one command:

```bash
./dikuclient --host mud.example.com --port 4000 --save-account
```

This will prompt you for:
- Account name
- Username (optional)
- Password (optional)

### Example Session

```
$ ./dikuclient --host aardmud.org --port 23 --save-account
Enter account name: AardMUD
Enter username (optional): mycharacter
Enter password (optional): mypassword
Account 'AardMUD' saved successfully.
```

## Using Saved Accounts

### Connect Using Account Name

```bash
./dikuclient --account AardMUD
```

The client will automatically:
1. Connect to the saved host and port
2. Wait for login prompts
3. Send username when prompted
4. Send password when prompted

### Select from Interactive Menu

```bash
./dikuclient
```

Then choose the account number from the list.

## Managing Accounts

### List All Accounts

```bash
./dikuclient --list-accounts
```

Example output:
```
Saved accounts:
  1. AardMUD (aardmud.org:23)
     Username: mycharacter
  2. DikuMUD (diku.mud.org:4000)
     Username: player1
  3. TestMUD (localhost:4000)
```

### Delete an Account

```bash
./dikuclient --delete-account AardMUD
```

This removes the account from your saved accounts list.

### Update an Account

To update an account, simply save it again with the same name:

```bash
./dikuclient --host aardmud.org --port 23 --save-account
```

Enter the same account name, and the existing account will be updated with the new details.

## Auto-Login Feature

When you save username and password with an account, the client will automatically detect common login prompts and send your credentials.

### Supported Prompts

The client recognizes these patterns (case-insensitive):
- Username prompts: "name:", "login:", "account:", "character:"
- Password prompts: "password:", "pass:"

### Auto-Login Sequence

1. Client connects to the MUD server
2. Waits for username prompt
3. Automatically sends username
4. Waits for password prompt
5. Automatically sends password
6. You're logged in and ready to play!

### Visual Feedback

The client shows gray messages when auto-login is happening:
```
[Auto-login: sending username 'mycharacter']
[Auto-login: sending password]
```

## Security Considerations

- Account credentials are stored in plain text in `~/.config/dikuclient/accounts.json`
- The file has permissions 0600 (readable/writable only by you)
- Keep your computer secure to protect your credentials
- Consider using different passwords for MUDs than for important accounts

## Example Account File

Here's what the accounts.json file looks like:

```json
{
  "accounts": [
    {
      "name": "AardMUD",
      "host": "aardmud.org",
      "port": 23,
      "username": "mycharacter",
      "password": "mypassword"
    },
    {
      "name": "LocalTest",
      "host": "localhost",
      "port": 4000,
      "username": "",
      "password": ""
    }
  ]
}
```

## Tips and Tricks

1. **Quick Connect**: Create accounts for your favorite MUDs and use `--account` flag for instant connection
2. **Multiple Characters**: Save multiple accounts for the same MUD with different usernames
3. **No Credentials**: You can save accounts without username/password for servers that don't require login
4. **Account Names**: Use descriptive names like "AardMUD-Main" and "AardMUD-Alt" for clarity

## Troubleshooting

### Auto-login not working?

The auto-login feature looks for common prompt patterns. If your MUD uses different prompts:
- The credentials are still saved
- You can manually type your username and password
- Consider reporting the prompt format so we can add support

### Can't find config file?

The config file is located at:
- Linux/Mac: `~/.config/dikuclient/accounts.json`
- Windows: `%USERPROFILE%\.config\dikuclient\accounts.json`

### Want to edit accounts manually?

You can directly edit the `accounts.json` file with any text editor. The format is straightforward JSON.
