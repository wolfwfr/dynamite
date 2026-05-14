package itemselection

import "charm.land/bubbles/v2/key"

// DetailsPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type DetailsPaneKeyMap struct {
	Zoom      key.Binding
	ToggleFmt key.Binding
	Copy      key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Zoom, km.ToggleFmt, km.Copy}
}

// FullHelp implements the KeyMap interface.
func (km *DetailsPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Zoom, km.ToggleFmt, km.Copy},
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
		Copy: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("shift+y", "copy"),
		),
	}
}

// ------------------------------------------ //

// ItemPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemPaneKeyMap struct {
	Search          key.Binding
	Zoom            key.Binding
	Esc             key.Binding
	ChCols          key.Binding
	ToggleFmt       key.Binding
	Scan            key.Binding
	ScanParameters  key.Binding
	Query           key.Binding
	QueryParameters key.Binding
	Copy            key.Binding
	Browser         key.Binding
	ColVis          key.Binding
	ColSort         key.Binding
	Reload          key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Search, km.Zoom, km.Reload, km.Esc, km.ToggleFmt, km.Scan, km.ScanParameters, km.Query, km.QueryParameters}
}

// FullHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Search, km.Zoom, km.Reload, km.Esc, km.ChCols, km.ToggleFmt, km.Scan, km.ScanParameters, km.Query, km.QueryParameters, km.Copy, km.Browser, km.ColVis, km.ColSort},
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
		Reload: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reload"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel/return"),
		),
		ChCols: key.NewBinding(
			key.WithKeys("W"),
			key.WithHelp("shift+w", "toggle column width"),
		),
		ToggleFmt: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("shift+j", "toggle json/yaml"),
		),
		Scan: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("shift+s", "scan"),
			key.WithDisabled(), // default to scan mode
		),
		ScanParameters: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("shift+s", "scan params"),
		),
		Query: key.NewBinding(
			key.WithKeys("Q"),
			key.WithHelp("shift+q", "query"),
		),
		QueryParameters: key.NewBinding(
			key.WithKeys("Q"),
			key.WithHelp("shift+q", "query params"),
			key.WithDisabled(), // defautl to scan mode
		),
		Copy: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("shift+y", "copy"),
		),
		Browser: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("shift+x", "open in browser"),
		),
		ColVis: key.NewBinding(
			key.WithKeys("V"),
			key.WithHelp("shift+v", "configure column visibility"),
		),
		ColSort: key.NewBinding(
			key.WithKeys("O"),
			key.WithHelp("shift+o", "configure column order (excl search)"),
		),
	}
}

// ------------------------------------------ //

// ItemViewKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemViewKeyMap struct {
	MoveFocus key.Binding
}

// DialogKeyMaps collects keys that toggle view-specific dailogs
type DialogKeyMaps struct {
	ColumnVisibility key.Binding
	ColumnSorting    key.Binding
	ScanParams       key.Binding
	QueryParams      key.Binding
	Copy             key.Binding
}

func (m *ItemSelection) DialogKeyMaps() DialogKeyMaps {
	return DialogKeyMaps{
		ColumnVisibility: m.itemsPane.KeyMap.ColVis,
		ColumnSorting:    m.itemsPane.KeyMap.ColSort,
		ScanParams:       m.itemsPane.KeyMap.ScanParameters,
		QueryParams:      m.itemsPane.KeyMap.QueryParameters,
		Copy:             m.itemsPane.KeyMap.Copy,
	}
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
