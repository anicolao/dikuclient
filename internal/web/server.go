package web

import (
	"fmt"
	"log"
	"net/http"
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
	return &Server{
		port:    port,
		handler: NewWebSocketHandlerWithLogging(enableLogs),
	}
}

// Start starts the HTTP server
func Start(port int) error {
	return StartWithLogging(port, false)
}

// StartWithLogging starts the HTTP server with logging option
func StartWithLogging(port int, enableLogs bool) error {
	server := NewServerWithLogging(port, enableLogs)

	// Serve static files from web/static directory
	http.Handle("/", http.FileServer(http.Dir("web/static")))

	// WebSocket endpoint
	http.HandleFunc("/ws", server.handler.HandleWebSocket)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
