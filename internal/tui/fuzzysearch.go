package tui

import (
	"github.com/sahilm/fuzzy"
	"go.seanlatimer.dev/ignr/internal/templates"
)

func FilterTemplates(query string, items []templates.Template) []templates.Template {
	if query == "" {
		return items
	}

	names := make([]string, 0, len(items))
	for _, t := range items {
		names = append(names, t.Name)
	}

	matches := fuzzy.FindFrom(query, stringSource(names))
	filtered := make([]templates.Template, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, items[match.Index])
	}

	return filtered
}

type stringSource []string

func (s stringSource) Len() int {
	return len(s)
}

func (s stringSource) String(i int) string {
	return s[i]
}
