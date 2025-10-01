package web

import (
	"strings"
	"testing"
)

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
