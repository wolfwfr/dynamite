package table

import "charm.land/bubbles/v2/key"

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	ScrollRight  key.Binding
	ScrollLeft   key.Binding
	ShiftRight   key.Binding
	ShiftLeft    key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	GotoLeft     key.Binding
	GotoRight    key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.LineUp, km.LineDown, km.ScrollLeft, km.ScrollRight}
}

// FullHelp implements the KeyMap interface.
func (km *KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.LineUp, km.LineDown, km.GotoTop, km.GotoBottom},
		{km.ScrollLeft}, {km.ScrollRight}, {km.ShiftLeft}, {km.ShiftRight}, {km.GotoLeft}, {km.GotoRight},
		{km.PageUp, km.PageDown, km.HalfPageUp, km.HalfPageDown},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		ScrollRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		ScrollLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		ShiftRight: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("shift+l", "half-width right"),
		),
		ShiftLeft: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("shift+h", "half-width left"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", "space"),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		GotoLeft: key.NewBinding(
			key.WithKeys("B", "0"),
			key.WithHelp("shift+b/0", "go to row beginning"),
		),
		GotoRight: key.NewBinding(
			key.WithKeys("end", "E", "$"),
			key.WithHelp("shift+e/$", "go to row end"),
		),
	}
}
