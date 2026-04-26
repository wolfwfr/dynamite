package itemselection

import (
	"context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
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

	content viewport.Model
}

func newDetailsPane(ctx context.Context, config *appconfig.Config) *detailsPane {
	step := 5
	c := viewport.New(viewport.WithHeight(20)) // content
	c.SoftWrap = false
	c.SetHorizontalStep(step)
	c.KeyMap.Left.SetHelp("←/h", "left")
	c.KeyMap.Right.SetHelp("→/l", "right")
	return &detailsPane{
		config:  config,
		content: c,
		KeyMap:  DefaultDetailsKeyMap(),
	}
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
		}
	case messages.PreviewItem:
		m.content.SetContent(msg.Item)
		return nil
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

func (m *detailsPane) applySize(height, width int) {
	// m.content.applySize(m.window.height-2-3, m.window.width/2-4)
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height)
	m.content.SetWidth(width)
}

func (m *detailsPane) View() string {
	if m.err != nil { // TODO: formatting
		return m.err.Error()
	}
	return m.content.View()
}
