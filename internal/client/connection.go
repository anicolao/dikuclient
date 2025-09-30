package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Telnet IAC (Interpret As Command) constants
const (
	IAC  = 255 // 0xFF
	WILL = 251 // 0xFB
	WONT = 252 // 0xFC
	DO   = 253 // 0xFD
	DONT = 254 // 0xFE
	GA   = 249 // 0xF9 - Go Ahead (marks end of prompt)
	SB   = 250 // 0xFA - Subnegotiation Begin
	SE   = 240 // 0xF0 - Subnegotiation End
)

// Telnet options
const (
	TELOPT_ECHO = 1
)

// Connection represents a connection to a MUD server
type Connection struct {
	conn         net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
	outChan      chan string
	inChan       chan string
	errChan      chan error
	echoChan     chan bool // Sends echo suppression state changes
	closeCh      chan struct{}
	mu           sync.RWMutex
	closed       bool
	serverEcho   bool   // Whether server is echoing (false = password mode)
	telnetBuffer []byte // Buffer for incomplete telnet sequences
}

// NewConnection creates a new MUD connection
func NewConnection(host string, port int) (*Connection, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c := &Connection{
		conn:       conn,
		reader:     bufio.NewReader(conn),
		writer:     bufio.NewWriter(conn),
		outChan:    make(chan string, 100),
		inChan:     make(chan string, 100),
		errChan:    make(chan error, 10),
		echoChan:   make(chan bool, 10),
		closeCh:    make(chan struct{}),
		serverEcho: true, // Assume server echoes initially
	}

	go c.readLoop()
	go c.writeLoop()

	return c, nil
}

// incompleteUTF8Tail returns the number of trailing bytes that form an incomplete UTF-8 sequence
func incompleteUTF8Tail(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	// Check last 1-4 bytes for incomplete UTF-8
	// UTF-8 encoding:
	// - 1 byte:  0xxxxxxx (0x00-0x7F)
	// - 2 bytes: 110xxxxx 10xxxxxx (0xC0-0xDF, 0x80-0xBF)
	// - 3 bytes: 1110xxxx 10xxxxxx 10xxxxxx (0xE0-0xEF, 0x80-0xBF, 0x80-0xBF)
	// - 4 bytes: 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx (0xF0-0xF7, 0x80-0xBF, 0x80-0xBF, 0x80-0xBF)

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
			// Check if the sequence is valid
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

// processTelnetData strips telnet IAC sequences and handles negotiation
// It properly handles telnet sequences that span buffer boundaries
func (c *Connection) processTelnetData(data []byte) []byte {
	// Prepend any buffered incomplete sequence from previous call
	if len(c.telnetBuffer) > 0 {
		data = append(c.telnetBuffer, data...)
		c.telnetBuffer = nil
	}

	result := make([]byte, 0, len(data))
	i := 0

	for i < len(data) {
		if data[i] == IAC {
			// Check if we have enough bytes for a complete sequence
			if i+1 >= len(data) {
				// Incomplete sequence - buffer it for next call
				c.telnetBuffer = append(c.telnetBuffer, data[i:]...)
				break
			}

			// Handle IAC sequences
			cmd := data[i+1]
			switch cmd {
			case IAC:
				// Escaped IAC (0xFF 0xFF) = literal 0xFF
				result = append(result, IAC)
				i += 2
			case WILL, WONT, DO, DONT:
				// Three-byte sequence: IAC WILL/WONT/DO/DONT <option>
				if i+2 >= len(data) {
					// Incomplete sequence - buffer it for next call
					c.telnetBuffer = append(c.telnetBuffer, data[i:]...)
					// Exit the outer loop
					i = len(data)
				} else {
					option := data[i+2]
					// Handle ECHO option
					if option == TELOPT_ECHO {
						c.mu.Lock()
						oldEcho := c.serverEcho
						if cmd == WILL {
							// Server will echo - this means server will handle echoing
							// For MUDs: WILL ECHO often means password mode (server echoes asterisks)
							// Client should suppress showing the actual input
							c.serverEcho = true
						} else if cmd == WONT {
							// Server won't echo - client should show input
							// This is normal mode where client displays what user types
							c.serverEcho = false
						}
						// Notify UI of echo state change
						// Send true when echo should be suppressed (when server WILL echo)
						if oldEcho != c.serverEcho {
							select {
							case c.echoChan <- c.serverEcho: // true when serverEcho is true (WILL/password)
							default:
							}
						}
						c.mu.Unlock()
					}
					i += 3
				}
			case GA:
				// Go Ahead - marks end of prompt, just skip it
				i += 2
			case SB:
				// Subnegotiation - skip until SE
				sbStart := i
				i += 2
				foundSE := false
				// Find IAC SE
				for i < len(data) {
					if data[i] == IAC {
						if i+1 >= len(data) {
							// Incomplete - buffer from start of SB and exit
							c.telnetBuffer = append(c.telnetBuffer, data[sbStart:]...)
							i = len(data)
							break
						}
						if data[i+1] == SE {
							i += 2 // Skip IAC SE
							foundSE = true
							break
						}
						// IAC followed by something other than SE (e.g., IAC IAC)
						// Skip both bytes
						i += 2
					} else {
						i++
					}
				}
				// If we didn't find SE and didn't buffer, we hit end of data
				if !foundSE && i >= len(data) && len(c.telnetBuffer) == 0 {
					// Buffer the entire incomplete subnegotiation
					c.telnetBuffer = append(c.telnetBuffer, data[sbStart:]...)
				}
			default:
				// Unknown two-byte sequence
				i += 2
			}
		} else {
			// Normal character
			result = append(result, data[i])
			i++
		}
	}

	// Check if result ends with incomplete UTF-8 sequence
	incompleteLen := incompleteUTF8Tail(result)
	if incompleteLen > 0 {
		// Buffer the incomplete UTF-8 bytes for next call
		splitPoint := len(result) - incompleteLen
		c.telnetBuffer = append(c.telnetBuffer, result[splitPoint:]...)
		result = result[:splitPoint]
	}

	return result
}

// readLoop continuously reads from the MUD server
func (c *Connection) readLoop() {
	defer func() {
		c.Close()
	}()

	buffer := make([]byte, 4096)
	var accumulated bytes.Buffer

	for {
		select {
		case <-c.closeCh:
			return
		default:
			// Set read timeout to check for partial data (prompts)
			c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			n, err := c.conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout - check if we have accumulated data to send
					if accumulated.Len() > 0 {
						data := accumulated.Bytes()
						accumulated.Reset()
						// Process telnet sequences
						cleaned := c.processTelnetData(data)
						// Strip \r characters
						dataStr := strings.ReplaceAll(string(cleaned), "\r", "")
						if dataStr != "" {
							c.outChan <- dataStr
						}
					}
					continue
				}
				if err != io.EOF {
					c.errChan <- fmt.Errorf("read error: %w", err)
				}
				return
			}

			if n > 0 {
				accumulated.Write(buffer[:n])

				// Check if we have complete lines
				data := accumulated.Bytes()
				dataStr := string(data)
				if strings.Contains(dataStr, "\n") {
					// Send complete lines immediately
					accumulated.Reset()
					// Process telnet sequences
					cleaned := c.processTelnetData(data)
					// Strip \r characters
					cleanedStr := strings.ReplaceAll(string(cleaned), "\r", "")
					if cleanedStr != "" {
						c.outChan <- cleanedStr
					}
				}
			}
		}
	}
}

// writeLoop continuously writes to the MUD server
func (c *Connection) writeLoop() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.closeCh:
			return
		case msg := <-c.inChan:
			_, err := c.writer.WriteString(msg + "\r\n")
			if err != nil {
				c.errChan <- fmt.Errorf("write error: %w", err)
				return
			}
			if err := c.writer.Flush(); err != nil {
				c.errChan <- fmt.Errorf("flush error: %w", err)
				return
			}
		}
	}
}

// Send sends a command to the MUD server
func (c *Connection) Send(msg string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.closed {
		c.inChan <- msg
	}
}

// Receive returns the output channel for reading server messages
func (c *Connection) Receive() <-chan string {
	return c.outChan
}

// EchoState returns the echo state channel (true = suppressed/password mode)
func (c *Connection) EchoState() <-chan bool {
	return c.echoChan
}

// Errors returns the error channel
func (c *Connection) Errors() <-chan error {
	return c.errChan
}

// Close closes the connection
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.closeCh)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// IsClosed returns whether the connection is closed
func (c *Connection) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}
