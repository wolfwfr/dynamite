package itemselection

import (
	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
)

type ItemSelection struct {
}

func NewItemSelection() *ItemSelection {
	return &ItemSelection{}
}

func (m *ItemSelection) Init() tea.Cmd {
	return nil
}

func (m *ItemSelection) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch s := msg.String(); s {
		case "esc":
			return m.escape()
		}
	}
	return nil
}

func (m *ItemSelection) escape() tea.Cmd {
	return func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Item_selection,
			NewView: messages.Table_selection,
		}
	}
}

func (m *ItemSelection) View() string {
	return "<item-selection-placeholder>"
}
