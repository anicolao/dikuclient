#!/bin/bash
# Build script for mobile platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}DikuClient Mobile Build Script${NC}"
echo "================================"
echo ""

# Check if gomobile is installed
if ! command -v gomobile &> /dev/null; then
    echo -e "${YELLOW}gomobile not found. Installing...${NC}"
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
fi

# Parse command line arguments
PLATFORM="${1:-all}"
BUILD_TYPE="${2:-debug}"

build_ios() {
    echo -e "${GREEN}Building iOS framework...${NC}"
    
    # Create output directory
    mkdir -p ios/Frameworks
    
    # Build iOS framework
    gomobile bind -target=ios -o ios/Frameworks/Dikuclient.xcframework github.com/anicolao/dikuclient/mobile
    
    echo -e "${GREEN}✓ iOS framework built: ios/Frameworks/Dikuclient.xcframework${NC}"
}

build_android() {
    echo -e "${GREEN}Building Android AAR...${NC}"
    
    # Create output directory
    mkdir -p android/app/libs
    
    # Build Android AAR
    gomobile bind -target=android -o android/app/libs/dikuclient.aar github.com/anicolao/dikuclient/mobile
    
    echo -e "${GREEN}✓ Android AAR built: android/app/libs/dikuclient.aar${NC}"
}

case "$PLATFORM" in
    ios)
        build_ios
        ;;
    android)
        build_android
        ;;
    all)
        build_ios
        build_android
        ;;
    *)
        echo -e "${RED}Error: Unknown platform '$PLATFORM'${NC}"
        echo "Usage: $0 [ios|android|all] [debug|release]"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo ""
echo "Next steps:"
echo "  - iOS: Open ios/DikuClient.xcodeproj in Xcode"
echo "  - Android: Open android/ in Android Studio"
echo ""
echo "See MOBILE_BUILD.md for detailed instructions."
