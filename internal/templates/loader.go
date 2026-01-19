package templates

import (
	"fmt"
	"os"
)

type LoadedTemplate struct {
	Template Template
	Content  string
}

func LoadTemplate(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}
	return string(data), nil
}

func LoadTemplates(templates []Template) ([]LoadedTemplate, error) {
	loaded := make([]LoadedTemplate, 0, len(templates))
	for _, t := range templates {
		content, err := LoadTemplate(t.Path)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, LoadedTemplate{
			Template: t,
			Content:  content,
		})
	}
	return loaded, nil
}
