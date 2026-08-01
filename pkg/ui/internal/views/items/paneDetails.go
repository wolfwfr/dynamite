package itemselection

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/atotto/clipboard"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
)

type detailsPane struct {
	// top-level context
	ctx context.Context

	// shared config
	config *appconfig.Config

	// logger
	logger *slog.Logger

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

	previewing messages.PreviewItem
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
		ctx:     ctx,
		config:  config,
		logger:  config.Logger.With(slog.String(logging.ViewKey, Log_ItemsView), slog.String(logging.PaneKey, "details-pane")),
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

func (m *detailsPane) exit() tea.Cmd {
	m.logger.Debug("exiting")
	m.content.SetContent("")
	return nil
}

func (m *detailsPane) Init() tea.Cmd {
	m.logger.Info("initialising...")
	m.previewing = messages.PreviewItem{}
	m.cleanSlate()
	m.logger.Info("initialisation complete")
	return nil
}

func (m *detailsPane) KeyMapExecutionSafe(k tea.KeyPressMsg) bool {
	return true
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
		m.previewing = msg
		m.content.SetYOffset(0)
		m.content.SetContent(msg.StyledItem)
		return nil
	case messages.CopyItem:
		return m.copy()
	}

	m.content, cmd = m.content.Update(msg)
	return
}

func (m *detailsPane) ToggleFmt() tea.Cmd {
	m.logger.Debug("emitting toggle-format message")
	return func() tea.Msg {
		return messages.ToggleJSONYAML{}
	}
}

func (m *detailsPane) Zoom() tea.Cmd {
	m.logger.Debug("emitting zoom message")
	return func() tea.Msg {
		return messages.ZoomToggleItemDetailsPane{}
	}
}

func (m *detailsPane) copy() tea.Cmd {
	m.logger.Debug("executing copy operation")
	c := m.previewing.RawItem
	if err := clipboard.WriteAll(c); err != nil {
		m.logger.Error("copy to clipboard returned an error", slog.Any("error", err))
		return func() tea.Msg {
			return messages.ToggleNotificationDialog{Error: fmt.Errorf("failed to copy: %w", err)}
		}
	}
	return notifyCopySuccess
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

func notifyCopySuccess() tea.Msg {
	return messages.ToggleNotificationDialog{Msg: "Copied!", Duration: 1 * time.Second}
}
