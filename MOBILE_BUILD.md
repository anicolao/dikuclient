# Mobile Build Instructions

This document describes how to build and test the DikuClient mobile apps for iOS and Android.

## Overview

The mobile implementation consists of:
- **Go Mobile Package** (`mobile/`): Go code that can be called from iOS/Android
- **iOS App** (`ios/`): Native SwiftUI app for iPhone/iPad
- **Android App** (`android/`): Native Kotlin app with Jetpack Compose

## Prerequisites

### For Go Mobile Bindings

```bash
# Install Go 1.24 or later
# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

### For iOS Development

- macOS with Xcode 15 or later
- iOS Simulator or physical iOS device
- Apple Developer account (for device testing)

### For Android Development

- Android Studio (latest version recommended)
- Android SDK with API 24+ (Android 7.0+)
- Android Emulator or physical Android device

## Building the Go Mobile Library

The Go mobile package provides bindings that can be called from native code.

### For iOS

```bash
cd /path/to/dikuclient

# Build iOS framework
gomobile bind -target=ios -o ios/Dikuclient.xcframework github.com/anicolao/dikuclient/mobile

# The framework will be created at ios/Dikuclient.xcframework
```

### For Android

```bash
cd /path/to/dikuclient

# Build Android AAR library
gomobile bind -target=android -o android/app/libs/dikuclient.aar github.com/anicolao/dikuclient/mobile

# The AAR will be created at android/app/libs/dikuclient.aar
```

## Building and Running the iOS App

### Using Xcode

1. Open the Xcode project:
   ```bash
   open ios/DikuClient.xcodeproj
   ```

2. Select a simulator or connected device from the device menu

3. Build and run:
   - Press `Cmd+R` or click the Play button
   - Or use: Product → Run

### Using Command Line

```bash
# List available simulators
xcrun simctl list devices

# Build for simulator
xcodebuild -project ios/DikuClient.xcodeproj \
  -scheme DikuClient \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  build

# Run on simulator (requires simulator to be booted)
xcrun simctl boot "iPhone 15"
xcrun simctl install booted /path/to/DikuClient.app
xcrun simctl launch booted com.dikuclient.ios
```

## Building and Running the Android App

### Using Android Studio

1. Open Android Studio

2. Select "Open" and navigate to the `android/` directory

3. Wait for Gradle sync to complete

4. Select an emulator or connected device from the device dropdown

5. Click the Run button (green play icon) or press `Shift+F10`

### Using Command Line

```bash
cd android

# Build debug APK
./gradlew assembleDebug

# The APK will be at: app/build/outputs/apk/debug/app-debug.apk

# Install on connected device or running emulator
./gradlew installDebug

# Or manually install
adb install app/build/outputs/apk/debug/app-debug.apk

# Launch the app
adb shell am start -n com.dikuclient/.MainActivity
```

## Testing on Emulators

### iOS Simulator

The iOS Simulator comes with Xcode and provides a fast way to test:

```bash
# Open Simulator
open -a Simulator

# Or launch specific simulator
xcrun simctl boot "iPhone 15"
open -a Simulator
```

Features tested in simulator:
- ✅ UI layout and navigation
- ✅ Text input and display
- ✅ Network connectivity (with localhost)
- ⚠️ Performance (slower than real device)
- ❌ Touch gestures (limited)

### Android Emulator

Create and run an Android emulator:

```bash
# List available AVDs (Android Virtual Devices)
emulator -list-avds

# Launch emulator
emulator -avd Pixel_5_API_34 &

# Or create new AVD in Android Studio:
# Tools → Device Manager → Create Device
```

Features tested in emulator:
- ✅ UI layout and navigation
- ✅ Text input and display  
- ✅ Network connectivity
- ⚠️ Performance (depends on host system)
- ⚠️ Touch gestures (mouse simulation)

## Testing on Real Devices

### iOS Physical Device

Requirements:
- Apple Developer account (free or paid)
- Device connected via USB or WiFi

Steps:
1. Connect your iOS device to your Mac
2. In Xcode, select your device from the device menu
3. Go to Signing & Capabilities tab
4. Select your team under Signing
5. Xcode will automatically provision your device
6. Build and run (`Cmd+R`)

### Android Physical Device

Requirements:
- Android device with Developer Options enabled
- USB debugging enabled

Steps:
1. Enable Developer Options:
   - Go to Settings → About Phone
   - Tap "Build Number" 7 times

2. Enable USB Debugging:
   - Go to Settings → Developer Options
   - Enable "USB debugging"

3. Connect device via USB

4. Verify connection:
   ```bash
   adb devices
   ```

5. Build and install:
   ```bash
   cd android
   ./gradlew installDebug
   ```

## Current Implementation Status

### ✅ Completed

- Go mobile package with minimal API
- iOS app structure with SwiftUI
  - Connection form UI
  - Terminal display view
  - Basic navigation
- Android app structure with Jetpack Compose
  - Connection form UI
  - Terminal display view
  - Basic navigation

### 🚧 Integration Required

To fully integrate the Go code, you need to:

1. **Build Go Mobile Framework/AAR** (see instructions above)

2. **Add to iOS Project**:
   - Drag `Dikuclient.xcframework` into Xcode project
   - Add to "Frameworks, Libraries, and Embedded Content"
   - Import in Swift: `import Dikuclient`
   - Uncomment Go function calls in `ClientViewModel.swift`

3. **Add to Android Project**:
   - Place `dikuclient.aar` in `android/app/libs/`
   - Add to `app/build.gradle`:
     ```gradle
     dependencies {
         implementation files('libs/dikuclient.aar')
     }
     ```
   - Import in Kotlin and call Go functions

4. **PTY Integration**:
   - iOS: Use `openpty()` to create pseudo-terminal
   - Android: Use JNI or Android APIs to create PTY
   - Pass PTY file descriptors to Go code

## Troubleshooting

### iOS Build Errors

**Error**: "Developer cannot be verified"
- Solution: Open Xcode preferences → Accounts → Add your Apple ID

**Error**: "No provisioning profile found"
- Solution: Select your team in Signing & Capabilities

**Error**: "Simulator not found"
- Solution: Install iOS simulators in Xcode preferences → Components

### Android Build Errors

**Error**: "SDK location not found"
- Solution: Create `local.properties` with: `sdk.dir=/path/to/Android/sdk`

**Error**: "Gradle sync failed"
- Solution: Ensure Android SDK and build tools are installed

**Error**: "ADB not found"
- Solution: Add Android SDK platform-tools to PATH

### Go Mobile Build Errors

**Error**: "go: could not create module cache: mkdir /nix/store/.../pkg: permission denied" (Nix users)
- Solution: The build script now automatically sets `GOMODCACHE` and `GOCACHE` to writable directories
- If you still encounter issues, manually set these environment variables:
  ```bash
  export GOMODCACHE="$HOME/go/pkg/mod"
  export GOCACHE="$HOME/.cache/go-build"
  ./scripts/build-mobile.sh all
  ```

## Next Steps

1. **Complete Go Mobile Integration**:
   - Build and link the Go mobile framework/AAR
   - Test Go function calls from native code

2. **Add PTY Support**:
   - Implement pseudo-terminal creation
   - Connect PTY to Go TUI code

3. **Enhanced Terminal Emulator**:
   - iOS: Integrate SwiftTerm for full ANSI support
   - Android: Use Termux terminal-view library

4. **Testing**:
   - Test on multiple iOS versions (15+)
   - Test on multiple Android versions (7+)
   - Test various screen sizes
   - Performance profiling

5. **Distribution**:
   - iOS: TestFlight beta → App Store
   - Android: Internal testing → Play Store or F-Droid

## Resources

- [Go Mobile Documentation](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)
- [Jetpack Compose Documentation](https://developer.android.com/jetpack/compose)
- [iOS Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)
- [Android Material Design](https://m3.material.io/)
