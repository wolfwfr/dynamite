package main

import tea "charm.land/bubbletea/v2"

type detailsPane struct{}

func (m *detailsPane) Init() tea.Cmd {
	return nil
}

func (m *detailsPane) Update(tea.Msg) (*detailsPane, tea.Cmd) {
	return m, nil
}

func (m *detailsPane) View() string {
	return "This is the future details-pane"
}
