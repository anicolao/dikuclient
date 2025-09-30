package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

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
	sessions map[*websocket.Conn]*Session
	mu       sync.RWMutex
}

// Session represents a WebSocket session with a PTY running the TUI
type Session struct {
	ws     *websocket.Conn
	ptmx   *os.File
	cmd    *exec.Cmd
	mu     sync.Mutex
	closed bool
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
	return &WebSocketHandler{
		sessions: make(map[*websocket.Conn]*Session),
	}
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

	// Create a new session
	session := &Session{
		ws: ws,
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
					case "connect":
						h.handleConnect(session, message)
						continue
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

// handleConnect starts the TUI client in a PTY
func (h *WebSocketHandler) handleConnect(session *Session, message []byte) {
	var connectMsg ConnectMessage
	if err := json.Unmarshal(message, &connectMsg); err != nil {
		session.sendError(fmt.Sprintf("Invalid connect message: %v", err))
		return
	}

	// Get the path to the dikuclient binary
	// In production, this should be configurable
	dikuclientPath, err := exec.LookPath("dikuclient")
	if err != nil {
		// Try relative path
		dikuclientPath = "./dikuclient"
	}

	// Start the TUI client
	cmd := exec.Command(dikuclientPath, "--host", connectMsg.Host, "--port", fmt.Sprintf("%d", connectMsg.Port))
	
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
				err = session.ws.WriteMessage(websocket.TextMessage, buf[:n])
				if err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					session.mu.Unlock()
					break
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
