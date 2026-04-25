package itemselection

import (
	"context"

	tea "charm.land/bubbletea/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
)

type ItemSelection struct {
	// top-level context
	ctx context.Context

	// shared config
	config *appconfig.Config
}

func NewItemSelection(ctx context.Context, config *appconfig.Config) *ItemSelection {
	return &ItemSelection{
		ctx:    ctx,
		config: config,
	}
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
