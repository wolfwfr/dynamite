package itemselection

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
)

type paneID int

const (
	itemsPaneID paneID = iota
	detailsPaneID
)

type ItemSelection struct {
	// shared config
	config *appconfig.Config

	// view window
	window struct {
		width  int
		height int
	}

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

var (
	borderStyle  = styles.BorderStyle
	focusedStyle = styles.FocusedBorderStyle
)

func (m *ItemSelection) renderBorder(paneID paneID, content string) string {
	if m.focused == paneID {
		return focusedStyle.Render(content)
	}
	return borderStyle.Render(content)
}

type Option func(t *ItemSelection)

func WithAdditionalKeys(keys keymaps.AdditionalKeys) Option {
	return func(t *ItemSelection) {
		t.AddKeyMap = keys
	}
}

func NewItemSelectionView(ctx context.Context, config *appconfig.Config, opts ...Option) *ItemSelection {
	i := &ItemSelection{
		config: config,
		KeyMap: DefaultItemViewKeyMap(),
	}
	for _, o := range opts {
		o(i)
	}

	i.itemsPane = newItemSelectionPane(ctx, config, withItemsPaneKeys(i.AddKeyMap))
	i.detailsPane = newDetailsPane(ctx, config, withDetailsPaneKeys(i.AddKeyMap))

	return i
}

func (m *ItemSelection) Init() tea.Cmd {
	return m.itemsPane.Init()
}

// update handles the message and if it does not detect a keypress that it can
// map itself proceeds to forward the message to the model's children
func (m *ItemSelection) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
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
	case messages.ZoomToggleItemSelectionPane, messages.ZoomToggleItemDetailsPane:
		cmd = m.handleZoom(msg)
	}

	return tea.Batch(cmd, m.forward(msg))
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
		panic("focused pane not found")
	}
}
func (m *ItemSelection) handleZoom(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.ZoomToggleItemSelectionPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = itemsPaneID
		m.focused = itemsPaneID
	case messages.ZoomToggleItemDetailsPane:
		m.zoomEnabled = !m.zoomEnabled
		m.zoomtarget = detailsPaneID
		m.focused = detailsPaneID
	}
	m.applySize()
	return nil
}

func (m *ItemSelection) moveFocus() {
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = itemsPaneID
	}
}

func (m *ItemSelection) applySize() {
	w := ternary(m.window.width, m.window.width/2, m.zoomEnabled)
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(w)

	focusedStyle = focusedStyle.
		Height(m.window.height - 2).
		Width(w)

	m.itemsPane.applySize(m.window.height-2-3, w-4)
	m.detailsPane.applySize(m.window.height-2-3, w-4)
}

func (m *ItemSelection) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		ternary(m.renderBorder(itemsPaneID, m.itemsPane.View()), "", !m.zoomEnabled || m.zoomtarget == itemsPaneID),
		ternary(m.renderBorder(detailsPaneID, m.detailsPane.View()), "", !m.zoomEnabled || m.zoomtarget == detailsPaneID),
	))
	return s.String()
}
