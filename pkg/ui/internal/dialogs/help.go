package dialogs

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

type helpStyles struct {
	dialog   lipgloss.Style
	title    lipgloss.Style
	fullHelp lipgloss.Style
	helpLine lipgloss.Style
}

func newHelpStyles() helpStyles {
	s := helpStyles{}
	s.dialog = styles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.fullHelp = lipgloss.NewStyle().Padding(1, 8, 1, 8)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

type helpKeyMap struct {
	close key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close}
}

// the Help dialog displays helpful instructions
type Help struct {
	activeView messages.View

	styles helpStyles

	keyMap helpKeyMap

	defaultDialogHeight int
	defaultDialogWidth  int

	window struct {
		width  int
		height int
	}

	dialog struct {
		width  int
		height int
	}

	Help help.Model

	// accessible views
	tableSelection help.KeyMap
	itemselection  help.KeyMap
}

func NewHelp(tableView, itemView help.KeyMap, close key.Binding) *Help {
	h := &Help{
		activeView: 0,

		styles: newHelpStyles(),

		defaultDialogHeight: 20,
		defaultDialogWidth:  50,

		keyMap: helpKeyMap{
			close: close,
		},

		Help: help.New(),

		tableSelection: tableView,
		itemselection:  itemView,
	}

	h.dialog.width = h.defaultDialogWidth
	h.dialog.height = h.defaultDialogHeight

	h.window.width = 150
	h.window.height = 100

	return h
}

func (m *Help) Init() tea.Cmd {
	return nil
}

func (m *Help) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleHelp()
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.SwitchView:
		m.activeView = msg.NewView

	}
	m.updateSize()
	return nil
}

func (m *Help) applySize(height, width int) {
	m.window.height = height
	m.window.width = width
	m.updateSize()
}

func (m *Help) updateSize() {
	// first reset widths for obtaining desired size
	m.Help.SetWidth(0)
	m.styles.dialog = m.styles.dialog.Width(0)
	m.styles.dialog = m.styles.dialog.Height(0)

	view := m.View()
	borderW := getBorderWidth(m.styles.dialog)
	paddinW := max(getPadWidth(m.styles.fullHelp), getPadWidth(m.styles.helpLine))

	viewWidth := lipgloss.Width(view)
	idealHelpWidth := viewWidth - borderW - paddinW

	m.dialog.width = min(viewWidth, m.window.width)
	helpWidth := min(idealHelpWidth, m.dialog.width-borderW-paddinW)

	m.dialog.height = min(m.defaultDialogHeight, m.window.height)

	// TODO: some widths mess up the layout, but it seems to be out of my
	// control. Maybe a bug in help.Model?
	m.Help.SetWidth(helpWidth)

	m.styles.dialog = m.styles.dialog.
		Height(m.dialog.height).
		Width(m.dialog.width)
}

func (m *Help) toggleHelp() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleHelp{}
	}
}

func (m *Help) View() string {
	title := "Help"

	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.title.Render(title),
			m.renderFullHelp(),
			m.styles.helpLine.Render(m.Help.ShortHelpView((m.keyMap.ShortHelp()))),
		),
	)
}

func (m *Help) renderFullHelp() string {
	var fullhelp string
	switch m.activeView {
	case messages.Item_selection:
		fullhelp = m.Help.FullHelpView(m.itemselection.FullHelp())
	case messages.Table_selection:
		fullhelp = m.Help.FullHelpView(m.tableSelection.FullHelp())
	}
	return m.styles.fullHelp.Render(fullhelp)
}
