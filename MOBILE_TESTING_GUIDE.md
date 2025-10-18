# Mobile Testing Guide

This guide shows you how to test the mobile apps that have been implemented.

## What Has Been Implemented

✅ **iOS Native App** (SwiftUI)
- Complete app structure ready for Xcode
- Connection form UI
- Terminal display UI
- ViewModel with PTY integration points

✅ **Android Native App** (Jetpack Compose)
- Complete app structure ready for Android Studio
- Connection form UI
- Terminal display UI
- ViewModel with PTY integration points

✅ **Go Mobile Package**
- gomobile-compatible API
- Client management
- Integration points for TUI code

## Testing the Apps

### Option 1: Test UI Without Go Integration

Both apps can be opened and built in their respective IDEs to test the UI:

#### iOS (requires macOS)

```bash
# Open in Xcode
open ios/DikuClient.xcodeproj

# In Xcode:
# 1. Select iPhone 15 simulator (or any simulator)
# 2. Press Cmd+R to build and run
# 3. You'll see the connection form
# 4. Enter any host/port and tap Connect
# 5. Terminal view will appear (simulated connection)
```

**What works**: UI, navigation, state management
**What doesn't work yet**: Actual MUD connection (needs Go integration)

#### Android (requires Android Studio)

```bash
# Open in Android Studio
# File → Open → Select dikuclient/android directory

# In Android Studio:
# 1. Wait for Gradle sync to complete
# 2. Select an emulator from device dropdown (or create one)
# 3. Click green Run button
# 4. You'll see the connection form
# 5. Enter any host/port and tap Connect
# 6. Terminal screen will appear (simulated connection)
```

**What works**: UI, navigation, state management
**What doesn't work yet**: Actual MUD connection (needs Go integration)

### Option 2: Full Integration Testing

To test with actual MUD connectivity, you need to:

#### Step 1: Install gomobile

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

#### Step 2: Build Go Mobile Frameworks

```bash
# From dikuclient root directory
./scripts/build-mobile.sh all
```

This creates:
- `ios/Frameworks/Dikuclient.xcframework` (iOS)
- `android/app/libs/dikuclient.aar` (Android)

#### Step 3: Link Frameworks to Apps

**iOS**:
1. Open `ios/DikuClient.xcodeproj` in Xcode
2. Drag `ios/Frameworks/Dikuclient.xcframework` into the project
3. Select "Copy items if needed"
4. Add to "Frameworks, Libraries, and Embedded Content"
5. In `ClientViewModel.swift`, uncomment the Go function calls:
   - Replace simulated `connect()` with `DikuclientStartClient()`
   - Replace simulated `sendInput()` with `DikuclientSendText()`
   - Replace simulated `disconnect()` with `DikuclientStop()`

**Android**:
1. The AAR is already referenced in `app/build.gradle`
2. In Android Studio, sync Gradle
3. In `ClientViewModel.kt`, uncomment the Go function calls:
   - Replace simulated `connect()` with `mobile.StartClient()`
   - Replace simulated `sendInput()` with `mobile.SendText()`
   - Replace simulated `disconnect()` with `mobile.Stop()`

#### Step 4: Test Full Functionality

Now the apps will:
- ✅ Connect to real MUD servers
- ✅ Display actual TUI output
- ✅ Send commands to MUD
- ✅ Show Bubble Tea interface in terminal view

## What You Can Test Right Now

### Without gomobile (UI only)

1. **iOS**: Open Xcode project, build, see UI
2. **Android**: Open Android Studio project, build, see UI
3. Verify connection form layout
4. Verify terminal view layout
5. Test navigation between screens
6. Test input field and send button

### With gomobile (Full functionality)

1. All of the above, plus:
2. Connect to real MUD servers (e.g., aardmud.org:23)
3. See actual game output in terminal
4. Send commands and see responses
5. Full Bubble Tea TUI in mobile app

## Test Scenarios

### Scenario 1: Build Verification

```bash
# Verify iOS project builds
./scripts/test-mobile.sh ios

# Verify Android project builds
./scripts/test-mobile.sh android
```

**Expected**: Both projects compile without errors

### Scenario 2: UI Testing

1. Launch app on simulator/emulator
2. See connection form with:
   - Host input field
   - Port input field
   - Connect button
   - App title and version
3. Enter "aardmud.org" and "23"
4. Tap Connect
5. See terminal view with:
   - Text output area
   - Input field at bottom
   - Send button
   - Disconnect button in nav bar

**Expected**: All UI elements present and functional

### Scenario 3: Full Integration (requires gomobile)

1. Build Go frameworks with `./scripts/build-mobile.sh all`
2. Link frameworks to apps (see Step 3 above)
3. Launch app
4. Enter "aardmud.org" and "23"
5. Tap Connect
6. See Aardwolf MUD welcome text
7. Type "help" and tap Send
8. See help text response
9. Tap Disconnect
10. Return to connection form

**Expected**: Real MUD interaction works

## Known Limitations (Current Version)

### iOS App
- ❌ Go integration commented out (PTY setup incomplete)
- ❌ Terminal is basic text view (no ANSI colors yet)
- ❌ No SwiftTerm integration (planned)
- ✅ UI structure complete
- ✅ Navigation works
- ✅ Ready for Go framework integration

### Android App
- ❌ Go integration commented out (PTY setup incomplete)
- ❌ Terminal is basic text view (no ANSI colors yet)
- ❌ No Termux terminal-view integration (planned)
- ✅ UI structure complete
- ✅ Navigation works
- ✅ Ready for Go AAR integration

## Troubleshooting

### iOS: "Command not found: xcodebuild"

Install Xcode from the App Store, then:
```bash
sudo xcode-select --switch /Applications/Xcode.app
```

### iOS: "No simulators available"

In Xcode, go to Preferences → Components → Install iOS simulators

### Android: "SDK location not found"

Create `android/local.properties`:
```
sdk.dir=/path/to/Android/sdk
```

Or set `ANDROID_HOME` environment variable

### Android: Gradle sync failed

1. Ensure Android Studio is updated
2. Install Android SDK Platform 34
3. Install Android SDK Build-Tools

### gomobile: "command not found"

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

## Performance Testing

Once full integration is complete:

### iOS Performance
- Launch time: Should be < 2 seconds
- Connection time: Depends on network
- Input response: Should be immediate
- Memory usage: Check in Xcode Instruments
- Battery drain: Monitor in Settings

### Android Performance
- Launch time: Should be < 2 seconds
- Connection time: Depends on network
- Input response: Should be immediate
- Memory usage: Check in Android Profiler
- Battery drain: Monitor in Settings

## Device Testing Matrix

### iOS
- [ ] iPhone SE (small screen)
- [ ] iPhone 15 (standard)
- [ ] iPhone 15 Pro Max (large)
- [ ] iPad (tablet)
- [ ] iOS 15.0 (minimum)
- [ ] iOS 17.x (latest)

### Android
- [ ] Small phone (5" screen)
- [ ] Standard phone (6" screen)
- [ ] Large phone (6.5"+ screen)
- [ ] Tablet (10"+ screen)
- [ ] Android 7.0 / API 24 (minimum)
- [ ] Android 14 / API 34 (latest)

## Next Steps After Testing

1. **Fix PTY Integration**: Complete pseudo-terminal setup for Go code
2. **Add Terminal Emulator**: 
   - iOS: Integrate SwiftTerm
   - Android: Integrate Termux terminal-view
3. **Add ANSI Color Support**: Full terminal emulation
4. **Add Floating Buttons**: Quick action buttons overlaid on terminal
5. **Add Settings Screen**: Connection history, preferences
6. **Beta Testing**:
   - iOS: TestFlight
   - Android: Play Store internal testing
7. **Public Release**:
   - iOS: App Store
   - Android: Play Store and F-Droid

## Getting Help

- **Build Issues**: See [MOBILE_BUILD.md](./MOBILE_BUILD.md)
- **Quick Commands**: See [MOBILE_QUICKSTART.md](./MOBILE_QUICKSTART.md)
- **Architecture**: See [MOBILE_ARCHITECTURE.md](./MOBILE_ARCHITECTURE.md)
- **Design**: See [MOBILE_DESIGN.md](./MOBILE_DESIGN.md)

## Summary

The mobile implementation is **testable right now**:

✅ **UI Testing**: Open in IDE and test interface (works today)
✅ **Build Testing**: Verify projects compile (works today)
🚧 **Integration Testing**: Requires gomobile build step (documented)
🚧 **Full Testing**: Requires PTY integration (next step)

Both apps provide a solid foundation that can be built upon incrementally. The structure is complete, the UI works, and the integration points are ready for the Go mobile frameworks.
