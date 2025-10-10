# DikuClient for iOS

Native iOS app for connecting to DikuMUD servers.

## Features

- Native SwiftUI interface
- Full-screen terminal display
- On-screen keyboard support
- Landscape and portrait orientation
- Works on iPhone and iPad

## Requirements

- iOS 15.0 or later
- Xcode 15 or later (for development)

## Quick Start

### For Users

1. Install from TestFlight (beta) or App Store (coming soon)
2. Launch the app
3. Enter your MUD server hostname and port
4. Tap "Connect"

### For Developers

See [MOBILE_BUILD.md](../MOBILE_BUILD.md) for detailed build instructions.

Quick steps:
```bash
# Open in Xcode
open DikuClient.xcodeproj

# Or build from command line
xcodebuild -project DikuClient.xcodeproj -scheme DikuClient -destination 'platform=iOS Simulator,name=iPhone 15' build
```

## Project Structure

```
ios/
├── DikuClient.xcodeproj/     # Xcode project
├── DikuClient/               # Swift source code
│   ├── DikuClientApp.swift   # App entry point
│   ├── ContentView.swift     # Main view (connection/terminal)
│   ├── ClientViewModel.swift # ViewModel (state management)
│   ├── TerminalView.swift    # Terminal display view
│   └── Info.plist           # App configuration
└── README.md                 # This file
```

## Testing

### On Simulator

1. Select a simulator device in Xcode
2. Press `Cmd+R` to build and run
3. Test connection form and terminal display

### On Physical Device

1. Connect your iPhone/iPad via USB
2. Select your device in Xcode
3. Configure code signing with your Apple ID
4. Build and run

## Known Limitations (Current Version)

- Go code integration pending (PTY setup required)
- Terminal emulator is basic text display (SwiftTerm integration planned)
- No ANSI color support yet (will use SwiftTerm)
- No floating action buttons yet (planned)

## Next Steps

1. Integrate Go mobile framework (`Dikuclient.xcframework`)
2. Add PTY (pseudo-terminal) support for Go TUI
3. Integrate SwiftTerm for full terminal emulation
4. Add floating buttons for quick commands
5. TestFlight beta testing
6. App Store submission

## Contributing

See main repository README for contribution guidelines.

## License

See [LICENSE](../LICENSE) in the root directory.
