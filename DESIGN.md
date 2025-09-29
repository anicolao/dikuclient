# DikuMUD Client Design Document

## Executive Summary

This document outlines the design for a modern, efficient DikuMUD client that operates in two modes:
1. **Terminal Mode**: Direct TUI (Text User Interface) in the user's terminal
2. **Web Mode**: TUI rendered in a web browser via WebSocket proxy

The client will be written in **Go** and provide a unified codebase that serves both use cases efficiently.

## Language Choice: Go

### Justification for Go over Rust

**Go is the recommended choice** for this project based on the following analysis:

#### Advantages of Go:
- **Rapid Development**: Go's simplicity and excellent standard library enable faster development cycles
- **Superior Web/Network Libraries**: Built-in HTTP/WebSocket support and mature ecosystem (gorilla/websocket, etc.)
- **Excellent Concurrency**: Goroutines provide lightweight concurrency perfect for handling multiple MUD connections
- **Cross-platform**: Easy compilation to multiple platforms with minimal dependencies
- **Terminal Libraries**: Mature TUI libraries like `tview`, `bubbletea`, and `termui`
- **Deployment**: Single binary deployment with no runtime dependencies
- **Community**: Large community with extensive MUD-related projects

#### Why Not Rust:
- **Steeper Learning Curve**: Memory management complexity could slow development
- **Web Integration**: While possible, Go's web ecosystem is more mature and simpler
- **TUI Ecosystem**: Fewer mature TUI libraries compared to Go
- **Development Speed**: Rust's compile times and complexity could hinder rapid iteration

## Overall Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        DikuMUD Client                          │
├─────────────────────────────────────────────────────────────────┤
│                     Application Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   TUI Engine    │  │  Command Parser │  │   Event System  │ │
│  │   (bubbletea)   │  │                 │  │                 │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                      Core Engine                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │    MUD Client   │  │  Session Mgmt   │  │   Plugin System │ │
│  │   Connection    │  │                 │  │                 │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    Interface Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐                     │
│  │  Terminal Mode  │  │    Web Mode     │                     │
│  │   (stdout)      │  │  (WebSocket)    │                     │
│  └─────────────────┘  └─────────────────┘                     │
└─────────────────────────────────────────────────────────────────┘
```

### Mode Selection

The application will determine its mode via command-line flags:

```bash
# Terminal mode (default)
./dikuclient --host mud.server.com --port 4000

# Web mode
./dikuclient --web --port 8080 --web-port 8080
```

## TUI Framework Selection: Bubble Tea

### Why Bubble Tea?

**Bubble Tea** (by Charm) is the recommended TUI framework:

- **Modern Architecture**: Elm-inspired, functional approach with clear separation of concerns
- **Excellent Performance**: Efficient rendering and minimal resource usage
- **Rich Components**: Comprehensive widget library (lists, forms, tables, etc.)
- **Active Development**: Well-maintained with regular updates
- **Web Compatibility**: Components can be adapted for web rendering
- **Testing**: Built-in testing capabilities

### TUI Design Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                         Main Window                            │
├─────────────────────────────────────────────────────────────────┤
│ Status Bar: [Connected] [Health: 100/100] [Mana: 50/50]        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐   │
│  │   Game Output   │  │         Side Panel              │   │
│  │                 │  │  ┌─────────────────────────────┐ │   │
│  │  [MUD Text]     │  │  │      Character Stats       │ │   │
│  │  [Combat]       │  │  │  HP: ████████████ 100%    │ │   │
│  │  [Chat]         │  │  │  MP: █████████████ 75%     │ │   │
│  │  [System]       │  │  └─────────────────────────────┘ │   │
│  │                 │  │  ┌─────────────────────────────┐ │   │
│  │                 │  │  │         Inventory           │ │   │
│  │                 │  │  │  • Sword (+5)              │ │   │
│  │                 │  │  │  • Health Potion (3)       │ │   │
│  │                 │  │  └─────────────────────────────┘ │   │
│  │                 │  │  ┌─────────────────────────────┐ │   │
│  │                 │  │  │          Map                │ │   │
│  │                 │  │  │    [M]─[M]─[M]             │ │   │
│  │                 │  │  │     │   │   │              │ │   │
│  │                 │  │  │    [M]─[@]─[M]             │ │   │
│  │                 │  │  └─────────────────────────────┘ │   │
│  └─────────────────┘  └─────────────────────────────────────┘   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ Input: > look north                                    [Enter] │
└─────────────────────────────────────────────────────────────────┘
```

### Key TUI Features

1. **Multi-pane Layout**: Resizable panes for different content types
2. **Syntax Highlighting**: Color-coded MUD output (combat, chat, etc.)
3. **Command History**: Up/down arrow navigation through previous commands
4. **Auto-completion**: Tab completion for common MUD commands
5. **Logging**: Automatic session logging with search capabilities
6. **Themes**: Multiple color schemes (dark, light, custom)
7. **Hotkeys**: Configurable keyboard shortcuts

## Web Browser Integration

### Architecture for Web Mode

The web mode will use a **hybrid approach** combining server-side rendering with WebSocket communication:

```
┌─────────────────┐    WebSocket    ┌─────────────────┐    TCP/Telnet    ┌─────────────────┐
│   Web Browser   │ ←──────────────→ │  DikuClient     │ ←───────────────→ │   MUD Server    │
│                 │                 │  (Web Mode)     │                  │                 │
│  • HTML/CSS/JS  │                 │                 │                  │                 │
│  • Canvas/Term  │                 │  • HTTP Server  │                  │                 │
│  • WebSocket    │                 │  • WS Handler   │                  │                 │
└─────────────────┘                 │  • TUI Engine   │                  │                 │
                                    └─────────────────┘                  └─────────────────┘
```

### Web Implementation Strategy

#### Option 1: Terminal Emulation (Recommended)
- **Approach**: Use a web-based terminal emulator (xterm.js) to render the TUI
- **Benefits**: Identical experience between terminal and web modes
- **Implementation**: 
  - TUI renders to a virtual terminal buffer
  - Buffer is streamed to browser via WebSocket
  - Keyboard input flows back through WebSocket

#### Option 2: DOM-based Rendering
- **Approach**: Convert TUI components to HTML/CSS equivalents
- **Benefits**: More native web experience, better accessibility
- **Drawbacks**: Requires maintaining two rendering backends

**Recommendation**: Option 1 provides the best code reuse and maintenance efficiency.

### Web Mode Components

1. **HTTP Server**: Serves the web interface
2. **WebSocket Handler**: Manages real-time communication
3. **Terminal Emulator**: xterm.js for rendering
4. **Session Manager**: Handles multiple concurrent users

### Web Interface Files

```
web/
├── static/
│   ├── index.html          # Main web interface
│   ├── app.js              # WebSocket client logic
│   ├── styles.css          # UI styling
│   └── xterm.js            # Terminal emulator
└── templates/
    └── client.html         # Go template for dynamic content
```

## Core Engine Design

### MUD Connection Module

```go
type MUDConnection struct {
    conn        net.Conn
    reader      *bufio.Reader
    writer      *bufio.Writer
    eventChan   chan MUDEvent
    commandChan chan string
    connected   bool
    mu          sync.RWMutex
}

type MUDEvent struct {
    Type      EventType
    Data      string
    Timestamp time.Time
}
```

### Session Management

```go
type Session struct {
    ID          string
    MUDConn     *MUDConnection
    UIRenderer  Renderer
    EventBus    *EventBus
    Config      *Config
    Logger      *Logger
}
```

### Plugin System

Support for extensibility through plugins:

```go
type Plugin interface {
    Name() string
    Initialize(session *Session) error
    HandleEvent(event MUDEvent) error
    Commands() []string
}
```

## Performance Considerations

### Efficiency Optimizations

1. **Connection Pooling**: Reuse TCP connections when possible
2. **Buffer Management**: Efficient text buffer with circular buffers for history
3. **Rendering Optimization**: Only update changed screen regions
4. **Memory Usage**: Bounded buffers to prevent memory leaks
5. **Concurrent Processing**: Separate goroutines for I/O, rendering, and event processing

### Scalability (Web Mode)

1. **Connection Limits**: Configurable maximum concurrent WebSocket connections
2. **Resource Monitoring**: Memory and CPU usage tracking
3. **Load Balancing**: Support for horizontal scaling (future enhancement)

## Configuration and Extensibility

### Configuration System

```yaml
# config.yaml
server:
  host: "mud.server.com"
  port: 4000
  
ui:
  theme: "dark"
  font_size: 12
  show_timestamps: true
  
web:
  port: 8080
  max_connections: 100
  
logging:
  level: "info"
  file: "dikuclient.log"
  
plugins:
  - name: "mapper"
    enabled: true
  - name: "triggers"
    enabled: true
```

### Plugin Examples

1. **Auto-mapper**: Builds maps based on MUD room descriptions
2. **Trigger System**: Custom responses to MUD events
3. **Combat Analyzer**: Parse and display combat statistics
4. **Chat Logger**: Enhanced chat logging and filtering

## Security Considerations

1. **Input Validation**: Sanitize all user input
2. **WebSocket Security**: Implement rate limiting and connection validation
3. **CORS Policy**: Proper cross-origin resource sharing configuration
4. **Authentication**: Optional user authentication for web mode
5. **SSL/TLS**: HTTPS/WSS support for secure connections

## Development Phases

### Phase 1: Core Foundation
- Basic TUI framework setup
- MUD connection handling
- Command input/output

### Phase 2: Enhanced TUI
- Multi-pane layout
- Syntax highlighting
- Configuration system

### Phase 3: Web Integration
- HTTP server setup
- WebSocket communication
- Terminal emulation

### Phase 4: Advanced Features
- Plugin system
- Mapping capabilities
- Performance optimizations

## File Structure

```
dikuclient/
├── cmd/
│   └── dikuclient/
│       └── main.go              # Entry point
├── internal/
│   ├── client/
│   │   ├── connection.go        # MUD connection logic
│   │   ├── session.go           # Session management
│   │   └── events.go            # Event system
│   ├── tui/
│   │   ├── app.go               # Main TUI application
│   │   ├── components/          # UI components
│   │   └── themes/              # Color schemes
│   ├── web/
│   │   ├── server.go            # HTTP server
│   │   ├── websocket.go         # WebSocket handler
│   │   └── static/              # Web assets
│   └── plugins/
│       ├── plugin.go            # Plugin interface
│       └── builtin/             # Built-in plugins
├── pkg/
│   └── mud/
│       ├── protocol.go          # MUD protocol handling
│       └── parser.go            # Text parsing utilities
├── web/
│   └── static/                  # Web interface files
├── configs/
│   └── default.yaml             # Default configuration
├── docs/
│   └── API.md                   # API documentation
└── README.md                    # Project overview
```

## Conclusion

This design provides a solid foundation for a modern, efficient DikuMUD client that serves both terminal and web users with a single codebase. The Go language choice ensures rapid development, excellent performance, and easy deployment, while the Bubble Tea framework provides a modern, maintainable TUI foundation that can be effectively bridged to web browsers through terminal emulation.

The modular architecture allows for future enhancements and community contributions through the plugin system, while the dual-mode operation ensures maximum accessibility for users across different environments.