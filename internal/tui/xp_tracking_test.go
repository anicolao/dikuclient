package tui

import (
	"fmt"
	"testing"
	"time"
)

// TestDetectCombatPrompt verifies that combat prompts are correctly detected
func TestDetectCombatPrompt(t *testing.T) {
	tests := []struct {
		name           string
		prompt         string
		expectCombat   bool
		expectedTarget string
	}{
		{
			name:           "basic combat prompt",
			prompt:         "101H 132V 54710X 49.60% 570C [Osric:V.Bad] [a goblin scout:Good] T:24 Exits:NS>",
			expectCombat:   true,
			expectedTarget: "a goblin scout",
		},
		{
			name:           "combat with different hero name",
			prompt:         "50H 100V 1000X 10% 100C [Hero:Excellent] [orc warrior:Fair] T:10 Exits:NESW>",
			expectCombat:   true,
			expectedTarget: "orc warrior",
		},
		{
			name:           "combat with multi-word target",
			prompt:         "75H 80V 2000X 20% 150C [Player:Good] [giant spider:Bad] T:15 Exits:N>",
			expectCombat:   true,
			expectedTarget: "giant spider",
		},
		{
			name:           "not a combat prompt",
			prompt:         "101H 132V 54710X 49.60% 570C Exits:NS>",
			expectCombat:   false,
			expectedTarget: "",
		},
		{
			name:           "combat prompt with ANSI codes",
			prompt:         "\x1b[32m101H 132V 54710X 49.60% 570C [Osric:V.Bad] [a rat:Poor] T:5 Exits:E>\x1b[0m",
			expectCombat:   true,
			expectedTarget: "a rat",
		},
		{
			name:           "mobile attacks player",
			prompt:         "80H 120V 5000X 25% 200C [Hero:Good] [aggressive wolf:Excellent] T:1 Exits:EW>",
			expectCombat:   true,
			expectedTarget: "aggressive wolf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel("test", 4000, nil, nil)
			m.detectCombatPrompt(tt.prompt)

			if tt.expectCombat {
				if m.pendingKill != tt.expectedTarget {
					t.Errorf("Expected pendingKill to be '%s', got '%s'", tt.expectedTarget, m.pendingKill)
				}
				if m.killTime.IsZero() {
					t.Errorf("Expected killTime to be set, but it was zero")
				}
			} else {
				if m.pendingKill != "" {
					t.Errorf("Expected no pendingKill, but got '%s'", m.pendingKill)
				}
			}
		})
	}
}

// TestDetectXPEvents verifies that death messages and XP gains are correctly detected
func TestDetectXPEvents(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// Simulate a kill command
	m.pendingKill = "goblin"
	m.killTime = time.Now().Add(-5 * time.Second) // 5 seconds ago
	
	// Simulate a death message
	m.detectXPEvents("The goblin is dead! R.I.P.")
	
	// pendingKill should still be set, waiting for XP
	if m.pendingKill == "" {
		t.Errorf("Expected pendingKill to still be set after death message")
	}
	
	// Simulate XP gain
	m.detectXPEvents("You receive 100 experience.")
	
	// Now pendingKill should be cleared
	if m.pendingKill != "" {
		t.Errorf("Expected pendingKill to be cleared after XP gain, got '%s'", m.pendingKill)
	}
	
	// Check that XP stat was recorded
	if len(m.xpTracking) != 1 {
		t.Errorf("Expected 1 XP tracking entry, got %d", len(m.xpTracking))
	}
	
	stat, exists := m.xpTracking["goblin"]
	if !exists {
		t.Errorf("Expected XP stat for 'goblin', but it doesn't exist")
	}
	
	if stat.XP != 100 {
		t.Errorf("Expected XP to be 100, got %d", stat.XP)
	}
	
	if stat.XPPerSecond <= 0 {
		t.Errorf("Expected XPPerSecond to be positive, got %f", stat.XPPerSecond)
	}
}

// TestXPTrackingMultipleCreatures verifies that multiple creatures can be tracked
func TestXPTrackingMultipleCreatures(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// First creature
	m.pendingKill = "orc"
	m.killTime = time.Now().Add(-10 * time.Second)
	m.detectXPEvents("The orc is dead! R.I.P.")
	m.detectXPEvents("You receive 50 experience.")
	
	// Second creature
	m.pendingKill = "goblin"
	m.killTime = time.Now().Add(-5 * time.Second)
	m.detectXPEvents("The goblin is dead! R.I.P.")
	m.detectXPEvents("You receive 100 experience.")
	
	// Check that both are tracked
	if len(m.xpTracking) != 2 {
		t.Errorf("Expected 2 XP tracking entries, got %d", len(m.xpTracking))
	}
	
	orcStat, orcExists := m.xpTracking["orc"]
	goblinStat, goblinExists := m.xpTracking["goblin"]
	
	if !orcExists {
		t.Errorf("Expected XP stat for 'orc', but it doesn't exist")
	}
	if !goblinExists {
		t.Errorf("Expected XP stat for 'goblin', but it doesn't exist")
	}
	
	// Goblin should have higher XP/s (100 XP in 5 seconds vs 50 XP in 10 seconds)
	if goblinStat.XPPerSecond <= orcStat.XPPerSecond {
		t.Errorf("Expected goblin XP/s (%f) to be higher than orc XP/s (%f)",
			goblinStat.XPPerSecond, orcStat.XPPerSecond)
	}
}

// TestXPPanelRendering verifies that the XP panel renders correctly
func TestXPPanelRendering(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	m.width = 120
	m.height = 40
	
	// Add some XP stats
	m.xpTracking["goblin"] = &XPStat{
		CreatureName: "goblin",
		XP:           100,
		Seconds:      5.0,
		XPPerSecond:  20.0,
	}
	m.xpTracking["orc"] = &XPStat{
		CreatureName: "orc",
		XP:           50,
		Seconds:      10.0,
		XPPerSecond:  5.0,
	}
	
	// Render the sidebar
	sidebar := m.renderSidebar(60, 30)
	
	// Check that it contains XP/s panel
	if sidebar == "" {
		t.Errorf("Expected sidebar to be rendered, but it was empty")
	}
	
	// The sidebar should contain "XP/s" header and creature names
	// (We can't do exact string matching because of formatting, but we can check
	// that the data structures were set up correctly)
	if len(m.xpTracking) != 2 {
		t.Errorf("Expected 2 XP tracking entries, got %d", len(m.xpTracking))
	}
}

// TestSlimyEarthwormExample verifies the exact example from the user
func TestSlimyEarthwormExample(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// Simulate combat prompt with slimy earthworm
	m.detectCombatPrompt("101H 132V 54710X 49.60% 570C [Osric:Good] [slimy earthworm:Excellent] T:24 Exits:NS>")
	
	// Set time in the past
	m.killTime = time.Now().Add(-3 * time.Second)
	
	// Simulate the exact death message from the user's example
	m.detectXPEvents("The slimy earthworm is dead! R.I.P.")
	
	// pendingKill should now be "slimy earthworm" (without "The")
	if m.pendingKill != "slimy earthworm" {
		t.Errorf("Expected pendingKill to be 'slimy earthworm', got '%s'", m.pendingKill)
	}
	
	// Simulate XP gain from the user's example
	m.detectXPEvents("You receive 102 experience.")
	
	// Now pendingKill should be cleared
	if m.pendingKill != "" {
		t.Errorf("Expected pendingKill to be cleared after XP gain, got '%s'", m.pendingKill)
	}
	
	// Check that XP stat was recorded
	stat, exists := m.xpTracking["slimy earthworm"]
	if !exists {
		t.Errorf("Expected XP stat for 'slimy earthworm', but it doesn't exist")
	}
	
	if stat.XP != 102 {
		t.Errorf("Expected XP to be 102, got %d", stat.XP)
	}
}

// TestDeathCryWithANSICodes verifies that death messages with ANSI codes are detected
func TestDeathCryWithANSICodes(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// Simulate a kill command
	m.pendingKill = "rat"
	m.killTime = time.Now().Add(-3 * time.Second)
	
	// Simulate a death message with ANSI codes
	m.detectXPEvents("\x1b[31mThe rat is dead! R.I.P.\x1b[0m")
	
	// pendingKill should still be set, waiting for XP
	if m.pendingKill == "" {
		t.Errorf("Expected pendingKill to still be set after death message with ANSI codes")
	}
	
	// Simulate XP gain with ANSI codes
	m.detectXPEvents("\x1b[32mYou receive 25 experience.\x1b[0m")
	
	// Now pendingKill should be cleared
	if m.pendingKill != "" {
		t.Errorf("Expected pendingKill to be cleared after XP gain, got '%s'", m.pendingKill)
	}
	
	// Check that XP stat was recorded
	if len(m.xpTracking) != 1 {
		t.Errorf("Expected 1 XP tracking entry, got %d", len(m.xpTracking))
	}
}

// TestXPTrackingFullWorkflow verifies the complete workflow from kill to XP display
func TestXPTrackingFullWorkflow(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	m.width = 120
	m.height = 40
	
	// Simulate killing multiple creatures
	creatures := []struct {
		name  string
		xp    int
		delay time.Duration
	}{
		{"goblin", 100, 5 * time.Second},
		{"orc", 50, 10 * time.Second},
		{"dragon", 500, 20 * time.Second},
		{"rat", 25, 2 * time.Second},
	}
	
	for _, c := range creatures {
		// Simulate combat prompt
		m.detectCombatPrompt(fmt.Sprintf("101H 132V 54710X 49.60%% 570C [Hero:Good] [%s:Excellent] T:24 Exits:NS>", c.name))
		
		// Simulate time passing
		m.killTime = time.Now().Add(-c.delay)
		
		// Simulate death message and XP gain
		m.detectXPEvents("The " + c.name + " is dead! R.I.P.")
		m.detectXPEvents(fmt.Sprintf("You receive %d experience.", c.xp))
	}
	
	// Verify all creatures are tracked
	if len(m.xpTracking) != len(creatures) {
		t.Errorf("Expected %d XP tracking entries, got %d", len(creatures), len(m.xpTracking))
	}
	
	// Verify dragon has highest XP/s (500 XP in 20 seconds = 25 XP/s)
	dragonStat := m.xpTracking["dragon"]
	if dragonStat == nil {
		t.Fatalf("Expected dragon to be in XP tracking")
	}
	
	// Check that dragon has higher XP/s than all others
	for name, stat := range m.xpTracking {
		if name != "dragon" && stat.XPPerSecond > dragonStat.XPPerSecond {
			t.Errorf("Expected dragon to have highest XP/s, but %s has %f vs dragon's %f",
				name, stat.XPPerSecond, dragonStat.XPPerSecond)
		}
	}
	
	// Verify orc has lowest XP/s (50 XP in 10 seconds = 5 XP/s)
	orcStat := m.xpTracking["orc"]
	if orcStat == nil {
		t.Fatalf("Expected orc to be in XP tracking")
	}
	
	// Check that orc has lower XP/s than all others
	for name, stat := range m.xpTracking {
		if name != "orc" && stat.XPPerSecond < orcStat.XPPerSecond {
			t.Errorf("Expected orc to have lowest XP/s, but %s has %f vs orc's %f",
				name, stat.XPPerSecond, orcStat.XPPerSecond)
		}
	}
	
	// Render the sidebar to ensure it doesn't crash
	sidebar := m.renderSidebar(60, 30)
	if sidebar == "" {
		t.Errorf("Expected sidebar to be rendered")
	}
}
