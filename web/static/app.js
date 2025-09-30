// WebSocket connection
let ws = null;
let connected = false;

// DOM elements
const hostInput = document.getElementById('host');
const portInput = document.getElementById('port');
const connectBtn = document.getElementById('connectBtn');
const disconnectBtn = document.getElementById('disconnectBtn');
const statusSpan = document.getElementById('status');
const outputDiv = document.getElementById('output');
const inputField = document.getElementById('input');
const sendBtn = document.getElementById('sendBtn');

// Connect to WebSocket server
connectBtn.addEventListener('click', () => {
    const host = hostInput.value.trim();
    const port = parseInt(portInput.value);

    if (!host || !port) {
        addOutput('ERROR: Please enter both host and port');
        return;
    }

    // Connect to local WebSocket server
    const wsUrl = `ws://${window.location.hostname}:${window.location.port || 8080}/ws`;
    addOutput(`Connecting to WebSocket server at ${wsUrl}...`);

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        addOutput('WebSocket connected. Connecting to MUD server...');
        // Send connect command to server
        ws.send(`CONNECT:${host}:${port}`);
    };

    ws.onmessage = (event) => {
        const message = event.data;
        
        if (message === 'CONNECTED') {
            connected = true;
            updateConnectionState(true);
            addOutput(`Connected to ${host}:${port}`);
        } else if (message.startsWith('ERROR:')) {
            addOutput(message);
            if (!connected) {
                ws.close();
            }
        } else {
            // Regular MUD output
            addOutput(message);
        }
    };

    ws.onerror = (error) => {
        addOutput('WebSocket error: ' + error);
    };

    ws.onclose = () => {
        connected = false;
        updateConnectionState(false);
        addOutput('Disconnected from server');
        ws = null;
    };
});

// Disconnect button
disconnectBtn.addEventListener('click', () => {
    if (ws) {
        ws.close();
    }
});

// Send command
function sendCommand() {
    const command = inputField.value.trim();
    if (!command || !connected || !ws) {
        return;
    }

    ws.send(command);
    addOutput(`> ${command}`, 'user-input');
    inputField.value = '';
}

sendBtn.addEventListener('click', sendCommand);

inputField.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        sendCommand();
    }
});

// Update connection state UI
function updateConnectionState(isConnected) {
    if (isConnected) {
        statusSpan.textContent = 'Connected';
        statusSpan.className = 'status connected';
        connectBtn.disabled = true;
        disconnectBtn.disabled = false;
        inputField.disabled = false;
        sendBtn.disabled = false;
        hostInput.disabled = true;
        portInput.disabled = true;
    } else {
        statusSpan.textContent = 'Disconnected';
        statusSpan.className = 'status disconnected';
        connectBtn.disabled = false;
        disconnectBtn.disabled = true;
        inputField.disabled = true;
        sendBtn.disabled = true;
        hostInput.disabled = false;
        portInput.disabled = false;
    }
}

// Add output to the display
function addOutput(text, className = '') {
    const line = document.createElement('div');
    line.className = 'output-line' + (className ? ' ' + className : '');
    line.textContent = text;
    outputDiv.appendChild(line);
    
    // Auto-scroll to bottom
    outputDiv.parentElement.scrollTop = outputDiv.parentElement.scrollHeight;
}

// Focus input field on load
window.addEventListener('load', () => {
    hostInput.focus();
});
