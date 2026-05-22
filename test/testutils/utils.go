package testutils

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// basic implementation, supports only lowercase basic characters
func CharsToMessages(in string) []tea.Msg {
	msgs := make([]tea.Msg, len(in))
	for i, c := range in {
		msgs[i] = tea.KeyPressMsg(tea.Key{Text: string(c)})
	}
	return msgs
}

// ExecuteCommand executes the given commands in linear fashion, for simplicity
// & greater reproducability
// TODO: execute linearly in DFS style
func ExecuteCommand(cmd tea.Cmd) []tea.Msg {
	var (
		msgs []tea.Msg
		i    = -1
		cmds = []tea.Cmd{cmd}
	)

	for {
		i++
		if i >= len(cmds) {
			break
		}

		cmd := cmds[i]
		if cmd == nil {
			continue
		}
		msg := cmd()

		if batch, ok := msg.(tea.BatchMsg); ok {
			cmds = append(cmds, batch...)
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func ExtractMessages[T any](cmd tea.Cmd) []T {
	var (
		targets []T
		i       = -1
		cmds    = []tea.Cmd{cmd}
	)

	for {
		i++
		if i >= len(cmds) {
			break
		}

		cmd := cmds[i]
		if cmd == nil {
			continue
		}
		msg := cmd()

		if pr, ok := msg.(T); ok {
			targets = append(targets, pr)
		}

		if batch, ok := msg.(tea.BatchMsg); ok {
			cmds = append(cmds, batch...)
		}
	}

	return targets
}

// convenience function for more concise test expressions
func FatalIf(t *testing.T, cond bool, msg ...any) {
	t.Helper()
	if cond {
		t.Fatal(msg...)
	}
}

// convenience function for more concise test expressions
func SkipIf(t *testing.T, cond bool, msg ...any) {
	t.Helper()
	if cond {
		t.Skip(msg...)
	}
}
