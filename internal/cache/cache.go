package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultCacheDirName  = "cache"
	defaultRepoDirName   = "github-gitignore"
	defaultRepoCloneURL  = "https://github.com/github/gitignore.git"
	defaultConfigDirName = "ignr"
)

type Status struct {
	Initialized bool
	Path        string
	HeadCommit  string
}

func GetCachePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	return filepath.Join(base, defaultConfigDirName, defaultCacheDirName, defaultRepoDirName), nil
}

func IsCacheInitialized() (bool, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return false, err
	}
	info, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}

	gitDir := filepath.Join(cachePath, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return true, nil
	}
	return false, nil
}

func InitializeCache() (string, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return "", err
	}

	initialized, err := IsCacheInitialized()
	if err != nil {
		return "", err
	}
	if initialized {
		return cachePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	if err := CloneRepo(defaultRepoCloneURL, cachePath); err != nil {
		return "", err
	}

	return cachePath, nil
}

func UpdateCache() (string, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return "", err
	}

	initialized, err := IsCacheInitialized()
	if err != nil {
		return "", err
	}
	if !initialized {
		return "", fmt.Errorf("cache not initialized; run init or generate first")
	}

	if err := PullRepo(cachePath); err != nil {
		return "", err
	}

	return cachePath, nil
}

func GetStatus() (Status, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return Status{}, err
	}

	initialized, err := IsCacheInitialized()
	if err != nil {
		return Status{}, err
	}

	status := Status{
		Initialized: initialized,
		Path:        cachePath,
	}

	if initialized {
		head, err := GetHeadCommit(cachePath)
		if err != nil {
			return Status{}, err
		}
		status.HeadCommit = head
	}

	return status, nil
}
