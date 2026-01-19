// Package cache provides git operations for cache management.
package cache

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

func CloneRepo(repoURL, dest string) error {
	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:           repoURL,
		Depth:         1,
		SingleBranch:  true,
		Progress:      nil,
	})
	if err != nil {
		return fmt.Errorf("git clone --depth 1 %s %s: %w", repoURL, dest, err)
	}
	return nil
}

func PullRepo(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("git pull --ff-only: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("git pull --ff-only: %w", err)
	}

	err = wt.Pull(&git.PullOptions{
		Depth: 1,
	})
	if err != nil {
		// NoErrAlreadyUpToDate is not actually an error, it means we're already up to date
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return fmt.Errorf("git pull --ff-only: %w", err)
	}
	return nil
}

func GetHeadCommit(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}

	return ref.Hash().String(), nil
}
