package tableselection

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
	tablesPaneID paneID = iota
	detailsPaneID
)

type paneProperties struct {
	height int
	width  int
	style  lipgloss.Style
}

type TableSelectionView struct {
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

	// key map
	keyMap *TableViewKeyMap

	// Additional Keys
	addKeyMap keymaps.AdditionalKeys

	// panes
	tablesPane *tableSelectionPane
	detailPane *detailsPane

	zoomEnabled bool

	focused    paneID
	zoomtarget paneID
}

func (m *TableSelectionView) renderBorder(paneID paneID, content string) string {
	st := m.panes[paneID].style
	if m.focused == paneID {
		return theme.FocusedBorderStyle.Inherit(st).Render(content)
	}
	return theme.BorderStyle.Inherit(st).Render(content)
}

type Option func(t *TableSelectionView)

func WithAdditionalKeys(keys keymaps.AdditionalKeys) Option {
	return func(t *TableSelectionView) {
		t.addKeyMap = keys
	}
}

func NewTableSelectionView(ctx context.Context, config *appconfig.Config, opts ...Option) *TableSelectionView {
	t := &TableSelectionView{
		logger: config.Logger.With(slog.String(logging.ViewKey, Log_TablesView)),
		config: config,
		keyMap: DefaultTableViewKeyMap(),
		panes:  make(map[paneID]paneProperties),
	}

	for _, o := range opts {
		o(t)
	}

	t.tablesPane = newTableSelectionPane(ctx, config, withTablePaneKeys(t.addKeyMap))
	t.detailPane = newDetailsPane(ctx, config, withDetailsPaneKeys(t.addKeyMap))

	return t
}

func (m *TableSelectionView) Init() tea.Cmd {
	m.logger.Info("initialising...")
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, m.tablesPane.Init())
	cmds = append(cmds, m.detailPane.Init())
	m.logger.Info("initialisation complete")
	return tea.Batch(cmds...)
}

// update handles the message and if it does not detect a keypress that it can
// map itself proceeds to forward the message to the model's children
func (m *TableSelectionView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.MoveFocus):
			m.moveFocus()
			return nil
		case key.Matches(msg, m.keyMap.MoveWidthLeft):
			return m.moveWidthLeft()
		case key.Matches(msg, m.keyMap.MoveWidthRight):
			return m.moveWidthRight()
		case key.Matches(msg, m.keyMap.Regions):
			return m.ToggleRegionsDialog()
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
		m.applySize()
	case messages.ZoomToggleTableSelectionPane, messages.ZoomToggleTableDetailsPane:
		cmd = m.handleZoom(msg)
	}

	return tea.Batch(cmd, m.forward(msg))
}

// forward takes a message and decides to broadcast or to forward only to focused
// children
func (m *TableSelectionView) forward(msg tea.Msg) tea.Cmd {
	if _, isKeyPress := msg.(tea.KeyPressMsg); isKeyPress {
		return m.routeToFocusedOnly(msg)
	}
	return m.broadcast(msg)
}

// broadcast takes a message and forwards it to all children
func (m TableSelectionView) broadcast(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, m.tablesPane.Update(msg))
	cmds = append(cmds, m.detailPane.Update(msg))
	return tea.Batch(cmds...)
}

// routeToFocusedOnly takes a message and only routes it to a single child, the
// active child with highest precedence (dialogs take precedence over views)
func (m *TableSelectionView) routeToFocusedOnly(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case tablesPaneID:
		return m.tablesPane.Update(msg)
	case detailsPaneID:
		return m.detailPane.Update(msg)
	default:
		panic("BUG: focused pane not found; report to maintainer")
	}
}

func (m *TableSelectionView) handleZoom(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.ZoomToggleTableSelectionPane:
		m.logger.Debug("zoom selection-pane")
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = tablesPaneID
		m.focused = tablesPaneID
		m.keyMap.MoveFocus.SetEnabled(!m.keyMap.MoveFocus.Enabled())
	case messages.ZoomToggleTableDetailsPane:
		m.logger.Debug("zoom details-pane")
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = detailsPaneID
		m.focused = detailsPaneID
		m.keyMap.MoveFocus.SetEnabled(!m.keyMap.MoveFocus.Enabled())
	}
	m.applySize()
	return nil
}

func (m TableSelectionView) ToggleRegionsDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleRegions{}
	}
}

func (m *TableSelectionView) moveWidthLeft() tea.Cmd {
	m.logger.Debug("moving width left",
		slog.Int("current_width", int(m.config.Tables.PrimaryWidth)),
	)
	m.config.Tables.PrimaryWidth = max(0, m.config.Tables.PrimaryWidth-5)
	m.applySize()
	return nil
}

func (m *TableSelectionView) moveWidthRight() tea.Cmd {
	m.logger.Debug("moving width right",
		slog.Int("current_width", int(m.config.Tables.PrimaryWidth)),
	)
	m.config.Tables.PrimaryWidth = min(100, m.config.Tables.PrimaryWidth+5)
	m.applySize()
	return nil
}

func (m *TableSelectionView) applySize() {
	var (
		borderH     = 2
		borderW     = 2
		homeGutterH = 1
		pct         = min(100.0, float64(m.config.Tables.PrimaryWidth))
		partWidth   = int(float64(m.window.width) / (100.0 / pct))
		tableswidth = u.Ternary(m.window.width, partWidth, m.zoomEnabled && m.zoomtarget == tablesPaneID)
		detailwidth = u.Ternary(m.window.width, m.window.width-tableswidth, m.zoomEnabled && m.zoomtarget == detailsPaneID)
		paddingR    = 1
	)
	// ensure full screen width is utilised,
	detailwidth = max(detailwidth, m.window.width-tableswidth)

	tb := m.panes[tablesPaneID]
	dt := m.panes[detailsPaneID]

	//heights
	tb.height = m.window.height - homeGutterH - borderH
	dt.height = m.window.height - homeGutterH - borderH

	// widths
	tb.width = tableswidth - borderW - paddingR
	dt.width = detailwidth - borderW - paddingR

	// styles
	tb.style = lipgloss.NewStyle().
		Inherit(tb.style).
		Height(m.window.height - homeGutterH).
		MaxHeight(m.window.height - homeGutterH).
		PaddingRight(paddingR).
		Width(tableswidth)
	dt.style = lipgloss.NewStyle().
		Inherit(dt.style).
		Height(m.window.height - homeGutterH).
		MaxHeight(m.window.height - homeGutterH).
		PaddingRight(paddingR).
		Width(detailwidth)

	// update
	m.panes[tablesPaneID] = tb
	m.panes[detailsPaneID] = dt

	// forward
	m.tablesPane.applySize(tb.height-1, tb.width) // subtracting 1 from height for slightly cleaner look
	m.detailPane.applySize(dt.height, dt.width)
}

func (m *TableSelectionView) moveFocus() {
	m.logger.Debug("moving focus", slog.Int("current_focus", int(m.focused)))
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = tablesPaneID
	}
}

func (m *TableSelectionView) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		u.Ternary(m.renderBorder(tablesPaneID, m.tablesPane.View()), "", !m.zoomEnabled || m.zoomtarget == tablesPaneID),
		u.Ternary(m.renderBorder(detailsPaneID, m.detailPane.View()), "", !m.zoomEnabled || m.zoomtarget == detailsPaneID),
	))
	return s.String()
}
