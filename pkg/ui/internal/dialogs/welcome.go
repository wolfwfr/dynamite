package dialogs

import tea "charm.land/bubbletea/v2"

// the Welcome dialog depicts a welcome message
type Welcome struct {
}

func (m *Welcome) Init() tea.Cmd {
	return nil
}

func (m *Welcome) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Welcome) View() string {
	return "<welcome-placeholder>"
}
