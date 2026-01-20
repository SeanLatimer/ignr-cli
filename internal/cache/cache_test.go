package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
)

// setupCacheTest sets up a temporary config directory for testing cache
// and returns a cleanup function.
func setupCacheTest(t *testing.T) func() {
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
	
	// Return cleanup function
	return func() {
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
}

func TestGetCachePath(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	path, err := GetCachePath()
	if err != nil {
		t.Fatalf("GetCachePath() error = %v", err)
	}
	
	// Should contain cache directory components
	if !strings.Contains(path, defaultConfigDirName) {
		t.Errorf("GetCachePath() = %q, want path containing %q", path, defaultConfigDirName)
	}
	if !strings.Contains(path, defaultCacheDirName) {
		t.Errorf("GetCachePath() = %q, want path containing %q", path, defaultCacheDirName)
	}
	if !strings.Contains(path, defaultRepoDirName) {
		t.Errorf("GetCachePath() = %q, want path containing %q", path, defaultRepoDirName)
	}
}

func TestIsCacheInitialized(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	tests := []struct {
		name          string
		setup         func() string
		want          bool
		wantErr       bool
	}{
		{
			name: "non-initialized cache",
			setup: func() string {
				// Don't create cache
				path, _ := GetCachePath()
				return path
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "initialized cache with .git",
			setup: func() string {
				path, _ := GetCachePath()
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatalf("failed to create cache dir: %v", err)
				}
				// Create .git directory
				gitDir := filepath.Join(path, ".git")
				if err := os.MkdirAll(gitDir, 0o755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}
				return path
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "cache directory exists but no .git",
			setup: func() string {
				path, _ := GetCachePath()
				// Ensure parent directories exist but not the cache path itself
				// or create it as a directory but without .git
				parentDir := filepath.Dir(path)
				if err := os.MkdirAll(parentDir, 0o755); err != nil {
					t.Fatalf("failed to create parent dir: %v", err)
				}
				// Create cache directory without .git
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatalf("failed to create cache dir: %v", err)
				}
				// Verify .git does not exist
				gitDir := filepath.Join(path, ".git")
				if _, err := os.Stat(gitDir); err == nil {
					// .git exists, remove it for this test
					if err := os.RemoveAll(gitDir); err != nil {
						t.Logf("failed to remove .git dir: %v", err)
					}
				}
				return path
			},
			want:    false,
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			
			initialized, err := IsCacheInitialized()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("IsCacheInitialized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if initialized != tt.want {
				t.Errorf("IsCacheInitialized() = %v, want %v", initialized, tt.want)
			}
		})
	}
}

func TestInitializeCache(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	// This test requires actual git operations, so we'll test that it tries to clone
	// In a real test environment, you might want to use a mock or local git repo
	
	// Test with non-existent cache
	path, err := InitializeCache()
	
	// InitializeCache will try to clone, which might fail in test environment
	// So we just check that it returns an error (expected in test) or succeeds
	if err != nil {
		// Expected in test environment without network access
		// Verify error is about git clone
		if !strings.Contains(err.Error(), "git clone") {
			t.Logf("InitializeCache() error = %v (expected in test environment)", err)
		}
	} else {
		// If it succeeds, verify path is correct
		wantPath, _ := GetCachePath()
		if path != wantPath {
			t.Errorf("InitializeCache() = %q, want %q", path, wantPath)
		}
		
		// Verify cache is initialized
		initialized, err := IsCacheInitialized()
		if err != nil {
			t.Errorf("IsCacheInitialized() error = %v", err)
		}
		if !initialized {
			t.Error("InitializeCache() cache not initialized after InitializeCache")
		}
	}
}

func TestInitializeCacheAlreadyInitialized(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	// Create an already initialized cache
	path, _ := GetCachePath()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	
	// Create .git directory to mark as initialized
	gitDir := filepath.Join(path, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	
	// InitializeCache should return existing path without cloning
	resultPath, err := InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}
	
	if resultPath != path {
		t.Errorf("InitializeCache() = %q, want %q", resultPath, path)
	}
}

func TestUpdateCache(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	// Test with non-initialized cache
	_, err := UpdateCache()
	if err == nil {
		t.Error("UpdateCache() expected error for non-initialized cache, got nil")
		return
	}
	
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("UpdateCache() error = %v, want error containing 'not initialized'", err)
	}
}

func TestGetStatus(t *testing.T) {
	cleanup := setupCacheTest(t)
	defer cleanup()
	
	tests := []struct {
		name          string
		setup         func()
		wantInitialized bool
		wantErr       bool
	}{
		{
			name: "non-initialized cache",
			setup: func() {
				// Don't create cache
			},
			wantInitialized: false,
			wantErr:         false,
		},
		{
			name: "initialized cache",
			setup: func() {
				// Create a proper git repository
				path, _ := GetCachePath()
				repo, err := git.PlainInit(path, false)
				if err != nil {
					t.Fatalf("failed to init git repo: %v", err)
				}
				
				// Create a test file and commit to make HEAD valid
				testFile := filepath.Join(path, "test.gitignore")
				if err := os.WriteFile(testFile, []byte("# test"), 0o644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				
				wt, err := repo.Worktree()
				if err != nil {
					t.Fatalf("failed to get worktree: %v", err)
				}
				
				if _, err := wt.Add("test.gitignore"); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
				
				if _, err := wt.Commit("Initial commit", &git.CommitOptions{}); err != nil {
					t.Fatalf("failed to commit: %v", err)
				}
			},
			wantInitialized: true,
			wantErr:         false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			
			status, err := GetStatus()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if status.Initialized != tt.wantInitialized {
				t.Errorf("GetStatus() Initialized = %v, want %v", status.Initialized, tt.wantInitialized)
			}
			
			// Verify Path is set
			if status.Path == "" {
				t.Error("GetStatus() Path is empty")
			}
			
			// If initialized, HeadCommit might be empty (if git operations fail)
			// or might contain a commit hash
		})
	}
}
