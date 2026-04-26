package tableselection

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

	// panes
	tablePane   *tableSelectionPane
	detailsPane *detailsPane
	focused     paneID
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

func NewTableSelectionView(ctx context.Context, config *appconfig.Config) *TableSelection {
	return &TableSelection{
		config:      config,
		tablePane:   newTableSelectionPane(ctx, config),
		detailsPane: newDetailsPane(ctx, config),
	}
}

func (m *TableSelection) Init() tea.Cmd {
	return m.tablePane.Init()
}

func (m *TableSelection) Update(msg tea.Msg) tea.Cmd {
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

	return m.foward(msg)
}

func (m *TableSelection) foward(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case tablePaneID:
		return m.tablePane.Update(msg)
	case detailsPaneID:
		return m.detailsPane.Update(msg)
	}
	return nil
}

func (m *TableSelection) applySize() {
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	focusedStyle = focusedStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	m.tablePane.applySize(m.window.height-2-3, m.window.width/2-4)
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
		m.renderBorder(tablePaneID, m.tablePane.View()),
		m.renderBorder(detailsPaneID, m.detailsPane.View()),
	))
	return s.String()
}

func ternary[T any](first T, second T, cond bool) T {
	if cond {
		return first
	}
	return second
}
