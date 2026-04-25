package itemselection

import tea "charm.land/bubbletea/v2"

type ItemSelection struct {
}

func NewItemSelection() *ItemSelection {
	return &ItemSelection{}
}

func (m *ItemSelection) Init() tea.Cmd {
	return nil
}

func (m *ItemSelection) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *ItemSelection) View() string {
	return "<item-selection-placeholder>"
}
