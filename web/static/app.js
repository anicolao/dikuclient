// WebSocket connection
let ws = null;
let connected = false;
let term = null;
let fitAddon = null;
let useFallback = false;
let fallbackContent = '';

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
        } else {
            initTerminal();
        }
        // Auto-connect on page load
        connectToServer();
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
            });
        } else {
            // Text data
            if (useFallback) {
                writeFallback(event.data);
            } else if (term) {
                term.write(event.data);
            }
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
    };
}
