package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"go.seanlatimer.dev/ignr/internal/presets"
)

type presetNameModel struct {
	prompt       string
	input        textinput.Model
	existingKeys map[string]struct{}
	allowExisting bool
	errMessage   string
	done         bool
	cancelled    bool
}

func ShowPresetNameInput(prompt string, existingKeys []string, allowExisting bool) (string, error) {
	input := textinput.New()
	input.Prompt = ""
	input.Focus()

	existing := make(map[string]struct{}, len(existingKeys))
	for _, key := range existingKeys {
		existing[strings.ToLower(key)] = struct{}{}
	}

	model := presetNameModel{
		prompt:        prompt,
		input:         input,
		existingKeys:  existing,
		allowExisting: allowExisting,
	}

	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return "", err
	}

	final := result.(presetNameModel)
	if final.cancelled {
		return "", ErrCancelled
	}
	return strings.TrimSpace(final.input.Value()), nil
}

func (m presetNameModel) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m presetNameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Initialize global styles instance (compat package handles adaptation)
		appStyles = newStyles()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if err := m.validate(value); err != nil {
				m.errMessage = err.Error()
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func isTextInputNavKey(key string) bool {
	switch key {
	case "backspace", "delete", "left", "right", "home", "end":
		return true
	default:
		return false
	}
}

func (m presetNameModel) View() tea.View {
	value := strings.TrimSpace(m.input.Value())
	keyPreview := presets.SluggifyName(value)
	lines := []string{
		getStyles().SelectedStyle.Render(m.prompt),
		getStyles().SearchInputStyle.Render(m.input.View()),
		getStyles().SubtleStyle.Render(fmt.Sprintf("Key: %s", keyPreview)),
	}
	if m.errMessage != "" {
		lines = append(lines, getStyles().ErrorStyle.Render(fmt.Sprintf("Error: %s", m.errMessage)))
	}
	lines = append(lines, getStyles().FooterStyle.Render("Enter confirm • Esc cancel"))
	v := tea.NewView("")
	v.SetContent(getStyles().BorderStyle.Render(strings.Join(lines, "\n")))
	return v
}

func (m presetNameModel) validate(value string) error {
	if value == "" {
		return fmt.Errorf("name is required")
	}
	key := presets.SluggifyName(value)
	if !m.allowExisting {
		if _, exists := m.existingKeys[strings.ToLower(key)]; exists {
			return fmt.Errorf("preset key already exists: %s", key)
		}
	}
	return nil
}
