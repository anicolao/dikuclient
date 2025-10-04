package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
)

// TestInventoryIntegration simulates receiving inventory output from a MUD server
func TestInventoryIntegration(t *testing.T) {
	// Create a model
	m := Model{
		output:       []string{},
		recentOutput: []string{},
	}

	// Simulate receiving the inventory command output line by line
	mudOutput := []string{
		"86H 109V 7563X 0.00% 79C T:3 Exits:D> i",
		"You are carrying:",
		"a sharp short sword",
		"a glowing scroll of recall..it glows dimly",
		"a torch [4]",
		"a rusty knife",
		"a bowl of Otik's spiced potatoes [2]",
		"an entire loaf of bread [4]",
		"",
		"86H 109V 7563X 0.00% 79C T:2 Exits:D>",
	}

	// Process each line as if coming from the MUD
	for _, line := range mudOutput {
		m.output = append(m.output, line)
		m.recentOutput = append(m.recentOutput, line)

		// Keep recentOutput to last 30 lines
		if len(m.recentOutput) > 30 {
			m.recentOutput = m.recentOutput[len(m.recentOutput)-30:]
		}
	}

	// Now detect inventory
	m.detectAndUpdateInventory()

	// Verify inventory was detected
	if len(m.inventory) != 6 {
		t.Errorf("Expected 6 items in inventory, got %d", len(m.inventory))
	}

	expectedItems := []string{
		"a sharp short sword",
		"a glowing scroll of recall..it glows dimly",
		"a torch [4]",
		"a rusty knife",
		"a bowl of Otik's spiced potatoes [2]",
		"an entire loaf of bread [4]",
	}

	for i, expected := range expectedItems {
		if i >= len(m.inventory) {
			t.Errorf("Missing item %d: %q", i, expected)
			continue
		}
		if m.inventory[i] != expected {
			t.Errorf("Item %d: expected %q, got %q", i, expected, m.inventory[i])
		}
	}

	// Verify timestamp was set
	if m.inventoryTime.IsZero() {
		t.Error("Inventory timestamp was not set")
	}

	if time.Since(m.inventoryTime) > time.Second {
		t.Error("Inventory timestamp is not recent")
	}
}

// TestInventoryRenderingWithItems tests that inventory panel renders correctly with items
func TestInventoryRenderingWithItems(t *testing.T) {
	m := Model{
		inventory: []string{
			"a sharp short sword",
			"a torch [4]",
			"a rusty knife",
		},
		inventoryTime: time.Now(),
		width:         100,
		height:        40,
		sidebarWidth:  30,
	}

	// Initialize inventory viewport
	m.inventoryViewport = viewport.New(m.sidebarWidth-4, 10)

	// Render the sidebar
	result := m.renderSidebar(m.sidebarWidth, m.height-10)

	// Check that the result contains inventory items
	if !strings.Contains(result, "Inventory") {
		t.Error("Sidebar should contain 'Inventory' header")
	}

	// Check that items are present
	for _, item := range m.inventory {
		if !strings.Contains(result, item) {
			t.Errorf("Sidebar should contain item: %q", item)
		}
	}

	// Check that timestamp is present
	timeStr := m.inventoryTime.Format("15:04:05")
	if !strings.Contains(result, timeStr) {
		t.Errorf("Sidebar should contain timestamp: %q", timeStr)
	}
}

// TestInventoryRenderingWithoutItems tests that inventory panel renders correctly without items
func TestInventoryRenderingWithoutItems(t *testing.T) {
	m := Model{
		inventory:    []string{},
		width:        100,
		height:       40,
		sidebarWidth: 30,
	}

	// Initialize inventory viewport
	m.inventoryViewport = viewport.New(m.sidebarWidth-4, 10)

	// Render the sidebar
	result := m.renderSidebar(m.sidebarWidth, m.height-10)

	// Check that the result contains inventory header
	if !strings.Contains(result, "Inventory") {
		t.Error("Sidebar should contain 'Inventory' header")
	}

	// Check that it shows "not populated"
	if !strings.Contains(result, "not populated") {
		t.Error("Sidebar should contain 'not populated' message")
	}
}
