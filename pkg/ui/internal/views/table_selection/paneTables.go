package tableselection

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
)

type tableSelectionPane struct {
	// top-level context
	ctx context.Context

	// cancel last call context (debounce)
	cancelLast  func()
	debounceDur time.Duration

	// standard timeout
	stdTO time.Duration

	// shared config
	config *appconfig.Config

	// errorText
	err error

	// pane's view window
	window struct {
		width  int
		height int
	}

	// fuzzy finding
	search *search.SearchBox

	content *table.Model

	tables           []string
	filteredTables   []int // indices referring to tables
	lastTableDetails int   // index
}

func newTableSelectionPane(ctx context.Context, config *appconfig.Config) *tableSelectionPane {
	t := table.New(
		table.WithColumns([]table.Column{{Title: "table-name", Width: 64}}),
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

	p := &tableSelectionPane{
		ctx:         ctx,
		cancelLast:  func() {}, // noop on init
		debounceDur: 50 * time.Millisecond,
		config:      config,
		stdTO:       5 * time.Second,
		// TODO: add table feature to hide header
		content: t,
	}
	p.search = search.NewSearchBox(
		search.SearchCallbacks{
			ToSearch: func() []string {
				return table.Rows(p.content.Rows()).ToStrings()
			},
			EmptyInput: func() tea.Cmd {
				p.content.ResetVirtualRows()
				return nil
			},
			Results: func(results []search.FilteredItem) {
				p.filteredTables = make([]int, len(results))
				rows := p.content.Rows()
				filtered := make([]table.Row, len(results))
				for i, match := range results {
					filtered[i] = rows[match.Index]
					p.filteredTables[i] = match.Index
				}
				p.content.SetVirtualRows(filtered)
			},
			Reset: func(searchHeight int) {
				p.filteredTables = make([]int, 0)
				p.content.ResetVirtualRows()
				p.content.SetHeight(p.content.Height() + searchHeight)
			},
			ViewBoxOpens: func(searchHeight int) {
				p.content.SetHeight(p.content.Height() - searchHeight)
			},
		},
	)
	return p
}

func (m *tableSelectionPane) cleanSlate() {
	m.err = nil
}

func (m *tableSelectionPane) Init() tea.Cmd {
	m.cleanSlate()
	if client := m.config.Client; client != nil {
		// TODO: async
		ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
		defer cc()
		tables, err := dynamodb.ListTables(client, ctx)
		if err != nil {
			m.err = err
			return nil
		}
		m.tables = tables
		rows := make([]table.Row, len(tables))
		for i := range tables {
			rows[i] = table.Row([]string{tables[i]})
		}
		m.content.SetRows(rows)
	}
	return m.MaybePreviewItem(true)
}

func (m *tableSelectionPane) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	if search.IsSearchBoxMessage(msg) || m.search.IsFocused() {
		cmds = append(cmds, m.search.Update(msg))
	} else {
		cmds = append(cmds, m.handleNavigation(msg))
	}
	cmds = append(cmds, m.MaybePreviewItem(false))
	return tea.Batch(cmds...)
}

// handleNavigation handles events when search is not active.
func (m *tableSelectionPane) handleNavigation(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg := msg.String(); msg {
		case "/":
			cmds = append(cmds, m.search.OpenSearchBox())
		case "enter":
			return m.selectTable()
		case "Z":
			return m.Zoom()
		case "esc":
			m.search.Reset()
		}
	}
	cmds = append(cmds, m.content.Update(msg))
	return tea.Batch(cmds...)
}

// force is used on new pane initialization because lastPreviewItem could be 0
func (m *tableSelectionPane) MaybePreviewItem(force bool) tea.Cmd {
	idx := m.content.Cursor()
	if len(m.filteredTables) > 0 { // cursor refers to filtered items
		idx = m.filteredTables[idx]
	}
	if idx == m.lastTableDetails && !force {
		return nil
	}
	m.lastTableDetails = idx
	table := m.tables[idx]

	// prepare debounce cancellation
	m.cancelLast()
	ctx, cc := context.WithCancel(m.ctx)
	m.cancelLast = cc

	return func() tea.Msg {
		time.Sleep(m.debounceDur)
		if ctx.Err() != nil { // context canceled
			return nil // debounce
		}

		ctx, cc := context.WithTimeout(ctx, m.stdTO)
		defer cc()
		details, err := dynamodb.DescribeTable(m.config.Client, ctx, table)
		if err != nil {
			return nil
		}

		return messages.TableDetails{
			Details: *details,
		}
	}
}

func (m *tableSelectionPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleTableSelectionPane{}
	}
}

func (m *tableSelectionPane) selectTable() tea.Cmd {
	switchView := func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Table_selection,
			NewView: messages.Item_selection,
		}
	}
	r := []string(m.content.SelectedRow())
	if len(r) == 0 {
		return nil // nothing to select
	}
	// TODO: table details should already be loaded as part of table navigation
	m.cleanSlate()
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	defer cc()
	details, err := dynamodb.DescribeTable(m.config.Client, ctx, r[0])
	if err != nil {
		m.err = err
		return nil
	}

	selectTable := func() tea.Msg {
		return messages.SelectTable{
			TableName:    r[0],
			TableDetails: *details,
		}
	}
	return tea.Batch(switchView, selectTable)
}

func (m *tableSelectionPane) applySize(height, width int) {
	searchBoxH := m.search.GetHeight()
	if !m.search.IsEnabled() {
		searchBoxH = 0
	}
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height - searchBoxH)
	m.content.SetWidth(width)
	m.search.SetWidth(width)
}

func (m *tableSelectionPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.content.View(),
		m.search.View(),
	)
}
