package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wolfwfr/dynamite/exp/wolf/table-view-proto-2/internal/table"
)

const (
	searchHeight int = 2
)

var (
	searchBox = lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4F4F4F")).
		PaddingLeft(2).
		Height(searchHeight)
)

type tablePane struct {
	table  table.Model
	search struct {
		enabled bool // enabled determines whether searchbox is visible
		active  bool // active determines whether searchbox is actively receiving input
		input   string
	}
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg := msg.String(); msg {
		case "/":
			m.openSearch()
		case "esc":
			if m.search.active {
				m.inactivateSearch()
			} else {
				m.resetSearch()
			}
		case "enter":
			if m.search.active {
				m.inactivateSearch()
			}
		default:
			if m.search.active {
				m.addSearch(msg)
			}
		}
	}
	if !m.search.active {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// TODO: temporary for testing; replace with text input bubble
func (m *tablePane) addSearch(s string) {
	if len(s) > 1 && s != "backspace" {
		return
	}
	if s == "backspace" && len(m.search.input) > 0 {
		m.search.input = m.search.input[:len(m.search.input)-1]
	}
	if s == "backspace" {
		return
	}
	m.search.input = m.search.input + s
}

func (m *tablePane) openSearch() {
	m.search.active = true
	if m.search.enabled {
		return
	}
	m.search.enabled = true
	m.table.SetHeight(m.table.Height() - searchHeight)
}

func (m *tablePane) inactivateSearch() {
	m.search.active = false
}

func (m *tablePane) resetSearch() {
	if !m.search.enabled {
		return
	}
	m.search.enabled = false
	m.search.active = false
	m.search.input = ""

	m.table.SetHeight(m.table.Height() + searchHeight)
}

func (m *tablePane) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.table.View(),
		m.renderSearchBox(),
	)
}

func (m *tablePane) renderSearchBox() string {
	if !m.search.enabled {
		return ""
	}
	inactive := " (inactive)"
	if m.search.active {
		inactive = ""
	}
	return searchBox.Render(fmt.Sprintf("Search%s > ", inactive) + m.search.input)
}

func (m *tablePane) applySize(height, width int) {
	searchBoxH := searchHeight
	if !m.search.enabled {
		searchBoxH = 0
	}
	m.table.SetHeight(height - searchBoxH)
	m.table.SetWidth(width)
	searchBox = searchBox.Width(width)
}
