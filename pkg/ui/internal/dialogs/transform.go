package dialogs

import tea "charm.land/bubbletea/v2"

// the Transform dialog enables the user to specify column transformations, such
// as unix-timestamp conversion
type Transform struct {
}

func (m *Transform) Init() tea.Cmd {
	return nil
}

func (m *Transform) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Transform) View() string {
	return "<transform-placeholder>"
}
