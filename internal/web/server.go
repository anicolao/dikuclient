package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
)

// Server represents the web server
type Server struct {
	port    int
	handler *WebSocketHandler
}

// NewServer creates a new web server
func NewServer(port int) *Server {
	return NewServerWithLogging(port, false)
}

// NewServerWithLogging creates a new web server with logging option
func NewServerWithLogging(port int, enableLogs bool) *Server {
	handler := NewWebSocketHandlerWithLogging(enableLogs)
	handler.SetPort(port)
	return &Server{
		port:    port,
		handler: handler,
	}
}

// Start starts the HTTP server
func Start(port int) error {
	return StartWithLogging(port, false)
}

// StartWithLogging starts the HTTP server with logging option
func StartWithLogging(port int, enableLogs bool) error {
	server := NewServerWithLogging(port, enableLogs)

	// Handle root with session management
	http.HandleFunc("/", server.handleRoot)

	// Serve static files (CSS, JS)
	http.Handle("/styles.css", http.FileServer(http.Dir("web/static")))
	http.Handle("/app.js", http.FileServer(http.Dir("web/static")))
	http.Handle("/storage.js", http.FileServer(http.Dir("web/static")))
	http.Handle("/datasync.js", http.FileServer(http.Dir("web/static")))
	http.Handle("/xterm.min.js", http.FileServer(http.Dir("web/static")))
	http.Handle("/xterm.min.css", http.FileServer(http.Dir("web/static")))
	http.Handle("/addon-fit.min.js", http.FileServer(http.Dir("web/static")))

	// WebSocket endpoints
	http.HandleFunc("/ws", server.handler.HandleWebSocket)
	http.HandleFunc("/data-ws", server.handler.HandleDataWebSocket)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleRoot serves the main page and handles session management
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Check if session ID is provided
	sessionID := r.URL.Query().Get("id")

	if sessionID == "" {
		// No session ID - generate a new GUID and redirect
		newSessionID := uuid.New().String()
		redirectURL := fmt.Sprintf("/?id=%s", newSessionID)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("New session created: %s", newSessionID)
		return
	}

	// Session ID provided - use it directly without validation
	// This allows users to share URLs with specific session IDs
	log.Printf("Using session ID from URL: %s", sessionID)

	// Serve the static index.html file
	http.ServeFile(w, r, filepath.Join("web", "static", "index.html"))
}
