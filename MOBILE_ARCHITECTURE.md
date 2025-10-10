# Mobile Architecture

## Overview

The mobile implementation follows the design specified in MOBILE_DESIGN.md, using gomobile to create native apps that embed the Go TUI code.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         User's Device                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                iOS App (SwiftUI)                         │   │
│  │  ┌───────────────────────────────────────────────────┐  │   │
│  │  │ ContentView (Connection or Terminal)              │  │   │
│  │  │  ┌─────────────────┐  ┌──────────────────────┐   │  │   │
│  │  │  │ Connection Form │  │ Terminal View        │   │  │   │
│  │  │  │  - Host input   │  │  - Output display    │   │  │   │
│  │  │  │  - Port input   │  │  - Input field       │   │  │   │
│  │  │  │  - Connect btn  │  │  - Send button       │   │  │   │
│  │  │  └─────────────────┘  └──────────────────────┘   │  │   │
│  │  └───────────┬───────────────────┬───────────────────┘  │   │
│  │              │                   │                       │   │
│  │  ┌───────────▼───────────────────▼───────────────────┐  │   │
│  │  │ ClientViewModel (State Management)                │  │   │
│  │  │  - Connection state                               │  │   │
│  │  │  - Terminal output                                │  │   │
│  │  │  - PTY management                                 │  │   │
│  │  └───────────┬───────────────────────────────────────┘  │   │
│  │              │ PTY (pseudo-terminal)                    │   │
│  │  ┌───────────▼───────────────────────────────────────┐  │   │
│  │  │ Go Mobile Framework (Dikuclient.xcframework)      │  │   │
│  │  │  ┌─────────────────────────────────────────────┐  │  │   │
│  │  │  │ mobile.StartClient(host, port, ptyFd)       │  │  │   │
│  │  │  │ mobile.SendText(input)                      │  │  │   │
│  │  │  │ mobile.Stop()                               │  │  │   │
│  │  │  └─────────────────────────────────────────────┘  │  │   │
│  │  └───────────┬───────────────────────────────────────┘  │   │
│  └──────────────┼──────────────────────────────────────────┘   │
│                 │                                               │
│  ┌──────────────┼──────────────────────────────────────────┐   │
│  │              │   Android App (Jetpack Compose)          │   │
│  │  ┌───────────▼───────────────────────────────────────┐  │   │
│  │  │ MainActivity (Compose UI)                         │  │   │
│  │  │  ┌─────────────────┐  ┌──────────────────────┐   │  │   │
│  │  │  │ Connection Form │  │ Terminal Screen      │   │  │   │
│  │  │  │  - Host input   │  │  - Output display    │   │  │   │
│  │  │  │  - Port input   │  │  - Input field       │   │  │   │
│  │  │  │  - Connect btn  │  │  - Send button       │   │  │   │
│  │  │  └─────────────────┘  └──────────────────────┘   │  │   │
│  │  └───────────┬───────────────────┬───────────────────┘  │   │
│  │              │                   │                       │   │
│  │  ┌───────────▼───────────────────▼───────────────────┐  │   │
│  │  │ ClientViewModel (State Management)                │  │   │
│  │  │  - Connection state                               │  │   │
│  │  │  - Terminal output                                │  │   │
│  │  │  - PTY management                                 │  │   │
│  │  └───────────┬───────────────────────────────────────┘  │   │
│  │              │ PTY (pseudo-terminal)                    │   │
│  │  ┌───────────▼───────────────────────────────────────┐  │   │
│  │  │ Go Mobile Library (dikuclient.aar)                │  │   │
│  │  │  ┌─────────────────────────────────────────────┐  │  │   │
│  │  │  │ mobile.StartClient(host, port, ptyFd)       │  │  │   │
│  │  │  │ mobile.SendText(input)                      │  │  │   │
│  │  │  │ mobile.Stop()                               │  │  │   │
│  │  │  └─────────────────────────────────────────────┘  │  │   │
│  │  └───────────┬───────────────────────────────────────┘  │   │
│  └──────────────┼──────────────────────────────────────────┘   │
│                 │                                               │
├─────────────────┼───────────────────────────────────────────────┤
│                 │ Shared Go Code Layer                          │
│  ┌──────────────▼───────────────────────────────────────────┐  │
│  │ mobile/ package (github.com/anicolao/dikuclient/mobile)  │  │
│  │  ┌─────────────────────────────────────────────────────┐ │  │
│  │  │ common.go                                           │ │  │
│  │  │  - StartClientWithPTY(host, port, ptyFd)           │ │  │
│  │  │  - SendInput(text)                                  │ │  │
│  │  │  - StopClient()                                     │ │  │
│  │  │  - Instance management (singleton)                  │ │  │
│  │  └─────────────────────────────────────────────────────┘ │  │
│  │  ┌─────────────────────────────────────────────────────┐ │  │
│  │  │ mobile.go (gomobile-compatible API)                │ │  │
│  │  │  - StartClient(host, port, ptyFd) string           │ │  │
│  │  │  - SendText(text) string                           │ │  │
│  │  │  - Stop() string                                    │ │  │
│  │  │  - CheckRunning() bool                             │ │  │
│  │  └─────────────────────────────────────────────────────┘ │  │
│  └──────────────┬───────────────────────────────────────────┘  │
│                 │ Reuses existing Go code                      │
│  ┌──────────────▼───────────────────────────────────────────┐  │
│  │ Existing dikuclient Go Code                             │  │
│  │  ┌─────────────────────────────────────────────────────┐ │  │
│  │  │ internal/tui/      - Bubble Tea TUI                 │ │  │
│  │  │ internal/client/   - MUD connection                 │ │  │
│  │  │ internal/mapper/   - Auto-mapper                    │ │  │
│  │  │ internal/triggers/ - Trigger system                 │ │  │
│  │  │ internal/aliases/  - Alias expansion                │ │  │
│  │  │ ... (other packages)                                │ │  │
│  │  └─────────────────────────────────────────────────────┘ │  │
│  └──────────────┬───────────────────────────────────────────┘  │
│                 │                                               │
│                 ▼ Network I/O                                   │
├─────────────────────────────────────────────────────────────────┤
│                    MUD Server (TCP)                             │
│              e.g., aardmud.org:23                               │
└─────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

### Native UI Layer (Platform-Specific)

**iOS (SwiftUI)**:
- `DikuClientApp.swift`: App entry point
- `ContentView.swift`: Main view controller (connection/terminal switcher)
- `TerminalView.swift`: Terminal display and input
- `ClientViewModel.swift`: State management, PTY handling, Go integration

**Android (Jetpack Compose)**:
- `MainActivity.kt`: Main activity with Compose UI
- `ClientViewModel.kt`: State management, PTY handling, Go integration
- Composable functions for connection and terminal screens

### Go Mobile Layer (Shared)

**mobile/common.go**:
- Core client instance management
- PTY communication handling
- Thread-safe singleton pattern
- Integration with existing TUI code

**mobile/mobile.go**:
- gomobile-compatible API surface
- Simple string-based error handling (no Go error type)
- Functions callable from Swift/Kotlin/Java

### Existing Go Code (Unchanged)

All existing dikuclient packages remain unchanged:
- `internal/tui/`: Bubble Tea TUI
- `internal/client/`: MUD TCP connection
- `internal/mapper/`: Auto-mapper
- `internal/triggers/`: Trigger system
- And all other packages...

## Data Flow

### Connection Flow

```
User taps "Connect"
  ↓
Native UI validates input
  ↓
ClientViewModel.connect(host, port)
  ↓
Native code creates PTY (openpty on iOS, JNI on Android)
  ↓
Call mobile.StartClient(host, port, ptyFd)
  ↓
Go code: StartClientWithPTY creates TUI model
  ↓
Bubble Tea program starts in goroutine
  ↓
Go code connects to MUD server
  ↓
TUI output written to PTY
  ↓
Native code reads from PTY
  ↓
Updates UI with terminal output
```

### Input Flow

```
User types command and taps "Send"
  ↓
Native UI captures input text
  ↓
ClientViewModel.sendInput(text)
  ↓
Call mobile.SendText(text)
  ↓
Go code writes to PTY input
  ↓
Bubble Tea receives input
  ↓
TUI processes command
  ↓
Sends to MUD server
  ↓
Response flows back through PTY to UI
```

## Build Process

### iOS Build

```
1. Developer runs: ./scripts/build-mobile.sh ios
   ↓
2. gomobile bind creates Dikuclient.xcframework
   ↓
3. Framework contains compiled Go code + C bindings
   ↓
4. Developer opens DikuClient.xcodeproj in Xcode
   ↓
5. Xcode links framework into app bundle
   ↓
6. Swift code can import and call Go functions
   ↓
7. Build produces DikuClient.app for iOS
```

### Android Build

```
1. Developer runs: ./scripts/build-mobile.sh android
   ↓
2. gomobile bind creates dikuclient.aar
   ↓
3. AAR contains compiled Go code + JNI bindings
   ↓
4. Developer opens android/ in Android Studio
   ↓
5. Gradle includes AAR as dependency
   ↓
6. Kotlin code can import and call Go functions
   ↓
7. Build produces app-debug.apk for Android
```

## Technology Stack

### iOS
- **Language**: Swift 5.0+
- **UI Framework**: SwiftUI
- **Minimum iOS**: 15.0
- **Build Tool**: Xcode 15+
- **Go Integration**: gomobile (creates .xcframework)

### Android
- **Language**: Kotlin 1.9+
- **UI Framework**: Jetpack Compose + Material 3
- **Minimum Android**: 7.0 (API 24)
- **Build Tool**: Gradle 8.2
- **Go Integration**: gomobile (creates .aar)

### Shared Go Code
- **Language**: Go 1.24+
- **TUI Framework**: Bubble Tea (charmbracelet)
- **Build Tool**: gomobile bind
- **Platforms**: iOS + Android

## Key Design Decisions

### 1. Minimal Changes to Existing Code
- **Decision**: Add new `mobile/` package, don't modify existing code
- **Rationale**: Preserve stability, ease of maintenance
- **Result**: 0 lines changed in existing code ✅

### 2. Native Apps vs Web/Hybrid
- **Decision**: Use native SwiftUI and Jetpack Compose
- **Rationale**: Best performance, platform conventions, floating buttons support
- **Trade-off**: More code vs better UX

### 3. PTY Communication
- **Decision**: Use pseudo-terminal for Go ↔ Native communication
- **Rationale**: Bubble Tea expects terminal I/O, minimal changes needed
- **Alternative**: Could use channels/pipes, but requires TUI refactoring

### 4. gomobile for Go Integration
- **Decision**: Use gomobile bind to create native frameworks
- **Rationale**: Official Go tool, proven approach, clean API
- **Trade-off**: Build complexity vs clean integration

### 5. Singleton Client Instance
- **Decision**: Only one client instance at a time
- **Rationale**: Simplifies state management, matches typical usage
- **Future**: Could support multiple instances if needed

## Testing Strategy

### Unit Tests (Future)
- Go mobile package functions
- Connection validation
- State management

### Integration Tests (Manual)
- Build iOS framework: `./scripts/build-mobile.sh ios`
- Build Android AAR: `./scripts/build-mobile.sh android`
- Test on iOS simulator: `./scripts/test-mobile.sh ios`
- Test on Android emulator: `./scripts/test-mobile.sh android`

### Real Device Testing
- iOS: TestFlight beta distribution
- Android: Internal testing track on Play Store
- Test various MUD servers
- Test network conditions
- Performance profiling

## Future Enhancements

### Phase 1 (Current)
- ✅ Basic connection form
- ✅ Simple terminal display
- ✅ PTY-based Go integration

### Phase 2 (Planned)
- [ ] Full terminal emulator (SwiftTerm for iOS, Termux view for Android)
- [ ] ANSI color support
- [ ] Scrollback buffer
- [ ] Copy/paste support

### Phase 3 (Advanced)
- [ ] Floating action buttons for quick commands
- [ ] Settings screen
- [ ] Multiple MUD profiles
- [ ] Persistent connection across app lifecycle
- [ ] Background execution (where supported)

## Performance Considerations

- **Go Code**: Compiled to native ARM64, excellent performance
- **UI Rendering**: Native SwiftUI/Compose, 60 FPS
- **Network**: Asynchronous I/O, no blocking
- **Memory**: Go's garbage collector + native memory management
- **Battery**: Terminal display is low-power, network is minimal

## Security Considerations

- **Network**: Plain TCP by default (MUD servers typically don't use TLS)
- **Storage**: No local credential storage yet (planned)
- **Permissions**: Only internet access required
- **Sandboxing**: Full iOS/Android app sandboxing

## Conclusion

This architecture provides a clean separation between:
1. **Native UI** (platform-specific, user-facing)
2. **Go Mobile Layer** (thin integration layer)
3. **Existing Go Code** (unchanged, reused)

The design allows testing on emulators and devices while maintaining the ability to enhance both the native UI and Go functionality independently.
