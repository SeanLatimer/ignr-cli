package presets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name      string
		setup     func() string
		wantCount int
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "file detection",
			setup: func() string {
				// Create a unique subdirectory for this test
				testDir := filepath.Join(tmpDir, "file-detection")
				if err := os.MkdirAll(testDir, 0o755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				// Create some files
				files := []string{"package.json", "go.mod", "requirements.txt"}
				for _, file := range files {
					filePath := filepath.Join(testDir, file)
					if err := os.WriteFile(filePath, []byte("# test"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return testDir
			},
			wantCount: 4, // Root directory + 3 files
			wantFiles: []string{"package.json", "go.mod", "requirements.txt"},
			wantErr:   false,
		},
		{
			name: "directory detection",
			setup: func() string {
				// Create a unique subdirectory for this test
				testDir := filepath.Join(tmpDir, "dir-detection")
				if err := os.MkdirAll(testDir, 0o755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				// Create directories
				dirs := []string{"node_modules", ".git", "src"}
				for _, dir := range dirs {
					dirPath := filepath.Join(testDir, dir)
					if err := os.MkdirAll(dirPath, 0o755); err != nil {
						t.Fatalf("failed to create directory: %v", err)
					}
				}
				return testDir
			},
			wantCount: 3, // Root directory + node_modules/ + src/ (.git excluded)
			wantFiles: []string{"node_modules/", "src/"},
			wantErr:   false,
		},
		{
			name: "exclude .git directory",
			setup: func() string {
				// Create a unique subdirectory for this test
				testDir := filepath.Join(tmpDir, "exclude-git")
				if err := os.MkdirAll(testDir, 0o755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				// Create .git directory
				gitDir := filepath.Join(testDir, ".git")
				if err := os.MkdirAll(gitDir, 0o755); err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}
				// Create a file in .git
				gitFile := filepath.Join(gitDir, "config")
				if err := os.WriteFile(gitFile, []byte("# git config"), 0o644); err != nil {
					t.Fatalf("failed to create git file: %v", err)
				}
				return testDir
			},
			wantCount: 1, // Root directory only (.git excluded)
			wantFiles: []string{},
			wantErr:   false,
		},
		{
			name: "empty directory",
			setup: func() string {
				// Create a unique empty directory
				testDir := filepath.Join(tmpDir, "empty")
				if err := os.MkdirAll(testDir, 0o755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				return testDir
			},
			wantCount: 1, // Root directory only
			wantFiles: []string{},
			wantErr:   false,
		},
		{
			name: "nested structure",
			setup: func() string {
				// Create a unique subdirectory for this test
				testDir := filepath.Join(tmpDir, "nested")
				if err := os.MkdirAll(testDir, 0o755); err != nil {
					t.Fatalf("failed to create test dir: %v", err)
				}
				// Create nested directories and files
				nestedDir := filepath.Join(testDir, "src", "app")
				if err := os.MkdirAll(nestedDir, 0o755); err != nil {
					t.Fatalf("failed to create nested directory: %v", err)
				}
				nestedFile := filepath.Join(nestedDir, "main.go")
				if err := os.WriteFile(nestedFile, []byte("package main"), 0o644); err != nil {
					t.Fatalf("failed to create nested file: %v", err)
				}
				rootFile := filepath.Join(testDir, "go.mod")
				if err := os.WriteFile(rootFile, []byte("module test"), 0o644); err != nil {
					t.Fatalf("failed to create root file: %v", err)
				}
				return testDir
			},
			wantCount: 5, // Root directory + src/ + app/ + go.mod + main.go
			wantFiles: []string{"src/", "go.mod", "main.go"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setup()
			
			detected, err := DetectFiles(repoPath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(detected) != tt.wantCount {
				t.Errorf("DetectFiles() = %d files, want %d (detected: %v)", len(detected), tt.wantCount, detected)
			}
			
			// Check for specific files (case-insensitive)
			detectedMap := make(map[string]bool)
			for _, file := range detected {
				detectedMap[strings.ToLower(file)] = true
			}
			
			for _, wantFile := range tt.wantFiles {
				if !detectedMap[strings.ToLower(wantFile)] {
					t.Errorf("DetectFiles() missing file: %q", wantFile)
				}
			}
			
			// Verify .git is excluded
			for _, file := range detected {
				if strings.Contains(strings.ToLower(file), ".git") {
					t.Errorf("DetectFiles() should not contain .git: %q", file)
				}
			}
		})
	}
}

func TestSuggestTemplates(t *testing.T) {
	tests := []struct {
		name        string
		detected    []string
		wantSuggest []string
		wantErr     bool
	}{
		{
			name:        "package.json suggests Node",
			detected:    []string{"package.json"},
			wantSuggest: []string{"Node"},
			wantErr:     false,
		},
		{
			name:        "go.mod suggests Go",
			detected:    []string{"go.mod"},
			wantSuggest: []string{"Go"},
			wantErr:     false,
		},
		{
			name:        "requirements.txt suggests Python",
			detected:    []string{"requirements.txt"},
			wantSuggest: []string{"Python"},
			wantErr:     false,
		},
		{
			name:        "multiple matches",
			detected:    []string{"package.json", "go.mod", "requirements.txt"},
			wantSuggest: []string{"Node", "Go", "Python"},
			wantErr:     false,
		},
		{
			name:        "case insensitive",
			detected:    []string{"PACKAGE.JSON", "GO.MOD"},
			wantSuggest: []string{"Node", "Go"},
			wantErr:     false,
		},
		{
			name:        "no matches",
			detected:    []string{"unknown.txt"},
			wantSuggest: []string{},
			wantErr:     false,
		},
		{
			name:        "empty detected",
			detected:    []string{},
			wantSuggest: []string{},
			wantErr:     false,
		},
		{
			name:        "directory match",
			detected:    []string{".idea/"},
			wantSuggest: []string{"IntelliJ"},
			wantErr:     false,
		},
		{
			name:        "pyproject.toml suggests Python",
			detected:    []string{"pyproject.toml"},
			wantSuggest: []string{"Python"},
			wantErr:     false,
		},
		{
			name:        "setup.py suggests Python",
			detected:    []string{"setup.py"},
			wantSuggest: []string{"Python"},
			wantErr:     false,
		},
		{
			name:        "cargo.toml suggests Rust",
			detected:    []string{"cargo.toml"},
			wantSuggest: []string{"Rust"},
			wantErr:     false,
		},
		{
			name:        "pom.xml suggests Maven",
			detected:    []string{"pom.xml"},
			wantSuggest: []string{"Maven"},
			wantErr:     false,
		},
		{
			name:        "build.gradle suggests Gradle",
			detected:    []string{"build.gradle"},
			wantSuggest: []string{"Gradle"},
			wantErr:     false,
		},
		{
			name:        "composer.json suggests Composer",
			detected:    []string{"composer.json"},
			wantSuggest: []string{"Composer"},
			wantErr:     false,
		},
		{
			name:        "TypeScript files",
			detected:    []string{"app.ts", "component.tsx"},
			wantSuggest: []string{"TypeScript"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, err := SuggestTemplates(tt.detected)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("SuggestTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(suggestions) != len(tt.wantSuggest) {
				t.Errorf("SuggestTemplates() = %d suggestions, want %d", len(suggestions), len(tt.wantSuggest))
			}
			
			// Check suggestions (order may vary)
			suggestMap := make(map[string]bool)
			for _, s := range suggestions {
				suggestMap[s] = true
			}
			
			for _, want := range tt.wantSuggest {
				if !suggestMap[want] {
					t.Errorf("SuggestTemplates() missing suggestion: %q", want)
				}
			}
		})
	}
}

func TestRuleMatches(t *testing.T) {
	tests := []struct {
		name     string
		rule     DetectionRule
		detected []string
		want     bool
	}{
		{
			name: "exact match",
			rule: DetectionRule{
				Patterns:  []string{"package.json"},
				Templates: []string{"Node"},
			},
			detected: []string{"package.json"},
			want:     true,
		},
		{
			name: "case insensitive match",
			rule: DetectionRule{
				Patterns:  []string{"package.json"},
				Templates: []string{"Node"},
			},
			detected: []string{"PACKAGE.JSON"},
			want:     true,
		},
		{
			name: "wildcard match",
			rule: DetectionRule{
				Patterns:  []string{"*.ts"},
				Templates: []string{"TypeScript"},
			},
			detected: []string{"app.ts"},
			want:     true,
		},
		{
			name: "no match",
			rule: DetectionRule{
				Patterns:  []string{"package.json"},
				Templates: []string{"Node"},
			},
			detected: []string{"go.mod"},
			want:     false,
		},
		{
			name: "multiple patterns, one matches",
			rule: DetectionRule{
				Patterns:  []string{"requirements.txt", "setup.py", "pyproject.toml"},
				Templates: []string{"Python"},
			},
			detected: []string{"go.mod", "setup.py"},
			want:     true,
		},
		{
			name: "multiple patterns, none match",
			rule: DetectionRule{
				Patterns:  []string{"requirements.txt", "setup.py"},
				Templates: []string{"Python"},
			},
			detected: []string{"go.mod", "package.json"},
			want:     false,
		},
		{
			name: "empty detected",
			rule: DetectionRule{
				Patterns:  []string{"package.json"},
				Templates: []string{"Node"},
			},
			detected: []string{},
			want:     false,
		},
		{
			name: "directory match",
			rule: DetectionRule{
				Patterns:  []string{".idea/"},
				Templates: []string{"IntelliJ"},
			},
			detected: []string{".idea/"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ruleMatches(tt.rule, tt.detected)
			if result != tt.want {
				t.Errorf("ruleMatches() = %v, want %v", result, tt.want)
			}
		})
	}
}
