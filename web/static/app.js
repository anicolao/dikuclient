// WebSocket connection
let ws = null;
let connected = false;
let term = null;
let fitAddon = null;
let useFallback = false;
let fallbackContent = '';

// Auto-login state
let lastOutput = '';  // Track terminal output for password prompt detection
let passwordSentForCurrentConnection = false;  // Track if password was sent for this connection
let currentUsername = '';  // Current username for auto-login
let autoLoginState = 0;  // 0=idle, 1=username sent, 2=password sent

// DOM elements
const terminalDiv = document.getElementById('terminal');

// Wait for page load to initialize terminal
window.addEventListener('load', () => {
    // Give scripts a moment to load
    setTimeout(() => {
        if (typeof Terminal === 'undefined') {
            console.error('xterm.js failed to load');
            useFallback = true;
            initFallbackTerminal();
            // Connect after fallback terminal is ready
            connectToServer();
        } else {
            // Initialize terminal first, then connect
            // The connectToServer() call is now inside initTerminal()
            initTerminal();
        }
    }, 100);
});

// Initialize xterm.js terminal
function initTerminal() {
    if (useFallback) {
        return;
    }

    try {
        term = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: 'Courier New, monospace',
            theme: {
                background: '#000000',
                foreground: '#d4d4d4',
                cursor: '#d4d4d4',
                cursorAccent: '#000000',
                selection: 'rgba(255, 255, 255, 0.3)',
                black: '#000000',
                red: '#cd3131',
                green: '#0dbc79',
                yellow: '#e5e510',
                blue: '#2472c8',
                magenta: '#bc3fbc',
                cyan: '#11a8cd',
                white: '#e5e5e5',
                brightBlack: '#666666',
                brightRed: '#f14c4c',
                brightGreen: '#23d18b',
                brightYellow: '#f5f543',
                brightBlue: '#3b8eea',
                brightMagenta: '#d670d6',
                brightCyan: '#29b8db',
                brightWhite: '#e5e5e5'
            }
        });

        if (typeof FitAddon !== 'undefined') {
            fitAddon = new FitAddon.FitAddon();
            term.loadAddon(fitAddon);
        }
        
        term.open(terminalDiv);
        
        // Handle terminal input
        term.onData(data => {
            if (ws && connected) {
                ws.send(data);
                
                // If this is Enter after a password prompt, save the password
                handleUserInput(data);
            }
        });
        
        // Handle window resize
        if (fitAddon) {
            window.addEventListener('resize', () => {
                fitAddon.fit();
                if (ws && connected) {
                    const dims = {
                        type: 'resize',
                        cols: term.cols,
                        rows: term.rows
                    };
                    ws.send(JSON.stringify(dims));
                }
            });
        }
        
        // Wait for terminal to be rendered in the DOM, then fit and connect
        // Use requestAnimationFrame to ensure the DOM has been updated
        requestAnimationFrame(() => {
            if (fitAddon) {
                fitAddon.fit();
                console.log('Terminal fitted to viewport:', term.cols, 'x', term.rows);
            }
            // Now that terminal is properly sized, establish WebSocket connection
            connectToServer();
        });
    } catch (e) {
        console.error('Failed to initialize xterm.js:', e);
        useFallback = true;
        initFallbackTerminal();
        // Connect after fallback terminal is ready
        connectToServer();
    }
}

// Initialize fallback terminal (simple text display)
function initFallbackTerminal() {
    terminalDiv.className = 'fallback';
    terminalDiv.setAttribute('tabindex', '0');
    
    // Handle keyboard input in fallback mode
    terminalDiv.addEventListener('keypress', (e) => {
        if (ws && connected) {
            ws.send(e.key);
            e.preventDefault();
        }
    });
    
    terminalDiv.addEventListener('keydown', (e) => {
        if (ws && connected) {
            // Handle special keys
            const specialKeys = {
                'Enter': '\r',
                'Backspace': '\b',
                'Tab': '\t',
                'Escape': '\x1b',
                'ArrowUp': '\x1b[A',
                'ArrowDown': '\x1b[B',
                'ArrowRight': '\x1b[C',
                'ArrowLeft': '\x1b[D'
            };
            
            if (specialKeys[e.key]) {
                ws.send(specialKeys[e.key]);
                e.preventDefault();
            }
        }
    });
}

// Write to terminal (fallback mode)
function writeFallback(data) {
    // Strip ANSI codes for simplicity in fallback mode
    const strippedData = data.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '');
    fallbackContent += strippedData;
    terminalDiv.textContent = fallbackContent;
    terminalDiv.scrollTop = terminalDiv.scrollHeight;
}

// Connect to WebSocket server
function connectToServer() {
    // Get session ID from URL query parameter
    const urlParams = new URLSearchParams(window.location.search);
    const sessionId = urlParams.get('id') || '';
    
    // Use wss:// for HTTPS pages, ws:// for HTTP pages
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    
    // Use the current host (includes port if non-standard) for reverse proxy compatibility
    const wsUrl = `${wsProtocol}//${window.location.host}/ws?id=${sessionId}`;
    
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        connected = true;
        
        // Send initial terminal size before any other messages
        if (term && fitAddon) {
            fitAddon.fit(); // Ensure terminal is properly sized
            const dims = {
                type: 'init',
                cols: term.cols,
                rows: term.rows
            };
            ws.send(JSON.stringify(dims));
            console.log(`Sent initial terminal size: ${term.cols}x${term.rows}`);
        } else if (useFallback) {
            // For fallback, estimate a reasonable size
            const dims = {
                type: 'init',
                cols: 80,
                rows: 24
            };
            ws.send(JSON.stringify(dims));
        }
        
        // Focus the terminal
        if (term) {
            term.focus();
        } else if (useFallback) {
            terminalDiv.focus();
        }
    };

    ws.onmessage = (event) => {
        // Handle both text and binary data
        if (event.data instanceof Blob) {
            // Binary data - convert to text
            event.data.arrayBuffer().then(buffer => {
                const bytes = new Uint8Array(buffer);
                const decoder = new TextDecoder('utf-8', { fatal: false });
                const text = decoder.decode(bytes);
                
                if (useFallback) {
                    writeFallback(text);
                } else if (term) {
                    term.write(text);
                }
                
                // Track output for auto-login
                handleTerminalOutput(text);
            });
        } else {
            // Text data
            if (useFallback) {
                writeFallback(event.data);
            } else if (term) {
                term.write(event.data);
            }
            
            // Track output for auto-login
            handleTerminalOutput(event.data);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
        connected = false;
        const msg = '\r\n\x1b[33mDisconnected from server\x1b[0m\r\n';
        if (useFallback) {
            writeFallback(msg);
        } else if (term) {
            term.write(msg);
        }
        ws = null;
        
        // Try to reconnect after 3 seconds
        setTimeout(() => {
            connectToServer();
        }, 3000);
        
        // Reset auto-login state on disconnect
        passwordSentForCurrentConnection = false;
        autoLoginState = 0;
        currentUsername = '';
    };
}

// Handle terminal output for auto-login detection
function handleTerminalOutput(text) {
    // Accumulate output
    lastOutput += text;
    
    // Keep only last 500 characters to avoid memory issues
    if (lastOutput.length > 500) {
        lastOutput = lastOutput.substring(lastOutput.length - 500);
    }
    
    // Try auto-login if we haven't completed it yet
    if (autoLoginState < 2) {
        tryAutoLogin();
    }
}

// Attempt auto-login by detecting prompts and sending credentials
async function tryAutoLogin() {
    // Get last line of output
    const lines = lastOutput.split(/\r?\n/);
    const lastLine = lines[lines.length - 1].toLowerCase();
    
    // Step 1: Check for username prompt
    if (autoLoginState === 0 && 
        (lastLine.includes('name') || lastLine.includes('login') || 
         lastLine.includes('account') || lastLine.includes('character'))) {
        
        // Try to load username from accounts
        const accounts = await loadAccountsForAutoLogin();
        if (accounts && accounts.length > 0) {
            const account = accounts[0];  // Use first account
            if (account.username) {
                currentUsername = account.username;
                ws.send(currentUsername + '\r');
                autoLoginState = 1;
                console.log(`Auto-login: sent username '${currentUsername}'`);
            }
        }
    }
    
    // Step 2: Check for password prompt
    if (autoLoginState === 1 && !passwordSentForCurrentConnection &&
        (lastLine.includes('password') || lastLine.includes('pass'))) {
        
        // Try to load password from IndexedDB
        const accounts = await loadAccountsForAutoLogin();
        if (accounts && accounts.length > 0) {
            const account = accounts[0];  // Use first account
            const accountKey = `${account.host}:${account.port}:${account.username}`;
            const password = await loadPassword(accountKey);
            
            if (password) {
                // Generate random number of bullets for display (password length ± 3)
                const bulletCount = password.length + Math.floor(Math.random() * 7) - 3;
                const bullets = '⚫'.repeat(Math.max(1, bulletCount));
                
                // Write bullets to terminal for visual feedback
                if (term) {
                    term.write(bullets + '\r\n');
                } else if (useFallback) {
                    writeFallback(bullets + '\r\n');
                }
                
                // Send password
                ws.send(password + '\r');
                passwordSentForCurrentConnection = true;
                autoLoginState = 2;
                console.log('Auto-login: sent password');
            }
        }
    }
}

// Load accounts for auto-login
async function loadAccountsForAutoLogin() {
    try {
        const file = await loadFile('accounts.json');
        if (file && file.content) {
            const data = JSON.parse(file.content);
            return data.accounts || [];
        }
    } catch (e) {
        console.error('Error loading accounts for auto-login:', e);
    }
    return null;
}

// Track user input for manual password entry
let userInputBuffer = '';
let lastPromptWasPassword = false;

// Handle user input to detect manually entered passwords
function handleUserInput(data) {
    // Track if last prompt was a password prompt
    const lines = lastOutput.split(/\r?\n/);
    const lastLine = lines[lines.length - 1].toLowerCase();
    const isPasswordPrompt = (lastLine.includes('password') || lastLine.includes('pass'));
    
    // If Enter key is pressed after password prompt
    if (data === '\r' && lastPromptWasPassword && userInputBuffer.length > 0) {
        // Save the password
        saveManualPassword(userInputBuffer);
        userInputBuffer = '';
        lastPromptWasPassword = false;
    } else if (data === '\r') {
        // Reset buffer on Enter
        userInputBuffer = '';
        lastPromptWasPassword = isPasswordPrompt;
    } else if (isPasswordPrompt && data.length === 1) {
        // Accumulate input during password entry
        userInputBuffer += data;
        lastPromptWasPassword = true;
    } else if (!isPasswordPrompt) {
        // Not a password prompt, reset
        userInputBuffer = '';
        lastPromptWasPassword = false;
    }
}

// Save manually entered password to IndexedDB
async function saveManualPassword(password) {
    try {
        const accounts = await loadAccountsForAutoLogin();
        if (accounts && accounts.length > 0) {
            const account = accounts[0];  // Save for first account
            const accountKey = `${account.host}:${account.port}:${account.username}`;
            
            // Check if password already exists
            const existing = await loadPassword(accountKey);
            if (!existing || existing === '') {
                await savePassword(accountKey, password);
                console.log(`Saved manually entered password for account: ${accountKey}`);
            }
        }
    } catch (e) {
        console.error('Error saving manual password:', e);
    }
}
