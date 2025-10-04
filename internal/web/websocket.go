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
	"sync"
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
	sessions      map[*websocket.Conn]*Session
	mu            sync.RWMutex
	enableLogs    bool   // Whether to enable logging for spawned TUI instances
	currentSessID string // Current session ID to use for new connections
	sessionIDMu   sync.RWMutex
	port          int    // Server port for generating share URLs
}

// Session represents a WebSocket session with a PTY running the TUI
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
		sessions:   make(map[*websocket.Conn]*Session),
		enableLogs: enableLogs,
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

	// Create a new session
	session := &Session{
		ws:        ws,
		sessionID: sessionID,
	}

	h.mu.Lock()
	h.sessions[ws] = session
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.sessions, ws)
		h.mu.Unlock()
		session.cleanup()
	}()

	// Wait for initial message with terminal size before starting TUI
	// This ensures the PTY is created with the correct size from the start
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

	// Auto-start the TUI client with initial size if available
	h.autoStartTUIWithSize(session, initialSize)

	// Handle incoming messages
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
						h.handleResize(session, message)
						continue
					}
				}
			}

			// Otherwise, it's terminal input - send to PTY
			if session.ptmx != nil && !session.closed {
				session.mu.Lock()
				_, err := session.ptmx.Write(message)
				session.mu.Unlock()
				if err != nil {
					log.Printf("Error writing to PTY: %v", err)
					break
				}
			}
		}
	}
}

// autoStartTUIWithSize automatically starts the TUI client in a PTY with optional initial size
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

// handleResize handles terminal resize requests
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

// forwardPTYOutput forwards output from PTY to WebSocket
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

// cleanup closes the PTY and terminates the process
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

// sendError sends an error message to the WebSocket client
func (s *Session) sendError(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	errorMsg := fmt.Sprintf("\r\n\x1b[31mERROR: %s\x1b[0m\r\n", message)
	err := s.ws.WriteMessage(websocket.TextMessage, []byte(errorMsg))
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}
