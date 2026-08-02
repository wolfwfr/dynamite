package common

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var textinputKeys []key.Binding

func init() {
	km := textinput.DefaultKeyMap()
	textinputKeys = make([]key.Binding, 16)
	textinputKeys[0] = km.CharacterForward
	textinputKeys[1] = km.CharacterBackward
	textinputKeys[2] = km.WordForward
	textinputKeys[3] = km.WordBackward
	textinputKeys[4] = km.DeleteWordBackward
	textinputKeys[5] = km.DeleteWordForward
	textinputKeys[6] = km.DeleteAfterCursor
	textinputKeys[7] = km.DeleteBeforeCursor
	textinputKeys[8] = km.DeleteCharacterBackward
	textinputKeys[9] = km.DeleteCharacterForward
	textinputKeys[10] = km.LineStart
	textinputKeys[11] = km.LineEnd
	textinputKeys[12] = km.Paste
	textinputKeys[13] = km.AcceptSuggestion
	textinputKeys[14] = km.NextSuggestion
	textinputKeys[15] = km.PrevSuggestion
}

func matchesTextInputKey(k tea.KeyPressMsg) bool {
	for _, kk := range textinputKeys {
		if key.Matches(k, kk) {
			return true
		}
	}
	return false
}

func IsTextInputKey(key tea.KeyPressMsg) bool {
	keyS := key.String()
	return SingleChar.MatchString(keyS) || Alphanum.MatchString(keyS) || keyS == string(tea.KeyBackspace) || matchesTextInputKey(key)
}
