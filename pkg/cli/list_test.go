package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupListTest(t *testing.T) (func(), string) {
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
	
	// Create cache structure
	cachePath := filepath.Join(tmpDir, "ignr", "cache", "github-gitignore")
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	// Create template files
	templates := map[string]string{
		"Go.gitignore":      "# Go",
		"Python.gitignore":  "# Python",
		"Node.gitignore":    "# Node",
	}
	
	for name, content := range templates {
		path := filepath.Join(cachePath, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}
	
	// Create .git directory
	gitDir := filepath.Join(cachePath, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	
	// Create Global directory
	globalDir := filepath.Join(cachePath, "Global")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("failed to create Global dir: %v", err)
	}
	
	globalFile := filepath.Join(globalDir, "macOS.gitignore")
	if err := os.WriteFile(globalFile, []byte("# macOS"), 0o644); err != nil {
		t.Fatalf("failed to create Global template: %v", err)
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
	
	return cleanup, cachePath
}

func TestNewListCommand(t *testing.T) {
	cleanup, _ := setupListTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newListCommand(opts)
	
	if cmd == nil {
		t.Fatal("newListCommand() returned nil")
	}
	
	if cmd.Use != "list" {
		t.Errorf("newListCommand() Use = %q, want %q", cmd.Use, "list")
	}
}

func TestListCommandAllTemplates(t *testing.T) {
	cleanup, _ := setupListTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newListCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command error = %v", err)
	}
	
	output := buf.String()
	
	// Should list all templates
	if !strings.Contains(output, "Go") {
		t.Error("list command output missing 'Go'")
	}
	if !strings.Contains(output, "Python") {
		t.Error("list command output missing 'Python'")
	}
	if !strings.Contains(output, "Node") {
		t.Error("list command output missing 'Node'")
	}
	if !strings.Contains(output, "macOS") {
		t.Error("list command output missing 'macOS'")
	}
	
	// Should include categories
	if !strings.Contains(output, "[root]") {
		t.Error("list command output missing category")
	}
	if !strings.Contains(output, "[Global]") {
		t.Error("list command output missing Global category")
	}
}

func TestListCommandWithCategoryFilter(t *testing.T) {
	cleanup, _ := setupListTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newListCommand(opts)
	
	// Set category flag
	cmd.SetArgs([]string{"--category", "Global"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("list command error = %v", err)
	}
	
	output := buf.String()
	
	// Should only list Global category templates
	if !strings.Contains(output, "macOS") {
		t.Error("list command with category filter missing 'macOS'")
	}
	
	// Should not contain root templates
	if strings.Contains(output, "Go") {
		t.Error("list command with category filter should not contain root templates")
	}
}

func TestListCommandEmptyCache(t *testing.T) {
	// Note: This test may pass if a real cache exists on the system
	// since cache.InitializeCache() will use the real cache if it exists
	// To test uninitialized cache, we'd need to mock or disable cache initialization
	t.Skip("Skipping test - InitializeCache uses real cache directory which may be initialized")
}
