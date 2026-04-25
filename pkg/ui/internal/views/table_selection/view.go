package tableselection

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
)

type TableSelection struct {
	// top-level context
	ctx context.Context

	// shared config
	config *appconfig.Config

	// errorText
	errText string

	// view window
	window struct {
		width  int
		height int
	}

	content *table.Model
}

func NewTableSelection(ctx context.Context, config *appconfig.Config) *TableSelection {
	return &TableSelection{
		ctx:    ctx,
		config: config,
		// TODO: add table feature to hide columns
		content: table.New(
			table.WithColumns([]table.Column{{Title: "table-name", Width: 64}}),
			table.WithFocused(true),
		),
	}
}

func (m *TableSelection) Init() tea.Cmd {
	if client := m.config.Client; client != nil {
		// TODO: async
		ctx, cc := context.WithTimeout(m.ctx, 5*time.Second)
		defer cc()
		tables, err := dynamodb.ListTables(client, ctx)
		if err != nil {
			m.errText = err.Error()
		}
		rows := make([]table.Row, len(tables))
		for i := range tables {
			rows[i] = table.Row([]string{tables[i]})
		}
		m.content.SetRows(rows)
	}
	return nil
}

func (m *TableSelection) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch s := msg.String(); s {
		case "enter":
			return m.selectTable()
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
		m.applySize()
	}

	return m.content.Update(msg)
}

func (m *TableSelection) selectTable() tea.Cmd {
	return func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Table_selection,
			NewView: messages.Item_selection,
		}
	}
}

func (m *TableSelection) applySize() {
	// m.content.applySize(m.window.height-2-3, m.window.width/2-4)
	m.content.SetHeight(m.window.height)
	m.content.SetWidth(m.window.width)
}

func (m *TableSelection) View() string {
	if m.errText != "" { // TODO: formatting
		return m.errText
	}
	return m.content.View()
	// return "<table-selection-placeholder>"
}
