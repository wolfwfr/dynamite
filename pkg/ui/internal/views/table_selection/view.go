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
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type paneID int

const (
	tablePaneID paneID = iota
	detailsPaneID
)

type TableSelection struct {
	// shared config
	config *appconfig.Config

	// view window
	window struct {
		width  int
		height int
	}

	// key map
	KeyMap *TableViewKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	// panes
	tablePane   *tableSelectionPane
	detailsPane *detailsPane

	zoomEnabled bool

	focused    paneID
	zoomtarget paneID
}

var (
	borderStyle  = styles.BorderStyle
	focusedStyle = styles.FocusedBorderStyle
)

func (m *TableSelection) renderBorder(paneID paneID, content string) string {
	if m.focused == paneID {
		return focusedStyle.Render(content)
	}
	return borderStyle.Render(content)
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
	}

	for _, o := range opts {
		o(t)
	}

	t.tablePane = newTableSelectionPane(ctx, config, withTablePaneKeys(t.AddKeyMap))
	t.detailsPane = newDetailsPane(ctx, config, withDetailsPaneKeys(t.AddKeyMap))

	return t
}

func (m *TableSelection) Init() tea.Cmd {
	return tea.Batch(m.tablePane.Init(), m.detailsPane.Init())
}

func (m *TableSelection) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.MoveFocus):
			m.moveFocus()
			return nil
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
		m.applySize()
		return nil
	case messages.ZoomToggleTableSelectionPane, messages.ZoomToggleTableDetailsPane:
		m.handleZoom(msg)
		return nil
	}

	return m.foward(msg)
}

func (m *TableSelection) handleZoom(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.ZoomToggleTableSelectionPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = tablePaneID
		m.focused = tablePaneID
	case messages.ZoomToggleTableDetailsPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = detailsPaneID
		m.focused = detailsPaneID
	}
	m.applySize()
	return nil
}

func (m *TableSelection) foward(msg tea.Msg) tea.Cmd {
	if _, isDetails := msg.(messages.TableDetails); isDetails {
		cmds := []tea.Cmd{}
		cmds = append(cmds, m.tablePane.Update(msg))
		cmds = append(cmds, m.detailsPane.Update(msg))
		return tea.Batch(cmds...)
	}

	switch m.focused {
	case tablePaneID:
		return m.tablePane.Update(msg)
	case detailsPaneID:
		return m.detailsPane.Update(msg)
	}
	return nil
}

func (m *TableSelection) applySize() {
	w := u.Ternary(m.window.width, m.window.width/2, m.zoomEnabled)
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(w)

	focusedStyle = focusedStyle.
		Height(m.window.height - 2).
		Width(w)

	m.tablePane.applySize(m.window.height-2-3, w-4)
	m.detailsPane.applySize(m.window.height-2-3, w-4)
}

func (m *TableSelection) moveFocus() {
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = tablePaneID
	}
}

func (m *TableSelection) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		u.Ternary(m.renderBorder(tablePaneID, m.tablePane.View()), "", !m.zoomEnabled || m.zoomtarget == tablePaneID),
		u.Ternary(m.renderBorder(detailsPaneID, m.detailsPane.View()), "", !m.zoomEnabled || m.zoomtarget == detailsPaneID),
	))
	return s.String()
}
