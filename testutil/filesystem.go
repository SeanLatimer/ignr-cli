// Package testutil provides utilities for testing file system operations.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
)

// CreateTempCache creates a temporary cache directory structure for testing.
func CreateTempCache(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache", "github-gitignore")
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create temp cache: %v", err)
	}
	return cachePath
}

// CreateTempTemplates creates temporary template files in the cache directory.
func CreateTempTemplates(t *testing.T, cachePath string, templates map[string]string) {
	t.Helper()
	
	// Create root templates
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	for name, content := range templates {
		// Determine if it's in a subdirectory
		parts := strings.Split(name, string(filepath.Separator))
		targetDir := cachePath
		
		// Handle category-based paths
		if len(parts) > 1 {
			targetDir = filepath.Join(cachePath, filepath.Join(parts[:len(parts)-1]...))
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				t.Fatalf("failed to create template directory: %v", err)
			}
		}
		
		fileName := parts[len(parts)-1]
		if !strings.HasSuffix(strings.ToLower(fileName), ".gitignore") {
			fileName += ".gitignore"
		}
		
		targetPath := filepath.Join(targetDir, fileName)
		if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write template file: %v", err)
		}
	}
}

// CreateTestRepo creates a temporary git repository structure.
func CreateTestRepo(t *testing.T) (string, error) {
	t.Helper()
	repoPath := t.TempDir()
	
	// Initialize git repository
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		return "", fmt.Errorf("failed to init git repo: %w", err)
	}
	
	// Create a test file
	testFile := filepath.Join(repoPath, "test.gitignore")
	if err := os.WriteFile(testFile, []byte("# test"), 0o644); err != nil {
		return "", fmt.Errorf("failed to write test file: %w", err)
	}
	
	// Add and commit
	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}
	
	if _, err := wt.Add("test.gitignore"); err != nil {
		return "", fmt.Errorf("failed to add file: %w", err)
	}
	
	if _, err := wt.Commit("Initial commit", &git.CommitOptions{}); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	
	return repoPath, nil
}

// SetupTestCache sets up a test cache directory with a git repository.
func SetupTestCache(t *testing.T) string {
	t.Helper()
	cachePath := CreateTempCache(t)
	
	// Initialize as git repository
	repo, err := git.PlainInit(cachePath, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	// Create a default .gitignore file
	testFile := filepath.Join(cachePath, "Go.gitignore")
	if err := os.WriteFile(testFile, []byte("# Go\n*.exe"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	
	// Add and commit
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	
	if _, err := wt.Add("Go.gitignore"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}
	
	if _, err := wt.Commit("Initial commit", &git.CommitOptions{}); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	
	return cachePath
}

// CreateTestConfig creates a test configuration file.
func CreateTestConfig(t *testing.T, dir string, config map[string]interface{}) string {
	t.Helper()
	configDir := filepath.Join(dir, "ignr")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	
	configPath := filepath.Join(configDir, "config.json")
	
	var content string
	if defaultOutput, ok := config["default_output"].(string); ok {
		content = fmt.Sprintf(`{"default_output": %q}`, defaultOutput)
	}
	if userTemplatePath, ok := config["user_template_path"].(string); ok {
		if content != "" {
			content = strings.TrimSuffix(content, "}")
			content += fmt.Sprintf(`,"user_template_path": %q}`, userTemplatePath)
		} else {
			content = fmt.Sprintf(`{"user_template_path": %q}`, userTemplatePath)
		}
	}
	
	if content == "" {
		content = "{}"
	}
	
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	
	return configPath
}

// CreateTestPresets creates a test presets file.
func CreateTestPresets(t *testing.T, dir string, presets []map[string]interface{}) string {
	t.Helper()
	configDir := filepath.Join(dir, "ignr")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	
	presetsPath := filepath.Join(configDir, "presets.yaml")
	
	var yamlContent strings.Builder
	yamlContent.WriteString("presets:\n")
	
	for _, preset := range presets {
		yamlContent.WriteString("  - ")
		if name, ok := preset["name"].(string); ok {
			yamlContent.WriteString(fmt.Sprintf("name: %q\n", name))
		}
		if key, ok := preset["key"].(string); ok {
			yamlContent.WriteString(fmt.Sprintf("    key: %q\n", key))
		}
		if templates, ok := preset["templates"].([]string); ok {
			yamlContent.WriteString("    templates:\n")
			for _, tmpl := range templates {
				yamlContent.WriteString(fmt.Sprintf("      - %q\n", tmpl))
			}
		}
	}
	
	if err := os.WriteFile(presetsPath, []byte(yamlContent.String()), 0o644); err != nil {
		t.Fatalf("failed to write presets file: %v", err)
	}
	
	return presetsPath
}

// CleanupTemp is a helper for cleanup (usually used with t.Cleanup).
func CleanupTemp(paths ...string) func() {
	return func() {
		for _, path := range paths {
			os.RemoveAll(path)
		}
	}
}
