package ticktimer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager(75)
	if m.TickInterval != 75 {
		t.Errorf("Expected tick interval 75, got %d", m.TickInterval)
	}
	if len(m.TickTriggers) != 0 {
		t.Errorf("Expected empty tick triggers, got %d", len(m.TickTriggers))
	}
}

func TestUpdateFromPrompt(t *testing.T) {
	m := NewManager(75)
	m.UpdateFromPrompt(24)

	if m.LastSeenValue != 24 {
		t.Errorf("Expected last seen value 24, got %d", m.LastSeenValue)
	}

	if m.LastUpdateTime.IsZero() {
		t.Error("Expected last update time to be set")
	}

	// Check that last tick time was calculated correctly
	// If T:24 with interval 75, tick was 75-24=51 seconds ago
	expectedTickTime := time.Now().Add(-51 * time.Second)
	diff := m.LastTickTime.Sub(expectedTickTime).Seconds()
	if diff > 1 || diff < -1 {
		t.Errorf("Expected last tick time around 51 seconds ago, got diff of %.2f", diff)
	}
}

func TestGetCurrentTickTime(t *testing.T) {
	m := NewManager(75)
	m.UpdateFromPrompt(24)

	// Immediately after update, should be close to 24
	currentTick := m.GetCurrentTickTime()
	if currentTick < 23 || currentTick > 25 {
		t.Errorf("Expected current tick time around 24, got %d", currentTick)
	}

	// Wait a second and check again
	time.Sleep(1 * time.Second)
	currentTick = m.GetCurrentTickTime()
	if currentTick < 22 || currentTick > 24 {
		t.Errorf("Expected current tick time around 23 after 1 second, got %d", currentTick)
	}
}

func TestAddTrigger(t *testing.T) {
	m := NewManager(75)

	err := m.AddTrigger(5, "cast 'heal'")
	if err != nil {
		t.Errorf("Failed to add trigger: %v", err)
	}

	if len(m.TickTriggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(m.TickTriggers))
	}

	if m.TickTriggers[0].TickTime != 5 {
		t.Errorf("Expected tick time 5, got %d", m.TickTriggers[0].TickTime)
	}

	if m.TickTriggers[0].Commands != "cast 'heal'" {
		t.Errorf("Expected commands 'cast 'heal'', got %s", m.TickTriggers[0].Commands)
	}

	// Test invalid tick time
	err = m.AddTrigger(100, "invalid")
	if err == nil {
		t.Error("Expected error for tick time > interval")
	}

	err = m.AddTrigger(-1, "invalid")
	if err == nil {
		t.Error("Expected error for negative tick time")
	}
}

func TestRemoveTrigger(t *testing.T) {
	m := NewManager(75)
	m.AddTrigger(5, "cast 'heal'")
	m.AddTrigger(10, "cast 'bless'")

	err := m.RemoveTrigger(0)
	if err != nil {
		t.Errorf("Failed to remove trigger: %v", err)
	}

	if len(m.TickTriggers) != 1 {
		t.Errorf("Expected 1 trigger after removal, got %d", len(m.TickTriggers))
	}

	if m.TickTriggers[0].TickTime != 10 {
		t.Errorf("Expected remaining trigger at tick time 10, got %d", m.TickTriggers[0].TickTime)
	}

	// Test invalid index
	err = m.RemoveTrigger(5)
	if err == nil {
		t.Error("Expected error for invalid index")
	}
}

func TestGetTriggersToFire(t *testing.T) {
	m := NewManager(75)
	m.AddTrigger(5, "cast 'heal'")
	m.AddTrigger(5, "say Healing!")
	m.AddTrigger(10, "cast 'bless'")

	// Set up tick time to be around 5
	m.UpdateFromPrompt(5)

	// Should fire triggers at tick time 5
	commands := m.GetTriggersToFire(0)
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands to fire, got %d", len(commands))
	}

	// Should not fire again with same lastFiredTickTime
	commands = m.GetTriggersToFire(5)
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands when lastFiredTickTime matches, got %d", len(commands))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	os.Setenv("DIKUCLIENT_CONFIG_DIR", tempDir)
	defer os.Unsetenv("DIKUCLIENT_CONFIG_DIR")

	// Create a new manager
	m := NewManager(75)
	m.filePath = filepath.Join(tempDir, "ticktimer_test_4000.json")
	m.UpdateFromPrompt(24)
	m.AddTrigger(5, "cast 'heal'")
	m.AddTrigger(10, "cast 'bless'")

	// Save
	err := m.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load
	m2, err := Load("test", 4000, 75)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if m2.TickInterval != 75 {
		t.Errorf("Expected tick interval 75, got %d", m2.TickInterval)
	}

	if m2.LastSeenValue != 24 {
		t.Errorf("Expected last seen value 24, got %d", m2.LastSeenValue)
	}

	if len(m2.TickTriggers) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(m2.TickTriggers))
	}
}
