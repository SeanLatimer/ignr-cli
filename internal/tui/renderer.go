package tui

import (
	"fmt"
	"strings"

	"go.seanlatimer.dev/ignr/internal/templates"
)

const (
	defaultListHeight = 10
)

type RenderState struct {
	Query     string
	Filtered  []templates.Template
	Cursor    int
	Selected  map[string]templates.Template
	Suggested map[string]bool
	Width     int
	Height    int
	Selection []templates.Template
}

func RenderUI(state RenderState) string {
	width := state.Width
	if width <= 0 {
		width = 80
	}

	lines := []string{}

	selectedLine := renderSelectedItems(state.Selection, width)
	if selectedLine != "" {
		lines = append(lines, selectedLine, "")
	}

	lines = append(lines, renderSearchBox(state, width)...)
	return strings.Join(lines, "\n")
}

func renderSelectedItems(selected []templates.Template, width int) string {
	if len(selected) == 0 {
		return ""
	}

	parts := make([]string, 0, len(selected))
	for _, t := range selected {
		parts = append(parts, displayName(t))
	}
	line := "Selected: " + strings.Join(parts, ", ")
	return truncateToWidth(getStyles().SelectedStyle.Render(line), width)
}

func renderSearchBox(state RenderState, width int) []string {
	queryLine := "Search: " + state.Query
	queryLine = truncateToWidth(getStyles().SearchInputStyle.Render(queryLine), width)

	list := renderList(state.Filtered, state.Cursor, state.Selected, state.Suggested, width)
	footer := renderFooter(width)

	lines := []string{queryLine}
	lines = append(lines, list...)
	lines = append(lines, footer)
	return lines
}

func renderList(items []templates.Template, cursor int, selected map[string]templates.Template, suggested map[string]bool, width int) []string {
	limit := defaultListHeight
	if len(items) < limit {
		limit = len(items)
	}

	lines := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		item := items[i]
		cursorMark := " "
		if i == cursor {
			cursorMark = ">"
		}
		selectMark := " "
		if _, ok := selected[item.Path]; ok {
			selectMark = "✓"
		}
		suggestMark := " "
		if suggested != nil && suggested[item.Path] {
			suggestMark = "*"
		}

		line := fmt.Sprintf("%s [%s%s] %s (%s)", cursorMark, selectMark, suggestMark, displayName(item), item.Category)
		if i == cursor {
			line = getStyles().SelectedStyle.Render(line)
		}
		if suggestMark != " " {
			line = getStyles().SuggestedStyle.Render(line)
		}
		line = truncateToWidth(line, width)
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		lines = append(lines, getStyles().FooterStyle.Render("(no matches)"))
	}

	return lines
}

func displayName(item templates.Template) string {
	if item.Source == templates.SourceUser {
		return getStyles().UserBadgeStyle.Render("(User)") + " " + item.Name
	}
	return item.Name
}

func renderFooter(width int) string {
	footer := "Enter/Space toggle • Tab confirm • Esc cancel"
	return truncateToWidth(getStyles().FooterStyle.Render(footer), width)
}

func truncateToWidth(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	if width <= 1 {
		return text[:width]
	}
	return text[:width-1] + "…"
}
