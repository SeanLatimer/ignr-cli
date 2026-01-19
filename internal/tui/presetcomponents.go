package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func renderAppHeader(title string) string {
	return getStyles().SelectedStyle.Render(title)
}

func renderAppFooter(help string) string {
	return getStyles().FooterStyle.Render(help)
}

func renderAppLayout(header, body, footer string) tea.View {
	v := tea.NewView("")
	content := joinBlocks(header, body, footer)
	v.SetContent(getStyles().BorderStyle.Render(content))
	return v
}

func renderStatus(message string, isError bool) string {
	if message == "" {
		return ""
	}
	if isError {
		return getStyles().ErrorStyle.Render(fmt.Sprintf("Error: %s", message))
	}
	return getStyles().SubtleStyle.Render(message)
}
