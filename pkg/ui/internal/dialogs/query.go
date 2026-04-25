package dialogs

import tea "charm.land/bubbletea/v2"

// the Query dialog enables the user to specify dynamo-db query parameters.
type Query struct {
}

func (m *Query) Init() tea.Cmd {
	return nil
}

func (m *Query) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m *Query) View() string {
	return "<query-placeholder>"
}
