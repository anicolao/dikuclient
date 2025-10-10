# DikuClient for Android

Native Android app for connecting to DikuMUD servers.

## Features

- Native Jetpack Compose interface
- Material Design 3
- Full-screen terminal display
- On-screen keyboard support
- Landscape and portrait orientation
- Works on phones and tablets

## Requirements

- Android 7.0 (API 24) or later
- Android Studio (for development)

## Quick Start

### For Users

1. Install from Play Store or F-Droid (coming soon)
2. Or download APK from GitHub Releases
3. Launch the app
4. Enter your MUD server hostname and port
5. Tap "Connect"

### For Developers

See [MOBILE_BUILD.md](../MOBILE_BUILD.md) for detailed build instructions.

Quick steps:
```bash
# Build debug APK
cd android
./gradlew assembleDebug

# Install on connected device/emulator
./gradlew installDebug

# Or manually install
adb install app/build/outputs/apk/debug/app-debug.apk
```

## Project Structure

```
android/
├── app/
│   ├── build.gradle            # App build configuration
│   ├── src/main/
│   │   ├── AndroidManifest.xml # App manifest
│   │   ├── kotlin/com/dikuclient/
│   │   │   ├── MainActivity.kt      # Main activity (Compose)
│   │   │   └── ClientViewModel.kt   # ViewModel (state)
│   │   └── res/
│   │       └── values/
│   │           ├── strings.xml      # String resources
│   │           └── themes.xml       # App theme
├── build.gradle                # Project build configuration
├── settings.gradle             # Project settings
└── README.md                   # This file
```

## Testing

### On Emulator

1. Create an AVD (Android Virtual Device) in Android Studio
2. Launch the emulator
3. Click Run in Android Studio
4. Test connection form and terminal display

### On Physical Device

1. Enable Developer Options and USB Debugging on your device
2. Connect via USB
3. Verify with `adb devices`
4. Click Run in Android Studio or use `./gradlew installDebug`

## Building Release APK

```bash
cd android

# Build release APK (unsigned)
./gradlew assembleRelease

# APK will be at: app/build/outputs/apk/release/app-release-unsigned.apk

# For signed APK, configure signing in app/build.gradle
```

## Known Limitations (Current Version)

- Go code integration pending (PTY setup required)
- Terminal emulator is basic text display (Termux terminal-view planned)
- No ANSI color support yet (will use terminal emulator library)
- No floating action buttons yet (planned)

## Next Steps

1. Integrate Go mobile AAR library (`dikuclient.aar`)
2. Add PTY (pseudo-terminal) support via JNI
3. Integrate Termux terminal-view library for full terminal emulation
4. Add floating buttons for quick commands
5. Internal testing track on Play Store
6. Public release on Play Store and F-Droid

## Contributing

See main repository README for contribution guidelines.

## License

See [LICENSE](../LICENSE) in the root directory.
