package tableselection

import "charm.land/bubbles/v2/key"

// DetailsPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type DetailsPaneKeyMap struct {
	Zoom key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Zoom}
}

// FullHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Zoom},
	}
}

// DefaultDetailsKeyMap returns a default set of keybindings.
func DefaultDetailsKeyMap() *DetailsPaneKeyMap {
	return &DetailsPaneKeyMap{
		Zoom: key.NewBinding(
			key.WithKeys("Z"),
			key.WithHelp("shift+z", "zoom"),
		),
	}
}

// ------------------------------------------ //

// TablePaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type TablePaneKeyMap struct {
	Select key.Binding
	Search key.Binding
	Zoom   key.Binding
	Copy   key.Binding
	Link   key.Binding
	Reload key.Binding
	Esc    key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *TablePaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Select, km.Search, km.Zoom, km.Reload, km.Esc}
}

// FullHelp implements the KeyMap interface.
func (km *TablePaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Select, km.Search, km.Zoom, km.Link, km.Reload, km.Copy, km.Esc},
	}
}

// DefaultTablePaneKeyMap returns a default set of keybindings.
func DefaultTablePaneKeyMap() *TablePaneKeyMap {
	return &TablePaneKeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Zoom: key.NewBinding(
			key.WithKeys("Z"),
			key.WithHelp("shift+z", "zoom"),
		),
		Copy: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("shift+y", "copy"),
		),
		Link: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("shift+L", "open in browser"),
		),
		Reload: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reload"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel search"),
		),
	}
}

// ------------------------------------------ //

// TableViewKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type TableViewKeyMap struct {
	MoveFocus key.Binding
	Regions   key.Binding
}

// DialogKeyMaps collects keys that toggle view-specific dailogs
type DialogKeyMaps struct {
	RegionDialog key.Binding
}

func (m *TableSelection) DialogKeyMaps() DialogKeyMaps {
	return DialogKeyMaps{
		RegionDialog: m.KeyMap.Regions,
	}
}

// ShortHelp implements the KeyMap interface.
func (km *TableViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.MoveFocus, km.Regions}
}

// FullHelp implements the KeyMap interface.
func (km *TableViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.MoveFocus, km.Regions},
	}
}

// DefaultTableViewKeyMap returns a default set of keybindings.
func DefaultTableViewKeyMap() *TableViewKeyMap {
	return &TableViewKeyMap{
		MoveFocus: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch panes"),
		),
		Regions: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("shift+r", "region select"),
		),
	}
}
