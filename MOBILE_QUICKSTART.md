# Mobile Quick Start Guide

Quick reference for building and testing the DikuClient mobile apps.

## Prerequisites

```bash
# Install Go Mobile (required)
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# For iOS: Install Xcode from App Store
# For Android: Install Android Studio
```

## Build Commands

### Build Go Mobile Frameworks

```bash
# Build for both platforms
./scripts/build-mobile.sh all

# Or build individually
./scripts/build-mobile.sh ios      # Creates ios/Frameworks/Dikuclient.xcframework
./scripts/build-mobile.sh android  # Creates android/app/libs/dikuclient.aar
```

### Build iOS App

```bash
# Method 1: Xcode (recommended)
open ios/DikuClient.xcodeproj
# Press Cmd+R to build and run

# Method 2: Command line
xcodebuild -project ios/DikuClient.xcodeproj \
  -scheme DikuClient \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  build
```

### Build Android App

```bash
# Method 1: Android Studio (recommended)
# Open android/ folder in Android Studio
# Click green Run button

# Method 2: Command line
cd android
./gradlew assembleDebug  # Creates APK
./gradlew installDebug   # Builds and installs on device
```

## Test Commands

```bash
# Test on iOS simulator
./scripts/test-mobile.sh ios

# Test on Android emulator/device
./scripts/test-mobile.sh android

# Test both
./scripts/test-mobile.sh all
```

## Common Tasks

### Start iOS Simulator

```bash
# List available simulators
xcrun simctl list devices

# Boot a simulator
xcrun simctl boot "iPhone 15"

# Open Simulator app
open -a Simulator
```

### Start Android Emulator

```bash
# List available emulators
emulator -list-avds

# Start an emulator
emulator -avd Pixel_5_API_34 &

# Or use Android Studio: Tools → Device Manager
```

### Install on Device

```bash
# iOS: Connect device, open Xcode, select device, press Cmd+R

# Android: Enable USB debugging, then:
cd android && ./gradlew installDebug
```

## Project Structure

```
dikuclient/
├── mobile/              # Go package (gomobile compatible)
├── ios/                 # iOS app (SwiftUI)
├── android/             # Android app (Jetpack Compose)
├── scripts/             # Build & test scripts
├── MOBILE_BUILD.md      # Detailed build instructions
└── MOBILE_QUICKSTART.md # This file
```

## File Locations

### iOS
- **Project**: `ios/DikuClient.xcodeproj`
- **Source**: `ios/DikuClient/*.swift`
- **Framework**: `ios/Frameworks/Dikuclient.xcframework` (after build)

### Android
- **Project**: `android/` (open this folder)
- **Source**: `android/app/src/main/kotlin/com/dikuclient/*.kt`
- **Library**: `android/app/libs/dikuclient.aar` (after build)

### Go Mobile
- **Source**: `mobile/*.go`
- **Build Output**: Platform-specific frameworks/AARs

## Troubleshooting

### gomobile not found
```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

### iOS build errors
- Ensure Xcode is installed and command line tools are set up
- Open Xcode Preferences → Accounts → Add Apple ID

### Android build errors
- Ensure Android Studio is installed
- Set ANDROID_HOME environment variable
- Install Android SDK and build tools

### No emulator/simulator
- **iOS**: Install iOS simulators in Xcode Preferences → Components
- **Android**: Create AVD in Android Studio Device Manager

## Getting Help

See detailed documentation:
- [MOBILE_BUILD.md](./MOBILE_BUILD.md) - Complete build guide
- [ios/README.md](./ios/README.md) - iOS-specific info
- [android/README.md](./android/README.md) - Android-specific info
- [MOBILE_DESIGN.md](./MOBILE_DESIGN.md) - Architecture and design

## Quick Links

- [Go Mobile Documentation](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)
- [Jetpack Compose](https://developer.android.com/jetpack/compose)
