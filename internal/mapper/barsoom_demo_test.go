package mapper

import (
	"fmt"
	"strings"
	"testing"
)

func TestBarsoomDemo(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("BARSOOM ROOM FORMAT DETECTION DEMO")
	fmt.Println(strings.Repeat("=", 70))
	
	// Test 1: Basic Barsoom room (new format: exits on >-- line)
	lines1 := []string{
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"--<",
		"Temple Square",
		"    You are standing in a large temple square. The ancient stones",
		"speak of a glorious past.",
		">-- Exits:NSE",
	}
	
	fmt.Println("\nTest 1: Basic Barsoom Room")
	fmt.Println("Input lines:")
	for i, line := range lines1 {
		fmt.Printf("  %d: %q\n", i, line)
	}
	fmt.Println()
	
	info1 := ParseRoomInfo(lines1, false)
	if info1 != nil {
		fmt.Printf("✓ Parsed successfully!\n")
		fmt.Printf("  Title: %s\n", info1.Title)
		fmt.Printf("  Description: %s\n", info1.Description)
		fmt.Printf("  Exits: %v\n", info1.Exits)
		fmt.Printf("  IsBarsoomRoom: %v\n", info1.IsBarsoomRoom)
		fmt.Printf("  Markers suppressed: lines %d and %d\n", info1.BarsoomStartIdx, info1.BarsoomEndIdx)
	} else {
		fmt.Println("✗ Failed to parse")
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 70))
	
	// Test 2: Regular room (non-Barsoom)
	lines2 := []string{
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"Temple Square",
		"    You are standing in a large temple square. The ancient stones",
		"speak of a glorious past.",
		"Exits: north, south, east",
	}
	
	fmt.Println("\nTest 2: Regular Room (non-Barsoom)")
	fmt.Println("Input lines:")
	for i, line := range lines2 {
		fmt.Printf("  %d: %q\n", i, line)
	}
	fmt.Println()
	
	info2 := ParseRoomInfo(lines2, false)
	if info2 != nil {
		fmt.Printf("✓ Parsed successfully!\n")
		fmt.Printf("  Title: %s\n", info2.Title)
		fmt.Printf("  Description: %s\n", info2.Description)
		fmt.Printf("  Exits: %v\n", info2.Exits)
		fmt.Printf("  IsBarsoomRoom: %v (no markers to suppress)\n", info2.IsBarsoomRoom)
	} else {
		fmt.Println("✗ Failed to parse")
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 70))
	
	// Test 3: Barsoom room with multiple paragraphs (new format: exits on >-- line)
	lines3 := []string{
		"119H 110V 3674X 0.00% 77C T:56 Exits:EW>",
		"--<",
		"Ancient Library",
		"Towering shelves filled with ancient tomes line the walls of this grand library.",
		"The musty smell of old parchment fills the air.",
		"",
		"A large reading table sits in the center of the room, covered with open books.",
		">-- Exits:W",
	}
	
	fmt.Println("\nTest 3: Barsoom Room with Multiple Paragraphs")
	fmt.Println("Input lines:")
	for i, line := range lines3 {
		fmt.Printf("  %d: %q\n", i, line)
	}
	fmt.Println()
	
	info3 := ParseRoomInfo(lines3, false)
	if info3 != nil {
		fmt.Printf("✓ Parsed successfully!\n")
		fmt.Printf("  Title: %s\n", info3.Title)
		fmt.Printf("  Description: %s\n", info3.Description)
		fmt.Printf("  Exits: %v\n", info3.Exits)
		fmt.Printf("  IsBarsoomRoom: %v\n", info3.IsBarsoomRoom)
		fmt.Printf("  Markers suppressed: lines %d and %d\n", info3.BarsoomStartIdx, info3.BarsoomEndIdx)
	} else {
		fmt.Println("✗ Failed to parse")
	}
	
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("SUMMARY:")
	fmt.Println("  • Barsoom rooms use --< and >-- markers to bracket descriptions")
	fmt.Println("  • These markers are detected and suppressed from display")
	fmt.Println("  • Room description is formatted and shown in a sticky top split")
	fmt.Println("  • Regular rooms continue to work normally without the split")
	fmt.Println(strings.Repeat("=", 70))
}
