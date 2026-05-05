package itemselection

import (
	"context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/atotto/clipboard"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
)

type detailsPane struct {
	// shared config
	config *appconfig.Config

	// errorText
	err error

	// pane's view window
	window struct {
		width  int
		height int
	}

	KeyMap *DetailsPaneKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	content viewport.Model
}

type detailsPaneOption func(p *detailsPane)

func withDetailsPaneKeys(keys keymaps.AdditionalKeys) detailsPaneOption {
	return func(t *detailsPane) {
		t.AddKeyMap = keys
	}
}

func newDetailsPane(ctx context.Context, config *appconfig.Config, opts ...detailsPaneOption) *detailsPane {
	step := 5
	c := viewport.New(viewport.WithHeight(20)) // content
	c.SoftWrap = false
	c.SetHorizontalStep(step)
	c.KeyMap.Left.SetHelp("←/h", "left")
	c.KeyMap.Right.SetHelp("→/l", "right")
	p := &detailsPane{
		config:  config,
		content: c,
		KeyMap:  DefaultDetailsKeyMap(),
	}
	for _, o := range opts {
		o(p)
	}

	if !keymaps.UniqueKeyMaps(p.KeyMap.ShortHelp(), p.AddKeyMap.Bindings()) {
		panic("overlapping keymaps!")
	}
	return p
}

func (m *detailsPane) cleanSlate() {
	m.err = nil
}

func (m *detailsPane) Init() tea.Cmd {
	m.cleanSlate()
	return nil
}

func (m *detailsPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Zoom):
			return m.Zoom()
		case key.Matches(msg, m.KeyMap.ToggleFmt):
			return m.ToggleFmt()
		case key.Matches(msg, m.KeyMap.Copy):
			return m.copy()
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	case messages.PreviewItem:
		m.content.SetContent(msg.Item)
		return nil
	case messages.CopyItem:
		return m.copy()
	}

	m.content, cmd = m.content.Update(msg)
	return
}

func (m *detailsPane) ToggleFmt() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleJSONYAML{}
	}
}

func (m *detailsPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleItemDetailsPane{}
	}
}

func (m *detailsPane) copy() tea.Cmd {
	c := m.content.GetContent()
	if err := clipboard.WriteAll(c); err != nil {
		// TODO: inform user of error (dialog?)
	}
	return nil
}

func (m *detailsPane) applySize(height, width int) {
	// m.content.applySize(m.window.height-2-3, m.window.width/2-4)
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height)
	m.content.SetWidth(width)
}

func (m *detailsPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return m.content.View()
}
