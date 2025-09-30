// WebSocket connection
let ws = null;
let connected = false;
let term = null;
let fitAddon = null;
let useFallback = false;
let fallbackContent = '';

// DOM elements
const hostInput = document.getElementById('host');
const portInput = document.getElementById('port');
const connectBtn = document.getElementById('connectBtn');
const disconnectBtn = document.getElementById('disconnectBtn');
const statusSpan = document.getElementById('status');
const terminalDiv = document.getElementById('terminal');

// Wait for page load to check if xterm loaded
window.addEventListener('load', () => {
    // Give scripts a moment to load
    setTimeout(() => {
        if (!window.xtermLoaded || typeof Terminal === 'undefined') {
            console.log('xterm.js not available, using fallback terminal');
            useFallback = true;
            initFallbackTerminal();
        }
        hostInput.focus();
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

        if (window.fitAddonLoaded && typeof FitAddon !== 'undefined') {
            fitAddon = new FitAddon.FitAddon();
            term.loadAddon(fitAddon);
        }
        
        term.open(terminalDiv);
        if (fitAddon) {
            fitAddon.fit();
        }
        
        // Handle terminal input
        term.onData(data => {
            if (ws && connected) {
                ws.send(data);
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
    } catch (e) {
        console.error('Failed to initialize xterm.js:', e);
        useFallback = true;
        initFallbackTerminal();
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
connectBtn.addEventListener('click', () => {
    const host = hostInput.value.trim();
    const port = parseInt(portInput.value);

    if (!host || !port) {
        const errorMsg = '\r\n\x1b[31mERROR: Please enter both host and port\x1b[0m\r\n';
        if (useFallback) {
            writeFallback(errorMsg);
        } else if (term) {
            term.writeln(errorMsg);
        }
        return;
    }

    // Initialize terminal if not already done
    if (!term && !useFallback) {
        initTerminal();
    } else if (useFallback && !terminalDiv.classList.contains('fallback')) {
        initFallbackTerminal();
    }

    // Connect to local WebSocket server
    const wsUrl = `ws://${window.location.hostname}:${window.location.port || 8080}/ws`;
    const connectMsg = `\r\nConnecting to WebSocket server at ${wsUrl}...\r\n`;
    
    if (useFallback) {
        writeFallback(connectMsg);
    } else if (term) {
        term.write(connectMsg);
    }

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        const msg = 'WebSocket connected. Starting TUI session...\r\n';
        if (useFallback) {
            writeFallback(msg);
        } else if (term) {
            term.write(msg);
        }
        
        // Send connect command with host and port
        const connectData = {
            type: 'connect',
            host: host,
            port: port,
            cols: term ? term.cols : 80,
            rows: term ? term.rows : 24
        };
        ws.send(JSON.stringify(connectData));
    };

    ws.onmessage = (event) => {
        // Write data directly to terminal
        if (useFallback) {
            writeFallback(event.data);
        } else if (term) {
            term.write(event.data);
        }
        
        // Update connection state if needed
        if (!connected) {
            connected = true;
            updateConnectionState(true);
        }
    };

    ws.onerror = (error) => {
        const errorMsg = '\r\n\x1b[31mWebSocket error\x1b[0m\r\n';
        if (useFallback) {
            writeFallback(errorMsg);
        } else if (term) {
            term.write(errorMsg);
        }
        console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
        connected = false;
        updateConnectionState(false);
        const msg = '\r\n\x1b[33mDisconnected from server\x1b[0m\r\n';
        if (useFallback) {
            writeFallback(msg);
        } else if (term) {
            term.write(msg);
        }
        ws = null;
    };
});

// Disconnect button
disconnectBtn.addEventListener('click', () => {
    if (ws) {
        ws.close();
    }
});

// Update connection state UI
function updateConnectionState(isConnected) {
    if (isConnected) {
        statusSpan.textContent = 'Connected';
        statusSpan.className = 'status connected';
        connectBtn.disabled = true;
        disconnectBtn.disabled = false;
        hostInput.disabled = true;
        portInput.disabled = true;
        if (term) {
            term.focus();
        } else if (useFallback) {
            terminalDiv.focus();
        }
    } else {
        statusSpan.textContent = 'Disconnected';
        statusSpan.className = 'status disconnected';
        connectBtn.disabled = false;
        disconnectBtn.disabled = true;
        hostInput.disabled = false;
        portInput.disabled = false;
    }
}
