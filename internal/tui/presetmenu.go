package tui

import (
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
)

type presetMenuItem struct {
	label string
}

func (i presetMenuItem) Title() string { return i.label }
func (i presetMenuItem) Description() string { return "" }
func (i presetMenuItem) FilterValue() string { return i.label }

type presetMenuModel struct {
	list      list.Model
	cancelled bool
}

func ShowPresetMenu() (string, error) {
	items := []list.Item{
		presetMenuItem{label: "Create new preset"},
		presetMenuItem{label: "Edit existing preset"},
		presetMenuItem{label: "Delete preset"},
		presetMenuItem{label: "List presets"},
		presetMenuItem{label: "Use preset"},
	}

	l := list.New(items, presetMenuDelegate{}, 0, 0)
	l.SetSize(60, len(items)+2)
	l.Title = "Preset Management"
	l.SetShowTitle(true)
	// Styles will be initialized when background color is detected
	// Note: This will be updated in Update() when BackgroundColorMsg arrives
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)

	model := presetMenuModel{list: l}
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return "", err
	}

	final := result.(presetMenuModel)
	if final.cancelled {
		return "", ErrCancelled
	}
	selected, ok := final.list.SelectedItem().(presetMenuItem)
	if !ok {
		return "", ErrCancelled
	}
	return selected.label, nil
}

func (m presetMenuModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m presetMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		// Update list styles now that styles are available
		m.list.Styles.Title = getStyles().SelectedStyle
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		width := msg.Width - 2
		height := msg.Height - 2
		if width < 10 {
			width = 10
		}
		if height < len(m.list.Items())+2 {
			height = len(m.list.Items()) + 2
		}
		m.list.SetSize(width, height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m presetMenuModel) View() tea.View {
	v := tea.NewView("")
	v.SetContent(getStyles().BorderStyle.Render(m.list.View()))
	return v
}

type presetMenuDelegate struct{}

func (d presetMenuDelegate) Height() int { return 1 }
func (d presetMenuDelegate) Spacing() int { return 0 }
func (d presetMenuDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d presetMenuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	menuItem, ok := item.(presetMenuItem)
	if !ok {
		return
	}
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}
	line := fmt.Sprintf("%s %s", cursor, menuItem.label)
	if index == m.Index() {
		line = getStyles().SelectedStyle.Render(line)
	}
	fmt.Fprint(w, line)
}
