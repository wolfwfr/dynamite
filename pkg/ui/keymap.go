package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
)

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type KeyMap struct {
	ForceQuit key.Binding
	Help      key.Binding
}

// ShortHelp implements the KeyMap interface.
func (km *KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.ForceQuit, km.Help}
}

// FullHelp implements the KeyMap interface.
func (km *KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.ForceQuit}, {km.Help},
	}
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// WARN: checked at home level; keys must not be required for anything else
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// DialogCloseKeymapFrom returns a keymap that is intended to close a dialog. It
// includes the first key mapped to the dialog to allow for closing it too. This
// ensures that fluid dialog UX.
func DialogCloseKeymapFrom(keymap key.Binding) key.Binding {
	k := keymap.Keys()
	kh := keymap.Help().Key
	if len(k) == 0 {
		return key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		)
	}
	return key.NewBinding(
		key.WithKeys(k[0], "esc", "q"),
		key.WithHelp(fmt.Sprintf("%s/esc/q", kh), "close"),
	)
}
