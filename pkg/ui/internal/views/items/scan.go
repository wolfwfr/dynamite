package itemselection

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
)

func (m *ItemSelectionPane) enableScanMode() tea.Cmd {
	if m.queryMode == messages.ScanMode {
		return nil
	}
	m.queryMode = messages.ScanMode
	m.KeyMap.Query.SetEnabled(true)
	m.KeyMap.Scan.SetEnabled(false)
	// TODO: impl
	return func() tea.Msg {
		return messages.SwitchQueryMode{
			OldMode: m.queryMode,
			NewMode: messages.ScanMode,
		}
	}
}
