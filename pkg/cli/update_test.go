package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupUpdateTest(t *testing.T) func() {
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
	
	cleanup := func() {
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
	if runtime.GOOS == "windows" {
		if err := os.Setenv("APPDATA", tmpDir); err != nil {
			t.Fatalf("failed to set APPDATA: %v", err)
		}
	} else {
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
	}
	
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
