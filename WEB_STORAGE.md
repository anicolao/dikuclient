# Web Mode Client-Side Storage

## Overview

When running in web mode, the dikuclient now provides client-side storage using IndexedDB, ensuring that user data persists even when the server session ends. This feature implements a dual-storage architecture where:

- **Server-side storage**: Files are stored in `.websessions/<session-id>/.config/dikuclient/` (temporary, cleaned up after session)
- **Client-side storage**: Files are stored in browser IndexedDB (persistent across sessions)

## Key Features

### 1. Bidirectional File Synchronization

A second WebSocket connection (`/data-ws`) runs transparently in the background to sync files between client and server:

- **Files synced**: `accounts.json`, `history.json`, `map.json`, `xps.json`
- **Automatic merging**: Smart conflict resolution when both client and server have data
- **Real-time updates**: Changes on either side are immediately propagated

### 2. Password Security

**Server-side**: Passwords are automatically stripped from `accounts.json` before saving to server storage
- Never writes passwords to server files
- Protects user credentials on shared/multi-user servers

**Client-side**: Passwords are stored only in the browser's IndexedDB
- Passwords detected during login are automatically saved to client storage
- Password hint mechanism transfers password from server TUI to client storage
- User account data differs: client has passwords, server doesn't

### 3. Data Merging on Login

When connecting with both client and server data present:

**Map data**: Merges room information from both sources
- Combines all rooms from client and server
- Prevents loss of exploration data
- Maintains room numbering consistency

**History**: Uses the more recent history
- Compares command lists
- Keeps the version with more entries
- Could be enhanced with timestamp-based merging

**Accounts**: Merges account list with client passwords
- Takes server account list as base
- Restores passwords from client storage
- Keeps credentials secure and available

## Technical Implementation

### Client-Side Components

**storage.js**: IndexedDB wrapper (per-session storage)
- `initDB()`: Initialize database for current session ID
- `saveFile(path, content, timestamp)`: Save file to IndexedDB (scoped to session)
- `loadFile(path)`: Load file from IndexedDB (scoped to session)
- `listFiles()`: List all stored files (scoped to session)
- `deleteFile(path)`: Remove file from storage (scoped to session)
- `saveLastSessionId(sessionId)`: Save session ID to cookie
- `getLastSessionId()`: Retrieve last used session ID from cookie

**datasync.js**: WebSocket synchronization
- Connects to `/data-ws` endpoint
- Handles file updates from server
- Sends client files to server (strips passwords from accounts.json)
- Implements data merging logic
- Processes password hints

### Server-Side Components

**internal/web/websocket.go**: Data WebSocket handler
- `HandleDataWebSocket()`: Main handler for data sync
- `handleClientFileUpdate()`: Process file updates from client
- `handleClientFileRequest()`: Respond to file requests
- `watchSessionFiles()`: Monitor and sync session files
- `watchPasswordHints()`: Watch for password hint files

**internal/tui/app.go**: Password detection
- `isPasswordPrompt()`: Detect password prompts
- `savePasswordForWebClient()`: Create password hint file
- Password hints sent to client via data sync

## Usage

### For Users

1. **Start web mode**:
   ```bash
   ./dikuclient --web --web-port 8080
   ```

2. **Access in browser**: http://localhost:8080
   - First visit: A new session ID (GUID) is generated
   - Subsequent visits: Last used session ID is remembered via cookie
   - Each session ID has its own isolated storage

3. **Automatic storage**:
   - All game data automatically saved to browser per session ID
   - Passwords stored securely in client-side IndexedDB
   - Data persists across browser sessions
   - Each session ID maintains separate configuration

4. **Multi-window support**:
   - Open multiple browser windows with different session IDs
   - Example: `http://localhost:8080/?id=session-1` and `http://localhost:8080/?id=session-2`
   - Each window has its own independent configuration and data
   - Useful for observing another player while playing yourself

5. **Moving between devices**:
   - Copy the session ID from URL: `http://localhost:8080/?id=<your-session-id>`
   - Open the same URL on a new device
   - You'll get that session's configuration, not the new device's previous config
   - Server-side data persists and syncs to the new client

6. **Data synchronization**:
   - Works transparently in the background
   - No user action required
   - Console logs show sync activity (check browser DevTools)

### For Developers

**Adding new synced files**:

1. Add filename to `filesToWatch` in `watchSessionFiles()` (websocket.go)
2. Add merge logic in `handleFileUpdate()` if needed (datasync.js)

**Password detection**:

Password prompts are detected by checking if output contains "pass" (case-insensitive). When detected:
1. Input not added to command history
2. Password saved to hint file
3. Hint file picked up by watcher
4. Sent to client via data WebSocket
5. Client saves to IndexedDB accounts.json

## Security Considerations

- **Client passwords**: Stored in IndexedDB (browser's encrypted storage)
- **Server passwords**: Never stored in web mode
- **CLI mode**: Passwords stored in files as before (unchanged behavior)
- **WebSocket security**: Uses same origin policy
- **Session isolation**: Each session has its own config directory

## File Locations

**Server (persistent per session ID)**:
```
.websessions/<session-id>/.config/dikuclient/
  ├── accounts.json (no passwords)
  ├── history.json
  ├── map.json
  ├── xps.json
  └── password_hint.json (temporary, deleted after sending)
```

**Client (persistent per session ID in IndexedDB)**:
```
IndexedDB: dikuclient-storage
  Store: files (compound key: [sessionId, path])
    Session <id-1>:
      ├── accounts.json (with passwords)
      ├── history.json
      ├── map.json
      └── xps.json
    Session <id-2>:
      ├── accounts.json (with passwords)
      ├── history.json
      ├── map.json
      └── xps.json

Cookie: dikuclient_last_session=<session-id> (30 day expiry)
```

## Troubleshooting

### Check if storage is working

Open browser DevTools console and run:
```javascript
// Check if IndexedDB is initialized
indexedDB.databases().then(dbs => console.log(dbs));

// Check current session ID
console.log('Current session:', getSessionId());

// Check stored files for current session
listFiles().then(files => console.log(files));

// Load a specific file for current session
loadFile('accounts.json').then(file => console.log(file));

// Check last used session ID
console.log('Last session:', getLastSessionId());
```

### Data sync not working

1. Check WebSocket connections in Network tab
2. Look for `/data-ws` connection
3. Check console for sync messages
4. Verify session ID in URL matches server logs

### Passwords not saving

1. Check if password prompt was detected (contains "pass")
2. Check console for "Saved password hint" messages
3. Verify account exists before entering password
4. Check IndexedDB for accounts.json with password field

## Browser Compatibility

- **IndexedDB**: Supported in all modern browsers
- **WebSocket**: Required for client operation
- **Storage quota**: Browsers typically allow 50MB+ for IndexedDB

## Future Enhancements

1. **File watching**: Implement fsnotify for real-time server file changes
2. **Conflict resolution UI**: Let users choose which version to keep
3. **Encryption**: Add optional client-side encryption for passwords
4. **Sync status indicator**: Show sync activity in the UI
5. **Export/Import**: Allow users to backup/restore client data
6. **Cross-device sync**: Optional cloud sync for multiple devices
