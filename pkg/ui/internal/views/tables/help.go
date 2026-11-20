package tableselection

import "charm.land/bubbles/v2/key"

// VIEW
func (m *TableSelectionView) ShortHelp() []key.Binding {
	ah := appendShortHelp
	switch m.focused {
	case tablesPaneID:
		return ah(ah(m.tablesPane.ShortHelp(), m.keyMap.ShortHelp()), m.tablesPane.addKeyMap.Bindings())
	case detailsPaneID:
		return ah(ah(m.detailPane.ShortHelp(), m.keyMap.ShortHelp()), m.detailPane.addKeyMap.Bindings())
	}
	return nil
}

// TABLE PANE
func (m *tableSelectionPane) ShortHelp() []key.Binding {
	return appendShortHelp(m.content.KeyMap.ShortHelp(), m.keyMap.ShortHelp())
}

// DETAILS PANE
func (m *detailsPane) ShortHelp() []key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right}
	return appendShortHelp(viewportHelp, m.keyMap.ShortHelp())
}

// VIEW
func (m *TableSelectionView) FullHelp() [][]key.Binding {
	switch m.focused {
	case tablesPaneID:
		return appendFullHelp(m.tablesPane.FullHelp(), m.keyMap.FullHelp())
	case detailsPaneID:
		return appendFullHelp(m.detailPane.FullHelp(), m.keyMap.FullHelp())
	}
	return nil
}

// TABLE PANE
func (m *tableSelectionPane) FullHelp() [][]key.Binding {
	return appendFullHelp(m.content.KeyMap.FullHelp(), m.keyMap.FullHelp())
}

// DETAILS PANE
func (m *detailsPane) FullHelp() [][]key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right, km.HalfPageUp, km.HalfPageDown, km.PageUp, km.PageDown}
	return appendFullHelp([][]key.Binding{viewportHelp}, m.keyMap.FullHelp())
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
