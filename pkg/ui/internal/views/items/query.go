package itemselection

import (
	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
)

func (m *ItemSelectionPane) enableQueryMode() tea.Cmd {
	if m.queryMode == messages.QueryMode {
		return nil
	}
	m.queryMode = messages.QueryMode
	m.KeyMap.Scan.SetEnabled(true)
	m.KeyMap.Query.SetEnabled(false)
	// TODO: impl
	return func() tea.Msg {
		return messages.SwitchQueryMode{
			OldMode: m.queryMode,
			NewMode: messages.QueryMode,
		}
	}
}
