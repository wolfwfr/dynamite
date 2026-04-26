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
	return &tableSelectionPane{
		ctx:    ctx,
		config: config,
		stdTO:  5 * time.Second,
		// TODO: add table feature to hide columns
		content: t,
	}
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

func (m *tableSelectionPane) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch s := msg.String(); s {
		case "enter":
			return m.selectTable()
		case "Z":
			return m.Zoom()
		}
	}

	return m.content.Update(msg)
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
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height)
	m.content.SetWidth(width)
}

func (m *tableSelectionPane) View() string {
	if m.err != nil { // TODO: formatting
		return m.err.Error()
	}
	return m.content.View()
}
