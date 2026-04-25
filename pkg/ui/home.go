package ui

import tea "charm.land/bubbletea/v2"

type Model struct {
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	v := tea.NewView("<placeholder>")
	v.AltScreen = true // fullscreen
	return v
}
