// IndexedDB wrapper for client-side file storage (per-session)

const DB_NAME = 'dikuclient-storage';
const DB_VERSION = 2; // Incremented to add passwords store
const STORE_NAME = 'files';
const PASSWORDS_STORE_NAME = 'passwords';

let db = null;
let currentSessionId = null;

// Get session ID from URL
function getSessionId() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('id') || '';
}

// Save last used session ID to cookie
function saveLastSessionId(sessionId) {
    // Cookie expires in 30 days
    const expires = new Date();
    expires.setDate(expires.getDate() + 30);
    document.cookie = `dikuclient_last_session=${sessionId}; expires=${expires.toUTCString()}; path=/; SameSite=Strict`;
}

// Get last used session ID from cookie
function getLastSessionId() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'dikuclient_last_session') {
            return value;
        }
    }
    return null;
}

// Initialize IndexedDB
async function initDB() {
    currentSessionId = getSessionId();
    if (!currentSessionId) {
        throw new Error('No session ID found');
    }
    
    // Save this session ID as the last used
    saveLastSessionId(currentSessionId);
    
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, DB_VERSION);
        
        request.onerror = () => reject(request.error);
        request.onsuccess = () => {
            db = request.result;
            resolve(db);
        };
        
        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            const oldVersion = event.oldVersion;
            
            // Create files store if it doesn't exist
            if (!db.objectStoreNames.contains(STORE_NAME)) {
                // Store format: { sessionId, path, content, timestamp }
                const store = db.createObjectStore(STORE_NAME, { keyPath: ['sessionId', 'path'] });
                store.createIndex('sessionId', 'sessionId', { unique: false });
            }
            
            // Create passwords store (added in version 2)
            if (oldVersion < 2 && !db.objectStoreNames.contains(PASSWORDS_STORE_NAME)) {
                // Store format: { sessionId, account, password }
                // account is "host:port:username"
                const passwordStore = db.createObjectStore(PASSWORDS_STORE_NAME, { keyPath: ['sessionId', 'account'] });
                passwordStore.createIndex('sessionId', 'sessionId', { unique: false });
            }
        };
    });
}

// Save file to IndexedDB (scoped to current session)
async function saveFile(path, content, timestamp) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([STORE_NAME], 'readwrite');
        const store = transaction.objectStore(STORE_NAME);
        const data = { sessionId: currentSessionId, path, content, timestamp };
        
        const request = store.put(data);
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
    });
}

// Load file from IndexedDB (scoped to current session)
async function loadFile(path) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([STORE_NAME], 'readonly');
        const store = transaction.objectStore(STORE_NAME);
        const request = store.get([currentSessionId, path]);
        
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

// List all files in IndexedDB (scoped to current session)
async function listFiles() {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([STORE_NAME], 'readonly');
        const store = transaction.objectStore(STORE_NAME);
        const index = store.index('sessionId');
        const request = index.getAllKeys(currentSessionId);
        
        request.onsuccess = () => {
            // Keys are [sessionId, path] tuples, extract just the paths
            const keys = request.result.map(key => key[1]);
            resolve(keys);
        };
        request.onerror = () => reject(request.error);
    });
}

// Delete file from IndexedDB (scoped to current session)
async function deleteFile(path) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([STORE_NAME], 'readwrite');
        const store = transaction.objectStore(STORE_NAME);
        const request = store.delete([currentSessionId, path]);
        
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
    });
}

// Save password to IndexedDB (scoped to current session)
async function savePassword(account, password) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([PASSWORDS_STORE_NAME], 'readwrite');
        const store = transaction.objectStore(PASSWORDS_STORE_NAME);
        const data = { sessionId: currentSessionId, account, password };
        
        const request = store.put(data);
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
    });
}

// Load password from IndexedDB (scoped to current session)
async function loadPassword(account) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([PASSWORDS_STORE_NAME], 'readonly');
        const store = transaction.objectStore(PASSWORDS_STORE_NAME);
        const request = store.get([currentSessionId, account]);
        
        request.onsuccess = () => {
            const result = request.result;
            resolve(result ? result.password : null);
        };
        request.onerror = () => reject(request.error);
    });
}

// List all passwords for current session
async function listPasswords() {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([PASSWORDS_STORE_NAME], 'readonly');
        const store = transaction.objectStore(PASSWORDS_STORE_NAME);
        const index = store.index('sessionId');
        const request = index.getAll(currentSessionId);
        
        request.onsuccess = () => {
            const passwords = {};
            for (const item of request.result) {
                passwords[item.account] = item.password;
            }
            resolve(passwords);
        };
        request.onerror = () => reject(request.error);
    });
}

// Delete password from IndexedDB (scoped to current session)
async function deletePassword(account) {
    if (!db) await initDB();
    
    return new Promise((resolve, reject) => {
        const transaction = db.transaction([PASSWORDS_STORE_NAME], 'readwrite');
        const store = transaction.objectStore(PASSWORDS_STORE_NAME);
        const request = store.delete([currentSessionId, account]);
        
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
    });
}

// Initialize DB on script load
initDB().catch(err => console.error('Failed to initialize IndexedDB:', err));
