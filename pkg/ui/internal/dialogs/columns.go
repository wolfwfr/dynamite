package dialogs

import tea "charm.land/bubbletea/v2"

// the Columns dialog enables the user to enable and disable visibility of
// individual columns
type Columns struct {
}

func (m *Columns) Init() tea.Cmd {
	return nil
}

func (m *Columns) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m *Columns) View() string {
	return "<columns-placeholder>"
}
