package main

import (
	"charm.land/bubbles/v2/textinput"
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

type search struct {
	enabled bool // enabled determines whether searchbox is visible
	active  bool // active determines whether searchbox is actively receiving input
	input   textinput.Model
}

type tablePane struct {
	table  table.Model
	search search
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

	searchInput := textinput.New()
	searchInput.Prompt = "Search > "
	searchInput.CharLimit = 64
	searchInput.Placeholder = "type to search..."

	m := tablePane{
		table: t,
		search: search{
			input: searchInput,
		},
	}
	return &m
}

func (m *tablePane) Init() tea.Cmd {
	return nil
}

func (m *tablePane) Update(msg tea.Msg) (*tablePane, tea.Cmd) {
	var cmd tea.Cmd

	if m.search.active {
		cmd = m.handleSearching(msg)
	} else {
		cmd = m.handleNavigation(msg)
	}
	return m, cmd
}

// handleNavigation handles events when search is active.
func (m *tablePane) handleSearching(msg tea.Msg) (cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch txt := msg.String(); txt {
		case "esc", "enter":
			m.inactivateSearch()
			fallthrough
		default:
			var newState textinput.Model
			newState, cmd = m.search.input.Update(msg)
			if newState.Value() != m.search.input.Value() {
				// apply filter
			}
			m.search.input = newState
			// m.addSearch(msg)
		}
	default:
		m.search.input, cmd = m.search.input.Update(msg)
	}
	return
}

// handleNavigation handles events when search is not active.
func (m *tablePane) handleNavigation(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg := msg.String(); msg {
		case "/":
			cmds = append(cmds, m.openSearch())
		case "esc":
			m.resetSearch()
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return tea.Batch(append(cmds, cmd)...)
}

func (m *tablePane) openSearch() tea.Cmd {
	cmds := []tea.Cmd{}
	m.search.active = true
	cmds = append(cmds, m.search.input.Focus())
	if m.search.enabled {
		return tea.Batch(cmds...)
	}
	cmds = append(cmds, func() tea.Msg { return textinput.Blink })
	m.search.enabled = true
	m.table.SetHeight(m.table.Height() - searchHeight)
	return tea.Batch(cmds...)
}

func (m *tablePane) inactivateSearch() {
	m.search.active = false
	m.search.input.Blur()
}

func (m *tablePane) resetSearch() {
	if !m.search.enabled {
		return
	}
	m.search.input.Reset()
	m.search.enabled = false
	m.search.active = false

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
	return lipgloss.NewStyle().PaddingTop(1).Render(m.search.input.View())
}

func (m *tablePane) applySize(height, width int) {
	searchBoxH := searchHeight
	if !m.search.enabled {
		searchBoxH = 0
	}
	m.table.SetHeight(height - searchBoxH)
	m.table.SetWidth(width)
	m.search.input.SetWidth(width)
	searchBox = searchBox.Width(width)
}
