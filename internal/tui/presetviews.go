package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"go.seanlatimer.dev/ignr/internal/config"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
)

// View constructors for preset management TUI.
func newCreateNameView(state *presetAppState) viewModel {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Preset name"
	input.Focus()
	return createNameView{
		state: state,
		input: input,
	}
}
func newCreateTemplatesView(state *presetAppState, name string) viewModel {
	return newTemplateSelectView(state, name, nil)
}
func newEditTemplatesView(state *presetAppState, preset presets.Preset) viewModel {
	return newTemplateSelectView(state, "", &preset)
}

type templateSelectView struct {
	state    *presetAppState
	selector selectorModel
	name     string
	preset   *presets.Preset
	err      string
}

func newTemplateSelectView(state *presetAppState, name string, preset *presets.Preset) templateSelectView {
	preselected := []string{}
	if preset != nil {
		preselected = preset.Templates
	}
	presetItems, presetLookup := buildPresetItems(nil)
	selected, selectedOrder, suggested := buildSelections(state.index, preselected, nil)
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search templates..."
	input.SetWidth(60)
	input.Blur() // Start unfocused so navigation works immediately

	l := list.New(templateListItems(state.templates, selected, suggested), templateListDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetSize(60, defaultListHeight+2)

	selector := selectorModel{
		all:           state.templates,
		filtered:      append(presetItems, state.templates...),
		searchInput:   input,
		list:          l,
		selected:      selected,
		selectedOrder: selectedOrder,
		presetItems:   presetItems,
		presetLookup:  presetLookup,
		index:         state.index,
		suggested:     suggested,
	}

	return templateSelectView{
		state:    state,
		selector: selector,
		name:     name,
		preset:   preset,
	}
}

func (t templateSelectView) Title() string {
	if t.preset != nil {
		return "Edit Templates"
	}
	return "Create Templates"
}

func (t templateSelectView) Init() tea.Cmd { return nil }

func (t templateSelectView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()

		switch keyStr {
		case "ctrl+c":
			return t, popView()
		case "esc":
			// Layered escape: unfocus -> clear -> pop view
			if t.selector.searchInput.Focused() {
				t.selector.searchInput.Blur()
				return t, nil
			}
			if t.selector.searchInput.Value() != "" {
				t.selector.searchInput.SetValue("")
				t.selector.applyFilter()
				return t, nil
			}
			return t, popView()
		case "tab", "ctrl+enter":
			if len(t.selector.selectedOrder) == 0 {
				t.err = "Select at least one template"
				return t, nil
			}
			templateNames := make([]string, 0, len(t.selector.selectedOrder))
			for _, tmpl := range t.selector.selectedOrder {
				templateNames = append(templateNames, tmpl.Name)
			}
			if t.preset == nil {
				if err := presets.CreatePreset(t.name, templateNames); err != nil {
					t.err = err.Error()
					return t, nil
				}
			} else {
				key := t.preset.Key
				if strings.TrimSpace(key) == "" {
					key = t.preset.Name
				}
				if err := presets.EditPreset(key, templateNames); err != nil {
					t.err = err.Error()
					return t, nil
				}
			}
			presetList, err := presets.ListPresets()
			if err == nil {
				t.state.presets = presetList
			}
			// Pop back to main view and show success message
			// Need to pop twice: once for templateSelectView, once for createNameView
			var successMsg string
			if t.preset == nil {
				successMsg = fmt.Sprintf("Created preset %q", t.name)
				// Creating new: pop twice to get back to list view
				return t, tea.Batch(popView(), popView(), func() tea.Msg {
					return refreshPresetsMsg{successMessage: successMsg}
				})
			} else {
				successMsg = fmt.Sprintf("Updated preset %q", t.preset.Name)
				// Editing: pop once to get back to list view
				return t, tea.Batch(popView(), func() tea.Msg {
					return refreshPresetsMsg{successMessage: successMsg}
				})
			}
		}
	}

	updated, cmd := t.selector.Update(msg)
	if sel, ok := updated.(selectorModel); ok {
		t.selector = sel
	}
	return t, cmd
}

func (t templateSelectView) View() tea.View {
	t.selector.errMessage = t.err
	v := tea.NewView("")
	v.SetContent(t.selector.Content())
	return v
}

func (t templateSelectView) Content() string {
	t.selector.errMessage = t.err
	return t.selector.Content()
}

type createNameView struct {
	state *presetAppState
	input textinput.Model
	err   string
}

func (c createNameView) Title() string { return "Create" }
func (c createNameView) Init() tea.Cmd { return nil }

func (c createNameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return c, popView()
		case "enter":
			name := strings.TrimSpace(c.input.Value())
			if name == "" {
				c.err = "Name is required"
				return c, nil
			}
			key := presets.SluggifyName(name)
			if presetKeyExists(stateKeys(c.state), key) {
				c.err = fmt.Sprintf("Key already exists: %s", key)
				return c, nil
			}
			return c, pushView(newCreateTemplatesView(c.state, name))
		}
	}

	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return c, cmd
}

func (c createNameView) View() tea.View {
	v := tea.NewView("")
	v.SetContent(c.Content())
	return v
}

func (c createNameView) Content() string {
	keyPreview := presets.SluggifyName(strings.TrimSpace(c.input.Value()))
	lines := []string{
		getStyles().SelectedStyle.Render("Create Preset"),
		getStyles().SearchInputStyle.Render(c.input.View()),
		getStyles().SubtleStyle.Render(fmt.Sprintf("Key: %s", keyPreview)),
	}
	if c.err != "" {
		lines = append(lines, getStyles().ErrorStyle.Render(c.err))
	}
	lines = append(lines, getStyles().FooterStyle.Render("Enter continue • Esc back"))
	return getStyles().BorderStyle.Render(strings.Join(lines, "\n"))
}

func stateKeys(state *presetAppState) []string {
	keys := make([]string, 0, len(state.presets))
	for _, preset := range state.presets {
		key := preset.Key
		if strings.TrimSpace(key) == "" {
			key = presets.SluggifyName(preset.Name)
		}
		keys = append(keys, key)
	}
	return keys
}

func presetKeyExists(keys []string, key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, existing := range keys {
		if strings.ToLower(existing) == key {
			return true
		}
	}
	return false
}

func (u *unifiedPresetListView) checkAndUsePreset(preset presets.Preset) (bool, string) {
	index := templates.BuildIndex(u.state.templates)
	selected := make([]templates.Template, 0, len(preset.Templates))
	for _, name := range preset.Templates {
		t, ok := templates.FindTemplate(index, name)
		if !ok {
			u.errMessage = fmt.Sprintf("template not found: %s", name)
			return false, ""
		}
		selected = append(selected, t)
	}

	target, err := resolveOutputPath()
	if err != nil {
		u.errMessage = err.Error()
		return false, ""
	}

	// Check if file exists and show confirmation
	if _, err := os.Stat(target); err == nil {
		// Need confirmation - set up state
		u.overwriteConfirm = &overwriteConfirmState{
			path:       target,
			templates:  selected,
			presetName: preset.Name,
		}
		return false, "" // Not done yet, waiting for confirmation
	}

	// No confirmation needed, proceed immediately
	return u.executePreset(target, selected, preset.Name), target
}

func (u *unifiedPresetListView) executePreset(target string, selected []templates.Template, presetName string) bool {
	loaded, err := templates.LoadTemplates(selected)
	if err != nil {
		u.errMessage = err.Error()
		return false
	}

	content := templates.MergeTemplates(loaded, templates.MergeOptions{
		Deduplicate: true,
		AddHeader:   true,
		Generator:   "ignr",
		Version:     "dev",
		Timestamp:   time.Now(),
	})
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		u.errMessage = err.Error()
		return false
	}

	u.statusMessage = fmt.Sprintf("Generated %s with preset %q", target, presetName)
	u.errMessage = ""
	return true
}

func resolveOutputPath() (string, error) {
	cfg, err := config.LoadConfig()
	if err == nil && strings.TrimSpace(cfg.DefaultOutput) != "" {
		return cfg.DefaultOutput, nil
	}
	return filepath.Join(".", ".gitignore"), nil
}

// --- Unified Preset List View ---

// presetTemplatesView displays the templates in a preset
type presetTemplatesView struct {
	state  *presetAppState
	preset presets.Preset
	width  int
	height int
}

func newPresetTemplatesView(state *presetAppState, preset presets.Preset) viewModel {
	return presetTemplatesView{
		state:  state,
		preset: preset,
	}
}

func (v presetTemplatesView) Title() string {
	return fmt.Sprintf("View: %s", v.preset.Name)
}

func (v presetTemplatesView) Init() tea.Cmd {
	return nil
}

func (v presetTemplatesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "enter", "q":
			return v, popView()
		}
	}
	return v, nil
}

func (v presetTemplatesView) View() tea.View {
	tv := tea.NewView("")
	tv.SetContent(v.Content())
	return tv
}

func (v presetTemplatesView) Content() string {
	contentWidth := v.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 80 {
		contentWidth = 80
	}

	fixedWidth := lipgloss.NewStyle().Width(contentWidth)

	var lines []string

	// Title
	key := v.preset.Key
	if strings.TrimSpace(key) == "" {
		key = presets.SluggifyName(v.preset.Name)
	}
	title := fmt.Sprintf("%s [%s]", v.preset.Name, key)
	lines = append(lines, fixedWidth.Render(getStyles().SelectedStyle.Render(title)))
	lines = append(lines, "")

	// Templates
	if len(v.preset.Templates) == 0 {
		lines = append(lines, fixedWidth.Render(getStyles().SubtleStyle.Render("  No templates")))
	} else {
		for _, templateName := range v.preset.Templates {
			// Find the template to get category info
			t, ok := templates.FindTemplate(v.state.index, templateName)
			var templateDisplay string
			if ok {
				templateDisplay = displayName(t)
			} else {
				templateDisplay = templateName
			}
			line := fmt.Sprintf("  • %s", templateDisplay)
			lines = append(lines, fixedWidth.Render(line))
		}
	}

	lines = append(lines, "")
	lines = append(lines, fixedWidth.Render(getStyles().FooterStyle.Render("Esc/Enter/Q back")))

	// Wrap in border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getStyles().Subtle).
		Width(contentWidth+4).
		Padding(0, 1)

	return containerStyle.Render(strings.Join(lines, "\n"))
}

// createPresetItem is a special list item for "[Create New Preset...]"
type createPresetItem struct{}

func (i createPresetItem) Title() string       { return "[Create New Preset...]" }
func (i createPresetItem) Description() string { return "" }
func (i createPresetItem) FilterValue() string { return "create new preset" }

// presetListItem wraps a preset for the unified list
type presetListItem struct {
	preset presets.Preset
}

func (i presetListItem) Title() string { return i.preset.Name }
func (i presetListItem) Description() string {
	return fmt.Sprintf("%d templates", len(i.preset.Templates))
}
func (i presetListItem) FilterValue() string {
	return i.preset.Name + " " + i.preset.Key
}

// unifiedPresetListView is the main view for preset management
type unifiedPresetListView struct {
	state               *presetAppState
	list                list.Model
	searchInput         textinput.Model
	allPresets          []presets.Preset
	lastQuery           string
	deleteConfirmPreset *presets.Preset
	overwriteConfirm    *overwriteConfirmState
	errMessage          string
	statusMessage       string
	width               int
	height              int
}

type overwriteConfirmState struct {
	path       string
	templates  []templates.Template
	presetName string
}

func newUnifiedPresetListView(state *presetAppState) unifiedPresetListView {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Type to search..."
	input.SetWidth(40)
	input.Blur() // Start unfocused so hotkeys work immediately

	items := buildUnifiedListItems(state.presets)
	listHeight := len(items)
	if listHeight > 15 {
		listHeight = 15
	}
	if listHeight < 3 {
		listHeight = 3
	}

	l := list.New(items, unifiedPresetDelegate{}, 50, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	return unifiedPresetListView{
		state:       state,
		list:        l,
		searchInput: input,
		allPresets:  state.presets,
	}
}

func buildUnifiedListItems(presetList []presets.Preset) []list.Item {
	items := make([]list.Item, 0, len(presetList)+1)
	items = append(items, createPresetItem{})
	for _, preset := range presetList {
		items = append(items, presetListItem{preset: preset})
	}
	return items
}

func (u unifiedPresetListView) Title() string { return "Presets" }
func (u unifiedPresetListView) Init() tea.Cmd { return nil }

func (u unifiedPresetListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshPresetsMsg:
		// Sync allPresets with state.presets and refresh the list
		u.allPresets = u.state.presets
		u.applyFilter()
		// Set success message if provided, otherwise clear messages
		if msg.successMessage != "" {
			u.statusMessage = msg.successMessage
			u.errMessage = ""
		} else {
			u.statusMessage = ""
			u.errMessage = ""
		}
		return u, nil

	case tea.WindowSizeMsg:
		u.width = msg.Width
		u.height = msg.Height

		// Calculate content width (matching Content() logic)
		contentWidth := msg.Width - 4
		if contentWidth < 40 {
			contentWidth = 40
		}
		if contentWidth > 80 {
			contentWidth = 80
		}

		// Update search input width
		u.searchInput.SetWidth(contentWidth - 4) // Account for "/ " prefix

		// Update list size
		listHeight := msg.Height - 9
		if listHeight < 5 {
			listHeight = 5
		}
		if listHeight > 20 {
			listHeight = 20
		}
		u.list.SetSize(contentWidth, listHeight)
		return u, nil

	case tea.KeyMsg:
		// Handle overwrite confirmation mode
		if u.overwriteConfirm != nil {
			return u.handleOverwriteConfirmation(msg)
		}
		// Handle delete confirmation mode
		if u.deleteConfirmPreset != nil {
			return u.handleDeleteConfirmation(msg)
		}

		keyStr := msg.String()

		// Global keys
		switch keyStr {
		case "ctrl+c":
			return u, popView()
		case "esc":
			// Layered escape: unfocus -> clear text -> exit
			if u.searchInput.Focused() {
				// Unfocus but keep the query for navigation
				u.searchInput.Blur()
				return u, nil
			}
			if u.searchInput.Value() != "" {
				// Clear search query
				u.searchInput.SetValue("")
				u.applyFilter()
				return u, nil
			}
			return u, popView()
		case "/":
			u.searchInput.Focus()
			return u, nil
		}

		// When search is focused, handle typing and enter
		if u.searchInput.Focused() {
			switch keyStr {
			case "enter":
				selected := u.list.SelectedItem()
				if _, ok := selected.(createPresetItem); ok {
					return u, pushView(newCreateNameView(u.state))
				}
				if preset := u.selectedPreset(); preset != nil {
					done, target := u.checkAndUsePreset(*preset)
					if done {
						// Successfully generated (no confirmation needed)
						return u, nil
					}
					if target != "" {
						// Error occurred
						return u, nil
					}
					// Waiting for confirmation
					return u, nil
				}
				return u, nil
			case "up", "k":
				var cmd tea.Cmd
				u.list, cmd = u.list.Update(msg)
				return u, cmd
			case "down", "j":
				var cmd tea.Cmd
				u.list, cmd = u.list.Update(msg)
				return u, cmd
			}
			// Let search input handle other keys
		} else {
			// Hotkeys when search not focused
			switch keyStr {
			case "c":
				return u, pushView(newCreateNameView(u.state))
			case "e":
				if preset := u.selectedPreset(); preset != nil {
					return u, pushView(newEditTemplatesView(u.state, *preset))
				}
				return u, nil
			case "d":
				if preset := u.selectedPreset(); preset != nil {
					u.deleteConfirmPreset = preset
					u.errMessage = ""
					return u, nil
				}
				return u, nil
			case "v":
				if preset := u.selectedPreset(); preset != nil {
					return u, pushView(newPresetTemplatesView(u.state, *preset))
				}
				return u, nil
			case "u", "enter":
				selected := u.list.SelectedItem()
				if _, ok := selected.(createPresetItem); ok {
					return u, pushView(newCreateNameView(u.state))
				}
				if preset := u.selectedPreset(); preset != nil {
					done, target := u.checkAndUsePreset(*preset)
					if done {
						// Successfully generated (no confirmation needed)
						return u, nil
					}
					if target != "" {
						// Error occurred
						return u, nil
					}
					// Waiting for confirmation
					return u, nil
				}
				return u, nil
			case "up", "k", "down", "j":
				var cmd tea.Cmd
				u.list, cmd = u.list.Update(msg)
				return u, cmd
			}
		}

		// Navigation keys for list
		switch keyStr {
		case "up", "k":
			var cmd tea.Cmd
			u.list, cmd = u.list.Update(msg)
			return u, cmd
		case "down", "j":
			var cmd tea.Cmd
			u.list, cmd = u.list.Update(msg)
			return u, cmd
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update search input only when focused
	if u.searchInput.Focused() {
		u.searchInput, cmd = u.searchInput.Update(msg)
		cmds = append(cmds, cmd)

		// Apply filter if query changed
		query := u.searchInput.Value()
		if query != u.lastQuery {
			u.lastQuery = query
			u.applyFilter()
		}
	}

	// Update list for navigation
	u.list, cmd = u.list.Update(msg)
	cmds = append(cmds, cmd)

	return u, tea.Batch(cmds...)
}

func (u unifiedPresetListView) handleOverwriteConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// User confirmed overwrite
		state := u.overwriteConfirm
		u.overwriteConfirm = nil
		// Execute the file write
		u.executePreset(state.path, state.templates, state.presetName)
		return u, nil
	case "n", "N", "esc", "ctrl+c":
		// User cancelled
		u.overwriteConfirm = nil
		return u, nil
	}
	return u, nil
}

func (u unifiedPresetListView) handleDeleteConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		preset := u.deleteConfirmPreset
		key := preset.Key
		if strings.TrimSpace(key) == "" {
			key = preset.Name
		}
		if err := presets.DeletePreset(key); err != nil {
			u.errMessage = err.Error()
			u.deleteConfirmPreset = nil
			return u, nil
		}
		// Refresh presets
		presetList, err := presets.ListPresets()
		if err == nil {
			u.state.presets = presetList
			u.allPresets = presetList
		}
		u.applyFilter()
		u.statusMessage = fmt.Sprintf("Deleted preset %q", preset.Name)
		u.deleteConfirmPreset = nil
		return u, nil
	case "n", "N", "esc", "ctrl+c":
		u.deleteConfirmPreset = nil
		return u, nil
	}
	return u, nil
}

func (u *unifiedPresetListView) applyFilter() {
	query := u.searchInput.Value()
	filtered := filterPresets(query, u.allPresets)
	items := buildUnifiedListItems(filtered)
	u.list.SetItems(items)

	// Adjust list height based on items
	listHeight := len(items)
	if listHeight > 15 {
		listHeight = 15
	}
	if listHeight < 3 {
		listHeight = 3
	}
	u.list.SetSize(50, listHeight)
}

func (u unifiedPresetListView) selectedPreset() *presets.Preset {
	selected := u.list.SelectedItem()
	if item, ok := selected.(presetListItem); ok {
		return &item.preset
	}
	return nil
}

func (u unifiedPresetListView) isCreateItemSelected() bool {
	_, ok := u.list.SelectedItem().(createPresetItem)
	return ok
}

func (u unifiedPresetListView) View() tea.View {
	v := tea.NewView("")
	v.SetContent(u.Content())
	return v
}

func (u unifiedPresetListView) Content() string {
	// Use terminal dimensions, with sensible minimums
	// Default to reasonable dimensions if not set
	width := u.width
	if width == 0 {
		width = 80
	}
	height := u.height
	if height == 0 {
		height = 24
	}

	contentWidth := width - 4 // Account for border and padding
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 80 {
		contentWidth = 80 // Cap width for readability
	}

	// Calculate list height based on terminal height
	// Reserve: title(1) + blank(1) + search(1) + blank(1) + blank(1) + status(1) + footer(1) + border(2) = 9 lines
	listHeight := height - 9
	if listHeight < 5 {
		listHeight = 5
	}
	if listHeight > 20 {
		listHeight = 20 // Cap for usability
	}

	// Fixed-width style for consistent layout
	fixedWidth := lipgloss.NewStyle().Width(contentWidth)

	var lines []string

	// Title
	lines = append(lines, fixedWidth.Render(getStyles().SelectedStyle.Render("Preset Management")))
	lines = append(lines, "")

	// Search input with label and focus indicator
	var searchLine string
	if u.searchInput.Focused() {
		// Focused: show input with cursor
		searchLine = getStyles().SelectedStyle.Render("/ ") + getStyles().SearchInputStyle.Render(u.searchInput.View())
	} else if u.searchInput.Value() != "" {
		// Has query but unfocused: show the query
		searchLine = getStyles().SubtleStyle.Render("/ ") + getStyles().SearchInputStyle.Render(u.searchInput.Value())
	} else {
		// No query, unfocused: show hint
		searchLine = getStyles().SubtleStyle.Render("/ Press / to search")
	}
	lines = append(lines, fixedWidth.Render(searchLine))
	lines = append(lines, "")

	// List items with calculated height
	items := u.list.Items()
	selectedIdx := u.list.Index()
	listLines := make([]string, 0, listHeight)

	for i, item := range items {
		if len(listLines) >= listHeight {
			break
		}
		cursor := "  "
		if i == selectedIdx {
			cursor = "> "
		}

		var line string
		switch it := item.(type) {
		case createPresetItem:
			line = cursor + it.Title()
			if i == selectedIdx {
				line = getStyles().PresetBadgeStyle.Render(line)
			} else {
				line = getStyles().SubtleStyle.Render(line)
			}
		case presetListItem:
			line = fmt.Sprintf("%s%s (%d templates)", cursor, it.preset.Name, len(it.preset.Templates))
			if i == selectedIdx {
				line = getStyles().SelectedStyle.Render(line)
			}
		}
		listLines = append(listLines, fixedWidth.Render(line))
	}

	// Pad list to fixed height for stable layout
	for len(listLines) < listHeight {
		listLines = append(listLines, fixedWidth.Render(""))
	}

	lines = append(lines, listLines...)
	lines = append(lines, "")

	// Status line (always present for stable height)
	var statusLine string
	if u.overwriteConfirm != nil {
		statusLine = getStyles().WarningStyle.Render(fmt.Sprintf("Overwrite %s? (Y/N)", u.overwriteConfirm.path))
	} else if u.deleteConfirmPreset != nil {
		statusLine = getStyles().WarningStyle.Render(fmt.Sprintf("Delete preset %q? (Y/N)", u.deleteConfirmPreset.Name))
	} else if u.errMessage != "" {
		// Truncate error message to fit content width
		msg := u.errMessage
		if len(msg) > contentWidth {
			msg = msg[:contentWidth-1] + "…"
		}
		statusLine = getStyles().ErrorStyle.Render(msg)
	} else if u.statusMessage != "" {
		// Truncate success message to fit content width
		msg := u.statusMessage
		if len(msg) > contentWidth {
			msg = msg[:contentWidth-1] + "…"
		}
		statusLine = getStyles().SuccessStyle.Render(msg)
	}
	lines = append(lines, fixedWidth.Render(statusLine))

	// Footer
	footer := u.buildFooter()
	lines = append(lines, fixedWidth.Render(getStyles().FooterStyle.Render(footer)))

	// Wrap in border with calculated width
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getStyles().Subtle).
		Width(contentWidth+4). // Account for border + padding
		Padding(0, 1)

	return containerStyle.Render(strings.Join(lines, "\n"))
}

func (u unifiedPresetListView) buildFooter() string {
	if u.overwriteConfirm != nil {
		return "Y confirm • N cancel"
	}
	if u.deleteConfirmPreset != nil {
		return "Y confirm • N cancel"
	}
	if u.searchInput.Focused() {
		return "Type to filter • ↑↓ navigate • Esc done"
	}
	// When there's a search query active
	if u.searchInput.Value() != "" {
		return "↑↓ navigate • Enter use • / edit search • Esc clear"
	}
	if u.isCreateItemSelected() {
		return "C/Enter create • / search • Esc exit"
	}
	return "C new • E edit • D del • V view • U/Enter use • / search"
}

// unifiedPresetDelegate renders items in the unified preset list
type unifiedPresetDelegate struct{}

func (d unifiedPresetDelegate) Height() int                               { return 1 }
func (d unifiedPresetDelegate) Spacing() int                              { return 0 }
func (d unifiedPresetDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d unifiedPresetDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	var line string
	switch item := listItem.(type) {
	case createPresetItem:
		line = cursor + item.Title()
		if index == m.Index() {
			line = getStyles().PresetBadgeStyle.Render(line)
		} else {
			line = getStyles().SubtleStyle.Render(line)
		}
	case presetListItem:
		line = fmt.Sprintf("%s%s (%d templates)", cursor, item.preset.Name, len(item.preset.Templates))
		if index == m.Index() {
			line = getStyles().SelectedStyle.Render(line)
	}
}

	_, _ = fmt.Fprint(w, line)
}
