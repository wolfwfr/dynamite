package itemselection

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
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
	focused     paneID
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

func NewItemSelectionView(ctx context.Context, config *appconfig.Config) *ItemSelection {
	return &ItemSelection{
		config:      config,
		itemsPane:   NewItemSelectionPane(ctx, config),
		detailsPane: newDetailsPane(ctx, config),
	}
}

func (m *ItemSelection) Init() tea.Cmd {
	return m.itemsPane.Init()
}

func (m *ItemSelection) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch s := msg.String(); s {
		case "tab", "shift+tab":
			m.moveFocus()
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
		m.applySize()
	}

	return m.forward(msg)
}

func (m *ItemSelection) forward(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case itemsPaneID:
		return m.itemsPane.Update(msg)
	case detailsPaneID:
		return m.detailsPane.Update(msg)
	}
	return nil
}

func (m *ItemSelection) moveFocus() {
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = itemsPaneID
	}
}

func (m *ItemSelection) applySize() {
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	focusedStyle = focusedStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	m.itemsPane.ApplySize(m.window.height-2-3, m.window.width/2-4)
}

func (m *ItemSelection) View() string {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderBorder(itemsPaneID, m.itemsPane.View()),
		m.renderBorder(detailsPaneID, m.detailsPane.View()),
	))
	return s.String()
}
