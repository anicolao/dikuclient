# Mobile Deployment Design Document

## Executive Summary

This document outlines strategies for deploying the DikuMUD client on iOS and Android mobile devices while minimizing code changes. The goal is to run the existing Go code directly on mobile devices with a full-screen terminal interface.

## Current Architecture

The dikuclient is built using:
- **Language**: Go 1.24
- **TUI Framework**: Bubble Tea (terminal UI)
- **UI Rendering**: lipgloss for styling
- **Network**: Standard Go TCP/telnet connection
- **Web Support**: WebSocket-based terminal emulation using xterm.js

Key characteristics:
- Single binary deployment
- Terminal-based interface (ANSI/VT100 escape sequences)
- No GUI framework dependencies
- Existing web mode that runs TUI in a PTY and streams to browser

## Mobile Requirements

1. **Run Go code natively** on the device (not via remote server)
2. **Full-screen terminal display** for the TUI
3. **Minimal code changes** to existing codebase
4. **Support both iOS and Android**
5. **Touchscreen input** support for typing commands
6. **Keyboard support** when available
7. **Standalone installation** - Users should be able to install DikuClient and have a working setup without needing to install any other programs
8. **Floating buttons capability** - The app architecture must support overlaying UI buttons on top of the terminal window for fast controls, with buttons potentially overlapping the terminal content

### Impact of Standalone Requirement

The standalone installation requirement is a critical constraint that significantly influences the deployment strategy:

**What it means**:
- Users install **one app** and can immediately use DikuClient
- No prerequisite apps or tools (like Termux) required
- No manual configuration or setup steps
- Professional, consumer-grade user experience

**What it rules out**:
- ❌ Termux-based distribution (requires installing Termux first)
- ❌ PWA with separate server process (complex background service management)
- ❌ Any solution requiring command-line setup or additional downloads

**What it requires**:
- ✅ Self-contained native apps for iOS and Android
- ✅ App store distribution (App Store / Play Store)
- ✅ Embedded terminal emulator within the app
- ✅ All dependencies bundled in the app package

This constraint shifts the recommendation from "quick and simple" (Termux) to "professional and complete" (native apps), requiring more initial development effort but providing a significantly better user experience.

### Impact of Floating Buttons Requirement

The floating buttons requirement ensures future extensibility for advanced mobile controls:

**What it means**:
- Ability to overlay interactive UI controls on top of the terminal display
- Buttons can overlap terminal content (not constrained to non-overlapping regions)
- Fast-access controls for common MUD commands or client features
- Native mobile UI patterns (floating action buttons, overlays)

**What it requires**:
- ✅ Native UI framework that supports layered views (Z-order control)
- ✅ Terminal emulator as a view component (not full-screen exclusive mode)
- ✅ Ability to handle touch events both for buttons and terminal
- ✅ Flexible layout system for positioning floating elements

**What it rules out**:
- ❌ Full-screen terminal-only solutions without UI layer capability
- ❌ Simple terminal emulation without native UI integration

**Why native apps satisfy this**:
- **iOS (SwiftUI/UIKit)**: ZStack in SwiftUI or view hierarchy in UIKit naturally supports overlaying views. SwiftTerm is a UIView component that can be placed in a view hierarchy with other UI elements layered on top.
- **Android (Jetpack Compose/Views)**: Box composable or FrameLayout naturally supports Z-ordering. Terminal views can have other composables/views rendered on top with full control over layering.

This requirement reinforces the need for native app development on both platforms, as both iOS and Android UI frameworks provide robust support for floating/overlaid UI elements.

## iOS Deployment Strategies

### Option 1: Go Mobile + Terminal Emulator (Recommended for iOS)

**Approach**: Use gomobile to build a native iOS app that embeds a terminal emulator view.

**Architecture**:
```
┌─────────────────────────────────────┐
│         iOS Native App              │
├─────────────────────────────────────┤
│  SwiftUI/UIKit Container (ZStack)   │
│  ┌───────────────────────────────┐  │
│  │  Terminal Emulator View       │  │
│  │  (SwiftTerm or custom)        │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  PTY (pseudo-terminal)  │  │  │
│  │  │  ┌───────────────────┐  │  │  │
│  │  │  │  Go Binary        │  │  │  │
│  │  │  │  (dikuclient TUI) │  │  │  │
│  │  │  └───────────────────┘  │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │  Floating Buttons Layer       │  │
│  │  (overlaid on terminal)       │  │
│  │  [↑] [↓] [⚔] [🛡] [🏃]       │  │
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │  On-screen Keyboard           │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Implementation Details**:

1. **Go Mobile Integration**:
   - Use `gomobile bind` to create iOS framework
   - Expose Go functions for terminal I/O
   - Create bridge between Go code and Swift

2. **Terminal Emulator**:
   - **SwiftTerm**: Open-source VT100/xterm emulator for iOS
     - Full ANSI color support
     - VT100/VT220/xterm compatibility
     - Native UIKit/SwiftUI integration
   - Alternative: Build custom terminal view with NSAttributedString

3. **PTY Integration**:
   - Run dikuclient in a pseudo-terminal (PTY)
   - SwiftTerm reads from PTY and renders
   - Keyboard input writes to PTY

4. **Code Changes Required**:
   - Minimal changes to dikuclient core
   - Add iOS-specific entry point (can reuse main.go logic)
   - Ensure proper signal handling for mobile environment
   - Add gomobile-compatible API layer

**Pros**:
- ✅ Full compatibility with existing TUI code
- ✅ All Bubble Tea features work unchanged
- ✅ ANSI colors and formatting preserved
- ✅ Native iOS performance
- ✅ Offline-capable (Go runs locally)
- ✅ SwiftTerm is mature and well-maintained
- ✅ **Floating buttons support**: SwiftUI ZStack or UIKit view hierarchy enables easy overlay of buttons on terminal

**Cons**:
- ⚠️ Requires iOS wrapper app development (Swift/SwiftUI)
- ⚠️ App Store submission required
- ⚠️ More complex build process (go + xcode)
- ⚠️ Larger app size (~10-15MB for Go runtime)

**Standalone Requirement**: ✅ **Satisfies** - Users install one app from App Store, no additional dependencies needed

**Floating Buttons Capability**: ✅ **Fully Supported** - SwiftUI's ZStack or UIKit's view layering allows floating buttons to be overlaid on the terminal view. Example:
```swift
ZStack {
    TerminalView()  // SwiftTerm terminal
    VStack {
        Spacer()
        HStack {
            FloatingButton(icon: "arrow.up") { sendCommand("north") }
            FloatingButton(icon: "arrow.down") { sendCommand("south") }
            // ... more buttons
        }
        .padding()
    }
}
```

**Build Process**:
```bash
# 1. Create Go mobile framework
gomobile bind -target=ios github.com/anicolao/dikuclient/mobile

# 2. Create Xcode project with SwiftTerm
# 3. Import Go framework
# 4. Build and sign iOS app
# 5. Deploy via TestFlight or App Store
```

### Option 2: Progressive Web App (PWA)

**Approach**: Use the existing web mode as a PWA installable on iOS.

**Architecture**:
```
┌─────────────────────────────────────┐
│      iOS Safari / Home Screen       │
├─────────────────────────────────────┤
│  PWA (Installed Web App)            │
│  ┌───────────────────────────────┐  │
│  │  xterm.js Terminal Emulator   │  │
│  │  (existing web interface)     │  │
│  └───────────────────────────────┘  │
│  ↓ WebSocket Connection             │
│  ┌───────────────────────────────┐  │
│  │  Go Web Server (localhost)    │  │
│  │  Running dikuclient TUI       │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Implementation Details**:

1. **Add PWA Manifest**:
   ```json
   {
     "name": "DikuMUD Client",
     "short_name": "DikuClient",
     "start_url": "/",
     "display": "standalone",
     "orientation": "portrait",
     "theme_color": "#000000",
     "background_color": "#000000",
     "icons": [...]
   }
   ```

2. **Service Worker**:
   - Cache static assets for offline use
   - Handle app lifecycle

3. **Background Server**:
   - Run Go server as iOS background process (challenging on iOS)
   - Alternative: Use local WebView with embedded server

**Pros**:
- ✅ Minimal code changes (just add manifest.json)
- ✅ No App Store approval needed
- ✅ Cross-platform (works on Android too)
- ✅ Easy updates
- ✅ Existing web mode already functional

**Cons**:
- ❌ Requires separate Go server process running
- ❌ iOS background process restrictions
- ❌ No truly native feel
- ❌ Network overhead (localhost WebSocket)
- ❌ Battery drain from running server
- ⚠️ **Limited floating buttons support** - Could overlay HTML buttons but would be non-native and less performant

**Not Recommended for iOS**: iOS restricts background processes, making this approach impractical. Also doesn't provide native floating buttons capability.

### Option 3: iOS Shell App (Alternative)

**Approach**: Create native iOS terminal app (like Blink Shell, iSH).

**Implementation**:
- Build custom iOS terminal emulator from scratch
- Compile dikuclient as static library
- Link directly into app

**Pros**:
- ✅ Full native integration
- ✅ Best performance
- ✅ Native floating buttons support

**Cons**:
- ❌ Requires significant iOS development work
- ❌ Most complex implementation
- ❌ Large maintenance burden
- ❌ Reinventing the wheel (SwiftTerm already exists)

**Not Recommended**: Too much custom development required when existing solutions (SwiftTerm) already provide all needed functionality including floating buttons support.

### iOS Recommendation: Option 1 (Go Mobile + SwiftTerm)

**Rationale**:
- Minimal changes to dikuclient code
- Leverages existing mature libraries (SwiftTerm)
- Native iOS performance and feel
- Full compatibility with Bubble Tea TUI
- **Full floating buttons support** via SwiftUI/UIKit layering
- Most maintainable long-term solution

**Estimated Effort**:
- Go mobile wrapper: 1-2 days
- iOS app with SwiftTerm: 3-5 days
- Testing and polish: 2-3 days
- Total: ~1-2 weeks

## Android Deployment Strategies

### Option 1: Termux Integration (Recommended for Android)

**Approach**: Distribute as a standard Go binary that runs in Termux terminal emulator.

**Architecture**:
```
┌─────────────────────────────────────┐
│          Android Device             │
├─────────────────────────────────────┐
│  Termux App (from F-Droid/Play)     │
│  ┌───────────────────────────────┐  │
│  │  Terminal Emulator            │  │
│  │  (Full VT100/xterm support)   │  │
│  │                               │  │
│  │  $ ./dikuclient --host ...    │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  Go Binary (ARM64)      │  │  │
│  │  │  dikuclient TUI         │  │  │
│  │  │  (Bubble Tea)           │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
│  • Full Linux environment           │
│  • Package manager (pkg install)    │
│  • Go runtime available             │
└─────────────────────────────────────┘
```

**Implementation Details**:

1. **Cross-Compile for Android**:
   ```bash
   # For ARM64 (most modern Android devices)
   GOOS=linux GOARCH=arm64 go build -o dikuclient-android-arm64 ./cmd/dikuclient
   
   # For 32-bit ARM (older devices)
   GOOS=linux GOARCH=arm go build -o dikuclient-android-arm ./cmd/dikuclient
   ```

2. **Distribution**:
   - Upload binary to GitHub releases
   - Users install via Termux:
     ```bash
     pkg install golang git
     go install github.com/anicolao/dikuclient/cmd/dikuclient@latest
     ```
   - Or download pre-built binary:
     ```bash
     curl -L https://github.com/anicolao/dikuclient/releases/latest/download/dikuclient-android-arm64 -o dikuclient
     chmod +x dikuclient
     ./dikuclient --host mud.server.com --port 4000
     ```

3. **Termux Features**:
   - Full VT100/xterm terminal emulation
   - Hardware keyboard support
   - On-screen keyboard with special keys
   - SSH support for remote connections
   - Notification support
   - Widget support

4. **Code Changes Required**:
   - **None!** Existing binary works as-is
   - Optional: Add Android-specific optimizations
   - Optional: Detect Termux environment and adjust defaults

**Pros**:
- ✅ **Zero code changes required**
- ✅ Full Bubble Tea TUI compatibility
- ✅ All ANSI colors and features work
- ✅ Native performance (compiled Go binary)
- ✅ Easy distribution (single binary)
- ✅ No Play Store approval needed
- ✅ Users already familiar with Termux
- ✅ Can use Termux:Widget for shortcuts
- ✅ SSH access for remote use

**Cons**:
- ⚠️ Requires users to install Termux
- ⚠️ Not a standalone app (runs in terminal)
- ⚠️ Less discoverable than Play Store app
- ❌ **Does NOT satisfy standalone requirement** - Users must install Termux first
- ❌ **Does NOT support floating buttons** - Terminal-only environment, no native UI layer for overlays

**User Experience**:
1. Install Termux from F-Droid or Play Store
2. Open Termux
3. Install dikuclient: `go install github.com/anicolao/dikuclient/cmd/dikuclient@latest`
4. Run: `dikuclient --host aardmud.org --port 23`
5. Enjoy full TUI experience

**Enhancement: Termux Shortcuts**:
Create a `dikuclient.shortcut` file:
```bash
#!/data/data/com.termux/files/usr/bin/bash
termux-wake-lock
dikuclient --account myaccount
```

Place in `~/.shortcuts/` and create home screen widget.

### Option 2: Native Android App with Go Mobile

**Approach**: Similar to iOS Option 1, use gomobile to create Android app.

**Architecture**:
```
┌─────────────────────────────────────┐
│      Native Android App             │
├─────────────────────────────────────┤
│  Jetpack Compose / Views (Box)      │
│  ┌───────────────────────────────┐  │
│  │  Terminal Emulator View       │  │
│  │  (Termux library or custom)   │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │  PTY (android-pty)      │  │  │
│  │  │  ┌───────────────────┐  │  │  │
│  │  │  │  Go Library       │  │  │  │
│  │  │  │  (dikuclient)     │  │  │  │
│  │  │  └───────────────────┘  │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │  Floating Buttons Layer       │  │
│  │  (overlaid on terminal)       │  │
│  │  [↑] [↓] [⚔] [🛡] [🏃]       │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Implementation Details**:

1. **Go Mobile for Android**:
   ```bash
   gomobile bind -target=android github.com/anicolao/dikuclient/mobile
   ```

2. **Terminal Emulator Libraries**:
   - Use Termux terminal emulator library
   - Or TerminalView (open-source Android library)

3. **Build Process**:
   - Create Android Studio project
   - Import Go mobile AAR
   - Implement terminal view
   - Handle PTY communication

**Pros**:
- ✅ Standalone app (no Termux needed)
- ✅ Play Store distribution possible
- ✅ Native Android feel
- ✅ Better discoverability
- ✅ Full Bubble Tea TUI compatibility
- ✅ All ANSI colors and features work
- ✅ **Floating buttons support**: Jetpack Compose Box or FrameLayout enables easy overlay of buttons on terminal

**Cons**:
- ⚠️ Requires Android app development (Kotlin/Java)
- ⚠️ More complex than Termux approach
- ⚠️ Larger app size (~15-20MB)
- ⚠️ Requires Play Store approval (or sideload)
- ⚠️ More maintenance burden

**Standalone Requirement**: ✅ **Satisfies** - Users install one app from Play Store, no additional dependencies needed

**Floating Buttons Capability**: ✅ **Fully Supported** - Jetpack Compose's Box or traditional FrameLayout allows floating buttons to be overlaid on the terminal view. Example:
```kotlin
Box {
    TerminalView()  // Terminal emulator
    Column(
        modifier = Modifier.align(Alignment.BottomCenter)
    ) {
        Row {
            FloatingButton("↑") { sendCommand("north") }
            FloatingButton("↓") { sendCommand("south") }
            // ... more buttons
        }
    }
}
```

**Estimated Effort**:
- Go mobile wrapper: 1-2 days
- Android app with terminal view: 3-5 days
- Testing and polish: 2-3 days
- Total: **1-2 weeks**

### Option 3: Progressive Web App (PWA)

**Approach**: Similar to iOS Option 2, but more viable on Android.

**Implementation Details**:

1. **Add PWA Manifest** (same as iOS)
2. **Service Worker for Offline**
3. **Background Service**:
   - Android allows background services more easily than iOS
   - Could use Foreground Service with notification
   - Or use WebView with embedded Go server

**Pros**:
- ✅ No app store needed
- ✅ Cross-platform
- ✅ Easy updates

**Cons**:
- ⚠️ Still requires separate server process
- ⚠️ Battery drain
- ⚠️ Complex background service management
- ⚠️ Not truly native
- ❌ **Does NOT satisfy standalone requirement** - Requires managing background server process
- ⚠️ **Limited floating buttons support** - Could overlay HTML buttons but would be non-native and less performant

**Not Recommended**: Complex architecture with poor user experience for standalone and floating buttons requirements.

### Android Recommendation: Option 2 (Native Android App with Go Mobile)

**Given the standalone and floating buttons requirements**, the Native Android App is the only viable option that satisfies all constraints.

**Rationale**:
- ✅ **Satisfies standalone requirement** - One app installation, no dependencies
- ✅ **Supports floating buttons** - Native UI framework with full overlay capability
- ✅ Play Store distribution for easy discovery and installation
- ✅ Native Android experience with proper app lifecycle management
- ✅ Full compatibility with existing TUI code
- ✅ All Bubble Tea features work unchanged
- ✅ Professional user experience

**Estimated Effort**:
- Go mobile wrapper: 1-2 days
- Android app with terminal emulator: 3-5 days
- Testing and polish: 2-3 days
- Total: **1-2 weeks**

**Why Not Termux (Option 1)?**
While the Termux approach requires zero code changes and can be implemented in half a day, it has two critical limitations:
1. **Violates standalone requirement** - Users must first install Termux (a separate app) before they can use DikuClient
2. **No floating buttons support** - Termux is a terminal-only environment with no native UI layer for overlaying buttons

These limitations make it unsuitable despite its simplicity.

**Implementation Path**: Similar to iOS, use gomobile to create an Android library, then build a native Android app with an embedded terminal emulator view (using Termux's open-source terminal library or TerminalView).

## Code Changes Summary

### Changes Required for iOS (Go Mobile + SwiftTerm)

**1. Create Mobile Package** (`mobile/ios.go`):
```go
package mobile

import (
    "github.com/anicolao/dikuclient/internal/client"
    "github.com/anicolao/dikuclient/internal/tui"
)

// StartClient starts the dikuclient and returns file descriptors for PTY
// This is called from iOS Swift code
func StartClient(host string, port int) (int, error) {
    // Similar to existing main.go but returns PTY file descriptor
    // iOS side creates PTY and passes to Go
    // Go writes TUI output to PTY
    return ptyFd, nil
}

// SendInput sends user input to the running client
func SendInput(input string) error {
    // Write to input channel
    return nil
}
```

**2. Adjust Main Logic** (optional, for clean separation):
- Extract core TUI logic into reusable function
- Keep cmd/dikuclient/main.go for CLI use
- Add mobile/ios.go for iOS integration

**3. Build Configuration**:
```bash
# Add gomobile.yml to .github/workflows for automated builds
# Build script for iOS framework
```

**Estimated Lines Changed**: ~200-300 new lines, 0 modified in existing code

### Changes Required for Android (Native App with Go Mobile)

**1. Create Mobile Package** (`mobile/android.go`):
```go
package mobile

import (
    "github.com/anicolao/dikuclient/internal/client"
    "github.com/anicolao/dikuclient/internal/tui"
)

// StartClient starts the dikuclient and returns file descriptors for PTY
// This is called from Android Kotlin/Java code
func StartClient(host string, port int) (int, error) {
    // Similar to iOS implementation
    // Android side creates PTY and passes to Go
    // Go writes TUI output to PTY
    return ptyFd, nil
}

// SendInput sends user input to the running client
func SendInput(input string) error {
    // Write to input channel
    return nil
}
```

**2. Shared Mobile Logic** (optional):
- Extract common mobile code into `mobile/common.go`
- Both iOS and Android can reuse the same core logic
- Only platform-specific PTY handling differs

**3. Build Configuration**:
```bash
# Build Android AAR library
gomobile bind -target=android -o dikuclient.aar github.com/anicolao/dikuclient/mobile

# Add to .github/workflows for automated Android builds
```

**Estimated Lines Changed**: ~200-300 new lines (similar to iOS), 0 modified in existing code

**Note**: The mobile wrapper code can be largely shared between iOS and Android, minimizing duplication.

**Optional Enhancements**:
1. **Detect Termux environment**:
   ```go
   func isTermux() bool {
       return os.Getenv("TERMUX_VERSION") != ""
   }
   ```

2. **Android-specific defaults**:
   ```go
   if isTermux() {
       // Maybe adjust default terminal size
       // Or add Termux-specific shortcuts
   }
   ```

**Estimated Lines Changed**: 0 required, ~50-100 optional

## Distribution Strategy

### iOS Distribution

**Option A: App Store**
- Requires Apple Developer account ($99/year)
- 1-2 week review process
- Widest reach

**Option B: TestFlight**
- Beta distribution (up to 10,000 users)
- Good for initial testing
- Easier approval than App Store

**Option C: Enterprise Distribution**
- For organizations
- No App Store review
- Requires Enterprise Developer account

### Android Distribution

**Option A: Direct Download (Termux Approach)**
- Upload binary to GitHub Releases
- Document installation in README
- No approval process
- Recommended approach

**Option B: F-Droid** (if building native app)
- Open-source app store
- Popular in Android community
- Free, but requires fully open-source app

**Option C: Play Store** (if building native app)
- One-time $25 fee
- Wider reach
- Review process

## Installation Documentation

### iOS (SwiftTerm App)

**README Addition**:
```markdown
## iOS Installation

### TestFlight Beta
1. Install TestFlight from App Store
2. Open TestFlight link: [beta link]
3. Install DikuMUD Client
4. Launch app and connect to your favorite MUD

### App Store (once released)
1. Search "DikuMUD Client" in App Store
2. Install
3. Launch and connect
```

### Android (Termux)

**README Addition**:
```markdown
## Android Installation

### Method 1: Using Termux (Recommended)

1. Install Termux from F-Droid (https://f-droid.org/packages/com.termux/)
   - Do NOT use Play Store version (outdated)
   
2. Open Termux and run:
   ```bash
   # Install Go
   pkg install golang git
   
   # Install dikuclient
   go install github.com/anicolao/dikuclient/cmd/dikuclient@latest
   
   # Run it
   dikuclient --host aardmud.org --port 23
   ```

### Method 2: Pre-built Binary

1. Install Termux from F-Droid
2. In Termux, run:
   ```bash
   # Download latest release
   curl -L https://github.com/anicolao/dikuclient/releases/latest/download/dikuclient-android-arm64 -o dikuclient
   chmod +x dikuclient
   ./dikuclient --host mud.server.com --port 4000
   ```

### Optional: Home Screen Shortcut

Create `~/.shortcuts/dikuclient.sh`:
```bash
#!/data/data/com.termux/files/usr/bin/bash
termux-wake-lock
dikuclient --account myaccount
```

Then add Termux:Widget to your home screen.

### Troubleshooting

- **Can't execute binary**: Run `chmod +x dikuclient`
- **Wrong architecture**: Download arm (32-bit) instead of arm64
- **Termux permissions**: Run `termux-setup-storage` for storage access
```

## Build Automation

### GitHub Actions Workflow

**`.github/workflows/mobile-release.yml`**:
```yaml
name: Mobile Release

on:
  release:
    types: [created]

jobs:
  build-android:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Build for Android ARM64
        run: |
          GOOS=linux GOARCH=arm64 go build -o dikuclient-android-arm64 ./cmd/dikuclient
          
      - name: Build for Android ARM
        run: |
          GOOS=linux GOARCH=arm go build -o dikuclient-android-arm ./cmd/dikuclient
      
      - name: Upload Release Assets
        uses: actions/upload-release-asset@v1
        with:
          upload_url: ${{ github.event.release.upload_url }}
          asset_path: ./dikuclient-android-arm64
          asset_name: dikuclient-android-arm64
          asset_content_type: application/octet-stream
  
  build-ios:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install gomobile
        run: |
          go install golang.org/x/mobile/cmd/gomobile@latest
          gomobile init
      
      - name: Build iOS Framework
        run: |
          gomobile bind -target=ios -o DikuClient.xcframework github.com/anicolao/dikuclient/mobile
      
      - name: Upload Framework
        uses: actions/upload-artifact@v3
        with:
          name: ios-framework
          path: DikuClient.xcframework
```

## Performance Considerations

### Battery Life

**Native App Approach** (iOS and Android):
- Go binary is efficient (~5-10 MB runtime)
- Terminal emulation is lightweight
- Minimal battery impact when idle
- Better battery management integration with native APIs
- Can use iOS/Android power management features properly

### Memory Usage

- Go runtime: ~5-10 MB
- TUI state: ~1-2 MB
- Native app overhead: ~10-15 MB
- Terminal emulator: ~5-10 MB
- Total: ~25-35 MB (very reasonable for modern devices)

### Network Usage

- Telnet protocol is extremely lightweight
- Typical MUD session: <100 KB/hour
- No HTTP overhead with direct connection
- Minimal battery impact from network

## Testing Strategy

### iOS Testing

1. **Simulator Testing**:
   - Test in iOS Simulator during development
   - Verify terminal rendering
   - Test keyboard input

2. **Device Testing**:
   - Test on physical iPhone/iPad
   - Various screen sizes (iPhone SE to Pro Max)
   - Verify performance and battery life

3. **TestFlight Beta**:
   - Release to small group of beta testers
   - Gather feedback on usability

### Android Testing

1. **Emulator Testing**:
   - Test in Android Studio emulator during development
   - Verify terminal rendering
   - Test keyboard input and on-screen keyboard

2. **Device Testing**:
   - Test on physical Android devices
   - Different Android versions (8-14)
   - ARM and ARM64 architectures
   - Various screen sizes (phones and tablets)
   - Test with different manufacturers (Samsung, Google, etc.)

3. **Terminal Emulator Testing**:
   - Verify ANSI color rendering
   - Test special keys (Ctrl, Esc, etc.)
   - Verify on-screen keyboard functionality

4. **Performance Testing**:
   - Monitor memory usage
   - Check battery drain
   - Test long-running sessions
   
5. **Beta Testing**:
   - Internal testing track on Play Store
   - Gather feedback before public release

## Timeline and Effort Estimates

### Android (Native App with Go Mobile)

| Task | Effort | Notes |
|------|--------|-------|
| Create mobile package | 1-2 days | Go mobile integration (shared with iOS) |
| Android app development | 3-5 days | Kotlin + Terminal View |
| Testing and debugging | 2-3 days | Multiple devices |
| Play Store submission | 1 day | If publishing |
| **Total** | **1-2 weeks** | Full-time effort |

### iOS (Go Mobile + SwiftTerm)

| Task | Effort | Notes |
|------|--------|-------|
| Create mobile package | 1-2 days | Go mobile integration (shared with Android) |
| iOS app development | 3-5 days | SwiftUI + SwiftTerm |
| Testing and debugging | 2-3 days | Multiple devices |
| App Store submission | 1 day | If publishing |
| **Total** | **1-2 weeks** | Full-time effort |

**Note**: If both platforms are developed together, the mobile package code can be shared, reducing overall effort to approximately **2-3 weeks total** instead of 4 weeks.

## Recommendations

### Given the Standalone and Floating Buttons Requirements

With the constraints that users should be able to install DikuClient without needing any other programs AND that the app must support overlaying floating buttons on the terminal, **both platforms require native app development**.

### Recommended Approach: Native Apps for Both Platforms

✅ **Build native apps using Go Mobile for both iOS and Android**

**Advantages of Parallel Development**:
- Shared mobile wrapper code (~200-300 lines can be reused)
- Consistent user experience across platforms
- Both satisfy standalone requirement
- **Both support floating buttons** via native UI frameworks (SwiftUI/UIKit and Jetpack Compose/Views)
- Professional app store presence
- Easier maintenance with unified codebase

**Development Strategy**:
1. **Phase 1: Shared Mobile Package** (1-2 days)
   - Create `mobile/common.go` with shared logic
   - Create `mobile/ios.go` for iOS-specific code
   - Create `mobile/android.go` for Android-specific code
   - Set up gomobile build for both platforms

2. **Phase 2: Platform-Specific Apps** (1-2 weeks each, can be parallel)
   - **iOS**: SwiftUI app with SwiftTerm integration
   - **Android**: Kotlin app with Terminal View integration
   - Both apps use the same Go mobile library

3. **Phase 3: Testing** (1 week)
   - TestFlight beta for iOS
   - Internal testing track for Android
   - Cross-platform bug fixes
   - Performance optimization

4. **Phase 4: Release** (1 week)
   - App Store submission (iOS)
   - Play Store submission (Android)
   - Documentation and marketing materials

**Total Estimated Effort**: 2-3 weeks if developed in parallel, 3-4 weeks if sequential

### Why Not Termux or PWA?

- **Termux (Option 1)**: While it requires zero code changes and minimal effort, it has two critical limitations:
  - **Violates standalone requirement** - Users must install Termux first
  - **No floating buttons support** - Terminal-only environment with no native UI layer
  
- **PWA (Option 3)**: While it works on both platforms, it has multiple limitations:
  - **Violates standalone requirement** - Requires managing background server process
  - **Limited floating buttons support** - Could overlay HTML buttons but non-native and less performant
  - Not truly native experience

### Incremental Rollout Strategy

If resources are limited, prioritize one platform first:

**Option A: iOS First**
- Larger App Store revenue potential
- More affluent user base
- Stricter quality standards benefit long-term

**Option B: Android First**  
- Larger user base globally
- Easier development and testing
- Less restrictive review process

However, given the code sharing potential, **developing both simultaneously is recommended** to maximize efficiency.

## Conclusion

### Summary of Recommendations

**Given the standalone installation requirement**, both platforms require native app development:

**For iOS**: Native app with Go Mobile + SwiftTerm
- ✅ Minimal code changes (200-300 lines of mobile wrapper)
- ✅ Satisfies standalone requirement (one app install)
- ✅ **Supports floating buttons** (SwiftUI ZStack/UIKit layering)
- ✅ Native experience with App Store distribution
- ✅ Full TUI compatibility
- Effort: 1-2 weeks

**For Android**: Native app with Go Mobile + Terminal View
- ✅ Minimal code changes (200-300 lines, shared with iOS)
- ✅ Satisfies standalone requirement (one app install)
- ✅ **Supports floating buttons** (Jetpack Compose Box/FrameLayout)
- ✅ Native experience with Play Store distribution
- ✅ Full TUI compatibility
- Effort: 1-2 weeks

**Both approaches maintain all core requirements**:
- ✅ **Run Go code directly on device** (not via remote server)
- ✅ **Full-screen terminal display** with complete ANSI support
- ✅ **Minimal changes to existing codebase** (~200-300 lines of mobile wrapper, no changes to core)
- ✅ **Standalone installation** (no additional apps required)
- ✅ **Floating buttons capability** (native UI framework overlay support)

### Key Insight: Code Sharing

The mobile wrapper code can be largely shared between iOS and Android platforms, making the combined effort **2-3 weeks** instead of 4 weeks if developed sequentially. This makes native app development for both platforms the most efficient approach.

### Next Steps

1. **Phase 1**: Create shared mobile package (1-2 days)
   - Implement `mobile/common.go` with core logic
   - Add platform-specific wrappers for iOS and Android
   - Set up gomobile build process

2. **Phase 2**: Develop native apps (1-2 weeks, can be parallel)
   - iOS: SwiftUI + SwiftTerm with floating button layer
   - Android: Kotlin + Terminal View with floating button layer
   - Both apps integrate the same Go mobile library

3. **Phase 3**: Beta testing (1 week)
   - TestFlight (iOS) and Internal Testing (Android)
   - Gather feedback and fix bugs
   - Performance optimization

4. **Phase 4**: Public release (1 week)
   - App Store and Play Store submissions
   - Documentation and user guides
   - Marketing and announcement

**Total Timeline**: Approximately 3-4 weeks to launch on both platforms with a standalone installation experience that satisfies all requirements, including floating buttons support for future enhancements.
