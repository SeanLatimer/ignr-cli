// Package testutil provides utilities for testing CLI commands.
package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.seanlatimer.dev/ignr/pkg/cli"
)

// ExecuteCommand executes a CLI command and captures its output.
// Returns stdout, stderr, exit code, and any execution error.
func ExecuteCommand(args []string, env map[string]string) (stdout, stderr string, exitCode int, err error) {
	// Build the command
	cmd := exec.Command("go", append([]string{"run", "./cmd/ignr"}, args...)...)
	
	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	
	runErr := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	
	if runErr != nil {
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
			err = runErr
		}
	}
	
	return stdout, stderr, exitCode, err
}

// ExecuteCommandInDir executes a CLI command in a specific directory.
func ExecuteCommandInDir(dir string, args []string, env map[string]string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.Command("go", append([]string{"run", "./cmd/ignr"}, args...)...)
	cmd.Dir = dir
	
	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	
	runErr := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	
	if runErr != nil {
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
			err = runErr
		}
	}
	
	return stdout, stderr, exitCode, err
}

// NewRootCommand creates a new root command for testing.
func NewRootCommand() *cli.Options {
	return &cli.Options{}
}

// MockTUI provides a way to mock TUI interactions for testing.
type MockTUI struct {
	SelectedTemplates []string
	Confirmed        bool
	PresetSelected   string
	PresetName       string
	Err              error
}

// CheckCommandOutput verifies that command output contains expected strings.
func CheckCommandOutput(output string, expected ...string) error {
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			return fmt.Errorf("output does not contain expected string: %q", exp)
		}
	}
	return nil
}
