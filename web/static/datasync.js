// Data synchronization between client (IndexedDB) and server (via WebSocket)

let dataWs = null;
let dataConnected = false;

// Connect to data WebSocket for file synchronization
function connectDataSync() {
    const urlParams = new URLSearchParams(window.location.search);
    const sessionId = urlParams.get('id') || '';
    
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/data-ws?id=${sessionId}`;
    
    dataWs = new WebSocket(wsUrl);
    
    dataWs.onopen = async () => {
        dataConnected = true;
        console.log('Data sync WebSocket connected');
        
        // Send passwords to server for auto-login
        await sendPasswordsToServer();
        
        // Send client files to server for merging
        await syncClientToServer();
    };
    
    dataWs.onmessage = async (event) => {
        try {
            const message = JSON.parse(event.data);
            await handleDataMessage(message);
        } catch (e) {
            console.error('Error handling data message:', e);
        }
    };
    
    dataWs.onerror = (error) => {
        console.error('Data WebSocket error:', error);
    };
    
    dataWs.onclose = () => {
        dataConnected = false;
        console.log('Data sync WebSocket closed');
        
        // Try to reconnect after 5 seconds
        setTimeout(() => {
            connectDataSync();
        }, 5000);
    };
}

// Handle incoming data messages
async function handleDataMessage(message) {
    switch (message.type) {
        case 'file_update':
            // Server sent a file update
            await handleFileUpdate(message);
            break;
        case 'file_request':
            // Server requests a file
            await handleFileRequest(message);
            break;
        case 'password_hint':
            // Server sent a password hint (for manually entered passwords)
            await handlePasswordHint(message);
            break;
        case 'merge_complete':
            console.log('Data merge complete:', message.files);
            break;
        default:
            console.log('Unknown data message type:', message.type);
    }
}

// Handle password hint from server
async function handlePasswordHint(message) {
    try {
        const hint = JSON.parse(message.content);
        const { account, password } = hint;
        
        if (account && password) {
            // Save password to client-side IndexedDB
            await savePassword(account, password);
            console.log(`[Client] Saved manually entered password for account: ${account}`);
            
            // Re-send passwords to server so it has the updated list
            await sendPasswordsToServer();
        }
    } catch (e) {
        console.error('[Client] Error handling password hint:', e);
    }
}



// Handle file update from server
async function handleFileUpdate(message) {
    const { path, content, timestamp } = message;
    
    // Check if we have a newer version locally
    const localFile = await loadFile(path);
    
    // Special handling for files that need merging
    if (path === 'map.json' && localFile && content) {
        // Merge map data
        const serverData = JSON.parse(content);
        const clientData = JSON.parse(localFile.content);
        const merged = await mergeMapData(clientData, serverData);
        const mergedContent = JSON.stringify(merged, null, 2);
        const mergedTimestamp = Math.max(timestamp, localFile.timestamp);
        await saveFile(path, mergedContent, mergedTimestamp);
        sendFileToServer(path, mergedContent, mergedTimestamp);
        console.log('Merged map data from client and server');
        return;
    }
    
    if (path === 'history.json' && localFile && content) {
        // Merge history - use more recent
        const serverData = JSON.parse(content);
        const clientData = JSON.parse(localFile.content);
        const merged = await mergeHistory(clientData, serverData);
        const mergedContent = JSON.stringify(merged, null, 2);
        const mergedTimestamp = Math.max(timestamp, localFile.timestamp);
        await saveFile(path, mergedContent, mergedTimestamp);
        sendFileToServer(path, mergedContent, mergedTimestamp);
        console.log('Merged history data from client and server');
        return;
    }
    
    if (path === 'accounts.json') {
        // Accounts file doesn't contain passwords anymore, so just normal merge
        // Passwords are stored separately in IndexedDB passwords table
        // Use server version if both exist (server is authoritative for account list)
        if (content) {
            await saveFile(path, content, timestamp);
            console.log('Updated accounts from server (passwords stored separately)');
            return;
        }
    }
    
    if (!localFile || localFile.timestamp < timestamp) {
        // Server version is newer or we don't have it, save it
        await saveFile(path, content, timestamp);
        console.log(`Updated local file: ${path}`);
    } else if (localFile.timestamp > timestamp) {
        // Local version is newer, send it to server
        sendFileToServer(path, localFile.content, localFile.timestamp);
    }
    // If timestamps are equal, no action needed
}

// Handle file request from server
async function handleFileRequest(message) {
    const { path } = message;
    const localFile = await loadFile(path);
    
    if (localFile) {
        sendFileToServer(path, localFile.content, localFile.timestamp);
    } else {
        // We don't have the file, send a response indicating that
        if (dataWs && dataConnected) {
            dataWs.send(JSON.stringify({
                type: 'file_not_found',
                path: path
            }));
        }
    }
}

// Send file to server
function sendFileToServer(path, content, timestamp) {
    if (dataWs && dataConnected) {
        // No password stripping needed - passwords are never in accounts.json anymore
        dataWs.send(JSON.stringify({
            type: 'file_update',
            path: path,
            content: content,
            timestamp: timestamp
        }));
        console.log(`Sent file to server: ${path}`);
    }
}

// Sync all client files to server
async function syncClientToServer() {
    const files = await listFiles();
    
    for (const path of files) {
        const file = await loadFile(path);
        if (file) {
            sendFileToServer(path, file.content, file.timestamp);
        }
    }
    
    if (files.length > 0) {
        console.log(`Synced ${files.length} files to server`);
    }
}

// Watch for file changes and sync them
window.addEventListener('storage-update', async (event) => {
    const { path, content, timestamp } = event.detail;
    await saveFile(path, content, timestamp);
    sendFileToServer(path, content, timestamp);
});

// Merge history data - keep the more recent one
async function mergeHistory(clientHistory, serverHistory) {
    if (!clientHistory && !serverHistory) {
        return { commands: [] };
    }
    
    if (!clientHistory) return serverHistory;
    if (!serverHistory) return clientHistory;
    
    // Simple approach: use the one with more recent last command
    // Could be enhanced with timestamp-based merging
    const clientLen = clientHistory.commands ? clientHistory.commands.length : 0;
    const serverLen = serverHistory.commands ? serverHistory.commands.length : 0;
    
    // Use the one with more commands (assumes more recent)
    return clientLen >= serverLen ? clientHistory : serverHistory;
}

// Send all passwords to server for auto-login
async function sendPasswordsToServer() {
    try {
        // Load all passwords from IndexedDB
        const passwords = await listPasswords();
        
        if (passwords && passwords.length > 0) {
            // Send to server
            if (dataWs && dataConnected) {
                dataWs.send(JSON.stringify({
                    type: 'passwords_init',
                    passwords: passwords
                }));
                console.log(`[Client] Sent ${passwords.length} passwords to server for auto-login`);
            }
        } else {
            console.log('[Client] No passwords to send to server');
        }
    } catch (e) {
        console.error('[Client] Error sending passwords to server:', e);
    }
}

// Merge map data - combine rooms from both
async function mergeMapData(clientMap, serverMap) {
    if (!clientMap && !serverMap) {
        return { rooms: {}, room_numbering: [] };
    }
    
    if (!clientMap) return serverMap;
    if (!serverMap) return clientMap;
    
    // Merge rooms - keep all rooms from both
    const merged = {
        rooms: { ...serverMap.rooms, ...clientMap.rooms },
        current_room_id: clientMap.current_room_id || serverMap.current_room_id,
        previous_room_id: clientMap.previous_room_id || serverMap.previous_room_id,
        last_direction: clientMap.last_direction || serverMap.last_direction,
        room_numbering: Array.from(new Set([
            ...(serverMap.room_numbering || []),
            ...(clientMap.room_numbering || [])
        ]))
    };
    
    console.log(`Merged map data: ${Object.keys(merged.rooms).length} rooms`);
    return merged;
}

// Handle data merging when both client and server have data
async function handleDataMerge() {
    const filesToMerge = [
        { path: 'history.json', merger: mergeHistory },
        { path: 'map.json', merger: mergeMapData }
    ];
    
    for (const { path, merger } of filesToMerge) {
        const clientFile = await loadFile(path);
        
        // Request server version
        if (dataWs && dataConnected) {
            dataWs.send(JSON.stringify({
                type: 'file_request',
                path: path
            }));
        }
        
        // Note: Actual merging happens in handleFileUpdate when server responds
    }
}

// Initialize data sync on page load
window.addEventListener('load', () => {
    // Wait a bit for the main terminal connection to establish first
    setTimeout(() => {
        connectDataSync();
    }, 1000);
});
