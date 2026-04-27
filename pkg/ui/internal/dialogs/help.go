package dialogs

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

var helpDialogStyle = styles.DialogStyle

type helpKeyMap struct {
	close key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close}
}

// the Help dialog displays helpful instructions
type Help struct {
	activeView messages.View

	keyMap helpKeyMap

	defaultDialogHeight int
	defaultDialogWidth  int

	width  int
	height int

	Help help.Model

	// accessible views
	tableSelection help.KeyMap
	itemselection  help.KeyMap
}

func NewHelp(tableView, itemView help.KeyMap) *Help {
	return &Help{
		activeView: 0,

		defaultDialogHeight: 20,
		defaultDialogWidth:  180,

		width:  20,
		height: 180,

		keyMap: helpKeyMap{
			close: key.NewBinding(
				key.WithKeys("?", "esc", "q"),
				key.WithHelp("?/esc/q", "close help"),
			),
		},

		Help: help.New(),

		tableSelection: tableView,
		itemselection:  itemView,
	}
}

func (m *Help) Init() tea.Cmd {
	return nil
}

func (m *Help) Width() int {
	return m.width
}

func (m *Help) Height() int {
	return m.height
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
	return nil
}

func (m *Help) applySize(height, width int) {
	m.width = m.defaultDialogWidth
	m.height = m.defaultDialogHeight
	helpDialogStyle = helpDialogStyle.
		Height(m.height).
		Width(m.width)

}

func (m *Help) toggleHelp() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleHelp{}
	}
}

func (m *Help) View() string {
	var fullhelp string
	switch m.activeView {
	case messages.Item_selection:
		fullhelp = m.Help.FullHelpView(m.itemselection.FullHelp())
	case messages.Table_selection:
		fullhelp = m.Help.FullHelpView(m.tableSelection.FullHelp())
	}

	helpHeight := height(fullhelp)
	padding := 1
	availableHeight := m.height - helpHeight - 2 - 2*padding
	nl := newLines(int(availableHeight / 2))

	return helpDialogStyle.Render("Help" + nl + fullhelp + nl + m.Help.ShortHelpView((m.keyMap.ShortHelp())))
}

func newLines(n int) string {
	s := strings.Builder{}
	for range n {
		s.WriteString("\n")
	}
	return s.String()
}

func height(in string) int {
	return strings.Count(in, "\n")
}
