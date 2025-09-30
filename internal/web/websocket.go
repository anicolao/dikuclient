package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/anicolao/dikuclient/internal/client"
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

// Session represents a WebSocket session with a MUD connection
type Session struct {
	ws   *websocket.Conn
	conn *client.Connection
	mu   sync.Mutex
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
		if session.conn != nil {
			session.conn.Close()
		}
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
			h.handleMessage(session, string(message))
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessage(session *Session, message string) {
	// Parse message format: {"type": "command", "data": "..."}
	// For now, we'll use a simple protocol
	
	// Check if it's a connect command
	if len(message) > 8 && message[:8] == "CONNECT:" {
		// Format: CONNECT:host:port
		parts := strings.Split(message[8:], ":")
		if len(parts) != 2 {
			session.sendError("Invalid connect command format. Expected: CONNECT:host:port")
			return
		}
		
		host := parts[0]
		port := 0
		_, err := fmt.Sscanf(parts[1], "%d", &port)
		if err != nil {
			session.sendError(fmt.Sprintf("Invalid port number: %v", err))
			return
		}

		// Create MUD connection
		conn, err := client.NewConnection(host, port)
		if err != nil {
			session.sendError(fmt.Sprintf("Failed to connect: %v", err))
			return
		}

		session.mu.Lock()
		session.conn = conn
		session.mu.Unlock()

		// Send success message
		session.send("CONNECTED")

		// Start forwarding MUD output to WebSocket
		go h.forwardMUDOutput(session)

		return
	}

	// If connected, send to MUD
	session.mu.Lock()
	conn := session.conn
	session.mu.Unlock()

	if conn != nil && !conn.IsClosed() {
		conn.Send(message)
	} else {
		session.sendError("Not connected to MUD server")
	}
}

// forwardMUDOutput forwards output from MUD to WebSocket
func (h *WebSocketHandler) forwardMUDOutput(session *Session) {
	session.mu.Lock()
	conn := session.conn
	session.mu.Unlock()

	if conn == nil {
		return
	}

	for {
		select {
		case msg := <-conn.Receive():
			session.send(msg)
		case err := <-conn.Errors():
			session.sendError(fmt.Sprintf("MUD error: %v", err))
			return
		}
	}
}

// send sends a message to the WebSocket client
func (s *Session) send(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.ws.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// sendError sends an error message to the WebSocket client
func (s *Session) sendError(message string) {
	s.send("ERROR:" + message)
}
