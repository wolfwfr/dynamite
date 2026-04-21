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
	f       FilterFunc
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
			f:     DefaultFilter,
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

	if _, ok := msg.(FilterMatchesMsg); ok || m.search.active {
		cmd = m.handleSearching(msg)
	} else {
		cmd = m.handleNavigation(msg)
	}
	// TODO: check current cursor position; if necessary async start loading
	// details & throw event when done, for async info see
	// [docs](https://github.com/charmbracelet/bubbletea/tree/main/tutorials/commands/).
	// IMPORTANT: to prevent many calls, add a configurable debounce (e.g. 50ms)
	// before making network call and cancel on next navigation event.
	// IMPORTANT: for the table-pane showing dynamo-db items, all items are
	// already in-memory, the details pane only shows the same item in a JSON
	// format; no need to make a new call.
	return m, cmd
}

// handleNavigation handles events when search is active.
func (m *tablePane) handleSearching(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch txt := msg.String(); txt {
		case "esc", "enter":
			m.inactivateSearch()
			fallthrough
		default:
			newQuery, cmd := m.search.input.Update(msg)
			cmds = append(cmds, cmd)
			if newQuery.Value() != m.search.input.Value() { // if new query
				cmds = append(cmds, m.Search(newQuery.Value()))
			}
			m.search.input = newQuery
		}
	case FilterMatchesMsg:
		if m.search.input.Value() == "" {
			m.table.ResetVirtualRows()
			break
		}
		rows := m.table.Rows()
		filtered := make([]table.Row, len(msg))
		for i, match := range msg {
			filtered[i] = rows[match.index]
		}
		m.table.SetVirtualRows(filtered)
	default:
		m.search.input, cmd = m.search.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

// search applies the current search query to all rows in the table and returns
// a cmd that the tea-framework can asynchronously process. Upon completion, it
// returns a tea.Msg containing the items remaining post-filter, which can be
// handled by replacing the table rows with the filtered items.
func (m *tablePane) Search(query string) tea.Cmd {
	rows := rowStrings(m.table.Rows())
	f := m.search.f
	// OPTIM: cancel on next text input for performance
	return func() tea.Msg { // will execute async
		ranks := f(query, rows)
		filtered := make([]filteredItem, len(ranks))
		for i, r := range ranks {
			item := filteredItem{
				index:   r.Index,
				item:    Item{Content: rows[r.Index]},
				matches: r.MatchedIndexes,
			}
			filtered[i] = item
		}
		return FilterMatchesMsg(filtered)
	}
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
	m.table.ResetVirtualRows()
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

func rowStrings(rows []table.Row) []string {
	res := make([]string, len(rows))
	for i := range rows {
		res[i] = rows[i].String()
	}
	return res
}
