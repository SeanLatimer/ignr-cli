package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"go.seanlatimer.dev/ignr/internal/templates"
)

type confirmModel struct {
	path        string
	templates   []templates.Template
	choice      bool
	done        bool
	cancelled   bool
	width       int
	height      int
	useAltScreen bool
}

func ConfirmOverwrite(path string, templates []templates.Template) (bool, error) {
	return ConfirmOverwriteWithOptions(path, templates, ConfirmOptions{
		UseAltScreen: true, // Default to alt screen for standalone use
	})
}

type ConfirmOptions struct {
	UseAltScreen bool
}

func ConfirmOverwriteWithOptions(path string, templates []templates.Template, opts ConfirmOptions) (bool, error) {
	model := confirmModel{
		path:        path,
		templates:   templates,
		useAltScreen: opts.UseAltScreen,
	}
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return false, err
	}
	final := result.(confirmModel)
	if final.cancelled {
		return false, nil
	}
	return final.choice, nil
}

func (m confirmModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		keyStr := strings.ToLower(msg.String())
		switch keyStr {
		case "y":
			m.choice = true
			m.done = true
			return m, tea.Quit
		case "n", "esc", "ctrl+c":
			m.choice = false
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.choice = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() tea.View {
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 80 {
		contentWidth = 80
	}

	fixedWidth := lipgloss.NewStyle().Width(contentWidth)

	var lines []string

	// Title
	lines = append(lines, fixedWidth.Render(getStyles().SelectedStyle.Render("Confirm Overwrite")))
	lines = append(lines, "")

	// File path
	lines = append(lines, fixedWidth.Render(fmt.Sprintf("Output file exists: %s", m.path)))
	lines = append(lines, "")

	// Templates being applied
	if len(m.templates) > 0 {
		lines = append(lines, fixedWidth.Render(getStyles().SubtleStyle.Render(fmt.Sprintf("Applying %d template(s):", len(m.templates)))))
		templateNames := make([]string, 0, len(m.templates))
		for _, tmpl := range m.templates {
			templateNames = append(templateNames, displayName(tmpl))
		}
		// Wrap template list across multiple lines
		templateList := strings.Join(templateNames, ", ")
		wrappedLines := wrapText(templateList, contentWidth-4, "  ")
		for _, line := range wrappedLines {
			lines = append(lines, fixedWidth.Render(line))
		}
		lines = append(lines, "")
	}

	// Question
	lines = append(lines, fixedWidth.Render("Overwrite? (y/N)"))

	// Footer
	lines = append(lines, fixedWidth.Render(getStyles().FooterStyle.Render("Y confirm • N cancel • Esc cancel")))

	// Wrap in border with AltScreen
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getStyles().Subtle).
		Width(contentWidth + 4).
		Padding(0, 1)

	content := containerStyle.Render(strings.Join(lines, "\n"))

	// Center content
	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	v := tea.NewView("")
	v.SetContent(content)
	v.AltScreen = m.useAltScreen
	v.WindowTitle = "Confirm Overwrite"
	return v
}

// wrapText wraps text to fit within the specified width, breaking at commas when possible
func wrapText(text string, width int, prefix string) []string {
	if width <= 0 {
		width = 40
	}
	if len(text) <= width-len(prefix) {
		return []string{prefix + text}
	}

	var lines []string
	var currentLine strings.Builder
	currentLine.WriteString(prefix)

	// Split by comma and space
	parts := strings.Split(text, ", ")
	for i, part := range parts {
		// For items after the first, we need to add a comma
		if i > 0 {
			// Check if adding ", " + part would exceed width
			if currentLine.Len()+len(", "+part) > width && currentLine.Len() > len(prefix) {
				// If just adding the comma fits, add it before breaking
				if currentLine.Len()+len(", ") <= width {
					currentLine.WriteString(", ")
				}
				// Current line already has content, break here
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(prefix)
			} else {
				// Add comma and space before the part on current line
				currentLine.WriteString(", ")
			}
		}
		// Add the part itself
		currentLine.WriteString(part)
	}

	if currentLine.Len() > len(prefix) {
		lines = append(lines, currentLine.String())
	}

	if len(lines) == 0 {
		return []string{prefix + text}
	}
	return lines
}
