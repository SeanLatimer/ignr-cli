package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplate(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	
	tests := []struct {
		name      string
		content   string
		setupFile bool
		wantErr   bool
	}{
		{
			name:      "valid file",
			content:   "# Test\ntest.txt",
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "missing file",
			setupFile: false,
			wantErr:   true,
		},
		{
			name:      "empty file",
			content:   "",
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "file with newlines",
			content:   "line1\nline2\nline3",
			setupFile: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.setupFile {
				filePath = filepath.Join(tmpDir, "test.gitignore")
				if err := os.WriteFile(filePath, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			} else {
				filePath = filepath.Join(tmpDir, "nonexistent.gitignore")
			}

			content, err := LoadTemplate(filePath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && content != tt.content {
				t.Errorf("LoadTemplate() = %q, want %q", content, tt.content)
			}
		})
	}
}

func TestLoadTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test template files
	template1 := Template{
		Name: "Go",
		Path: filepath.Join(tmpDir, "Go.gitignore"),
	}
	template2 := Template{
		Name: "Python",
		Path: filepath.Join(tmpDir, "Python.gitignore"),
	}
	
	// Write template files
	if err := os.WriteFile(template1.Path, []byte("# Go\ngo.mod"), 0o644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}
	if err := os.WriteFile(template2.Path, []byte("# Python\n*.pyc"), 0o644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	tests := []struct {
		name      string
		templates []Template
		wantErr   bool
		wantCount int
	}{
		{
			name:      "multiple templates",
			templates: []Template{template1, template2},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "single template",
			templates: []Template{template1},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "empty list",
			templates: []Template{},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "missing template",
			templates: []Template{
				{Name: "Missing", Path: filepath.Join(tmpDir, "missing.gitignore")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaded, err := LoadTemplates(tt.templates)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && len(loaded) != tt.wantCount {
				t.Errorf("LoadTemplates() = %d templates, want %d", len(loaded), tt.wantCount)
			}
		})
	}
}

func TestLoadTemplatesErrorPropagation(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create one valid and one invalid template
	validTemplate := Template{
		Name: "Go",
		Path: filepath.Join(tmpDir, "Go.gitignore"),
	}
	invalidTemplate := Template{
		Name: "Missing",
		Path: filepath.Join(tmpDir, "missing.gitignore"),
	}
	
	if err := os.WriteFile(validTemplate.Path, []byte("# Go"), 0o644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	loaded, err := LoadTemplates([]Template{validTemplate, invalidTemplate})
	
	if err == nil {
		t.Error("LoadTemplates() expected error, got nil")
	}
	
	if loaded != nil {
		t.Errorf("LoadTemplates() expected nil on error, got %v", loaded)
	}
}
