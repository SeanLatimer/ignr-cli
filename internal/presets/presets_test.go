package presets

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
)

// setupPresetTest sets up a temporary config directory for testing presets
// and returns a cleanup function.
func setupPresetTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	
	// Save original values
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	originalConfigHome := xdg.ConfigHome
	
	// Set XDG_CONFIG_HOME environment variable
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	
	// Directly override xdg.ConfigHome since xdg reads env vars at init time
	xdg.ConfigHome = tmpDir
	
	// Return cleanup function
	return func() {
		// Restore xdg.ConfigHome
		xdg.ConfigHome = originalConfigHome
		
		// Restore environment variable
		if originalXDGConfig != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDGConfig); err != nil {
				t.Logf("failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}
}

func TestSluggifyName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal name",
			input:    "My Project",
			expected: "my-project",
		},
		{
			name:     "name with spaces",
			input:    "Web Development",
			expected: "web-development",
		},
		{
			name:     "name with underscores",
			input:    "my_project",
			expected: "my-project",
		},
		{
			name:     "name with hyphens",
			input:    "my-project",
			expected: "my-project",
		},
		{
			name:     "multiple consecutive spaces",
			input:    "My   Project",
			expected: "my-project",
		},
		{
			name:     "special characters",
			input:    "My@Project#123",
			expected: "myproject123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "preset",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "preset",
		},
		{
			name:     "case normalization",
			input:    "MY PROJECT",
			expected: "my-project",
		},
		{
			name:     "mixed case",
			input:    "MyProject",
			expected: "myproject",
		},
		{
			name:     "numbers",
			input:    "Project123",
			expected: "project123",
		},
		{
			name:     "leading/trailing spaces",
			input:    "  My Project  ",
			expected: "my-project",
		},
		{
			name:     "multiple hyphens normalization",
			input:    "my---project",
			expected: "my-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SluggifyName(tt.input)
			if result != tt.expected {
				t.Errorf("SluggifyName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreatePreset(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	tests := []struct {
		name        string
		presetName  string
		templates   []string
		wantErr     bool
		errContains string
	}{
		{
			name:       "new preset creation",
			presetName: "My Project",
			templates:  []string{"Go", "Python"},
			wantErr:    false,
		},
		{
			name:       "empty template list",
			presetName: "Empty Preset",
			templates:  []string{},
			wantErr:    false,
		},
		{
			name:        "duplicate key",
			presetName:  "My Project 2", // Different name to avoid conflict with previous test
			templates:   []string{"Go"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreatePreset(tt.presetName, tt.templates)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePreset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CreatePreset() error = %v, want error containing %q", err, tt.errContains)
				}
			}
			
			// Verify preset was created
			if !tt.wantErr {
				preset, found, err := FindPreset(tt.presetName)
				if err != nil {
					t.Errorf("FindPreset() error = %v", err)
					return
				}
				if !found {
					t.Error("CreatePreset() preset was not created")
					return
				}
				if preset.Name != tt.presetName {
					t.Errorf("CreatePreset() preset name = %q, want %q", preset.Name, tt.presetName)
				}
				if len(preset.Templates) != len(tt.templates) {
					t.Errorf("CreatePreset() templates count = %d, want %d", len(preset.Templates), len(tt.templates))
				}
			}
		})
	}
}

func TestCreatePresetDuplicateKey(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	// Create first preset
	if err := CreatePreset("My Project", []string{"Go"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	// Try to create duplicate
	err := CreatePreset("My Project", []string{"Python"})
	if err == nil {
		t.Error("CreatePreset() expected error for duplicate key, got nil")
		return
	}
	
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("CreatePreset() error = %v, want error containing 'already exists'", err)
	}
}

func TestEditPreset(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	// Create a preset first
	presetName := "My Project"
	if err := CreatePreset(presetName, []string{"Go"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	// Edit the preset
	newTemplates := []string{"Go", "Python", "Node"}
	err := EditPreset(presetName, newTemplates)
	if err != nil {
		t.Fatalf("EditPreset() error = %v", err)
	}

	// Verify the preset was updated
	preset, found, err := FindPreset(presetName)
	if err != nil {
		t.Fatalf("FindPreset() error = %v", err)
	}
	if !found {
		t.Fatal("EditPreset() preset not found after edit")
	}
	if len(preset.Templates) != len(newTemplates) {
		t.Errorf("EditPreset() templates count = %d, want %d", len(preset.Templates), len(newTemplates))
	}
}

func TestEditPresetNotFound(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	err := EditPreset("Nonexistent", []string{"Go"})
	if err == nil {
		t.Error("EditPreset() expected error for nonexistent preset, got nil")
		return
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("EditPreset() error = %v, want error containing 'not found'", err)
	}
}

func TestDeletePreset(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	// Create a preset first
	presetName := "My Project"
	if err := CreatePreset(presetName, []string{"Go"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	// Delete the preset
	err := DeletePreset(presetName)
	if err != nil {
		t.Fatalf("DeletePreset() error = %v", err)
	}

	// Verify the preset was deleted
	_, found, err := FindPreset(presetName)
	if err != nil {
		t.Fatalf("FindPreset() error = %v", err)
	}
	if found {
		t.Error("DeletePreset() preset still exists after deletion")
	}
}

func TestDeletePresetNotFound(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	err := DeletePreset("Nonexistent")
	if err == nil {
		t.Error("DeletePreset() expected error for nonexistent preset, got nil")
		return
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("DeletePreset() error = %v, want error containing 'not found'", err)
	}
}

func TestFindPreset(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	presetName := "My Project"
	presetKey := "my-project"
	templates := []string{"Go", "Python"}
	
	if err := CreatePreset(presetName, templates); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	tests := []struct {
		name     string
		search   string
		wantName string
		wantOk   bool
	}{
		{
			name:     "find by key",
			search:   presetKey,
			wantName: presetName,
			wantOk:   true,
		},
		{
			name:     "find by name",
			search:   presetName,
			wantName: presetName,
			wantOk:   true,
		},
		{
			name:     "case insensitive by key",
			search:   "MY-PROJECT",
			wantName: presetName,
			wantOk:   true,
		},
		{
			name:     "case insensitive by name",
			search:   "my project",
			wantName: presetName,
			wantOk:   true,
		},
		{
			name:     "not found",
			search:   "Nonexistent",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset, found, err := FindPreset(tt.search)
			if err != nil {
				t.Errorf("FindPreset() error = %v", err)
				return
			}
			if found != tt.wantOk {
				t.Errorf("FindPreset() found = %v, want %v", found, tt.wantOk)
				return
			}
			if found && preset.Name != tt.wantName {
				t.Errorf("FindPreset() name = %q, want %q", preset.Name, tt.wantName)
			}
		})
	}
}

func TestListPresets(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	// Test empty list
	presets, err := ListPresets()
	if err != nil {
		t.Fatalf("ListPresets() error = %v", err)
	}
	if len(presets) != 0 {
		t.Errorf("ListPresets() = %d presets, want 0", len(presets))
	}

	// Create some presets
	preset1 := "Project 1"
	preset2 := "Project 2"
	
	if err := CreatePreset(preset1, []string{"Go"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}
	if err := CreatePreset(preset2, []string{"Python"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	// List presets
	presets, err = ListPresets()
	if err != nil {
		t.Fatalf("ListPresets() error = %v", err)
	}
	if len(presets) != 2 {
		t.Errorf("ListPresets() = %d presets, want 2", len(presets))
	}
}

func TestPresetTimestampGeneration(t *testing.T) {
	cleanup := setupPresetTest(t)
	defer cleanup()

	presetName := "My Project"
	if err := CreatePreset(presetName, []string{"Go"}); err != nil {
		t.Fatalf("CreatePreset() error = %v", err)
	}

	preset, found, err := FindPreset(presetName)
	if err != nil {
		t.Fatalf("FindPreset() error = %v", err)
	}
	if !found {
		t.Fatal("FindPreset() preset not found")
	}

	// Verify timestamps are set
	if preset.Created == "" {
		t.Error("CreatePreset() Created timestamp is empty")
	}
	if preset.Updated == "" {
		t.Error("CreatePreset() Updated timestamp is empty")
	}

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, preset.Created)
	if err != nil {
		t.Errorf("CreatePreset() Created timestamp format invalid: %v", err)
	}
}
