// package keymaps defines resources for sharing additional key-maps across
// different app layers
package keymaps

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// ------------------------------------------ //

type AdditionalKey struct {
	Binding key.Binding
	Call    tea.Cmd
}

type AdditionalKeys []AdditionalKey

func (a AdditionalKeys) ShortHelp() []key.Binding {
	keys := make([]key.Binding, len(a))
	for i, b := range a {
		keys[i] = b.Binding
	}
	return keys
}

func (a AdditionalKeys) Bindings() []key.Binding {
	return a.ShortHelp()
}

func (a AdditionalKeys) Matches(msg tea.Msg) (bool, tea.Cmd) {
	keyPress, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return false, nil
	}
	for _, b := range a {
		if key.Matches(keyPress, b.Binding) {
			return true, b.Call
		}
	}
	return false, nil
}

func UniqueKeyMaps(keymaps ...[]key.Binding) bool {
	seen := map[string]struct{}{}
	for _, m := range keymaps {
		for _, b := range m {
			if !b.Enabled() {
				continue
			}
			for _, k := range b.Keys() {
				if _, ok := seen[k]; ok {
					return false
				}
				seen[k] = struct{}{}
			}
		}
	}
	return true
}
