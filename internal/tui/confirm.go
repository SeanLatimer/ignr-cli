package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	path      string
	choice    bool
	done      bool
	cancelled bool
}

func ConfirmOverwrite(path string) (bool, error) {
	model := confirmModel{path: path}
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
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.choice = true
			m.done = true
			return m, tea.Quit
		case "n", "N", "esc", "ctrl+c":
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

func (m confirmModel) View() string {
	return fmt.Sprintf("Output file exists:\n%s\n\nOverwrite? (y/N)", m.path)
}
