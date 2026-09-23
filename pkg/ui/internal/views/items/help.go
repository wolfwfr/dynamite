package itemselection

import "charm.land/bubbles/v2/key"

// VIEW
func (m *ItemSelectionView) ShortHelp() []key.Binding {
	ah := appendShortHelp
	switch m.focused {
	case itemsPaneID:
		return ah(ah(m.itemsPane.ShortHelp(), m.keyMap.ShortHelp()), m.itemsPane.addKeyMap.Bindings())
	case detailsPaneID:
		return ah(ah(m.detailsPane.ShortHelp(), m.keyMap.ShortHelp()), m.detailsPane.addKeyMap.Bindings())
	}
	return nil
}

// ITEM PANE
func (m *ItemSelectionPane) ShortHelp() []key.Binding {
	return appendShortHelp(m.table.GetKeyMap().ShortHelp(), m.keyMap.ShortHelp())
}

// DETAILS PANE
func (m *detailsPane) ShortHelp() []key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right}
	return appendShortHelp(viewportHelp, m.keyMap.ShortHelp())
}

// VIEW
func (m *ItemSelectionView) FullHelp() [][]key.Binding {
	switch m.focused {
	case itemsPaneID:
		return appendFullHelp(m.itemsPane.FullHelp(), m.keyMap.FullHelp())
	case detailsPaneID:
		return appendFullHelp(m.detailsPane.FullHelp(), m.keyMap.FullHelp())
	}
	return nil
}

// ITEM PANE
func (m *ItemSelectionPane) FullHelp() [][]key.Binding {
	return appendFullHelp(m.table.GetKeyMap().FullHelp(), m.keyMap.FullHelp())
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
