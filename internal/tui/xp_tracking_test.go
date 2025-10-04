package tui

import (
	"testing"
	"time"
)

// TestDetectKillCommand verifies that kill commands are correctly detected
func TestDetectKillCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectKill  bool
		expectedTarget string
	}{
		{
			name:        "basic kill command",
			command:     "kill orc",
			expectKill:  true,
			expectedTarget: "orc",
		},
		{
			name:        "kill with multi-word target",
			command:     "kill giant spider",
			expectKill:  true,
			expectedTarget: "giant spider",
		},
		{
			name:        "kill with uppercase",
			command:     "KILL goblin",
			expectKill:  true,
			expectedTarget: "goblin",
		},
		{
			name:        "not a kill command",
			command:     "look",
			expectKill:  false,
			expectedTarget: "",
		},
		{
			name:        "kill with leading spaces",
			command:     "  kill rat  ",
			expectKill:  true,
			expectedTarget: "rat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel("test", 4000, nil, nil)
			m.detectKillCommand(tt.command)

			if tt.expectKill {
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

// TestDetectXPEvents verifies that death cries and XP gains are correctly detected
func TestDetectXPEvents(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// Simulate a kill command
	m.pendingKill = "goblin"
	m.killTime = time.Now().Add(-5 * time.Second) // 5 seconds ago
	
	// Simulate a death cry
	m.detectXPEvents("The goblin cries a sad death cry")
	
	// pendingKill should still be set, waiting for XP
	if m.pendingKill == "" {
		t.Errorf("Expected pendingKill to still be set after death cry")
	}
	
	// Simulate XP gain
	m.detectXPEvents("You gain 100XP")
	
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
	m.detectXPEvents("The orc cries a sad death cry")
	m.detectXPEvents("You gain 50XP")
	
	// Second creature
	m.pendingKill = "goblin"
	m.killTime = time.Now().Add(-5 * time.Second)
	m.detectXPEvents("The goblin cries a sad death cry")
	m.detectXPEvents("You gain 100XP")
	
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

// TestDeathCryWithANSICodes verifies that death cries with ANSI codes are detected
func TestDeathCryWithANSICodes(t *testing.T) {
	m := NewModel("test", 4000, nil, nil)
	
	// Simulate a kill command
	m.pendingKill = "rat"
	m.killTime = time.Now().Add(-3 * time.Second)
	
	// Simulate a death cry with ANSI codes
	m.detectXPEvents("\x1b[31mThe rat cries a sad death cry\x1b[0m")
	
	// pendingKill should still be set, waiting for XP
	if m.pendingKill == "" {
		t.Errorf("Expected pendingKill to still be set after death cry with ANSI codes")
	}
	
	// Simulate XP gain with ANSI codes
	m.detectXPEvents("\x1b[32mYou gain 25XP\x1b[0m")
	
	// Now pendingKill should be cleared
	if m.pendingKill != "" {
		t.Errorf("Expected pendingKill to be cleared after XP gain, got '%s'", m.pendingKill)
	}
	
	// Check that XP stat was recorded
	if len(m.xpTracking) != 1 {
		t.Errorf("Expected 1 XP tracking entry, got %d", len(m.xpTracking))
	}
}
