package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverUserTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid user template directory",
			setup: func() string {
				userPath := filepath.Join(tmpDir, "user-templates-1")
				if err := os.MkdirAll(userPath, 0o755); err != nil {
					t.Fatalf("failed to create user template dir: %v", err)
				}
				// Create a template file
				templateFile := filepath.Join(userPath, "Custom.gitignore")
				if err := os.WriteFile(templateFile, []byte("# Custom\ncustom.txt"), 0o644); err != nil {
					t.Fatalf("failed to create template file: %v", err)
				}
				return userPath
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			wantCount: 0,
			wantErr:   false, // Should return empty list, not error
		},
		{
			name: "empty directory",
			setup: func() string {
				emptyDir := filepath.Join(tmpDir, "empty")
				if err := os.MkdirAll(emptyDir, 0o755); err != nil {
					t.Fatalf("failed to create empty dir: %v", err)
				}
				return emptyDir
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "multiple templates",
			setup: func() string {
				userPath := filepath.Join(tmpDir, "user-templates-2")
				if err := os.MkdirAll(userPath, 0o755); err != nil {
					t.Fatalf("failed to create user template dir: %v", err)
				}
				// Create multiple template files
				templates := []string{"Custom1.gitignore", "Custom2.gitignore"}
				for _, tmpl := range templates {
					templateFile := filepath.Join(userPath, tmpl)
					if err := os.WriteFile(templateFile, []byte("# Custom"), 0o644); err != nil {
						t.Fatalf("failed to create template file: %v", err)
					}
				}
				return userPath
			},
			wantCount: 2, // discoverTemplates does recursive walk, but should only find these 2
			wantErr:   false,
		},
		{
			name: "empty path",
			setup: func() string {
				return ""
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "templates in subdirectories",
			setup: func() string {
				userPath := filepath.Join(tmpDir, "user-templates-3")
				subDir := filepath.Join(userPath, "subdir")
				if err := os.MkdirAll(subDir, 0o755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				templateFile := filepath.Join(subDir, "Sub.gitignore")
				if err := os.WriteFile(templateFile, []byte("# Sub"), 0o644); err != nil {
					t.Fatalf("failed to create template file: %v", err)
				}
				return userPath
			},
			wantCount: 1, // discoverTemplates does recursive walk, so finds subdirectory templates
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.setup()
			
			templates, err := DiscoverUserTemplates(testPath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverUserTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(templates) != tt.wantCount {
				t.Errorf("DiscoverUserTemplates() = %d templates, want %d", len(templates), tt.wantCount)
			}
			
			// Verify all templates have CategoryUser
			for _, tmpl := range templates {
				if tmpl.Category != CategoryUser {
					t.Errorf("DiscoverUserTemplates() template %q has category %q, want %q", 
						tmpl.Name, tmpl.Category, CategoryUser)
				}
				if tmpl.Source != SourceUser {
					t.Errorf("DiscoverUserTemplates() template %q has source %q, want %q", 
						tmpl.Name, tmpl.Source, SourceUser)
				}
			}
		})
	}
}

func TestDiscoverUserTemplatesExcludesNonGitignoreFiles(t *testing.T) {
	tmpDir := t.TempDir()
	userPath := filepath.Join(tmpDir, "user-templates")
	
	if err := os.MkdirAll(userPath, 0o755); err != nil {
		t.Fatalf("failed to create user template dir: %v", err)
	}
	
	// Create a gitignore file and a non-gitignore file
	gitignoreFile := filepath.Join(userPath, "Custom.gitignore")
	nonGitignoreFile := filepath.Join(userPath, "README.md")
	
	if err := os.WriteFile(gitignoreFile, []byte("# Custom"), 0o644); err != nil {
		t.Fatalf("failed to create gitignore file: %v", err)
	}
	if err := os.WriteFile(nonGitignoreFile, []byte("# README"), 0o644); err != nil {
		t.Fatalf("failed to create non-gitignore file: %v", err)
	}
	
	templates, err := DiscoverUserTemplates(userPath)
	if err != nil {
		t.Fatalf("DiscoverUserTemplates() error = %v", err)
	}
	
	// Should only find the gitignore file (discoverTemplates does recursive walk)
	if len(templates) != 1 {
		t.Errorf("DiscoverUserTemplates() = %d templates, want 1", len(templates))
	}
	
	if templates[0].Name != "Custom" {
		t.Errorf("DiscoverUserTemplates() = %q, want %q", templates[0].Name, "Custom")
	}
}
