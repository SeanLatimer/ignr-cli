package cli

import (
	"bytes"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	opts := &Options{}
	cmd := NewRootCommand(opts)
	
	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}
	
	if cmd.Use != "ignr" {
		t.Errorf("NewRootCommand() Use = %q, want %q", cmd.Use, "ignr")
	}
	
	// Check that subcommands are registered
	commands := cmd.Commands()
	commandNames := make(map[string]bool)
	for _, c := range commands {
		commandNames[c.Name()] = true
	}
	
	expectedCommands := []string{"list", "search", "generate", "preset", "update"}
	for _, name := range expectedCommands {
		if !commandNames[name] {
			t.Errorf("NewRootCommand() missing subcommand: %q", name)
		}
	}
}

func TestRootCommandVersion(t *testing.T) {
	opts := &Options{}
	cmd := NewRootCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	cmd.SetArgs([]string{"--version"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command --version error = %v", err)
	}
	
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("ignr")) {
		t.Error("root command --version output missing 'ignr'")
	}
}

func TestRootCommandHelp(t *testing.T) {
	opts := &Options{}
	cmd := NewRootCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command --help error = %v", err)
	}
	
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Offline-first gitignore generator")) {
		t.Error("root command --help output missing description")
	}
}

func TestRootCommandInvalidCommand(t *testing.T) {
	opts := &Options{}
	cmd := NewRootCommand(opts)
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	cmd.SetArgs([]string{"invalid-command"})
	
	err := cmd.Execute()
	
	// Should error for invalid command
	if err == nil {
		t.Error("root command expected error for invalid command, got nil")
	}
}

func TestRootCommandFlags(t *testing.T) {
	opts := &Options{}
	cmd := NewRootCommand(opts)
	
	// Test config flag
	cmd.SetArgs([]string{"--config", "/test/config.json", "list"})
	
	// Command should parse flags without error (even if command fails)
	err := cmd.Execute()
	// List might fail due to missing cache, but flags should be parsed
	if err != nil {
		// Verify the error is not about flag parsing
		if opts.ConfigPath != "/test/config.json" {
			t.Errorf("root command ConfigPath = %q, want %q", opts.ConfigPath, "/test/config.json")
		}
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute function works (it will fail due to missing cache,
	// but we can verify it doesn't panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()
	
	// This will likely fail due to missing cache, but should not panic
	_ = Execute()
}
