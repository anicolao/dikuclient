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
)

// Connection represents a connection to a MUD server
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	outChan  chan string
	inChan   chan string
	errChan  chan error
	closeCh  chan struct{}
	mu       sync.RWMutex
	closed   bool
}

// NewConnection creates a new MUD connection
func NewConnection(host string, port int) (*Connection, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c := &Connection{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		outChan: make(chan string, 100),
		inChan:  make(chan string, 100),
		errChan: make(chan error, 10),
		closeCh: make(chan struct{}),
	}

	go c.readLoop()
	go c.writeLoop()

	return c, nil
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
						data := accumulated.String()
						accumulated.Reset()
						// Strip \r characters
						data = strings.ReplaceAll(data, "\r", "")
						if data != "" {
							c.outChan <- data
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
				data := accumulated.String()
				if strings.Contains(data, "\n") {
					// Send complete lines immediately
					accumulated.Reset()
					// Strip \r characters
					data = strings.ReplaceAll(data, "\r", "")
					if data != "" {
						c.outChan <- data
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
