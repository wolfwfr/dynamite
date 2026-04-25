package tableselection

import tea "charm.land/bubbletea/v2"

type TableSelection struct {
}

func NewTableSelection() *TableSelection {
	return &TableSelection{}
}

func (m *TableSelection) Init() tea.Cmd {
	return nil
}

func (m *TableSelection) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *TableSelection) View() string {
	return "<table-selection-placeholder>"
}
