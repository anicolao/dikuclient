package xpstats

import (
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m.Stats == nil {
		t.Error("Expected Stats map to be initialized")
	}
	if len(m.Stats) != 0 {
		t.Errorf("Expected empty Stats map, got %d items", len(m.Stats))
	}
}

func TestUpdateStat(t *testing.T) {
	m := NewManager()

	// First update - should store directly
	m.UpdateStat("goblin", 20.0)
	
	stat, exists := m.GetStat("goblin")
	if !exists {
		t.Fatal("Expected stat for 'goblin' to exist")
	}
	
	if stat.XPPerSecond != 20.0 {
		t.Errorf("Expected XPPerSecond to be 20.0, got %f", stat.XPPerSecond)
	}
	
	if stat.SampleCount != 1 {
		t.Errorf("Expected SampleCount to be 1, got %d", stat.SampleCount)
	}

	// Second update - should use EMA
	m.UpdateStat("goblin", 30.0)
	
	stat, _ = m.GetStat("goblin")
	
	// EMA with alpha=0.25: 0.25*30 + 0.75*20 = 7.5 + 15 = 22.5
	expected := 22.5
	if stat.XPPerSecond != expected {
		t.Errorf("Expected XPPerSecond to be %f, got %f", expected, stat.XPPerSecond)
	}
	
	if stat.SampleCount != 2 {
		t.Errorf("Expected SampleCount to be 2, got %d", stat.SampleCount)
	}
}

func TestUpdateStatMultipleSamples(t *testing.T) {
	m := NewManager()

	// Simulate multiple kills with varying XP/s
	samples := []float64{20.0, 25.0, 22.0, 30.0, 28.0}
	
	for _, sample := range samples {
		m.UpdateStat("orc", sample)
	}
	
	stat, exists := m.GetStat("orc")
	if !exists {
		t.Fatal("Expected stat for 'orc' to exist")
	}
	
	if stat.SampleCount != len(samples) {
		t.Errorf("Expected SampleCount to be %d, got %d", len(samples), stat.SampleCount)
	}
	
	// EMA should be somewhere in the range, weighted toward recent values
	if stat.XPPerSecond < 20.0 || stat.XPPerSecond > 30.0 {
		t.Errorf("Expected XPPerSecond to be in reasonable range, got %f", stat.XPPerSecond)
	}
}

func TestPersistence(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	xpsPath := filepath.Join(tmpDir, "xps.json")

	// Create and populate manager
	m1 := NewManager()
	m1.filePath = xpsPath
	m1.UpdateStat("goblin", 20.0)
	m1.UpdateStat("orc", 15.0)
	m1.UpdateStat("goblin", 25.0)

	// Save to disk
	if err := m1.Save(); err != nil {
		t.Fatalf("Failed to save XP stats: %v", err)
	}

	// Load from disk
	m2, err := LoadFromPath(xpsPath)
	if err != nil {
		t.Fatalf("Failed to load XP stats: %v", err)
	}

	// Verify loaded data
	if len(m2.Stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(m2.Stats))
	}

	goblinStat, exists := m2.GetStat("goblin")
	if !exists {
		t.Fatal("Expected stat for 'goblin' to exist after loading")
	}

	if goblinStat.SampleCount != 2 {
		t.Errorf("Expected SampleCount to be 2, got %d", goblinStat.SampleCount)
	}

	orcStat, exists := m2.GetStat("orc")
	if !exists {
		t.Fatal("Expected stat for 'orc' to exist after loading")
	}

	if orcStat.XPPerSecond != 15.0 {
		t.Errorf("Expected XPPerSecond to be 15.0, got %f", orcStat.XPPerSecond)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	xpsPath := filepath.Join(tmpDir, "nonexistent.json")

	m, err := LoadFromPath(xpsPath)
	if err != nil {
		t.Fatalf("Expected no error loading non-existent file, got: %v", err)
	}

	if len(m.Stats) != 0 {
		t.Errorf("Expected empty stats for new manager, got %d items", len(m.Stats))
	}

	if m.filePath != xpsPath {
		t.Errorf("Expected filePath to be set to %s, got %s", xpsPath, m.filePath)
	}
}

func TestGetAllStats(t *testing.T) {
	m := NewManager()
	m.UpdateStat("goblin", 20.0)
	m.UpdateStat("orc", 15.0)
	m.UpdateStat("rat", 10.0)

	allStats := m.GetAllStats()
	if len(allStats) != 3 {
		t.Errorf("Expected 3 stats, got %d", len(allStats))
	}

	expectedCreatures := []string{"goblin", "orc", "rat"}
	for _, creature := range expectedCreatures {
		if _, exists := allStats[creature]; !exists {
			t.Errorf("Expected stat for '%s' to exist", creature)
		}
	}
}

func TestEMABehavior(t *testing.T) {
	m := NewManager()

	// Test that EMA gives more weight to recent samples
	// First, establish a baseline
	for i := 0; i < 5; i++ {
		m.UpdateStat("test", 10.0)
	}

	stat, _ := m.GetStat("test")
	oldValue := stat.XPPerSecond

	// Now add a significantly different value
	m.UpdateStat("test", 50.0)

	stat, _ = m.GetStat("test")
	newValue := stat.XPPerSecond

	// New value should be between old and the new sample
	// With alpha=0.25: newValue = 0.25*50 + 0.75*10 = 12.5 + 7.5 = 20.0
	// But since we had 5 samples at 10.0, it should be closer to 10 than 50
	if newValue <= oldValue {
		t.Error("Expected EMA to increase after higher sample")
	}

	if newValue >= 50.0 {
		t.Error("Expected EMA to not jump to new value immediately")
	}
}
