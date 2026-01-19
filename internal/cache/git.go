package cache

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func CloneRepo(repoURL, dest string) error {
	_, err := runGit("", "clone", "--depth", "1", repoURL, dest)
	return err
}

func PullRepo(repoPath string) error {
	_, err := runGit(repoPath, "pull", "--ff-only")
	return err
}

func GetHeadCommit(repoPath string) (string, error) {
	out, err := runGit(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func runGit(workingDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}

	return stdout.String(), nil
}
