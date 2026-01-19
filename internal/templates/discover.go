// Package templates provides discovery and merging for gitignore templates.
package templates

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Category string

const (
	CategoryRoot      Category = "root"
	CategoryGlobal    Category = "Global"
	CategoryCommunity Category = "community"
	CategoryUser      Category = "user"
)

type TemplateSource string

const (
	SourceCache TemplateSource = "cache"
	SourceUser  TemplateSource = "user"
)

type Template struct {
	Name     string
	Category Category
	Path     string
	Source   TemplateSource
}

type Index struct {
	ByName map[string]Template
	List   []Template
}

func DiscoverTemplates(cachePath string) ([]Template, error) {
	return discoverTemplates(cachePath, SourceCache, categorize)
}

func discoverTemplates(rootPath string, source TemplateSource, categorizePath func(string) Category) ([]Template, error) {
	var templates []Template

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".gitignore") {
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}

		category := categorizePath(rel)
		name := normalizeName(d.Name())

		templates = append(templates, Template{
			Name:     name,
			Category: category,
			Path:     path,
			Source:   source,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return templates, nil
}

func BuildIndex(templates []Template) Index {
	index := Index{
		ByName: make(map[string]Template, len(templates)),
		List:   templates,
	}

	for _, t := range templates {
		key := strings.ToLower(t.Name)
		if _, exists := index.ByName[key]; !exists {
			index.ByName[key] = t
		}
	}

	return index
}

func FindTemplate(index Index, name string) (Template, bool) {
	key := strings.ToLower(normalizeName(name))
	t, ok := index.ByName[key]
	return t, ok
}

func normalizeName(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".gitignore") {
		return name[:len(name)-len(".gitignore")]
	}
	return name
}

func categorize(relPath string) Category {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) == 0 {
		return CategoryRoot
	}
	switch strings.ToLower(parts[0]) {
	case "global":
		return CategoryGlobal
	case "community":
		return CategoryCommunity
	default:
		return CategoryRoot
	}
}
