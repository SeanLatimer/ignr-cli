package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"go.seanlatimer.dev/ignr/internal/cache"
	"go.seanlatimer.dev/ignr/internal/config"
	"go.seanlatimer.dev/ignr/internal/presets"
	"go.seanlatimer.dev/ignr/internal/templates"
)

type viewModel interface {
	tea.Model
	Title() string
}

type presetAppState struct {
	presets   []presets.Preset
	templates []templates.Template
	index     templates.Index
}

type presetAppModel struct {
	stack     []viewModel
	width     int
	height    int
	state     *presetAppState
}

type pushViewMsg struct {
	view viewModel
}

type popViewMsg struct{}

type refreshPresetsMsg struct {
	successMessage string
}

type quitAppMsg struct{}

func ShowPresetApp() error {
	app, err := newPresetAppModel()
	if err != nil {
		return err
	}
	program := tea.NewProgram(app)
	_, err = program.Run()
	return err
}

func newPresetAppModel() (presetAppModel, error) {
	presetList, err := presets.ListPresets()
	if err != nil {
		return presetAppModel{}, err
	}

	items, err := loadAllTemplates()
	if err != nil {
		return presetAppModel{}, err
	}

	index := templates.BuildIndex(items)
	state := &presetAppState{
		presets:   presetList,
		templates: items,
		index:     index,
	}
	root := newUnifiedPresetListView(state)

	return presetAppModel{
		stack:     []viewModel{root},
		state:     state,
	}, nil
}

func loadAllTemplates() ([]templates.Template, error) {
	cachePath, err := cache.InitializeCache()
	if err != nil {
		return nil, err
	}

	items, err := templates.DiscoverTemplates(cachePath)
	if err != nil {
		return nil, err
	}

	userPath, err := config.GetUserTemplatePath()
	if err != nil {
		return nil, err
	}
	userItems, err := templates.DiscoverUserTemplates(userPath)
	if err != nil {
		return nil, err
	}

	return append(items, userItems...), nil
}

func (m presetAppModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m presetAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		return m, nil
	case pushViewMsg:
		m.stack = append(m.stack, msg.view)
		return m, nil
	case popViewMsg:
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
			// Send refresh message to the new current view
			return m, func() tea.Msg { return refreshPresetsMsg{successMessage: ""} }
		}
		return m, tea.Quit
	case quitAppMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	current := m.currentView()
	if current == nil {
		return m, tea.Quit
	}
	updated, cmd := current.Update(msg)
	if view, ok := updated.(viewModel); ok {
		m.stack[len(m.stack)-1] = view
		return m, cmd
	}
	return m, cmd
}

func (m presetAppModel) View() tea.View {
	current := m.currentView()
	if current == nil {
		v := tea.NewView("")
		v.SetContent("No view available")
		return v
	}
	content := ""
	if provider, ok := current.(interface{ Content() string }); ok {
		content = provider.Content()
	}
	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	v := tea.NewView("")
	v.SetContent(content)
	v.AltScreen = true
	v.WindowTitle = fmt.Sprintf("Preset Management • %s", current.Title())
	return v
}

func (m presetAppModel) currentView() viewModel {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

func pushView(view viewModel) tea.Cmd {
	return func() tea.Msg {
		return pushViewMsg{view: view}
	}
}

func popView() tea.Cmd {
	return func() tea.Msg {
		return popViewMsg{}
	}
}

func quitApp() tea.Cmd {
	return func() tea.Msg {
		return quitAppMsg{}
	}
}
