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

## iOS Deployment Strategies

### Option 1: Go Mobile + Terminal Emulator (Recommended for iOS)

**Approach**: Use gomobile to build a native iOS app that embeds a terminal emulator view.

**Architecture**:
```
┌─────────────────────────────────────┐
│         iOS Native App              │
├─────────────────────────────────────┤
│  SwiftUI/UIKit Container            │
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

**Cons**:
- ⚠️ Requires iOS wrapper app development (Swift/SwiftUI)
- ⚠️ App Store submission required
- ⚠️ More complex build process (go + xcode)
- ⚠️ Larger app size (~10-15MB for Go runtime)

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

**Not Recommended for iOS**: iOS restricts background processes, making this approach impractical.

### Option 3: iOS Shell App (Alternative)

**Approach**: Create native iOS terminal app (like Blink Shell, iSH).

**Implementation**:
- Build custom iOS terminal emulator from scratch
- Compile dikuclient as static library
- Link directly into app

**Pros**:
- ✅ Full native integration
- ✅ Best performance

**Cons**:
- ❌ Requires significant iOS development work
- ❌ Most complex implementation
- ❌ Large maintenance burden

**Not Recommended**: Too much custom development required.

### iOS Recommendation: Option 1 (Go Mobile + SwiftTerm)

**Rationale**:
- Minimal changes to dikuclient code
- Leverages existing mature libraries (SwiftTerm)
- Native iOS performance and feel
- Full compatibility with Bubble Tea TUI
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
│  Jetpack Compose / Views            │
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

**Cons**:
- ⚠️ Requires Android app development (Kotlin/Java)
- ⚠️ More complex than Termux approach
- ⚠️ Larger app size
- ⚠️ Requires Play Store approval (or sideload)
- ⚠️ More maintenance burden

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

**Viable but Not Optimal**: Works better on Android than iOS, but Termux is simpler.

### Android Recommendation: Option 1 (Termux)

**Rationale**:
- **Zero code changes** - existing binary works perfectly
- Termux is popular and well-maintained
- Full terminal emulation with all features
- Easy distribution (GitHub releases)
- No Play Store approval needed
- Best performance (native Go binary)
- Users who want CLI MUD client likely already use Termux

**Estimated Effort**:
- Cross-compile binary: 30 minutes
- Create installation docs: 1 hour
- Test on Android devices: 2-3 hours
- Total: **Half a day**

**Alternative for Broader Reach**: If targeting non-technical users who won't install Termux, implement Option 2 (Native App) with estimated effort of 1-2 weeks.

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

### Changes Required for Android (Termux)

**None!** The existing binary works as-is on Android via Termux.

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

**Termux Approach**:
- Go binary is efficient
- Terminal emulation is lightweight
- Minimal battery impact when idle
- Can use `termux-wake-lock` to prevent sleep during long sessions

**Native App Approach**:
- Similar efficiency
- Better battery management integration
- Can use Android/iOS power management APIs

### Memory Usage

- Go runtime: ~5-10 MB
- TUI state: ~1-2 MB
- Termux overhead: ~20-30 MB
- Total: ~30-40 MB (very reasonable)

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

1. **Termux Testing**:
   - Test on various Android devices
   - Different Android versions (8-14)
   - ARM and ARM64 architectures
   - Various screen sizes

2. **Terminal Emulator Testing**:
   - Verify ANSI color rendering
   - Test special keys (Ctrl, Esc, etc.)
   - Verify on-screen keyboard functionality

3. **Performance Testing**:
   - Monitor memory usage
   - Check battery drain
   - Test long-running sessions

## Timeline and Effort Estimates

### Android (Termux Approach)

| Task | Effort | Notes |
|------|--------|-------|
| Cross-compile binaries | 30 min | One-time setup |
| Test on Android devices | 2-3 hours | Various devices |
| Write documentation | 1 hour | README updates |
| Setup GitHub Actions | 1 hour | Automated builds |
| **Total** | **4-5 hours** | Less than a day |

### iOS (Go Mobile + SwiftTerm)

| Task | Effort | Notes |
|------|--------|-------|
| Create mobile package | 1-2 days | Go mobile integration |
| iOS app development | 3-5 days | SwiftUI + SwiftTerm |
| Testing and debugging | 2-3 days | Multiple devices |
| App Store submission | 1 day | If publishing |
| **Total** | **1-2 weeks** | Full-time effort |

## Recommendations

### Immediate Action (Android)

✅ **Implement Termux distribution immediately**
- Requires almost no code changes
- Can be done in half a day
- Provides immediate Android support
- Low maintenance burden
- Users can start using today

**Steps**:
1. Add cross-compilation to build process
2. Update README with Android/Termux instructions
3. Upload binaries to GitHub releases
4. Announce Android support

### Future Enhancement (iOS)

✅ **Plan iOS native app for later**
- Start with TestFlight beta
- Gauge user interest
- Build incrementally
- Consider hiring iOS developer if needed

**Phased Approach**:
- Phase 1: Research and prototype (1 week)
- Phase 2: Alpha testing (2 weeks)
- Phase 3: Beta via TestFlight (1 month)
- Phase 4: App Store release (if warranted)

### Alternative: PWA for Both Platforms

⚠️ **Consider PWA as fallback**
- If iOS native app proves too complex
- Or as interim solution
- Works on both platforms
- But not recommended as primary approach due to limitations

## Conclusion

### Summary of Recommendations

**For Android**: Use Termux distribution
- ✅ Zero code changes
- ✅ Immediate availability
- ✅ Full feature support
- ✅ Easy maintenance
- Effort: Half a day

**For iOS**: Build native app with Go Mobile + SwiftTerm
- ✅ Minimal code changes (200-300 lines)
- ✅ Native experience
- ✅ Full feature compatibility
- ✅ App Store distribution
- Effort: 1-2 weeks

Both approaches maintain the core principle of **running Go code directly on device** with **full-screen terminal display** while making **minimal changes to existing codebase**.

### Next Steps

1. **Immediate**: Implement Android/Termux support (this week)
2. **Short-term**: Prototype iOS app (next month)
3. **Medium-term**: Beta test iOS app via TestFlight
4. **Long-term**: Release on App Store if successful

The Termux approach for Android can be implemented immediately with minimal effort, while the iOS approach requires more investment but provides a polished, native experience. Both solutions run the Go code directly on the device and provide a full-screen terminal experience as required.
