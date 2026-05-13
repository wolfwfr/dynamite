package tableselection

import "charm.land/bubbles/v2/key"

// VIEW
func (m *TableSelection) ShortHelp() []key.Binding {
	ah := appendShortHelp
	switch m.focused {
	case tablesPaneID:
		return ah(ah(m.tablesPane.ShortHelp(), m.KeyMap.ShortHelp()), m.tablesPane.AddKeyMap.Bindings())
	case detailPaneID:
		return ah(ah(m.detailPane.ShortHelp(), m.KeyMap.ShortHelp()), m.detailPane.AddKeyMap.Bindings())
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
	case tablesPaneID:
		return appendFullHelp(m.tablesPane.FullHelp(), m.KeyMap.FullHelp())
	case detailPaneID:
		return appendFullHelp(m.detailPane.FullHelp(), m.KeyMap.FullHelp())
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
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right, km.HalfPageUp, km.HalfPageDown, km.PageUp, km.PageDown}
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
