package tui

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/sahilm/fuzzy"
	"go.seanlatimer.dev/ignr/internal/presets"
)

type presetSelectorItem struct {
	preset presets.Preset
}

func (i presetSelectorItem) Title() string { return i.preset.Name }
func (i presetSelectorItem) Description() string {
	return fmt.Sprintf("%d templates", len(i.preset.Templates))
}
func (i presetSelectorItem) FilterValue() string {
	return i.preset.Name + " " + i.preset.Key
}

type presetSelectorModel struct {
	all        []presets.Preset
	list       list.Model
	input      textinput.Model
	lastQuery  string
	cancelled  bool
	errMessage string
	width      int
	height     int
}

func ShowPresetSelector(items []presets.Preset) (presets.Preset, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search presets..."
	input.SetWidth(50)
	input.Blur() // Start unfocused so navigation works immediately

	l := list.New(presetItems(items), presetSelectorDelegate{}, 0, 0)
	l.SetSize(50, defaultListHeight+2)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	model := presetSelectorModel{
		all:   items,
		list:  l,
		input: input,
	}

	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return presets.Preset{}, err
	}

	final := result.(presetSelectorModel)
	if final.cancelled {
		return presets.Preset{}, ErrCancelled
	}
	selected, ok := final.list.SelectedItem().(presetSelectorItem)
	if !ok {
		return presets.Preset{}, fmt.Errorf("no preset selected")
	}
	return selected.preset, nil
}

func (m presetSelectorModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m presetSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentWidth := msg.Width - 4
		if contentWidth < 40 {
			contentWidth = 40
		}
		if contentWidth > 60 {
			contentWidth = 60
		}

		listHeight := msg.Height - 8
		if listHeight < 5 {
			listHeight = 5
		}
		if listHeight > 15 {
			listHeight = 15
		}

		m.input.SetWidth(contentWidth - 4)
		m.list.SetSize(contentWidth, listHeight)

	case tea.KeyMsg:
		keyStr := msg.String()

		switch keyStr {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			// Layered escape: unfocus -> clear -> cancel
			if m.input.Focused() {
				m.input.Blur()
				return m, nil
			}
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.list.SetItems(presetItems(m.all))
				return m, nil
			}
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		case "/":
			m.input.Focus()
			return m, nil
		}

		// Navigation works regardless of focus
		if keyStr == "up" || keyStr == "k" || keyStr == "down" || keyStr == "j" {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update search input only when focused
	if m.input.Focused() {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)

		query := m.input.Value()
		if query != m.lastQuery {
			m.lastQuery = query
			m.list.SetItems(presetItems(filterPresets(query, m.all)))
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m presetSelectorModel) View() tea.View {
	v := tea.NewView("")
	v.SetContent(m.Content())
	return v
}

func (m presetSelectorModel) Content() string {
	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 60 {
		contentWidth = 60
	}

	fixedWidth := lipgloss.NewStyle().Width(contentWidth)

	var lines []string

	// Title
	lines = append(lines, fixedWidth.Render(getStyles().SelectedStyle.Render("Select Preset")))
	lines = append(lines, "")

	// Search input
	var searchLine string
	if m.input.Focused() {
		searchLine = getStyles().SelectedStyle.Render("/ ") + getStyles().SearchInputStyle.Render(m.input.View())
	} else if m.input.Value() != "" {
		searchLine = getStyles().SubtleStyle.Render("/ ") + getStyles().SearchInputStyle.Render(m.input.Value())
	} else {
		searchLine = getStyles().SubtleStyle.Render("/ Press / to search")
	}
	lines = append(lines, fixedWidth.Render(searchLine))
	lines = append(lines, "")

	// List
	lines = append(lines, m.list.View())
	lines = append(lines, "")

	// Error message
	if m.errMessage != "" {
		lines = append(lines, fixedWidth.Render(getStyles().ErrorStyle.Render(m.errMessage)))
	}

	// Footer
	var footer string
	if m.input.Focused() {
		footer = "Type to filter • ↑↓ navigate • Esc done"
	} else if m.input.Value() != "" {
		footer = "Enter select • V view • / edit search • Esc clear"
	} else {
		footer = "Enter select • V view • / search • Esc cancel"
	}
	lines = append(lines, fixedWidth.Render(getStyles().FooterStyle.Render(footer)))

	// Wrap in border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getStyles().Subtle).
		Width(contentWidth + 4).
		Padding(0, 1)

	return containerStyle.Render(strings.Join(lines, "\n"))
}

type presetSelectorDelegate struct{}

func (d presetSelectorDelegate) Height() int { return 1 }
func (d presetSelectorDelegate) Spacing() int { return 0 }
func (d presetSelectorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d presetSelectorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(presetSelectorItem)
	if !ok {
		return
	}
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}
	line := fmt.Sprintf("%s %s (%d templates)", cursor, item.preset.Name, len(item.preset.Templates))
	if index == m.Index() {
		line = getStyles().SelectedStyle.Render(line)
	}
	fmt.Fprint(w, line)
}

func presetItems(items []presets.Preset) []list.Item {
	results := make([]list.Item, 0, len(items))
	for _, preset := range items {
		results = append(results, presetSelectorItem{preset: preset})
	}
	return results
}

func filterPresets(query string, items []presets.Preset) []presets.Preset {
	if query == "" {
		return items
	}

	entries := make([]string, 0, len(items))
	for _, preset := range items {
		entries = append(entries, preset.Name+" "+preset.Key)
	}

	matches := fuzzy.FindFrom(query, stringSource(entries))
	filtered := make([]presets.Preset, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, items[match.Index])
	}
	return filtered
}
