package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
)

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	Quit    key.Binding
	Help    key.Binding
	Regions key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.Help, km.Regions}
}

// FullHelp implements the KeyMap interface.
func (km *KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.Quit}, {km.Help}, {km.Regions},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Regions: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("shift+r", "region select"),
		),
	}
}

// DialogCloseKeymapFrom returns a keymap that is intended to close a dialog. It
// includes the first key mapped to the dialog to allow for closing it too. This
// ensures that fluid dialog UX.
func DialogCloseKeymapFrom(keymap key.Binding) key.Binding {
	k := keymap.Keys()
	if len(k) == 0 {
		return key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		)
	}
	return key.NewBinding(
		key.WithKeys(k[0], "esc", "q"),
		key.WithHelp(fmt.Sprintf("%s/esc/q", k[0]), "close"),
	)
}
