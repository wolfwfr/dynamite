package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wolfwfr/dynamite/exp/wolf/table-view-proto-2/internal/table"
)

type tablePane struct {
	table table.Model
}

func newTablePane(cols []table.Column, rows []table.Row) *tablePane {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := tablePane{
		table: t,
	}
	return &m
}

func (m *tablePane) Init() tea.Cmd {
	return nil
}

func (m *tablePane) Update(msg tea.Msg) (*tablePane, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle()

// var baseStyle = lipgloss.NewStyle().
// 	BorderStyle(lipgloss.NormalBorder()).
// 	BorderForeground(lipgloss.Color("240"))

func (m *tablePane) View() string {
	// return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
	return baseStyle.Render(m.table.View())
}

func (m *tablePane) applySize(height, width int) {
	m.table.SetHeight(height)
	m.table.SetWidth(width)
}
