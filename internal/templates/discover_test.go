package templates

import (
	"path/filepath"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with .gitignore suffix",
			input:    "Python.gitignore",
			expected: "Python",
		},
		{
			name:     "lowercase with suffix",
			input:    "python.gitignore",
			expected: "python",
		},
		{
			name:     "without suffix",
			input:    "Python",
			expected: "Python",
		},
		{
			name:     "case insensitive suffix",
			input:    "Python.GITIGNORE",
			expected: "Python",
		},
		{
			name:     "mixed case",
			input:    "pYtHoN.gitignore",
			expected: "pYtHoN",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only suffix",
			input:    ".gitignore",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCategorize(t *testing.T) {
	// Use filepath operations to ensure OS-appropriate separators
	tests := []struct {
		name     string
		input    string
		expected Category
	}{
		{
			name:     "root category",
			input:    "Go.gitignore",
			expected: CategoryRoot,
		},
		{
			name:     "global category",
			input:    filepath.Join("Global", "macOS.gitignore"),
			expected: CategoryGlobal,
		},
		{
			name:     "community category",
			input:    filepath.Join("community", "Ruby.gitignore"),
			expected: CategoryCommunity,
		},
		{
			name:     "case insensitive global",
			input:    filepath.Join("global", "macos.gitignore"),
			expected: CategoryGlobal,
		},
		{
			name:     "case insensitive community",
			input:    filepath.Join("COMMUNITY", "ruby.gitignore"),
			expected: CategoryCommunity,
		},
		{
			name:     "nested path",
			input:    filepath.Join("Global", "Sub", "macOS.gitignore"),
			expected: CategoryGlobal,
		},
		{
			name:     "empty path",
			input:    "",
			expected: CategoryRoot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorize(tt.input)
			if result != tt.expected {
				t.Errorf("categorize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildIndex(t *testing.T) {
	tests := []struct {
		name      string
		templates []Template
		check     func(Index) bool
	}{
		{
			name: "basic index",
			templates: []Template{
				{Name: "Go", Path: "/go.gitignore"},
				{Name: "Python", Path: "/python.gitignore"},
			},
			check: func(idx Index) bool {
				if len(idx.List) != 2 {
					return false
				}
				if len(idx.ByName) != 2 {
					return false
				}
				_, ok1 := idx.ByName["go"]
				_, ok2 := idx.ByName["python"]
				return ok1 && ok2
			},
		},
		{
			name: "case insensitive keys",
			templates: []Template{
				{Name: "Go", Path: "/go.gitignore"},
				{Name: "GO", Path: "/GO.gitignore"},
			},
			check: func(idx Index) bool {
				// First one should win
				return len(idx.ByName) == 1
			},
		},
		{
			name:      "empty templates",
			templates: []Template{},
			check: func(idx Index) bool {
				return len(idx.List) == 0 && len(idx.ByName) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIndex(tt.templates)
			if !tt.check(result) {
				t.Errorf("BuildIndex() did not meet expected conditions")
			}
		})
	}
}

func TestFindTemplate(t *testing.T) {
	index := BuildIndex([]Template{
		{Name: "Go", Path: "/go.gitignore"},
		{Name: "Python", Path: "/python.gitignore"},
		{Name: "Node.js", Path: "/node.js.gitignore"},
	})

	tests := []struct {
		name     string
		search   string
		want     Template
		wantOk   bool
	}{
		{
			name:   "exact match",
			search: "Go",
			want:   Template{Name: "Go", Path: "/go.gitignore"},
			wantOk: true,
		},
		{
			name:   "case insensitive match",
			search: "go",
			want:   Template{Name: "Go", Path: "/go.gitignore"},
			wantOk: true,
		},
		{
			name:   "with suffix",
			search: "Go.gitignore",
			want:   Template{Name: "Go", Path: "/go.gitignore"},
			wantOk: true,
		},
		{
			name:   "case insensitive with suffix",
			search: "GO.GITIGNORE",
			want:   Template{Name: "Go", Path: "/go.gitignore"},
			wantOk: true,
		},
		{
			name:   "not found",
			search: "Rust",
			wantOk: false,
		},
		{
			name:   "empty search",
			search: "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindTemplate(index, tt.search)
			if ok != tt.wantOk {
				t.Errorf("FindTemplate() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if ok && got.Name != tt.want.Name {
				t.Errorf("FindTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeNameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "special characters preserved",
			input:    "Node.js.gitignore",
			expected: "Node.js",
		},
		{
			name:     "multiple dots",
			input:    "my.file.gitignore",
			expected: "my.file",
		},
		{
			name:     "uppercase suffix",
			input:    "Python.GITIGNORE",
			expected: "Python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
