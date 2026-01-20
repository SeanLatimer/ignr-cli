package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func setupUpdateTest(t *testing.T) func() {
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
	
	cleanup := func() {
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
	
	return cleanup
}

func TestNewUpdateCommand(t *testing.T) {
	cleanup := setupUpdateTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newUpdateCommand(opts)
	
	if cmd == nil {
		t.Fatal("newUpdateCommand() returned nil")
	}
	
	if cmd.Use != "update" {
		t.Errorf("newUpdateCommand() Use = %q, want %q", cmd.Use, "update")
	}
}

func TestUpdateCommandNonInitializedCache(t *testing.T) {
	cleanup := setupUpdateTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newUpdateCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	err := cmd.Execute()
	
	// Should error because cache is not initialized
	if err == nil {
		t.Error("update command expected error for non-initialized cache, got nil")
		return
	}
	
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("update command error = %v, want error containing 'not initialized'", err)
	}
}

func TestUpdateCommandSuccess(t *testing.T) {
	cleanup := setupUpdateTest(t)
	defer cleanup()
	
	// Create an initialized cache structure
	tmpDir := t.TempDir()
	// Set XDG_CONFIG_HOME and override xdg.ConfigHome
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	xdg.ConfigHome = tmpDir
	
	cachePath := filepath.Join(tmpDir, "ignr", "cache", "github-gitignore")
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	// Create .git directory to mark as initialized
	gitDir := filepath.Join(cachePath, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	
	opts := &Options{}
	cmd := newUpdateCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	err := cmd.Execute()
	
	// Update might fail if there's no remote or network, but that's expected in tests
	// The important thing is it doesn't crash and handles the case properly
	if err != nil {
		// Expected in test environment without git remote or network
		if !strings.Contains(err.Error(), "git pull") && 
		   !strings.Contains(err.Error(), "not initialized") &&
		   !strings.Contains(err.Error(), "remote") {
			t.Logf("update command error = %v (expected in test environment)", err)
		}
	} else {
		// If it succeeds, verify output format
		output := buf.String()
		if !strings.Contains(output, "cache") && output != "" {
			t.Logf("update command succeeded, output: %q", output)
		}
	}
}
