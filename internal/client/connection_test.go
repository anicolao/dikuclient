package client

import (
	"bytes"
	"testing"
)

func TestProcessTelnetData_CompleteSsequences(t *testing.T) {
	conn := &Connection{}

	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "plain text",
			input:    []byte("Hello, World!"),
			expected: []byte("Hello, World!"),
		},
		{
			name:     "escaped IAC",
			input:    []byte{IAC, IAC, 'A', 'B'},
			expected: []byte{IAC, 'A', 'B'},
		},
		{
			name:     "IAC GA (Go Ahead)",
			input:    []byte{'A', IAC, GA, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC WILL ECHO",
			input:    []byte{'A', IAC, WILL, TELOPT_ECHO, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC WONT ECHO",
			input:    []byte{'A', IAC, WONT, TELOPT_ECHO, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC DO option",
			input:    []byte{'A', IAC, DO, 3, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC DONT option",
			input:    []byte{'A', IAC, DONT, 3, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC SB subnegotiation IAC SE",
			input:    []byte{'A', IAC, SB, 1, 2, 3, IAC, SE, 'B'},
			expected: []byte{'A', 'B'},
		},
		{
			name:     "IAC SB with escaped IAC inside",
			input:    []byte{'A', IAC, SB, 1, IAC, IAC, 2, IAC, SE, 'B'},
			expected: []byte{'A', 'B'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset telnet buffer for each test
			conn.telnetBuffer = nil
			result := conn.processTelnetData(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("processTelnetData() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessTelnetData_BoundarySpanning(t *testing.T) {
	conn := &Connection{}

	tests := []struct {
		name     string
		chunks   [][]byte
		expected string
	}{
		{
			name: "IAC at end of buffer",
			chunks: [][]byte{
				[]byte("Hello" + string([]byte{IAC})),
				[]byte{GA, 'W', 'o', 'r', 'l', 'd'},
			},
			expected: "HelloWorld",
		},
		{
			name: "IAC WILL at boundary",
			chunks: [][]byte{
				[]byte("Test" + string([]byte{IAC})),
				[]byte{WILL, TELOPT_ECHO, '!'},
			},
			expected: "Test!",
		},
		{
			name: "IAC WILL incomplete - needs option",
			chunks: [][]byte{
				[]byte("Test" + string([]byte{IAC, WILL})),
				[]byte{TELOPT_ECHO, '!'},
			},
			expected: "Test!",
		},
		{
			name: "IAC SB spanning multiple chunks",
			chunks: [][]byte{
				[]byte("A" + string([]byte{IAC, SB})),
				[]byte{1, 2, 3},
				[]byte{IAC, SE, 'B'},
			},
			expected: "AB",
		},
		{
			name: "Multiple IAC sequences with boundaries",
			chunks: [][]byte{
				[]byte{IAC},
				[]byte{GA},
				[]byte{'H', 'i', IAC},
				[]byte{WILL},
				[]byte{TELOPT_ECHO},
				[]byte{'!'},
			},
			expected: "Hi!",
		},
		{
			name: "Escaped IAC at boundary",
			chunks: [][]byte{
				[]byte("Test" + string([]byte{IAC})),
				[]byte{IAC, 'O', 'K'},
			},
			expected: "Test" + string([]byte{IAC}) + "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset telnet buffer for each test
			conn.telnetBuffer = nil

			var result []byte
			for _, chunk := range tt.chunks {
				processed := conn.processTelnetData(chunk)
				result = append(result, processed...)
			}

			if string(result) != tt.expected {
				t.Errorf("processTelnetData() = %q, want %q", string(result), tt.expected)
			}

			// Verify buffer is empty at end
			if len(conn.telnetBuffer) > 0 {
				t.Errorf("telnetBuffer not empty at end: %v", conn.telnetBuffer)
			}
		})
	}
}

func TestProcessTelnetData_PartialSequenceAtEnd(t *testing.T) {
	conn := &Connection{}

	// First chunk ends with partial IAC sequence
	chunk1 := []byte("Hello" + string([]byte{IAC}))
	result1 := conn.processTelnetData(chunk1)

	if string(result1) != "Hello" {
		t.Errorf("First chunk: got %q, want %q", string(result1), "Hello")
	}

	// Buffer should contain the IAC
	if len(conn.telnetBuffer) != 1 || conn.telnetBuffer[0] != IAC {
		t.Errorf("Buffer should contain IAC, got %v", conn.telnetBuffer)
	}

	// Second chunk completes the sequence
	chunk2 := []byte{GA, 'W', 'o', 'r', 'l', 'd'}
	result2 := conn.processTelnetData(chunk2)

	if string(result2) != "World" {
		t.Errorf("Second chunk: got %q, want %q", string(result2), "World")
	}

	// Buffer should be empty now
	if len(conn.telnetBuffer) != 0 {
		t.Errorf("Buffer should be empty, got %v", conn.telnetBuffer)
	}
}

func TestProcessTelnetData_UTF8Boundaries(t *testing.T) {
	conn := &Connection{}

	tests := []struct {
		name     string
		chunks   [][]byte
		expected string
	}{
		{
			name: "2-byte UTF-8 at boundary (é = 0xC3 0xA9)",
			chunks: [][]byte{
				[]byte("Hello " + string([]byte{0xC3})),
				[]byte{0xA9, '!'},
			},
			expected: "Hello é!",
		},
		{
			name: "3-byte UTF-8 at boundary (€ = 0xE2 0x82 0xAC)",
			chunks: [][]byte{
				[]byte("Price: " + string([]byte{0xE2})),
				[]byte{0x82, 0xAC},
			},
			expected: "Price: €",
		},
		{
			name: "3-byte UTF-8 split 2-1",
			chunks: [][]byte{
				[]byte("Test " + string([]byte{0xE2, 0x82})),
				[]byte{0xAC, ' ', 'O', 'K'},
			},
			expected: "Test € OK",
		},
		{
			name: "4-byte UTF-8 at boundary (😀 = 0xF0 0x9F 0x98 0x80)",
			chunks: [][]byte{
				[]byte("Emoji " + string([]byte{0xF0})),
				[]byte{0x9F, 0x98, 0x80},
			},
			expected: "Emoji 😀",
		},
		{
			name: "4-byte UTF-8 split 2-2",
			chunks: [][]byte{
				[]byte("Hi " + string([]byte{0xF0, 0x9F})),
				[]byte{0x98, 0x80, '!'},
			},
			expected: "Hi 😀!",
		},
		{
			name: "4-byte UTF-8 split 3-1",
			chunks: [][]byte{
				[]byte("Wow " + string([]byte{0xF0, 0x9F, 0x98})),
				[]byte{0x80},
			},
			expected: "Wow 😀",
		},
		{
			name: "Multiple multi-byte characters",
			chunks: [][]byte{
				[]byte("Test " + string([]byte{0xC3})),
				[]byte{0xA9, ' ', 0xE2},
				[]byte{0x82, 0xAC, ' ', 0xF0},
				[]byte{0x9F, 0x98, 0x80},
			},
			expected: "Test é € 😀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset telnet buffer for each test
			conn.telnetBuffer = nil

			var result []byte
			for _, chunk := range tt.chunks {
				processed := conn.processTelnetData(chunk)
				result = append(result, processed...)
			}

			if string(result) != tt.expected {
				t.Errorf("processTelnetData() = %q (bytes: %v), want %q", string(result), result, tt.expected)
			}

			// Verify buffer is empty at end
			if len(conn.telnetBuffer) > 0 {
				t.Errorf("telnetBuffer not empty at end: %v", conn.telnetBuffer)
			}
		})
	}
}
