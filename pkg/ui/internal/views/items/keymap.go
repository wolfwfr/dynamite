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
			key.WithKeys("z"),
			key.WithHelp("z", "zoom"),
		),
		ToggleFmt: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("shift+j", "toggle json/yaml"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy"),
		),
	}
}

// ------------------------------------------ //

// ItemPaneKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemPaneKeyMap struct {
	Search           key.Binding
	Zoom             key.Binding
	Esc              key.Binding
	Back             key.Binding
	Continue         key.Binding
	ColWidth         key.Binding
	ToggleFmt        key.Binding
	Scan             key.Binding
	ScanParameters   key.Binding
	Query            key.Binding
	QueryParameters  key.Binding
	FilterParameters key.Binding
	Copy             key.Binding
	Browser          key.Binding
	ColVis           key.Binding
	ColSort          key.Binding
	ColTransform     key.Binding
	Reload           key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Back, km.Search, km.Zoom, km.Reload, km.Esc, km.ToggleFmt, km.Scan, km.ScanParameters, km.Query, km.QueryParameters, km.FilterParameters}
}

// FullHelp implements the KeyMap interface.
func (km *ItemPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Search, km.Zoom, km.Reload, km.Esc, km.Continue, km.Back, km.ColWidth, km.ToggleFmt, km.Scan, km.ScanParameters, km.Query, km.QueryParameters, km.FilterParameters, km.Copy, km.Browser, km.ColVis, km.ColSort, km.ColTransform},
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
			key.WithKeys("z"),
			key.WithHelp("z", "zoom"),
		),
		Reload: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reload"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back"),
		),
		Continue: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "continue paging"),
		),
		ColWidth: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle column width"),
		),
		ToggleFmt: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("shift+j", "toggle json/yaml"),
		),
		Scan: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "scan"),
			key.WithDisabled(), // default to scan mode
		),
		ScanParameters: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "scan params"),
		),
		Query: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "query"),
		),
		QueryParameters: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "query params"),
			key.WithDisabled(), // defautl to scan mode
		),
		FilterParameters: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter params"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy"),
		),
		Browser: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "open in browser"),
		),
		ColVis: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "configure column visibility"),
		),
		ColSort: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "configure column order (excl search)"),
		),
		ColTransform: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "transform unix timestamps"),
		),
	}
}

// ------------------------------------------ //

// ItemViewKeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type ItemViewKeyMap struct {
	MoveFocus      key.Binding
	MoveWidthLeft  key.Binding
	MoveWidthRight key.Binding
}

// DialogKeyMaps collects keys that toggle view-specific dailogs
type DialogKeyMaps struct {
	ColumnVisibility key.Binding
	ColumnSorting    key.Binding
	ColumnTransform  key.Binding
	ColumnWidth      key.Binding
	ScanParams       key.Binding
	QueryParams      key.Binding
	FilterParams     key.Binding
	Copy             key.Binding
}

func (m *ItemSelection) DialogKeyMaps() DialogKeyMaps {
	return DialogKeyMaps{
		ColumnVisibility: m.itemsPane.KeyMap.ColVis,
		ColumnSorting:    m.itemsPane.KeyMap.ColSort,
		ColumnTransform:  m.itemsPane.KeyMap.ColTransform,
		ColumnWidth:      m.itemsPane.KeyMap.ColWidth,
		ScanParams:       m.itemsPane.KeyMap.ScanParameters,
		QueryParams:      m.itemsPane.KeyMap.QueryParameters,
		FilterParams:     m.itemsPane.KeyMap.FilterParameters,
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
		{km.MoveFocus, km.MoveWidthLeft, km.MoveWidthRight},
	}
}

// DefaultItemViewKeyMap returns a default set of keybindings.
func DefaultItemViewKeyMap() *ItemViewKeyMap {
	return &ItemViewKeyMap{
		MoveFocus: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch panes"),
		),
		MoveWidthLeft: key.NewBinding(
			key.WithKeys("shift+left"),
			key.WithHelp("shift+←", "move width left"),
		),
		MoveWidthRight: key.NewBinding(
			key.WithKeys("shift+right"),
			key.WithHelp("shift+→", "move width right"),
		),
	}
}
