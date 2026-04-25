package tableselection

import (
	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
)

type TableSelection struct {
}

func NewTableSelection() *TableSelection {
	return &TableSelection{}
}

func (m *TableSelection) Init() tea.Cmd {
	return nil
}

func (m *TableSelection) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch s := msg.String(); s {
		case "enter":
			return m.selectTable()
		}
	}
	return nil
}

func (m *TableSelection) selectTable() tea.Cmd {
	return func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Table_selection,
			NewView: messages.Item_selection,
		}
	}
}

func (m *TableSelection) View() string {
	return "<table-selection-placeholder>"
}
