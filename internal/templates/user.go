package templates

import (
	"os"
	"strings"
)

func DiscoverUserTemplates(userPath string) ([]Template, error) {
	if strings.TrimSpace(userPath) == "" {
		return nil, nil
	}

	if _, err := os.Stat(userPath); err != nil {
		if os.IsNotExist(err) {
			return []Template{}, nil
		}
		return nil, err
	}

	return discoverTemplates(userPath, SourceUser, func(string) Category {
		return CategoryUser
	})
}
