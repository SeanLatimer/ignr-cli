// Package tui provides the interactive terminal UI for template selection.
package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
)

var ErrCancelled = errors.New("selection cancelled")

type selectorModel struct {
	all            []templates.Template
	filtered       []templates.Template
	query          string
	cursor         int
	selected       map[string]templates.Template
	selectedOrder  []templates.Template
	width          int
	height         int
	done           bool
	cancelled      bool
	presets        []presets.Preset
	presetItems    []templates.Template
	presetLookup   map[string]presets.Preset
	showingPresets bool
	index          templates.Index
	suggested      map[string]bool
}

func ShowInteractiveSelector(items []templates.Template, presetList []presets.Preset, suggestedNames []string) ([]templates.Template, error) {
	presetItems, presetLookup := buildPresetItems(presetList)
	index := templates.BuildIndex(items)
	selected, selectedOrder, suggested := buildSuggestedSelections(index, suggestedNames)
	model := selectorModel{
		all:           items,
		filtered:      append(presetItems, items...),
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
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "tab", "ctrl+enter", "ctrl+j":
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "p":
			m.showingPresets = !m.showingPresets
			m.applyFilter()
			m.cursor = clampCursor(m.cursor, len(m.filtered))
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.applyFilter()
				m.cursor = clampCursor(m.cursor, len(m.filtered))
			}
		case "enter", " ":
			m.toggleSelection()
		default:
			if msg.Type == tea.KeyRunes {
				m.query += msg.String()
				m.applyFilter()
				m.cursor = clampCursor(m.cursor, len(m.filtered))
			}
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	state := RenderState{
		Query:     m.query,
		Filtered:  m.filtered,
		Cursor:    m.cursor,
		Selected:  m.selected,
		Suggested: m.suggested,
		Width:     m.width,
		Height:    m.height,
		Selection: m.selectedOrder,
	}
	return RenderUI(state)
}

func (m *selectorModel) toggleSelection() {
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return
	}
	item := m.filtered[m.cursor]
	if preset, ok := m.presetLookup[item.Path]; ok {
		m.applyPresetSelection(preset)
		return
	}
	if _, exists := m.selected[item.Path]; exists {
		delete(m.selected, item.Path)
		m.selectedOrder = removeSelected(m.selectedOrder, item.Path)
		return
	}
	m.selected[item.Path] = item
	m.selectedOrder = append(m.selectedOrder, item)
}

func (m *selectorModel) applyFilter() {
	presetFiltered := FilterTemplates(m.query, m.presetItems)
	if m.showingPresets {
		m.filtered = presetFiltered
		return
	}
	templateFiltered := FilterTemplates(m.query, m.all)
	m.filtered = append(presetFiltered, templateFiltered...)
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

func buildSuggestedSelections(index templates.Index, suggestedNames []string) (map[string]templates.Template, []templates.Template, map[string]bool) {
	selected := make(map[string]templates.Template)
	selectedOrder := make([]templates.Template, 0, len(suggestedNames))
	suggested := make(map[string]bool)

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
