package itemselection

import "charm.land/bubbles/v2/key"

// DetailsPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type DetailsPaneKeyMap struct {
	Zoom      key.Binding
	ToggleFmt key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Zoom, km.ToggleFmt}
}

// FullHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Zoom}, {km.ToggleFmt},
	}
}

// DefaultDetailsKeyMap returns a default set of keybindings.
func DefaultDetailsKeyMap() *DetailsPaneKeyMap {
	return &DetailsPaneKeyMap{
		Zoom: key.NewBinding(
			key.WithKeys("Z"),
			key.WithHelp("shift+z", "zoom"),
		),
		ToggleFmt: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("shift+j", "toggle json/yaml"),
		),
	}
}

// ------------------------------------------ //

// ItemPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemPaneKeyMap struct {
	Search key.Binding
	Zoom   key.Binding
	Esc    key.Binding
	ChCols key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Search, km.Zoom, km.Esc, km.ChCols}
}

// FullHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Search}, {km.Zoom}, {km.Esc}, {km.ChCols},
	}
}

// DefaultItemPaneKeyMap returns a default set of keybindings.
func DefaultItemPaneKeyMap() *ItemPaneKeyMap {
	return &ItemPaneKeyMap{
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Zoom: key.NewBinding(
			key.WithKeys("Z"),
			key.WithHelp("shift+z", "zoom"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel search/return"),
		),
		ChCols: key.NewBinding(
			key.WithKeys("W"),
			key.WithHelp("shift+w", "toggle dynamic column width"),
		),
	}
}

// ------------------------------------------ //

// ItemViewKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemViewKeyMap struct {
	MoveFocus key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *ItemViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.MoveFocus}
}

// FullHelp implements the KeyMap interface.
func (km *ItemViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.MoveFocus},
	}
}

// DefaultItemViewKeyMap returns a default set of keybindings.
func DefaultItemViewKeyMap() *ItemViewKeyMap {
	return &ItemViewKeyMap{
		MoveFocus: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch panes"),
		),
	}
}
