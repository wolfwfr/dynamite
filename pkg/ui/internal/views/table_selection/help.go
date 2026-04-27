package tableselection

import "charm.land/bubbles/v2/key"

// VIEW
func (m *TableSelection) ShortHelp() []key.Binding {
	ah := appendShortHelp
	switch m.focused {
	case tablePaneID:
		return ah(ah(m.tablePane.ShortHelp(), m.KeyMap.ShortHelp()), m.tablePane.AddKeyMap.Bindings())
	case detailsPaneID:
		return ah(ah(m.detailsPane.ShortHelp(), m.KeyMap.ShortHelp()), m.detailsPane.AddKeyMap.Bindings())
	}
	return nil
}

// TABLE PANE
func (m *tableSelectionPane) ShortHelp() []key.Binding {
	return appendShortHelp(m.content.KeyMap.ShortHelp(), m.KeyMap.ShortHelp())
}

// DETAILS PANE
func (m *detailsPane) ShortHelp() []key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right}
	return appendShortHelp(viewportHelp, m.KeyMap.ShortHelp())
}

// VIEW
func (m *TableSelection) FullHelp() [][]key.Binding {
	switch m.focused {
	case tablePaneID:
		return appendFullHelp(m.tablePane.FullHelp(), m.KeyMap.FullHelp())
	case detailsPaneID:
		return appendFullHelp(m.detailsPane.FullHelp(), m.KeyMap.FullHelp())
	}
	return nil
}

// TABLE PANE
func (m *tableSelectionPane) FullHelp() [][]key.Binding {
	return appendFullHelp(m.content.KeyMap.FullHelp(), m.KeyMap.FullHelp())
}

// DETAILS PANE
func (m *detailsPane) FullHelp() [][]key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right}
	return appendFullHelp([][]key.Binding{viewportHelp}, m.KeyMap.FullHelp())
}

func appendShortHelp(help []key.Binding, extra []key.Binding) []key.Binding {
	res := make([]key.Binding, len(help)+len(extra))
	copy(res[:len(help)], help)
	copy(res[len(help):], extra)
	return res
}

func appendFullHelp(help [][]key.Binding, extra [][]key.Binding) [][]key.Binding {
	res := make([][]key.Binding, len(help)+len(extra))
	copy(res[:len(help)], help)
	copy(res[len(help):], extra)
	return res
}
