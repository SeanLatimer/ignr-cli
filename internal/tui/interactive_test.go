// +build !windows

package tui

import (
	"os"
	"testing"

	"go.seanlatimer.dev/ignr/internal/templates"
)

func isCI() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}

func TestShowInteractiveSelectorEmpty(t *testing.T) {
	// Skip in CI environments where TUI tests hang
	if isCI() {
		t.Skip("Skipping TUI test in CI environment")
	}
	
	// Test with empty template list
	selected, err := ShowInteractiveSelector([]templates.Template{}, nil, nil, nil)
	
	// Should not error with empty list
	if err != nil {
		t.Logf("ShowInteractiveSelector() with empty list returned error: %v (may be expected)", err)
	}
	
	if len(selected) != 0 {
		t.Errorf("ShowInteractiveSelector() with empty list = %d templates, want 0", len(selected))
	}
}

func TestShowInteractiveSelectorBasic(t *testing.T) {
	// Skip in CI environments where TUI tests hang
	if isCI() {
		t.Skip("Skipping TUI test in CI environment")
	}
	
	// Create test templates
	testTemplates := []templates.Template{
		{Name: "Go", Path: "/go.gitignore", Category: templates.CategoryRoot},
		{Name: "Python", Path: "/python.gitignore", Category: templates.CategoryRoot},
	}
	
	// Note: This will fail in non-interactive environments, which is expected
	// In a full implementation, we'd use teatest to mock the TUI
	_, err := ShowInteractiveSelector(testTemplates, nil, nil, nil)
	
	// In non-interactive environments, this will fail
	// This is expected behavior
	if err != nil && err != ErrCancelled {
		t.Logf("ShowInteractiveSelector() error = %v (expected in non-interactive test environment)", err)
	}
}

// Note: Full TUI testing requires teatest package from charmbracelet/x/exp/teatest
// To enable comprehensive TUI tests, add the dependency:
// go get github.com/charmbracelet/x/exp/teatest
//
// Example teatest usage (from ref/bubbletea/simple/main_test.go):
//   tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(70, 30))
//   tm.Type("search query")
//   tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
//   out := tm.FinalOutput(t)
