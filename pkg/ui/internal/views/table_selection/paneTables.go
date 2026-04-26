package tableselection

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
)

type tableSelectionPane struct {
	// top-level context
	ctx context.Context

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
		ctx:    ctx,
		config: config,
		stdTO:  5 * time.Second,
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
				rows := p.content.Rows()
				filtered := make([]table.Row, len(results))
				for i, match := range results {
					filtered[i] = rows[match.Index]
				}
				p.content.SetVirtualRows(filtered)
			},
			Reset: func(searchHeight int) {
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
		rows := make([]table.Row, len(tables))
		for i := range tables {
			rows[i] = table.Row([]string{tables[i]})
		}
		m.content.SetRows(rows)
	}
	return nil
}

func (m *tableSelectionPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	if search.IsSearchBoxMessage(msg) || m.search.IsFocused() {
		cmd = m.search.Update(msg)
	} else {
		cmd = m.handleNavigation(msg)
	}
	return
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
	var cmd tea.Cmd
	cmd = m.content.Update(msg)
	return tea.Batch(append(cmds, cmd)...)
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
	var details *types.DescribeTableResponse
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	defer cc()
	var err error
	details, err = dynamodb.DescribeTable(m.config.Client, ctx, r[0])
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
