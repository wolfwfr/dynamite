package dialogs

import tea "charm.land/bubbletea/v2"

// the Regions dialog enables the user to select an AWS-region
type Regions struct {
}

func (m *Regions) Init() tea.Cmd {
	return nil
}

func (m *Regions) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Regions) View() string {
	return "<regions-placeholder>"
}
