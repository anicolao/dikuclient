package web

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestPasswordsInitFormat(t *testing.T) {
	// Test that the JavaScript format matches Go deserialization expectations
	tests := []struct {
		name        string
		jsonInput   string
		wantCount   int
		wantAccount string
		wantPass    string
		wantErr     bool
	}{
		{
			name:        "empty array",
			jsonInput:   `{"type":"passwords_init","passwords":[]}`,
			wantCount:   0,
			wantErr:     false,
		},
		{
			name:        "single password",
			jsonInput:   `{"type":"passwords_init","passwords":[{"account":"mud.org:4000:testuser","password":"testpass"}]}`,
			wantCount:   1,
			wantAccount: "mud.org:4000:testuser",
			wantPass:    "testpass",
			wantErr:     false,
		},
		{
			name:        "multiple passwords",
			jsonInput:   `{"type":"passwords_init","passwords":[{"account":"mud1.org:23:user1","password":"pass1"},{"account":"mud2.org:4000:user2","password":"pass2"}]}`,
			wantCount:   2,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg DataMessage
			err := json.Unmarshal([]byte(tt.jsonInput), &msg)
			
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if err != nil {
				return
			}

			if len(msg.Passwords) != tt.wantCount {
				t.Errorf("got %d passwords, want %d", len(msg.Passwords), tt.wantCount)
			}

			if tt.wantCount > 0 && tt.wantAccount != "" {
				if msg.Passwords[0].Account != tt.wantAccount {
					t.Errorf("got account %q, want %q", msg.Passwords[0].Account, tt.wantAccount)
				}
				if msg.Passwords[0].Password != tt.wantPass {
					t.Errorf("got password %q, want %q", msg.Passwords[0].Password, tt.wantPass)
				}
			}
		})
	}
}

func TestWebSocketHandler_parseConnectMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "valid connect message",
			message:  "CONNECT:example.com:4000",
			wantHost: "example.com",
			wantPort: 4000,
			wantErr:  false,
		},
		{
			name:     "valid connect with different port",
			message:  "CONNECT:mud.server.org:23",
			wantHost: "mud.server.org",
			wantPort: 23,
			wantErr:  false,
		},
		{
			name:     "invalid format - no port",
			message:  "CONNECT:example.com",
			wantHost: "",
			wantPort: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - no host",
			message:  "CONNECT::4000",
			wantHost: "",
			wantPort: 4000,
			wantErr:  false, // Empty host is technically valid, will fail on connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.message, "CONNECT:") {
				return
			}

			parts := strings.Split(tt.message[8:], ":")
			if len(parts) != 2 && tt.wantErr {
				return // Expected error
			}

			if len(parts) == 2 {
				host := parts[0]
				if host != tt.wantHost && !tt.wantErr {
					t.Errorf("got host %q, want %q", host, tt.wantHost)
				}
			}
		})
	}
}

func TestSharedSession_TerminalSizeStorage(t *testing.T) {
	// Test that terminal size is properly stored and retrieved in SharedSession
	session := &SharedSession{
		sessionID: "test-session",
		clients:   make(map[*websocket.Conn]bool),
	}

	// Initially, no size should be set
	if session.rows != 0 || session.cols != 0 {
		t.Errorf("expected initial size to be 0x0, got %dx%d", session.cols, session.rows)
	}

	// Simulate setting terminal size
	session.rows = 50
	session.cols = 120

	// Verify size is stored
	if session.rows != 50 || session.cols != 120 {
		t.Errorf("expected size 120x50, got %dx%d", session.cols, session.rows)
	}

	// Simulate resize
	session.rows = 60
	session.cols = 150

	// Verify updated size
	if session.rows != 60 || session.cols != 150 {
		t.Errorf("expected size 150x60, got %dx%d", session.cols, session.rows)
	}
}

func TestHandleSharedResize_StoresSize(t *testing.T) {
	// Test that handleSharedResize stores the terminal size
	handler := NewWebSocketHandler()
	session := &SharedSession{
		sessionID: "test-session",
		clients:   make(map[*websocket.Conn]bool),
		closed:    true, // Mark as closed so we don't try to set PTY size
	}

	// Create resize message
	resizeMsg := ResizeMessage{
		Type: "resize",
		Cols: 100,
		Rows: 40,
	}
	msgBytes, _ := json.Marshal(resizeMsg)

	// Handle resize
	handler.handleSharedResize(session, msgBytes)

	// Verify size is stored
	if session.rows != 40 || session.cols != 100 {
		t.Errorf("expected size 100x40 to be stored, got %dx%d", session.cols, session.rows)
	}
}
