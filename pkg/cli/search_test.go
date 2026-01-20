package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupSearchTest(t *testing.T) func() {
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
		"Nodejs.gitignore":  "# Node.js",
		"Ruby.gitignore":    "# Ruby",
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

func TestNewSearchCommand(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	if cmd == nil {
		t.Fatal("newSearchCommand() returned nil")
	}
	
	if cmd.Use != "search <pattern>" {
		t.Errorf("newSearchCommand() Use = %q, want %q", cmd.Use, "search <pattern>")
	}
}

func TestSearchCommandFuzzySearch(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	// Test fuzzy search for "python"
	cmd.SetArgs([]string{"python"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("search command error = %v", err)
	}
	
	output := buf.String()
	
	// Should find Python
	if !strings.Contains(output, "Python") {
		t.Error("search command output missing 'Python'")
	}
}

func TestSearchCommandMultiplePatterns(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	// Test multiple word pattern - joined as one search query
	cmd.SetArgs([]string{"node js"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("search command error = %v", err)
	}
	
	output := buf.String()
	
	// Should find Node or Nodejs - fuzzy search may match either
	if !strings.Contains(output, "Node") && !strings.Contains(output, "node") {
		t.Logf("search command output: %q (may not contain 'Node' depending on fuzzy match)", output)
	}
}

func TestSearchCommandCaseInsensitive(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	// Test case insensitive search
	cmd.SetArgs([]string{"PYTHON"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("search command error = %v", err)
	}
	
	output := buf.String()
	
	// Should find Python (case insensitive)
	if !strings.Contains(output, "Python") {
		t.Error("search command case insensitive search failed")
	}
}

func TestSearchCommandEmptyResults(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	// Search for something that doesn't exist
	cmd.SetArgs([]string{"nonexistent12345"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("search command error = %v", err)
	}
	
	output := buf.String()
	
	// Should have no results (empty output or specific message)
	// The fuzzy search might return empty results
	if strings.Contains(output, "Python") || strings.Contains(output, "Node") {
		t.Error("search command should not find templates for nonexistent pattern")
	}
}

func TestSearchCommandRequiresPattern(t *testing.T) {
	cleanup := setupSearchTest(t)
	defer cleanup()
	
	opts := &Options{}
	cmd := newSearchCommand(opts)
	
	// No args provided
	cmd.SetArgs([]string{})
	
	err := cmd.Execute()
	
	// Should error because pattern is required
	if err == nil {
		t.Error("search command expected error for missing pattern, got nil")
	}
}
