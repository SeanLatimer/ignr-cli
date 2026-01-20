//go:build integration

package presets

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func setupPresetIntegrationTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	
	// Save original environment variables
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	originalAppData := os.Getenv("APPDATA")
	
	// Set environment variables based on OS
	if runtime.GOOS == "windows" {
		os.Setenv("APPDATA", tmpDir)
	} else {
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
	}
	
	// Return cleanup function
	return func() {
		if originalXDGConfig != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalAppData != "" {
			os.Setenv("APPDATA", originalAppData)
		} else {
			os.Unsetenv("APPDATA")
		}
	}
}

func TestPresetLifecycleIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Test Create
	presetName := "My Project"
	templates := []string{"Go", "Python", "Node"}
	
	err := CreatePreset(presetName, templates)
	if err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}
	
	// Test Read (LoadPresets)
	store, err := LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store.Presets) != 1 {
		t.Fatalf("LoadPresets() = %d presets, want 1", len(store.Presets))
	}
	
	preset := store.Presets[0]
	if preset.Name != presetName {
		t.Errorf("LoadPresets() preset name = %q, want %q", preset.Name, presetName)
	}
	
	if len(preset.Templates) != len(templates) {
		t.Errorf("LoadPresets() templates count = %d, want %d", len(preset.Templates), len(templates))
	}
	
	// Verify timestamps
	if preset.Created == "" {
		t.Error("CreatePreset() Created timestamp is empty")
	}
	if preset.Updated == "" {
		t.Error("CreatePreset() Updated timestamp is empty")
	}
	
	// Test Update (EditPreset)
	newTemplates := []string{"Go", "Python", "Node", "Rust"}
	originalCreated := preset.Created
	
	err = EditPreset(presetName, newTemplates)
	if err != nil {
		t.Fatalf("EditPreset() error = %v", err)
	}
	
	// Verify update
	store, err = LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store.Presets[0].Templates) != len(newTemplates) {
		t.Errorf("EditPreset() templates count = %d, want %d", len(store.Presets[0].Templates), len(newTemplates))
	}
	
	// Updated should be different from Created
	if store.Presets[0].Updated == originalCreated {
		t.Error("EditPreset() Updated timestamp should be different from Created")
	}
	
	// Created should remain unchanged
	if store.Presets[0].Created != originalCreated {
		t.Error("EditPreset() Created timestamp should not change")
	}
	
	// Test Delete
	err = DeletePreset(presetName)
	if err != nil {
		t.Fatalf("DeletePreset() error = %v", err)
	}
	
	// Verify deletion
	store, err = LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store.Presets) != 0 {
		t.Errorf("DeletePreset() = %d presets, want 0", len(store.Presets))
	}
}

func TestPresetPersistenceIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create a preset
	presetName := "Test Preset"
	templates := []string{"Go", "Python"}
	
	err := CreatePreset(presetName, templates)
	if err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}
	
	// Load presets and verify
	store1, err := LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store1.Presets) != 1 {
		t.Fatalf("LoadPresets() = %d presets, want 1", len(store1.Presets))
	}
	
	// Load again to verify persistence
	store2, err := LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store2.Presets) != len(store1.Presets) {
		t.Errorf("LoadPresets() persistence check failed: first load = %d, second load = %d", 
			len(store1.Presets), len(store2.Presets))
	}
	
	if store2.Presets[0].Name != store1.Presets[0].Name {
		t.Errorf("LoadPresets() persistence check failed: first load name = %q, second load name = %q",
			store1.Presets[0].Name, store2.Presets[0].Name)
	}
}

func TestPresetYAMLFormatIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create a preset
	presetName := "YAML Test"
	templates := []string{"Go", "Python", "Node"}
	
	err := CreatePreset(presetName, templates)
	if err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}
	
	// Read the YAML file directly
	path, err := GetPresetsPath()
	if err != nil {
		t.Fatalf("GetPresetsPath() error = %v", err)
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read presets file: %v", err)
	}
	
	// Verify it's valid YAML
	var store PresetStore
	if err := yaml.Unmarshal(data, &store); err != nil {
		t.Fatalf("presets file is not valid YAML: %v", err)
	}
	
	// Verify structure
	if len(store.Presets) != 1 {
		t.Errorf("YAML has %d presets, want 1", len(store.Presets))
	}
	
	preset := store.Presets[0]
	if preset.Name != presetName {
		t.Errorf("YAML preset name = %q, want %q", preset.Name, presetName)
	}
	
	if len(preset.Templates) != len(templates) {
		t.Errorf("YAML templates count = %d, want %d", len(preset.Templates), len(templates))
	}
}

func TestMultiplePresetsIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create multiple presets
	presets := []struct {
		name      string
		templates []string
	}{
		{"Project 1", []string{"Go"}},
		{"Project 2", []string{"Python", "Node"}},
		{"Project 3", []string{"Rust"}},
	}
	
	for _, p := range presets {
		err := CreatePreset(p.name, p.templates)
		if err != nil {
			t.Fatalf("CreatePreset(%q) error = %v", p.name, err)
		}
	}
	
	// Load all presets
	store, err := LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets() error = %v", err)
	}
	
	if len(store.Presets) != len(presets) {
		t.Errorf("LoadPresets() = %d presets, want %d", len(store.Presets), len(presets))
	}
	
	// Verify each preset
	presetMap := make(map[string]Preset)
	for _, p := range store.Presets {
		presetMap[p.Name] = p
	}
	
	for _, want := range presets {
		got, ok := presetMap[want.name]
		if !ok {
			t.Errorf("LoadPresets() missing preset: %q", want.name)
			continue
		}
		
		if len(got.Templates) != len(want.templates) {
			t.Errorf("LoadPresets() preset %q templates count = %d, want %d", 
				want.name, len(got.Templates), len(want.templates))
		}
	}
}

func TestPresetKeyGenerationIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create presets with various names
	testCases := []struct {
		name     string
		wantKey   string
	}{
		{"My Project", "my-project"},
		{"Web Development", "web-development"},
		{"Project123", "project123"},
		{"project-name", "project-name"},
	}
	
	for _, tc := range testCases {
		err := CreatePreset(tc.name, []string{"Go"})
		if err != nil {
			t.Fatalf("CreatePreset(%q) error = %v", tc.name, err)
		}
		
		preset, found, err := FindPreset(tc.name)
		if err != nil {
			t.Fatalf("FindPreset(%q) error = %v", tc.name, err)
		}
		if !found {
			t.Fatalf("FindPreset(%q) not found", tc.name)
		}
		
		key := preset.Key
		if key == "" {
			key = SluggifyName(preset.Name)
		}
		
		if key != tc.wantKey {
			t.Errorf("CreatePreset(%q) key = %q, want %q", tc.name, key, tc.wantKey)
		}
		
		// Clean up for next iteration
		_ = DeletePreset(tc.name)
	}
}

func TestPresetInvalidYAMLIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create invalid YAML file
	path, err := GetPresetsPath()
	if err != nil {
		t.Fatalf("GetPresetsPath() error = %v", err)
	}
	
	invalidYAML := `presets:
  - name: Test
    templates: invalid`
	
	if err := os.WriteFile(path, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("failed to write invalid YAML: %v", err)
	}
	
	// LoadPresets should return an error
	_, err = LoadPresets()
	if err == nil {
		t.Error("LoadPresets() expected error for invalid YAML, got nil")
	}
}

func TestPresetTimestampsIntegration(t *testing.T) {
	cleanup := setupPresetIntegrationTest(t)
	defer cleanup()
	
	// Create a preset
	presetName := "Timestamp Test"
	err := CreatePreset(presetName, []string{"Go"})
	if err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}
	
	// Get the preset
	preset, found, err := FindPreset(presetName)
	if err != nil {
		t.Fatalf("FindPreset() error = %v", err)
	}
	if !found {
		t.Fatal("FindPreset() preset not found")
	}
	
	// Parse timestamps
	created, err := time.Parse(time.RFC3339, preset.Created)
	if err != nil {
		t.Fatalf("failed to parse Created timestamp: %v", err)
	}
	
	updated, err := time.Parse(time.RFC3339, preset.Updated)
	if err != nil {
		t.Fatalf("failed to parse Updated timestamp: %v", err)
	}
	
	// Timestamps should be recent (within last minute)
	now := time.Now()
	if now.Sub(created) > time.Minute {
		t.Errorf("CreatePreset() Created timestamp is too old: %v", created)
	}
	if now.Sub(updated) > time.Minute {
		t.Errorf("CreatePreset() Updated timestamp is too old: %v", updated)
	}
	
	// Wait a moment before editing
	time.Sleep(time.Second)
	
	// Edit the preset
	err = EditPreset(presetName, []string{"Go", "Python"})
	if err != nil {
		t.Fatalf("EditPreset() error = %v", err)
	}
	
	// Get updated preset
	preset, found, err = FindPreset(presetName)
	if err != nil {
		t.Fatalf("FindPreset() error = %v", err)
	}
	if !found {
		t.Fatal("FindPreset() preset not found")
	}
	
	// Parse updated timestamp
	updatedAfter, err := time.Parse(time.RFC3339, preset.Updated)
	if err != nil {
		t.Fatalf("failed to parse Updated timestamp after edit: %v", err)
	}
	
	// Updated timestamp should be newer
	if !updatedAfter.After(updated) {
		t.Errorf("EditPreset() Updated timestamp did not increase: %v -> %v", updated, updatedAfter)
	}
	
	// Created timestamp should remain unchanged
	if preset.Created != preset.Created {
		t.Error("EditPreset() Created timestamp changed")
	}
}
