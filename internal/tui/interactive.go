// Package tui provides the interactive terminal UI for template selection.
package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
)

var ErrCancelled = errors.New("selection cancelled")

type selectorModel struct {
	all            []templates.Template
	filtered       []templates.Template
	searchInput    textinput.Model
	list           list.Model
	lastQuery      string
	selected       map[string]templates.Template
	selectedOrder  []templates.Template
	width          int
	height         int
	done           bool
	cancelled      bool
	errMessage     string
	presets        []presets.Preset
	presetItems    []templates.Template
	presetLookup   map[string]presets.Preset
	showingPresets bool
	index          templates.Index
	suggested      map[string]bool
}

func ShowInteractiveSelector(items []templates.Template, presetList []presets.Preset, preselectedNames []string, suggestedNames []string) ([]templates.Template, error) {
	presetItems, presetLookup := buildPresetItems(presetList)
	index := templates.BuildIndex(items)
	selected, selectedOrder, suggested := buildSelections(index, preselectedNames, suggestedNames)
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search templates..."
	input.SetWidth(60)
	input.Blur() // Start unfocused so hotkeys work immediately

	l := list.New(templateListItemsWithPresets(append(presetItems, items...), selected, suggested, presetLookup, index), templateListDelegate{}, 0, 0)
	l.SetSize(60, defaultListHeight+2)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	model := selectorModel{
		all:           items,
		filtered:      append(presetItems, items...),
		searchInput:   input,
		list:          l,
		selected:      selected,
		selectedOrder: selectedOrder,
		presets:       presetList,
		presetItems:   presetItems,
		presetLookup:  presetLookup,
		index:         index,
		suggested:     suggested,
	}

	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return nil, err
	}

	final := result.(selectorModel)
	if final.cancelled {
		return nil, ErrCancelled
	}
	return final.selectedOrder, nil
}

func (m selectorModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate dimensions
		contentWidth := msg.Width - 4
		if contentWidth < 40 {
			contentWidth = 40
		}
		if contentWidth > 80 {
			contentWidth = 80
		}

		listHeight := msg.Height - 10
		if listHeight < 5 {
			listHeight = 5
		}
		if listHeight > 20 {
			listHeight = 20
		}

		m.searchInput.SetWidth(contentWidth - 4)
		m.list.SetSize(contentWidth, listHeight)

	case tea.KeyMsg:
		keyStr := msg.String()
		key := msg.Key()

		// Handle space separately - check both String() and Key().Text
		if keyStr == " " || key.Text == " " {
			if !m.searchInput.Focused() {
				m.toggleSelection()
				return m, nil
			}
			// If search is focused, let search input handle it
		}

		switch keyStr {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			// Layered escape: unfocus -> clear -> cancel
			if m.searchInput.Focused() {
				m.searchInput.Blur()
				return m, nil
			}
			if m.searchInput.Value() != "" {
				m.searchInput.SetValue("")
				m.applyFilter()
				return m, nil
			}
			m.cancelled = true
			return m, tea.Quit
		case "tab", "ctrl+enter", "ctrl+j":
			m.done = true
			return m, tea.Quit
		case "/":
			m.searchInput.Focus()
			return m, nil
		case "p":
			if len(m.presetItems) > 0 && !m.searchInput.Focused() {
				m.showingPresets = !m.showingPresets
				m.applyFilter()
				return m, nil
			}
		case "enter":
			if !m.searchInput.Focused() {
				m.toggleSelection()
				return m, nil
			}
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
	if m.searchInput.Focused() {
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)

		query := m.searchInput.Value()
		if query != m.lastQuery {
			m.lastQuery = query
			m.applyFilter()
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m selectorModel) View() tea.View {
	v := tea.NewView("")
	v.SetContent(m.Content())
	return v
}

func (m selectorModel) Content() string {
	// Calculate content width
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
	lines = append(lines, fixedWidth.Render(getStyles().SelectedStyle.Render("Template Selection")))
	lines = append(lines, "")

	// Selected items summary (always show to prevent layout shifts)
	var selectedLine string
	if len(m.selectedOrder) > 0 {
		selectedLine = renderSelectedItems(m.selectedOrder, contentWidth)
	} else {
		selectedLine = "Selected: None"
	}
	lines = append(lines, fixedWidth.Render(getStyles().SubtleStyle.Render(selectedLine)))
	lines = append(lines, "")

	// Search input
	var searchLine string
	if m.searchInput.Focused() {
		searchLine = getStyles().SelectedStyle.Render("/ ") + getStyles().SearchInputStyle.Render(m.searchInput.View())
	} else if m.searchInput.Value() != "" {
		searchLine = getStyles().SubtleStyle.Render("/ ") + getStyles().SearchInputStyle.Render(m.searchInput.Value())
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
	if m.searchInput.Focused() {
		footer = "Type to filter • ↑↓ navigate • Esc done"
	} else if m.searchInput.Value() != "" {
		footer = "Enter/Space toggle • Tab confirm • / edit search • Esc clear"
	} else {
		footer = "Enter/Space toggle • Tab confirm • / search • Esc cancel"
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

func (m *selectorModel) toggleSelection() {
	current := m.list.SelectedItem()
	if current == nil {
		return
	}
	item := current.(templateListItem).template
	if preset, ok := m.presetLookup[item.Path]; ok {
		m.applyPresetSelection(preset)
		m.list.SetItems(templateListItemsWithPresets(m.filtered, m.selected, m.suggested, m.presetLookup, m.index))
		return
	}
	if _, exists := m.selected[item.Path]; exists {
		delete(m.selected, item.Path)
		m.selectedOrder = removeSelected(m.selectedOrder, item.Path)
		m.list.SetItems(templateListItemsWithPresets(m.filtered, m.selected, m.suggested, m.presetLookup, m.index))
		return
	}
	m.selected[item.Path] = item
	m.selectedOrder = append(m.selectedOrder, item)
	m.list.SetItems(templateListItemsWithPresets(m.filtered, m.selected, m.suggested, m.presetLookup, m.index))
}

func (m *selectorModel) applyFilter() {
	query := m.searchInput.Value()
	presetFiltered := FilterTemplates(query, m.presetItems)
	if m.showingPresets {
		m.filtered = presetFiltered
		m.list.SetItems(templateListItemsWithPresets(m.filtered, m.selected, m.suggested, m.presetLookup, m.index))
		return
	}
	templateFiltered := FilterTemplates(query, m.all)
	m.filtered = append(presetFiltered, templateFiltered...)
	m.list.SetItems(templateListItemsWithPresets(m.filtered, m.selected, m.suggested, m.presetLookup, m.index))
}

func (m *selectorModel) applyPresetSelection(preset presets.Preset) {
	selectedTemplates := make([]templates.Template, 0, len(preset.Templates))
	allSelected := len(preset.Templates) > 0
	for _, name := range preset.Templates {
		t, ok := templates.FindTemplate(m.index, name)
		if !ok {
			continue
		}
		selectedTemplates = append(selectedTemplates, t)
		if _, exists := m.selected[t.Path]; !exists {
			allSelected = false
		}
	}

	if allSelected {
		for _, t := range selectedTemplates {
			delete(m.selected, t.Path)
			m.selectedOrder = removeSelected(m.selectedOrder, t.Path)
		}
		return
	}

	for _, t := range selectedTemplates {
		if _, exists := m.selected[t.Path]; exists {
			continue
		}
		m.selected[t.Path] = t
		m.selectedOrder = append(m.selectedOrder, t)
	}
}

func buildPresetItems(presetList []presets.Preset) ([]templates.Template, map[string]presets.Preset) {
	items := make([]templates.Template, 0, len(presetList))
	lookup := make(map[string]presets.Preset, len(presetList))
	for _, preset := range presetList {
		path := "preset:" + preset.Name
		items = append(items, templates.Template{
			Name:     fmt.Sprintf("[Preset] %s (%d templates)", preset.Name, len(preset.Templates)),
			Category: templates.Category("preset"),
			Path:     path,
		})
		lookup[path] = preset
	}
	return items, lookup
}

func buildSelections(index templates.Index, preselectedNames []string, suggestedNames []string) (map[string]templates.Template, []templates.Template, map[string]bool) {
	selected := make(map[string]templates.Template)
	selectedOrder := make([]templates.Template, 0, len(preselectedNames))
	suggested := make(map[string]bool)

	for _, name := range preselectedNames {
		t, ok := templates.FindTemplate(index, name)
		if !ok {
			continue
		}
		if _, exists := selected[t.Path]; exists {
			continue
		}
		selected[t.Path] = t
		selectedOrder = append(selectedOrder, t)
	}

	for _, name := range suggestedNames {
		t, ok := templates.FindTemplate(index, name)
		if !ok {
			continue
		}
		suggested[t.Path] = true
		if _, exists := selected[t.Path]; exists {
			continue
		}
		selected[t.Path] = t
		selectedOrder = append(selectedOrder, t)
	}

	return selected, selectedOrder, suggested
}

func removeSelected(items []templates.Template, path string) []templates.Template {
	for i, item := range items {
		if item.Path == path {
			return append(items[:i], items[i+1:]...)
		}
	}
	return items
}

func clampCursor(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	if cursor < 0 {
		return 0
	}
	return cursor
}

type templateListItem struct {
	template templates.Template
	selected bool
	suggested bool
}

func (i templateListItem) Title() string { return displayName(i.template) }
func (i templateListItem) Description() string { return string(i.template.Category) }
func (i templateListItem) FilterValue() string { return displayName(i.template) }

type templateListDelegate struct{}

func (d templateListDelegate) Height() int { return 1 }
func (d templateListDelegate) Spacing() int { return 0 }
func (d templateListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d templateListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(templateListItem)
	if !ok {
		return
	}
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}
	selectMark := " "
	if item.selected {
		selectMark = "✓"
	}
	suggestMark := " "
	if item.suggested {
		suggestMark = "*"
	}
	line := fmt.Sprintf("%s [%s%s] %s", cursor, selectMark, suggestMark, displayName(item.template))
	if index == m.Index() {
		line = getStyles().SelectedStyle.Render(line)
	}
	fmt.Fprint(w, line)
}

func templateListItems(items []templates.Template, selected map[string]templates.Template, suggested map[string]bool) []list.Item {
	return templateListItemsWithPresets(items, selected, suggested, nil, templates.Index{})
}

func templateListItemsWithPresets(items []templates.Template, selected map[string]templates.Template, suggested map[string]bool, presetLookup map[string]presets.Preset, index templates.Index) []list.Item {
	results := make([]list.Item, 0, len(items))
	for _, item := range items {
		isSelected := false
		
		// Check if it's a preset - if so, check if all its templates are selected
		if len(presetLookup) > 0 && len(index.ByName) > 0 {
			if preset, ok := presetLookup[item.Path]; ok {
				// It's a preset - check if all its templates are selected
				allSelected := len(preset.Templates) > 0
				for _, templateName := range preset.Templates {
					t, ok := templates.FindTemplate(index, templateName)
					if !ok {
						allSelected = false
						break
					}
					if _, exists := selected[t.Path]; !exists {
						allSelected = false
						break
					}
				}
				isSelected = allSelected
			} else {
				// Regular template
				_, isSelected = selected[item.Path]
			}
		} else {
			// No preset lookup - treat as regular template
			_, isSelected = selected[item.Path]
		}
		
		isSuggested := false
		if suggested != nil {
			isSuggested = suggested[item.Path]
		}
		results = append(results, templateListItem{
			template:  item,
			selected:  isSelected,
			suggested: isSuggested,
		})
	}
	return results
}

func joinBlocks(blocks ...string) string {
	parts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if block == "" {
			continue
		}
		parts = append(parts, block)
	}
	return strings.Join(parts, "\n")
}
