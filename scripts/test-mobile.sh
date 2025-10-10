#!/bin/bash
# Test script for mobile platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}DikuClient Mobile Test Script${NC}"
echo "=============================="
echo ""

PLATFORM="${1:-all}"

test_ios() {
    echo -e "${GREEN}Testing iOS app on simulator...${NC}"
    
    if ! command -v xcodebuild &> /dev/null; then
        echo -e "${RED}Error: xcodebuild not found. Is Xcode installed?${NC}"
        exit 1
    fi
    
    # Check if project exists
    if [ ! -f "ios/DikuClient.xcodeproj/project.pbxproj" ]; then
        echo -e "${RED}Error: iOS project not found${NC}"
        exit 1
    fi
    
    # Build for simulator
    echo "Building for iOS Simulator..."
    xcodebuild -project ios/DikuClient.xcodeproj \
        -scheme DikuClient \
        -destination 'platform=iOS Simulator,name=iPhone 15' \
        build
    
    echo -e "${GREEN}✓ iOS build successful${NC}"
    echo ""
    echo "To run on simulator:"
    echo "  1. Open Simulator app"
    echo "  2. Run: open ios/DikuClient.xcodeproj"
    echo "  3. Press Cmd+R in Xcode"
}

test_android() {
    echo -e "${GREEN}Testing Android app...${NC}"
    
    if [ ! -f "android/gradlew" ]; then
        echo -e "${RED}Error: Android project not found${NC}"
        exit 1
    fi
    
    cd android
    
    # Check if emulator is running
    ADB_DEVICES=$(adb devices 2>/dev/null | grep -v "List" | grep "device$" | wc -l)
    
    if [ "$ADB_DEVICES" -eq 0 ]; then
        echo -e "${YELLOW}Warning: No Android device/emulator detected${NC}"
        echo "Building debug APK only..."
        ./gradlew assembleDebug
        echo -e "${GREEN}✓ Debug APK built: android/app/build/outputs/apk/debug/app-debug.apk${NC}"
    else
        echo "Building and installing on device/emulator..."
        ./gradlew installDebug
        echo -e "${GREEN}✓ App installed on device${NC}"
        
        echo ""
        echo "Starting app..."
        adb shell am start -n com.dikuclient/.MainActivity
    fi
    
    cd ..
}

case "$PLATFORM" in
    ios)
        test_ios
        ;;
    android)
        test_android
        ;;
    all)
        test_ios
        echo ""
        test_android
        ;;
    *)
        echo -e "${RED}Error: Unknown platform '$PLATFORM'${NC}"
        echo "Usage: $0 [ios|android|all]"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Testing complete!${NC}"
