package tableselection

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type paneID int

const (
	tablesPaneID paneID = iota
	detailPaneID
)

type paneProperties struct {
	height int
	width  int
	style  lipgloss.Style
}

type TableSelection struct {
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
	KeyMap *TableViewKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	// panes
	tablesPane *tableSelectionPane
	detailPane *detailsPane

	zoomEnabled bool

	focused    paneID
	zoomtarget paneID
}

var (
	unfocusedBorderStyle = styles.BorderStyle
	focusedBorderStyle   = styles.FocusedBorderStyle
)

func (m *TableSelection) renderBorder(paneID paneID, content string) string {
	st := m.panes[paneID].style
	if m.focused == paneID {
		return focusedBorderStyle.Inherit(st).Render(content)
	}
	return unfocusedBorderStyle.Inherit(st).Render(content)
}

type Option func(t *TableSelection)

func WithAdditionalKeys(keys keymaps.AdditionalKeys) Option {
	return func(t *TableSelection) {
		t.AddKeyMap = keys
	}
}

func NewTableSelectionView(ctx context.Context, config *appconfig.Config, opts ...Option) *TableSelection {
	t := &TableSelection{
		config: config,
		KeyMap: DefaultTableViewKeyMap(),
		panes:  make(map[paneID]paneProperties),
	}

	for _, o := range opts {
		o(t)
	}

	t.tablesPane = newTableSelectionPane(ctx, config, withTablePaneKeys(t.AddKeyMap))
	t.detailPane = newDetailsPane(ctx, config, withDetailsPaneKeys(t.AddKeyMap))

	return t
}

func (m *TableSelection) Init() tea.Cmd {
	return tea.Batch(m.tablesPane.Init(), m.detailPane.Init())
}

// update handles the message and if it does not detect a keypress that it can
// map itself proceeds to forward the message to the model's children
func (m *TableSelection) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.MoveFocus):
			m.moveFocus()
			return nil
		case key.Matches(msg, m.KeyMap.Regions):
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
func (m *TableSelection) forward(msg tea.Msg) tea.Cmd {
	if _, isKeyPress := msg.(tea.KeyPressMsg); isKeyPress {
		return m.routeToFocusedOnly(msg)
	}
	return m.broadcast(msg)
}

// broadcast takes a message and forwards it to all children
func (m TableSelection) broadcast(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, m.tablesPane.Update(msg))
	cmds = append(cmds, m.detailPane.Update(msg))
	return tea.Batch(cmds...)
}

// routeToFocusedOnly takes a message and only routes it to a single child, the
// active child with highest precedence (dialogs take precedence over views)
func (m *TableSelection) routeToFocusedOnly(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case tablesPaneID:
		return m.tablesPane.Update(msg)
	case detailPaneID:
		return m.detailPane.Update(msg)
	default:
		panic("focused pane not found")
	}
}

func (m *TableSelection) handleZoom(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.ZoomToggleTableSelectionPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = tablesPaneID
		m.focused = tablesPaneID
		m.KeyMap.MoveFocus.SetEnabled(!m.KeyMap.MoveFocus.Enabled())
	case messages.ZoomToggleTableDetailsPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = detailPaneID
		m.focused = detailPaneID
		m.KeyMap.MoveFocus.SetEnabled(!m.KeyMap.MoveFocus.Enabled())
	}
	m.applySize()
	return nil
}

func (m TableSelection) ToggleRegionsDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleRegions{}
	}
}

func (m *TableSelection) applySize() {
	var (
		borderH     = 2
		borderW     = 2
		homeGutterH = 1
		tableswidth = u.Ternary(m.window.width, m.window.width/2, m.zoomEnabled && m.zoomtarget == tablesPaneID)
		detailwidth = u.Ternary(m.window.width, m.window.width/2, m.zoomEnabled && m.zoomtarget == detailPaneID)
		paddingR    = 1
	)
	// ensure full screen width is utilised,
	detailwidth = max(detailwidth, m.window.width-tableswidth)

	tb := m.panes[tablesPaneID]
	dt := m.panes[detailPaneID]

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
	m.panes[detailPaneID] = dt

	// forward
	m.tablesPane.applySize(tb.height, tb.width)
	m.detailPane.applySize(dt.height, dt.width)
}

func (m *TableSelection) moveFocus() {
	m.focused++
	if m.focused > detailPaneID {
		m.focused = tablesPaneID
	}
}

func (m *TableSelection) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		u.Ternary(m.renderBorder(tablesPaneID, m.tablesPane.View()), "", !m.zoomEnabled || m.zoomtarget == tablesPaneID),
		u.Ternary(m.renderBorder(detailPaneID, m.detailPane.View()), "", !m.zoomEnabled || m.zoomtarget == detailPaneID),
	))
	return s.String()
}
