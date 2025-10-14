//go:build !windows
// +build !windows

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
	// Check if user wants a new session via /new path
	if r.URL.Path == "/new" {
		// Generate a new GUID and redirect
		newSessionID := uuid.New().String()
		redirectURL := fmt.Sprintf("/?id=%s", newSessionID)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("New session created (via /new path): %s", newSessionID)
		return
	}

	// Check if session ID is provided
	sessionID := r.URL.Query().Get("id")
	
	// Check for server and port parameters
	server := r.URL.Query().Get("server")
	port := r.URL.Query().Get("port")

	if sessionID == "" {
		// No session ID in URL - check for last session cookie
		if cookie, err := r.Cookie("dikuclient_last_session"); err == nil && cookie.Value != "" {
			// Redirect to last used session (preserve server/port if provided)
			redirectURL := fmt.Sprintf("/?id=%s", cookie.Value)
			if server != "" {
				redirectURL += fmt.Sprintf("&server=%s", server)
			}
			if port != "" {
				redirectURL += fmt.Sprintf("&port=%s", port)
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			log.Printf("Redirecting to last session: %s", cookie.Value)
			return
		}
		
		// No cookie or empty cookie - generate a new GUID and redirect
		newSessionID := uuid.New().String()
		redirectURL := fmt.Sprintf("/?id=%s", newSessionID)
		if server != "" {
			redirectURL += fmt.Sprintf("&server=%s", server)
		}
		if port != "" {
			redirectURL += fmt.Sprintf("&port=%s", port)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("New session created: %s", newSessionID)
		return
	}

	// Check if user explicitly wants a new session via ?id=new
	if sessionID == "new" {
		// Generate a new GUID and redirect
		newSessionID := uuid.New().String()
		redirectURL := fmt.Sprintf("/?id=%s", newSessionID)
		if server != "" {
			redirectURL += fmt.Sprintf("&server=%s", server)
		}
		if port != "" {
			redirectURL += fmt.Sprintf("&port=%s", port)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("New session created (explicit): %s", newSessionID)
		return
	}

	// Session ID provided - use it directly without validation
	// This allows users to share URLs with specific session IDs
	if server != "" && port != "" {
		log.Printf("Using session ID from URL with server %s:%s: %s", server, port, sessionID)
		// Store server/port in session metadata
		s.handler.SetSessionServer(sessionID, server, port)
	} else {
		log.Printf("Using session ID from URL: %s", sessionID)
	}

	// Serve the static index.html file
	http.ServeFile(w, r, filepath.Join("web", "static", "index.html"))
}
