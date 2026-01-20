package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setupConfigTest sets up a temporary config directory for testing
// and returns a cleanup function.
func setupConfigTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	
	// Save original environment variables
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	originalAppData := os.Getenv("APPDATA")
	
	// Set environment variables based on OS
	if runtime.GOOS == "windows" {
		if err := os.Setenv("APPDATA", tmpDir); err != nil {
			t.Fatalf("failed to set APPDATA: %v", err)
		}
	} else {
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
	}
	
	// Return cleanup function
	return func() {
		if originalXDGConfig != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDGConfig); err != nil {
				t.Logf("failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
		if originalAppData != "" {
			if err := os.Setenv("APPDATA", originalAppData); err != nil {
				t.Logf("failed to restore APPDATA: %v", err)
			}
		} else {
			if err := os.Unsetenv("APPDATA"); err != nil {
				t.Logf("failed to unset APPDATA: %v", err)
			}
		}
	}
}

func TestGetConfigDir(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}
	
	// Should contain the config directory name
	if !strings.Contains(dir, configDirName) {
		t.Errorf("GetConfigDir() = %q, want path containing %q", dir, configDirName)
	}
	
	// Should be a directory
	info, err := os.Stat(dir)
	if err == nil && !info.IsDir() {
		t.Errorf("GetConfigDir() = %q, is not a directory", dir)
	}
}

func TestGetConfigPath(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	
	// Should end with config file name
	if !strings.HasSuffix(path, configFileName) {
		t.Errorf("GetConfigPath() = %q, want path ending with %q", path, configFileName)
	}
	
	// Should contain config directory
	if !strings.Contains(path, configDirName) {
		t.Errorf("GetConfigPath() = %q, want path containing %q", path, configDirName)
	}
}

func TestGetUserTemplatePath(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	tests := []struct {
		name          string
		configData    Config
		wantPath      func(string) bool
		wantErr       bool
	}{
		{
			name:       "default path",
			configData: Config{},
			wantPath: func(path string) bool {
				return strings.Contains(path, "templates")
			},
			wantErr: false,
		},
		{
			name: "custom path",
			configData: Config{
				UserTemplatePath: "/custom/templates",
			},
			wantPath: func(path string) bool {
				return path == "/custom/templates"
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save config
			if err := SaveConfig(tt.configData); err != nil {
				t.Fatalf("SaveConfig() error = %v", err)
			}
			
			path, err := GetUserTemplatePath()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserTemplatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && !tt.wantPath(path) {
				t.Errorf("GetUserTemplatePath() = %q, did not match expected path", path)
			}
			
			// Verify directory was created
			if !tt.wantErr {
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("GetUserTemplatePath() directory does not exist: %v", err)
				} else if !info.IsDir() {
					t.Errorf("GetUserTemplatePath() path is not a directory: %q", path)
				}
			}
		})
	}
}

func TestGetPresetsPath(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	path, err := GetPresetsPath()
	if err != nil {
		t.Fatalf("GetPresetsPath() error = %v", err)
	}
	
	// Should end with presets.yaml
	if !strings.HasSuffix(path, "presets.yaml") {
		t.Errorf("GetPresetsPath() = %q, want path ending with presets.yaml", path)
	}
	
	// Should create presets file if missing
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("GetPresetsPath() presets file does not exist: %v", err)
	} else if info.IsDir() {
		t.Errorf("GetPresetsPath() path is a directory, want file")
	}
	
	// Verify file content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read presets file: %v", err)
	}
	
	if !strings.Contains(string(data), "presets:") {
		t.Errorf("GetPresetsPath() presets file content = %q, want containing 'presets:'", string(data))
	}
}

func TestLoadConfig(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	tests := []struct {
		name      string
		setup     func()
		want      Config
		wantErr   bool
	}{
		{
			name: "valid config",
			setup: func() {
				cfg := Config{
					DefaultOutput:    ".gitignore",
					UserTemplatePath: "/custom/templates",
				}
				if err := SaveConfig(cfg); err != nil {
					t.Fatalf("SaveConfig() error = %v", err)
				}
			},
			want: Config{
				DefaultOutput:    ".gitignore",
				UserTemplatePath: "/custom/templates",
			},
			wantErr: false,
		},
		{
			name: "missing config file",
			setup: func() {
				// Ensure config directory exists but remove config file if it exists
				path, err := GetConfigPath()
				if err == nil {
					if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
						t.Logf("failed to remove config file: %v", err)
					}
				}
			},
			want:    Config{}, // Should return empty config, not error
			wantErr: false,
		},
		{
			name: "empty config",
			setup: func() {
				cfg := Config{}
				if err := SaveConfig(cfg); err != nil {
					t.Fatalf("SaveConfig() error = %v", err)
				}
			},
			want:    Config{},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			
			cfg, err := LoadConfig()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if cfg.DefaultOutput != tt.want.DefaultOutput {
					t.Errorf("LoadConfig() DefaultOutput = %q, want %q", cfg.DefaultOutput, tt.want.DefaultOutput)
				}
				if cfg.UserTemplatePath != tt.want.UserTemplatePath {
					t.Errorf("LoadConfig() UserTemplatePath = %q, want %q", cfg.UserTemplatePath, tt.want.UserTemplatePath)
				}
			}
		})
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	// Create invalid JSON file
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	
	invalidJSON := `{"default_output": "test", invalid}`
	if err := os.WriteFile(path, []byte(invalidJSON), 0o644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}
	
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig() expected error for invalid JSON, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "basic config",
			cfg: Config{
				DefaultOutput: ".gitignore",
			},
		},
		{
			name: "full config",
			cfg: Config{
				DefaultOutput:    ".gitignore",
				UserTemplatePath: "/custom/templates",
			},
		},
		{
			name: "empty config",
			cfg:  Config{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveConfig(tt.cfg)
			if err != nil {
				t.Fatalf("SaveConfig() error = %v", err)
			}
			
			// Verify config was saved
			path, err := GetConfigPath()
			if err != nil {
				t.Fatalf("GetConfigPath() error = %v", err)
			}
			
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}
			
			// Verify JSON format
			var loaded Config
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}
			
			if loaded.DefaultOutput != tt.cfg.DefaultOutput {
				t.Errorf("SaveConfig() DefaultOutput = %q, want %q", loaded.DefaultOutput, tt.cfg.DefaultOutput)
			}
			if loaded.UserTemplatePath != tt.cfg.UserTemplatePath {
				t.Errorf("SaveConfig() UserTemplatePath = %q, want %q", loaded.UserTemplatePath, tt.cfg.UserTemplatePath)
			}
		})
	}
}

func TestSaveConfigCreatesDirectory(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	cfg := Config{
		DefaultOutput: ".gitignore",
	}
	
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	
	// Verify directory was created
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}
	
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("SaveConfig() config directory does not exist: %v", err)
	} else if !info.IsDir() {
		t.Errorf("SaveConfig() config path is not a directory: %q", dir)
	}
}

func TestConfigJSONFormatting(t *testing.T) {
	cleanup := setupConfigTest(t)
	defer cleanup()
	
	cfg := Config{
		DefaultOutput:    ".gitignore",
		UserTemplatePath: "/custom/templates",
	}
	
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	
	// Read and verify formatting
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	
	// Should be indented JSON
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		t.Errorf("SaveConfig() JSON should be formatted with indentation, got %d lines", len(lines))
	}
	
	// Verify it's valid JSON
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("SaveConfig() saved invalid JSON: %v", err)
	}
}
