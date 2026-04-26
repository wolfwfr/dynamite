package itemselection

import "charm.land/bubbles/v2/key"

// VIEW
func (m *ItemSelection) ShortHelp() []key.Binding {
	switch m.focused {
	case itemsPaneID:
		return augmentShortHelp(m.itemsPane.ShortHelp(), m.KeyMap.ShortHelp())
	case detailsPaneID:
		return augmentShortHelp(m.detailsPane.ShortHelp(), m.KeyMap.ShortHelp())
	}
	return nil
}

// ITEM PANE
func (m *ItemSelectionPane) ShortHelp() []key.Binding {
	return augmentShortHelp(m.content.KeyMap.ShortHelp(), m.KeyMap.ShortHelp())
}

// DETAILS PANE
func (m *detailsPane) ShortHelp() []key.Binding {
	km := m.content.KeyMap
	viewportHelp := []key.Binding{km.Up, km.Down, km.Left, km.Right}
	return augmentShortHelp(viewportHelp, m.KeyMap.ShortHelp())
}

// VIEW
func (m *ItemSelection) FullHelp() [][]key.Binding {
	switch m.focused {
	case itemsPaneID:
		return augmentFullHelp(m.itemsPane.FullHelp(), m.KeyMap.FullHelp())
	case detailsPaneID:
		return m.detailsPane.FullHelp()
	}
	return nil
}

// ITEM PANE
func (m *ItemSelectionPane) FullHelp() [][]key.Binding {
	return augmentFullHelp(m.content.KeyMap.FullHelp(), m.KeyMap.FullHelp())
}

// DETAILS PANE
func (m *detailsPane) FullHelp() [][]key.Binding {
	return nil
}

func augmentShortHelp(help []key.Binding, extra []key.Binding) []key.Binding {
	res := make([]key.Binding, len(help)+len(extra))
	copy(res[:len(help)], help)
	copy(res[len(help):], extra)
	return res
}

func augmentFullHelp(help [][]key.Binding, extra [][]key.Binding) [][]key.Binding {
	res := make([][]key.Binding, len(help)+len(extra))
	copy(res[:len(help)], help)
	copy(res[len(help):], extra)
	return res
}
