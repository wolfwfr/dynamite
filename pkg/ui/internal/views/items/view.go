package itemselection

import (
	"context"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type paneID int

const (
	itemsPaneID paneID = iota
	detailsPaneID
)

type paneProperties struct {
	height int
	width  int
	style  lipgloss.Style
}

type ItemSelection struct {
	// logger
	logger *slog.Logger

	// shared config
	config *appconfig.Config

	// view window
	window struct {
		width  int
		height int
	}

	// pane-properties
	panes map[paneID]paneProperties

	// panes
	itemsPane   *ItemSelectionPane
	detailsPane *detailsPane

	zoomEnabled bool

	KeyMap *ItemViewKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	focused    paneID
	zoomtarget paneID
}

func (m *ItemSelection) renderBorder(paneID paneID, content string) string {
	st := m.panes[paneID].style
	if m.focused == paneID {
		return theme.FocusedBorderStyle.Inherit(st).Render(content)
	}
	return theme.BorderStyle.Inherit(st).Render(content)
}

type Option func(t *ItemSelection)

func WithAdditionalKeys(keys keymaps.AdditionalKeys) Option {
	return func(t *ItemSelection) {
		t.AddKeyMap = keys
	}
}

func NewItemSelectionView(ctx context.Context, config *appconfig.Config, opts ...Option) *ItemSelection {
	i := &ItemSelection{
		logger: config.Logger.With(slog.String(logging.ViewKey, Log_ItemsView)),
		config: config,
		KeyMap: DefaultItemViewKeyMap(),
		panes:  make(map[paneID]paneProperties),
	}
	for _, o := range opts {
		o(i)
	}

	i.itemsPane = newItemSelectionPane(ctx, config, withItemsPaneKeys(i.AddKeyMap))
	i.detailsPane = newDetailsPane(ctx, config, withDetailsPaneKeys(i.AddKeyMap))

	return i
}

func (m *ItemSelection) Init() tea.Cmd {
	m.logger.Info("initialising...")
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.itemsPane.Init())
	cmds = append(cmds, m.detailsPane.Init())
	m.logger.Info("initialisation complete")
	return tea.Batch(cmds...)
}

// update handles the message and if it does not detect a keypress that it can
// map itself proceeds to forward the message to the model's children
func (m *ItemSelection) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Back):
			if m.keyExecutionSafe(msg) {
				return m.back()
			}
		case key.Matches(msg, m.KeyMap.MoveFocus):
			m.moveFocus()
			return nil
		case key.Matches(msg, m.KeyMap.MoveWidthLeft):
			return m.moveWidthLeft()
		case key.Matches(msg, m.KeyMap.MoveWidthRight):
			return m.moveWidthRight()
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
		m.applySize()
	case messages.ZoomToggleItemSelectionPane, messages.ZoomToggleItemDetailsPane:
		cmd = m.handleZoom(msg)
	}

	return tea.Batch(cmd, m.forward(msg))
}

func (m *ItemSelection) keyExecutionSafe(k tea.KeyPressMsg) bool {
	switch m.focused {
	case itemsPaneID:
		return m.itemsPane.KeyMapExecutionSafe(k)
	case detailsPaneID:
		return m.detailsPane.KeyMapExecutionSafe(k)
	}
	return true
}

// forward takes a message and decides to broadcast or to forward only to focused
// children
func (m *ItemSelection) forward(msg tea.Msg) tea.Cmd {
	if _, isKeyPress := msg.(tea.KeyPressMsg); isKeyPress {
		return m.routeToFocusedOnly(msg)
	}
	return m.broadcast(msg)
}

// broadcast takes a message and forwards it to all children
func (m ItemSelection) broadcast(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, m.itemsPane.Update(msg))
	cmds = append(cmds, m.detailsPane.Update(msg))
	return tea.Batch(cmds...)
}

// routeToFocusedOnly takes a message and only routes it to a single child, the
// active child with highest precedence (dialogs take precedence over views)
func (m *ItemSelection) routeToFocusedOnly(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case itemsPaneID:
		return m.itemsPane.Update(msg)
	case detailsPaneID:
		return m.detailsPane.Update(msg)
	default:
		panic("BUG: focused pane not found; report to maintainer")
	}
}

func (m *ItemSelection) handleZoom(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.ZoomToggleItemSelectionPane:
		m.logger.Debug("zoom selection-pane")
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = itemsPaneID
		m.focused = itemsPaneID
		m.KeyMap.MoveFocus.SetEnabled(!m.KeyMap.MoveFocus.Enabled())
	case messages.ZoomToggleItemDetailsPane:
		m.logger.Debug("zoom details-pane")
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = detailsPaneID
		m.focused = detailsPaneID
		m.KeyMap.MoveFocus.SetEnabled(!m.KeyMap.MoveFocus.Enabled())
	}
	m.applySize()
	return nil
}

func (m *ItemSelection) back() tea.Cmd {
	m.logger.Debug("backing out of items-view")
	// switch to previous view
	switchView := func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Item_selection,
			NewView: messages.Table_selection,
		}
	}

	itemsCmd := m.itemsPane.exit()
	detailsCmd := m.detailsPane.exit()
	return tea.Batch(switchView, itemsCmd, detailsCmd)
}

func (m *ItemSelection) moveFocus() {
	m.logger.Debug("moving focus", slog.Int("current_focus", int(m.focused)))
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = itemsPaneID
	}
}

func (m *ItemSelection) moveWidthLeft() tea.Cmd {
	m.logger.Debug("moving width left",
		slog.Int("current_width", int(m.config.Items.PrimaryWidth)),
	)
	m.config.Items.PrimaryWidth = max(0, m.config.Items.PrimaryWidth-5)
	m.applySize()
	return nil
}

func (m *ItemSelection) moveWidthRight() tea.Cmd {
	m.logger.Debug("moving width right",
		slog.Int("current_width", int(m.config.Items.PrimaryWidth)),
	)
	m.config.Items.PrimaryWidth = min(100, m.config.Items.PrimaryWidth+5)
	m.applySize()
	return nil
}

func (m *ItemSelection) applySize() {
	var (
		borderH     = 2
		borderW     = 2
		homeGutterH = 1
		pct         = min(100.0, float64(m.config.Items.PrimaryWidth))
		partWidth   = int(float64(m.window.width) / (100.0 / pct))
		itemswidth  = u.Ternary(m.window.width, partWidth, m.zoomEnabled && m.zoomtarget == itemsPaneID)
		detailwidth = u.Ternary(m.window.width, m.window.width-itemswidth, m.zoomEnabled && m.zoomtarget == detailsPaneID)
		paddingR    = 1
	)

	// ensure full screen width is utilised,
	detailwidth = max(detailwidth, m.window.width-itemswidth)

	tb := m.panes[itemsPaneID]
	dt := m.panes[detailsPaneID]

	//heights
	tb.height = m.window.height - homeGutterH - borderH
	dt.height = m.window.height - homeGutterH - borderH

	// widths
	tb.width = itemswidth - borderW - paddingR
	dt.width = detailwidth - borderW - paddingR

	// styles
	tb.style = lipgloss.NewStyle().
		Inherit(tb.style).
		Height(m.window.height - homeGutterH).
		MaxHeight(m.window.height - homeGutterH).
		PaddingRight(paddingR).
		Width(itemswidth)
	dt.style = lipgloss.NewStyle().
		Inherit(dt.style).
		Height(m.window.height - homeGutterH).
		MaxHeight(m.window.height - homeGutterH).
		PaddingRight(paddingR).
		Width(detailwidth)

	// update
	m.panes[itemsPaneID] = tb
	m.panes[detailsPaneID] = dt

	// forward
	m.itemsPane.applySize(tb.height, tb.width)
	m.detailsPane.applySize(dt.height, dt.width)
}

func (m *ItemSelection) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		ternary(m.renderBorder(itemsPaneID, m.itemsPane.View()), "", !m.zoomEnabled || m.zoomtarget == itemsPaneID),
		ternary(m.renderBorder(detailsPaneID, m.detailsPane.View()), "", !m.zoomEnabled || m.zoomtarget == detailsPaneID),
	))
	return s.String()
}
