package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now - should be configured in production
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	sessions       map[*websocket.Conn]*ClientConnection
	sharedSessions map[string]*SharedSession // Maps session ID to shared session
	mu             sync.RWMutex
	enableLogs     bool   // Whether to enable logging for spawned TUI instances
	currentSessID  string // Current session ID to use for new connections
	sessionIDMu    sync.RWMutex
	port           int // Server port for generating share URLs
	passwordStore  map[string]map[string]string // sessionID -> (account -> password), kept in memory only
	passwordMu     sync.RWMutex
}

// SharedSession represents a shared PTY session that multiple clients can connect to
type SharedSession struct {
	sessionID  string
	ptmx       *os.File
	cmd        *exec.Cmd
	clients    map[*websocket.Conn]bool
	mu         sync.RWMutex
	closed     bool
	utf8Buffer []byte // Buffer for incomplete UTF-8 sequences at PTY read boundaries
}

// ClientConnection represents a single WebSocket client connection to a shared session
type ClientConnection struct {
	ws            *websocket.Conn
	sharedSession *SharedSession
	sessionID     string
}

// Session represents a WebSocket session with a PTY running the TUI (kept for compatibility)
type Session struct {
	ws         *websocket.Conn
	ptmx       *os.File
	cmd        *exec.Cmd
	mu         sync.Mutex
	closed     bool
	utf8Buffer []byte // Buffer for incomplete UTF-8 sequences at PTY read boundaries
	sessionID  string // Session ID for this connection
}

// ConnectMessage represents the connection request from client
type ConnectMessage struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port int    `json:"port"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// ResizeMessage represents a terminal resize request
type ResizeMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return NewWebSocketHandlerWithLogging(false)
}

// NewWebSocketHandlerWithLogging creates a new WebSocket handler with logging option
func NewWebSocketHandlerWithLogging(enableLogs bool) *WebSocketHandler {
	return &WebSocketHandler{
		sessions:       make(map[*websocket.Conn]*ClientConnection),
		sharedSessions: make(map[string]*SharedSession),
		enableLogs:     enableLogs,
		passwordStore:  make(map[string]map[string]string),
	}
}

// SetSessionID sets the current session ID for the next connection
func (h *WebSocketHandler) SetSessionID(sessionID string) {
	h.sessionIDMu.Lock()
	defer h.sessionIDMu.Unlock()
	h.currentSessID = sessionID
}

// GetSessionID gets the current session ID
func (h *WebSocketHandler) GetSessionID() string {
	h.sessionIDMu.RLock()
	defer h.sessionIDMu.RUnlock()
	return h.currentSessID
}

// SetPort sets the server port for generating share URLs
func (h *WebSocketHandler) SetPort(port int) {
	h.port = port
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer ws.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		log.Printf("Warning: No session ID in WebSocket URL, using default")
		sessionID = "default"
	}

	// Wait for initial message with terminal size
	messageType, message, err := ws.ReadMessage()
	if err != nil {
		log.Printf("Error reading initial message: %v", err)
		return
	}

	var initialSize *ResizeMessage
	if messageType == websocket.TextMessage {
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok && msgType == "init" {
				var initMsg ResizeMessage
				if err := json.Unmarshal(message, &initMsg); err == nil {
					initialSize = &initMsg
					log.Printf("Initial terminal size: %dx%d", initMsg.Cols, initMsg.Rows)
				}
			}
		}
	}

	// Get or create shared session
	h.mu.Lock()
	sharedSession, exists := h.sharedSessions[sessionID]
	if !exists {
		// Create new shared session
		sharedSession = &SharedSession{
			sessionID: sessionID,
			clients:   make(map[*websocket.Conn]bool),
		}
		h.sharedSessions[sessionID] = sharedSession
		log.Printf("Created new shared session: %s", sessionID)
	} else {
		log.Printf("Joining existing shared session: %s", sessionID)
	}
	h.mu.Unlock()

	// Add this client to the shared session
	sharedSession.mu.Lock()
	sharedSession.clients[ws] = true
	needsStart := sharedSession.ptmx == nil
	sharedSession.mu.Unlock()

	// Create client connection
	client := &ClientConnection{
		ws:            ws,
		sharedSession: sharedSession,
		sessionID:     sessionID,
	}

	h.mu.Lock()
	h.sessions[ws] = client
	h.mu.Unlock()

	defer func() {
		// Remove client from shared session
		sharedSession.mu.Lock()
		delete(sharedSession.clients, ws)
		clientCount := len(sharedSession.clients)
		sharedSession.mu.Unlock()

		// Remove from handler's session map
		h.mu.Lock()
		delete(h.sessions, ws)
		h.mu.Unlock()

		// If no more clients, cleanup the shared session
		if clientCount == 0 {
			log.Printf("Last client disconnected from session %s, cleaning up", sessionID)
			sharedSession.cleanup()
			h.mu.Lock()
			delete(h.sharedSessions, sessionID)
			h.mu.Unlock()
		} else {
			log.Printf("Client disconnected from session %s, %d clients remaining", sessionID, clientCount)
		}
	}()

	// Start TUI if this is the first client for this session
	if needsStart {
		h.startSharedTUI(sharedSession, initialSize)
		// Start forwarding PTY output to all clients
		go h.forwardSharedPTYOutput(sharedSession)
	}

	// Handle incoming messages from this client
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			// Try to parse as JSON for control messages
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err == nil {
				if msgType, ok := msg["type"].(string); ok {
					switch msgType {
					case "resize":
						h.handleSharedResize(sharedSession, message)
						continue
					}
				}
			}

			// Otherwise, it's terminal input - send to PTY
			sharedSession.mu.RLock()
			ptmx := sharedSession.ptmx
			closed := sharedSession.closed
			sharedSession.mu.RUnlock()

			if ptmx != nil && !closed {
				_, err := ptmx.Write(message)
				if err != nil {
					log.Printf("Error writing to PTY: %v", err)
					break
				}
			}
		}
	}
}

// startSharedTUI starts the TUI for a shared session
func (h *WebSocketHandler) startSharedTUI(sharedSession *SharedSession, initialSize *ResizeMessage) {
	// Create session directory
	sessionDir := filepath.Join(".websessions", sharedSession.sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		h.sendErrorToSession(sharedSession, fmt.Sprintf("Failed to create session directory: %v", err))
		return
	}

	// Get the path to the dikuclient binary
	dikuclientPath, err := exec.LookPath("dikuclient")
	if err != nil {
		// Try relative path from current working directory
		cwd, _ := os.Getwd()
		dikuclientPath = filepath.Join(cwd, "dikuclient")
		// Check if it exists
		if _, err := os.Stat(dikuclientPath); err != nil {
			dikuclientPath = "./dikuclient"
		}
	}

	// Build command arguments - no host/port, just optional logging
	args := []string{}
	if h.enableLogs {
		args = append(args, "--log-all")
	}

	// Start the TUI client
	cmd := exec.Command(dikuclientPath, args...)

	// Get absolute path for session directory
	absSessionDir, err := filepath.Abs(sessionDir)
	if err != nil {
		h.sendErrorToSession(sharedSession, fmt.Sprintf("Failed to get absolute path: %v", err))
		return
	}

	// Set working directory to session directory
	cmd.Dir = absSessionDir

	// Set environment variables for session sharing and config
	configDir := filepath.Join(absSessionDir, ".config", "dikuclient")
	serverURL := fmt.Sprintf("http://localhost:%d", h.port)
	envVars := []string{
		fmt.Sprintf("DIKUCLIENT_CONFIG_DIR=%s", configDir),
		fmt.Sprintf("DIKUCLIENT_WEB_SESSION_ID=%s", sharedSession.sessionID),
		fmt.Sprintf("DIKUCLIENT_WEB_SERVER_URL=%s", serverURL),
		"TERM=xterm-kitty",        // Ensure consistent color support regardless of server terminal
		"COLORTERM=truecolor",      // Enable 24-bit true color support
	}
	
	// Add passwords from memory as environment variable
	if passwordsEnv := h.getPasswordsEnv(sharedSession.sessionID); passwordsEnv != "" {
		envVars = append(envVars, fmt.Sprintf("DIKUCLIENT_WEB_PASSWORDS=%s", passwordsEnv))
	}
	
	cmd.Env = append(os.Environ(), envVars...)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		h.sendErrorToSession(sharedSession, fmt.Sprintf("Failed to start TUI: %v", err))
		return
	}

	// Set initial PTY size from client if available, otherwise use defaults
	rows := uint16(24)
	cols := uint16(80)
	if initialSize != nil && initialSize.Rows > 0 && initialSize.Cols > 0 {
		rows = uint16(initialSize.Rows)
		cols = uint16(initialSize.Cols)
	}
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})

	sharedSession.mu.Lock()
	sharedSession.ptmx = ptmx
	sharedSession.cmd = cmd
	sharedSession.mu.Unlock()

	log.Printf("Started shared TUI session for %s with size %dx%d", sharedSession.sessionID, cols, rows)
}

// autoStartTUIWithSize automatically starts the TUI client in a PTY with optional initial size (deprecated)
func (h *WebSocketHandler) autoStartTUIWithSize(session *Session, initialSize *ResizeMessage) {
	// Create session directory
	sessionDir := filepath.Join(".websessions", session.sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		session.sendError(fmt.Sprintf("Failed to create session directory: %v", err))
		return
	}

	// Get the path to the dikuclient binary
	dikuclientPath, err := exec.LookPath("dikuclient")
	if err != nil {
		// Try relative path from current working directory
		cwd, _ := os.Getwd()
		dikuclientPath = filepath.Join(cwd, "dikuclient")
		// Check if it exists
		if _, err := os.Stat(dikuclientPath); err != nil {
			dikuclientPath = "./dikuclient"
		}
	}

	// Build command arguments - no host/port, just optional logging
	args := []string{}
	if h.enableLogs {
		args = append(args, "--log-all")
	}

	// Start the TUI client
	cmd := exec.Command(dikuclientPath, args...)

	// Get absolute path for session directory
	absSessionDir, err := filepath.Abs(sessionDir)
	if err != nil {
		session.sendError(fmt.Sprintf("Failed to get absolute path: %v", err))
		return
	}

	// Set working directory to session directory
	cmd.Dir = absSessionDir

	// Set environment variables for session sharing and config
	configDir := filepath.Join(absSessionDir, ".config", "dikuclient")
	serverURL := fmt.Sprintf("http://localhost:%d", h.port)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DIKUCLIENT_CONFIG_DIR=%s", configDir),
		fmt.Sprintf("DIKUCLIENT_WEB_SESSION_ID=%s", session.sessionID),
		fmt.Sprintf("DIKUCLIENT_WEB_SERVER_URL=%s", serverURL),
		"TERM=xterm-kitty",        // Ensure consistent color support regardless of server terminal
		"COLORTERM=truecolor",      // Enable 24-bit true color support
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		session.sendError(fmt.Sprintf("Failed to start TUI: %v", err))
		return
	}

	// Set initial PTY size from client if available, otherwise use defaults
	rows := uint16(24)
	cols := uint16(80)
	if initialSize != nil && initialSize.Rows > 0 && initialSize.Cols > 0 {
		rows = uint16(initialSize.Rows)
		cols = uint16(initialSize.Cols)
	}
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})

	session.mu.Lock()
	session.ptmx = ptmx
	session.cmd = cmd
	session.mu.Unlock()

	// Start forwarding PTY output to WebSocket
	go h.forwardPTYOutput(session)

	log.Printf("Started TUI session for %s with size %dx%d (no host/port - interactive mode)", session.sessionID, cols, rows)
}

// autoStartTUI automatically starts the TUI client in a PTY with no arguments (deprecated)
// Kept for compatibility, calls autoStartTUIWithSize with nil initial size
func (h *WebSocketHandler) autoStartTUI(session *Session) {
	h.autoStartTUIWithSize(session, nil)
}

// handleConnect starts the TUI client in a PTY (deprecated - kept for compatibility)
func (h *WebSocketHandler) handleConnect(session *Session, message []byte) {
	var connectMsg ConnectMessage
	if err := json.Unmarshal(message, &connectMsg); err != nil {
		session.sendError(fmt.Sprintf("Invalid connect message: %v", err))
		return
	}

	// Create session directory
	sessionDir := filepath.Join(".websessions", session.sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		session.sendError(fmt.Sprintf("Failed to create session directory: %v", err))
		return
	}

	// Get the path to the dikuclient binary
	// In production, this should be configurable
	dikuclientPath, err := exec.LookPath("dikuclient")
	if err != nil {
		// Try relative path
		dikuclientPath = "./dikuclient"
	}

	// Build command arguments
	args := []string{"--host", connectMsg.Host, "--port", fmt.Sprintf("%d", connectMsg.Port)}
	if h.enableLogs {
		args = append(args, "--log-all")
	}

	// Start the TUI client
	cmd := exec.Command(dikuclientPath, args...)

	// Get absolute path for session directory
	absSessionDir, err := filepath.Abs(sessionDir)
	if err != nil {
		session.sendError(fmt.Sprintf("Failed to get absolute path: %v", err))
		return
	}

	// Set working directory to session directory
	cmd.Dir = absSessionDir

	// Set environment variables for session sharing and config
	configDir := filepath.Join(absSessionDir, ".config", "dikuclient")
	serverURL := fmt.Sprintf("http://localhost:%d", h.port)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DIKUCLIENT_CONFIG_DIR=%s", configDir),
		fmt.Sprintf("DIKUCLIENT_WEB_SESSION_ID=%s", session.sessionID),
		fmt.Sprintf("DIKUCLIENT_WEB_SERVER_URL=%s", serverURL),
		"TERM=xterm-kitty",        // Ensure consistent color support regardless of server terminal
		"COLORTERM=truecolor",      // Enable 24-bit true color support
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		session.sendError(fmt.Sprintf("Failed to start TUI: %v", err))
		return
	}

	// Set PTY size
	if connectMsg.Cols > 0 && connectMsg.Rows > 0 {
		pty.Setsize(ptmx, &pty.Winsize{
			Rows: uint16(connectMsg.Rows),
			Cols: uint16(connectMsg.Cols),
		})
	}

	session.mu.Lock()
	session.ptmx = ptmx
	session.cmd = cmd
	session.mu.Unlock()

	// Start forwarding PTY output to WebSocket
	go h.forwardPTYOutput(session)

	log.Printf("Started TUI session for %s:%d", connectMsg.Host, connectMsg.Port)
}

// handleSharedResize handles terminal resize requests for shared sessions
func (h *WebSocketHandler) handleSharedResize(sharedSession *SharedSession, message []byte) {
	var resizeMsg ResizeMessage
	if err := json.Unmarshal(message, &resizeMsg); err != nil {
		log.Printf("Invalid resize message: %v", err)
		return
	}

	sharedSession.mu.Lock()
	defer sharedSession.mu.Unlock()

	if sharedSession.ptmx != nil && !sharedSession.closed {
		pty.Setsize(sharedSession.ptmx, &pty.Winsize{
			Rows: uint16(resizeMsg.Rows),
			Cols: uint16(resizeMsg.Cols),
		})
	}
}

// handleResize handles terminal resize requests (deprecated)
func (h *WebSocketHandler) handleResize(session *Session, message []byte) {
	var resizeMsg ResizeMessage
	if err := json.Unmarshal(message, &resizeMsg); err != nil {
		log.Printf("Invalid resize message: %v", err)
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.ptmx != nil && !session.closed {
		pty.Setsize(session.ptmx, &pty.Winsize{
			Rows: uint16(resizeMsg.Rows),
			Cols: uint16(resizeMsg.Cols),
		})
	}
}

// incompleteUTF8Tail returns the number of trailing bytes that form an incomplete UTF-8 sequence
func incompleteUTF8Tail(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	// Check last 1-4 bytes for incomplete UTF-8
	maxCheck := 4
	if len(data) < maxCheck {
		maxCheck = len(data)
	}

	// Start from the end and look for the beginning of a UTF-8 sequence
	for i := 1; i <= maxCheck; i++ {
		pos := len(data) - i
		b := data[pos]

		// Check if this is a start byte
		if b < 0x80 {
			// ASCII character, complete
			return 0
		} else if b >= 0xC0 && b < 0xE0 {
			// Start of 2-byte sequence
			expected := 2
			if i < expected {
				return i // Incomplete
			}
			if i == expected && utf8.Valid(data[pos:]) {
				return 0
			}
			return i
		} else if b >= 0xE0 && b < 0xF0 {
			// Start of 3-byte sequence
			expected := 3
			if i < expected {
				return i // Incomplete
			}
			if i == expected && utf8.Valid(data[pos:]) {
				return 0
			}
			return i
		} else if b >= 0xF0 && b < 0xF8 {
			// Start of 4-byte sequence
			expected := 4
			if i < expected {
				return i // Incomplete
			}
			if i == expected && utf8.Valid(data[pos:]) {
				return 0
			}
			return i
		}
		// Continue looking backwards (this byte is a continuation byte 0x80-0xBF)
	}

	return 0
}

// forwardSharedPTYOutput forwards output from PTY to all connected WebSocket clients
func (h *WebSocketHandler) forwardSharedPTYOutput(sharedSession *SharedSession) {
	sharedSession.mu.RLock()
	ptmx := sharedSession.ptmx
	sharedSession.mu.RUnlock()

	if ptmx == nil {
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from PTY: %v", err)
			}
			break
		}

		if n > 0 {
			sharedSession.mu.Lock()
			if !sharedSession.closed {
				// Prepend any buffered UTF-8 bytes from previous read
				data := buf[:n]
				if len(sharedSession.utf8Buffer) > 0 {
					data = append(sharedSession.utf8Buffer, data...)
					sharedSession.utf8Buffer = nil
				}

				// Check if data ends with incomplete UTF-8 sequence
				incompleteLen := incompleteUTF8Tail(data)
				if incompleteLen > 0 {
					// Buffer the incomplete UTF-8 bytes for next read
					splitPoint := len(data) - incompleteLen
					sharedSession.utf8Buffer = append(sharedSession.utf8Buffer, data[splitPoint:]...)
					data = data[:splitPoint]
				}

				// Broadcast to all connected clients
				if len(data) > 0 {
					for ws := range sharedSession.clients {
						err := ws.WriteMessage(websocket.BinaryMessage, data)
						if err != nil {
							log.Printf("Error writing to WebSocket client: %v", err)
							// Note: Don't break here, try to send to other clients
						}
					}
				}
			}
			sharedSession.mu.Unlock()
		}
	}

	// PTY closed, clean up
	sharedSession.cleanup()
}

// forwardPTYOutput forwards output from PTY to WebSocket (deprecated)
func (h *WebSocketHandler) forwardPTYOutput(session *Session) {
	session.mu.Lock()
	ptmx := session.ptmx
	session.mu.Unlock()

	if ptmx == nil {
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from PTY: %v", err)
			}
			break
		}

		if n > 0 {
			session.mu.Lock()
			if !session.closed {
				// Prepend any buffered UTF-8 bytes from previous read
				data := buf[:n]
				if len(session.utf8Buffer) > 0 {
					data = append(session.utf8Buffer, data...)
					session.utf8Buffer = nil
				}

				// Check if data ends with incomplete UTF-8 sequence
				incompleteLen := incompleteUTF8Tail(data)
				if incompleteLen > 0 {
					// Buffer the incomplete UTF-8 bytes for next read
					splitPoint := len(data) - incompleteLen
					session.utf8Buffer = append(session.utf8Buffer, data[splitPoint:]...)
					data = data[:splitPoint]
				}

				// Only send if we have data after UTF-8 boundary handling
				if len(data) > 0 {
					// Use BinaryMessage to avoid UTF-8 validation issues
					// PTY output may contain binary telnet sequences
					err = session.ws.WriteMessage(websocket.BinaryMessage, data)
					if err != nil {
						log.Printf("Error writing to WebSocket: %v", err)
						session.mu.Unlock()
						break
					}
				}
			}
			session.mu.Unlock()
		}
	}

	// PTY closed, clean up
	session.cleanup()
}

// cleanup closes the PTY and terminates the process for a shared session
func (s *SharedSession) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	if s.ptmx != nil {
		s.ptmx.Close()
		s.ptmx = nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}

// cleanup closes the PTY and terminates the process (deprecated)
func (s *Session) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	if s.ptmx != nil {
		s.ptmx.Close()
		s.ptmx = nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}

// sendErrorToSession sends an error message to all clients in a shared session
func (h *WebSocketHandler) sendErrorToSession(sharedSession *SharedSession, message string) {
	sharedSession.mu.RLock()
	defer sharedSession.mu.RUnlock()

	errorMsg := fmt.Sprintf("\r\n\x1b[31mERROR: %s\x1b[0m\r\n", message)
	for ws := range sharedSession.clients {
		err := ws.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		if err != nil {
			log.Printf("Error sending error message to client: %v", err)
		}
	}
}

// sendError sends an error message to the WebSocket client (deprecated)
func (s *Session) sendError(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	errorMsg := fmt.Sprintf("\r\n\x1b[31mERROR: %s\x1b[0m\r\n", message)
	err := s.ws.WriteMessage(websocket.TextMessage, []byte(errorMsg))
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

// DataMessage represents file synchronization messages
type PasswordEntry struct {
	Account  string `json:"account"`  // Format: host:port:username
	Password string `json:"password"` // Password for the account
}

type DataMessage struct {
	Type      string          `json:"type"`       // "file_update", "file_request", "file_not_found", "merge_complete", "passwords_init"
	Path      string          `json:"path"`       // File path relative to config directory
	Content   string          `json:"content"`    // File content (JSON string)
	Timestamp int64           `json:"timestamp"`  // Unix timestamp in milliseconds
	Files     []string        `json:"files,omitempty"` // List of files (for merge_complete)
	Passwords []PasswordEntry `json:"passwords,omitempty"` // Password entries (for passwords_init)
}

// DataConnection represents a data synchronization WebSocket connection
type DataConnection struct {
	ws        *websocket.Conn
	sessionID string
	mu        sync.Mutex
	done      chan struct{}
}

// HandleDataWebSocket handles data synchronization WebSocket connections
func (h *WebSocketHandler) HandleDataWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		log.Printf("Data WebSocket connection rejected: no session ID")
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade data WebSocket: %v", err)
		return
	}

	conn := &DataConnection{
		ws:        ws,
		sessionID: sessionID,
		done:      make(chan struct{}),
	}

	log.Printf("Data WebSocket connected for session %s", sessionID)

	// Start file watcher for this session
	go h.watchSessionFiles(conn)
	go h.watchPasswordHints(conn)

	// Handle incoming messages
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Data WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var msg DataMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Invalid data message: %v", err)
				continue
			}

			h.handleDataMessage(conn, &msg)
		}
	}

	// Signal goroutines to stop
	close(conn.done)
	
	log.Printf("Data WebSocket closed for session %s", sessionID)
}

// handleDataMessage processes incoming data synchronization messages
func (h *WebSocketHandler) handleDataMessage(conn *DataConnection, msg *DataMessage) {
	sessionDir := filepath.Join(".websessions", conn.sessionID, ".config", "dikuclient")

	switch msg.Type {
	case "file_update":
		// Client sent a file update
		h.handleClientFileUpdate(conn, sessionDir, msg)
	case "file_request":
		// Client requests a file
		h.handleClientFileRequest(conn, sessionDir, msg)
	case "file_not_found":
		log.Printf("Client doesn't have file: %s", msg.Path)
	case "passwords_init":
		// Client sent passwords for auto-login
		h.handlePasswordsInit(conn, msg)
	default:
		log.Printf("Unknown data message type: %s", msg.Type)
	}
}

// handleClientFileUpdate processes file updates from client
func (h *WebSocketHandler) handleClientFileUpdate(conn *DataConnection, sessionDir string, msg *DataMessage) {
	filePath := filepath.Join(sessionDir, msg.Path)
	
	// Prevent writing .passwords file in web mode (passwords should never reach server)
	if msg.Path == ".passwords" {
		log.Printf("Blocked attempt to write .passwords file in web mode")
		return
	}

	// Check if file exists and its modification time
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		// File exists, check if client version is newer
		serverTime := fileInfo.ModTime().UnixMilli()
		if msg.Timestamp <= serverTime {
			// Server version is newer or same, don't overwrite
			log.Printf("Skipping server update for %s (server version is newer)", msg.Path)
			return
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		log.Printf("Failed to create directory for %s: %v", msg.Path, err)
		return
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(msg.Content), 0600); err != nil {
		log.Printf("Failed to write file %s: %v", msg.Path, err)
		return
	}

	// Set modification time
	modTime := time.UnixMilli(msg.Timestamp)
	if err := os.Chtimes(filePath, modTime, modTime); err != nil {
		log.Printf("Failed to set file time for %s: %v", msg.Path, err)
	}

	log.Printf("Updated server file from client: %s", msg.Path)
}

// handlePasswordsInit processes passwords from client for auto-login
func (h *WebSocketHandler) handlePasswordsInit(conn *DataConnection, msg *DataMessage) {
	if len(msg.Passwords) == 0 {
		return
	}

	h.passwordMu.Lock()
	defer h.passwordMu.Unlock()

	// Initialize password map for this session if needed
	if h.passwordStore[conn.sessionID] == nil {
		h.passwordStore[conn.sessionID] = make(map[string]string)
	}

	// Store passwords in memory
	for _, entry := range msg.Passwords {
		h.passwordStore[conn.sessionID][entry.Account] = entry.Password
	}

	log.Printf("Stored %d passwords in memory for session %s (never written to disk)", len(msg.Passwords), conn.sessionID)
}

// getPasswordsEnv formats passwords as environment variable string
// Format: host:port:username|password entries separated by newlines
func (h *WebSocketHandler) getPasswordsEnv(sessionID string) string {
	h.passwordMu.RLock()
	defer h.passwordMu.RUnlock()

	passwords, ok := h.passwordStore[sessionID]
	if !ok || len(passwords) == 0 {
		return ""
	}

	var entries []string
	for account, password := range passwords {
		entries = append(entries, fmt.Sprintf("%s|%s", account, password))
	}

	return strings.Join(entries, "\n")
}

// handleClientFileRequest processes file requests from client
func (h *WebSocketHandler) handleClientFileRequest(conn *DataConnection, sessionDir string, msg *DataMessage) {
	filePath := filepath.Join(sessionDir, msg.Path)

	data, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist, send not found
		conn.sendMessage(&DataMessage{
			Type: "file_not_found",
			Path: msg.Path,
		})
		return
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat file %s: %v", msg.Path, err)
		return
	}

	// Send file to client
	conn.sendMessage(&DataMessage{
		Type:      "file_update",
		Path:      msg.Path,
		Content:   string(data),
		Timestamp: fileInfo.ModTime().UnixMilli(),
	})
}

// watchSessionFiles watches for changes in session files and syncs to client
func (h *WebSocketHandler) watchSessionFiles(conn *DataConnection) {
	sessionDir := filepath.Join(".websessions", conn.sessionID, ".config", "dikuclient")
	
	// Files to watch
	filesToWatch := []string{
		"accounts.json",
		"history.json",
		"map.json",
		"xps.json",
	}

	// Initial sync - send all existing files to client
	for _, fileName := range filesToWatch {
		filePath := filepath.Join(sessionDir, fileName)
		if data, err := os.ReadFile(filePath); err == nil {
			fileInfo, _ := os.Stat(filePath)
			conn.sendMessage(&DataMessage{
				Type:      "file_update",
				Path:      fileName,
				Content:   string(data),
				Timestamp: fileInfo.ModTime().UnixMilli(),
			})
		}
	}

	// Note: Full file watching implementation would require fsnotify or polling
	// For now, we rely on the initial sync and client-initiated updates
	log.Printf("Initial file sync complete for session %s", conn.sessionID)
}

// watchPasswordHints watches for password hint files and sends them to client
func (h *WebSocketHandler) watchPasswordHints(conn *DataConnection) {
	sessionDir := filepath.Join(".websessions", conn.sessionID)
	hintFile := filepath.Join(sessionDir, ".password_hint")

	// Poll for password hints every 100ms
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if hint file exists
			if data, err := os.ReadFile(hintFile); err == nil {
				// Send hint to client
				conn.sendMessage(&DataMessage{
					Type:    "password_hint",
					Content: string(data),
				})

				// Delete hint file after sending
				os.Remove(hintFile)
				log.Printf("[Server] Sent password hint to client for session %s", conn.sessionID)
			}
		case <-conn.done:
			return
		}
	}
}

// sendMessage sends a data message to the client
func (conn *DataConnection) sendMessage(msg *DataMessage) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := conn.ws.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
