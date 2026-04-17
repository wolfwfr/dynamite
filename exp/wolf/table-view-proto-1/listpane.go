package main

import tea "charm.land/bubbletea/v2"

type listPane struct {
}

func newListPane() listPane {
	m := listPane{}
	return m
}

func (m listPane) Init() tea.Cmd {
	return nil
}

func (m listPane) Update(tea.Msg) (listPane, tea.Cmd) {
	return m, nil
}

func (m listPane) View() string {
	s := "This is going to be a list pane"
	return s
}
