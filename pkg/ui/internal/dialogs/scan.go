package dialogs

import tea "charm.land/bubbletea/v2"

// the Scan dialog enables the user to specify dynamo-db scan parameters.
type Scan struct {
}

func (m *Scan) Init() tea.Cmd {
	return nil
}

func (m *Scan) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Scan) View() string {
	return "<scan-placeholder>"
}
