//go:build integration

package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverTemplatesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache", "github-gitignore")
	
	// Create directory structure mimicking github/gitignore repo
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	// Create root templates
	rootTemplates := map[string]string{
		"Go.gitignore":      "# Go\ngo.mod\nvendor/",
		"Python.gitignore":  "# Python\n*.pyc\n__pycache__/",
		"Node.gitignore":    "# Node\nnode_modules/\n*.log",
	}
	
	for name, content := range rootTemplates {
		path := filepath.Join(cachePath, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}
	
	// Create Global templates
	globalDir := filepath.Join(cachePath, "Global")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("failed to create Global dir: %v", err)
	}
	
	globalTemplates := map[string]string{
		"macOS.gitignore": "# macOS\n.DS_Store",
		"Windows.gitignore": "# Windows\nThumbs.db",
	}
	
	for name, content := range globalTemplates {
		path := filepath.Join(globalDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}
	
	// Create community templates
	communityDir := filepath.Join(cachePath, "community")
	if err := os.MkdirAll(communityDir, 0o755); err != nil {
		t.Fatalf("failed to create community dir: %v", err)
	}
	
	communityTemplates := map[string]string{
		"Ruby.gitignore": "# Ruby\nGemfile.lock",
	}
	
	for name, content := range communityTemplates {
		path := filepath.Join(communityDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}
	
	// Create .git directory (should be excluded)
	gitDir := filepath.Join(cachePath, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	
	// Create a file in .git (should be excluded)
	gitFile := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitFile, []byte("# git config"), 0o644); err != nil {
		t.Fatalf("failed to create git file: %v", err)
	}
	
	// Discover templates
	templates, err := DiscoverTemplates(cachePath)
	if err != nil {
		t.Fatalf("DiscoverTemplates() error = %v", err)
	}
	
	// Should find all templates except those in .git
	expectedCount := len(rootTemplates) + len(globalTemplates) + len(communityTemplates)
	if len(templates) != expectedCount {
		t.Errorf("DiscoverTemplates() = %d templates, want %d", len(templates), expectedCount)
	}
	
	// Verify categories
	categoryCounts := make(map[Category]int)
	for _, tmpl := range templates {
		categoryCounts[tmpl.Category]++
	}
	
	if categoryCounts[CategoryRoot] != len(rootTemplates) {
		t.Errorf("DiscoverTemplates() root category = %d, want %d", categoryCounts[CategoryRoot], len(rootTemplates))
	}
	if categoryCounts[CategoryGlobal] != len(globalTemplates) {
		t.Errorf("DiscoverTemplates() Global category = %d, want %d", categoryCounts[CategoryGlobal], len(globalTemplates))
	}
	if categoryCounts[CategoryCommunity] != len(communityTemplates) {
		t.Errorf("DiscoverTemplates() community category = %d, want %d", categoryCounts[CategoryCommunity], len(communityTemplates))
	}
	
	// Verify .git directory was excluded
	for _, tmpl := range templates {
		if strings.Contains(tmpl.Path, ".git") {
			t.Errorf("DiscoverTemplates() found template in .git directory: %q", tmpl.Path)
		}
	}
	
	// Verify template names are normalized
	for _, tmpl := range templates {
		if strings.HasSuffix(tmpl.Name, ".gitignore") {
			t.Errorf("DiscoverTemplates() template name has .gitignore suffix: %q", tmpl.Name)
		}
	}
	
	// Verify all templates have SourceCache
	for _, tmpl := range templates {
		if tmpl.Source != SourceCache {
			t.Errorf("DiscoverTemplates() template %q has source %q, want %q", tmpl.Name, tmpl.Source, SourceCache)
		}
	}
}

func TestDiscoverTemplatesNestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache", "github-gitignore")
	
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	// Create nested directory structure
	nestedDir := filepath.Join(cachePath, "Global", "Sub", "Nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	
	nestedFile := filepath.Join(nestedDir, "Nested.gitignore")
	if err := os.WriteFile(nestedFile, []byte("# Nested"), 0o644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}
	
	// Discover templates
	templates, err := DiscoverTemplates(cachePath)
	if err != nil {
		t.Fatalf("DiscoverTemplates() error = %v", err)
	}
	
	// Should find the nested template
	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "Nested" {
			found = true
			// Should still be categorized as Global based on first directory
			if tmpl.Category != CategoryGlobal {
				t.Errorf("DiscoverTemplates() nested template category = %q, want %q", tmpl.Category, CategoryGlobal)
			}
			break
		}
	}
	
	if !found {
		t.Error("DiscoverTemplates() did not find nested template")
	}
}

func TestDiscoverTemplatesCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache", "github-gitignore")
	
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		t.Fatalf("failed to create cache path: %v", err)
	}
	
	// Create templates with various case combinations
	templates := map[string]string{
		"go.gitignore":      "# go lowercase",
		"PYTHON.GITIGNORE":  "# PYTHON uppercase",
		"Node.GitIgnore":    "# Node mixed case",
	}
	
	for name, content := range templates {
		path := filepath.Join(cachePath, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create template file: %v", err)
		}
	}
	
	// Discover templates
	discovered, err := DiscoverTemplates(cachePath)
	if err != nil {
		t.Fatalf("DiscoverTemplates() error = %v", err)
	}
	
	// Should find all templates regardless of case
	if len(discovered) != len(templates) {
		t.Errorf("DiscoverTemplates() = %d templates, want %d", len(discovered), len(templates))
	}
	
	// Verify names are normalized (should not have .gitignore suffix)
	for _, tmpl := range discovered {
		if strings.HasSuffix(strings.ToLower(tmpl.Name), ".gitignore") {
			t.Errorf("DiscoverTemplates() template name has .gitignore suffix: %q", tmpl.Name)
		}
	}
}
