package dialogs

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/atotto/clipboard"
	regular "github.com/wolfwfr/dynamite/pkg/ui/internal/components/regular_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

type copyKeyMap struct {
	close key.Binding
	enter key.Binding
}

func (h copyKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter}
}

// the CopyDialog dialog enables the user to select a column for copying its contents
type CopyDialog struct {
	selected string

	keyMap copyKeyMap

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

	state struct {
		TableARN   string
		AllColumns []string // matching by index
		ColValues  []string // matching by index
	}

	styles copyStyles

	content list.Model
}

type copyStyles struct {
	regular.Styles
	dialog   lipgloss.Style
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
}

func newCopyStyles(darkBG bool) copyStyles {
	var s copyStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)

	s.dialog = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

func NewCopyDialog(close key.Binding) *CopyDialog {
	c := &CopyDialog{
		keyMap: copyKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
		},

		defaultDialogHeight: 46,
		defaultDialogWidth:  55,
	}
	c.dialog.width = c.defaultDialogWidth
	c.dialog.height = c.defaultDialogHeight

	c.window.width = 150
	c.window.height = 100

	{ // list
		l := list.New([]list.Item{}, regular.ItemDelegate{}, c.dialog.width, c.dialog.height)
		l.Title = "Copy Column Value"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		// replace '?' with 'm'
		l.KeyMap.ShowFullHelp.SetKeys("m")
		l.KeyMap.ShowFullHelp.SetHelp("m", "more")
		l.KeyMap.CloseFullHelp.SetKeys("m")
		l.KeyMap.CloseFullHelp.SetHelp("m", "close help")
		l.KeyMap.Quit.SetKeys(c.keyMap.close.Keys()...)
		l.KeyMap.Quit.SetHelp(c.keyMap.close.Help().Key, c.keyMap.close.Help().Desc)

		c.content = l
	}

	c.updateStyles(true) // default to dark styles.
	c.updateSize()

	return c
}

func (m *CopyDialog) newDelegate(s *copyStyles) regular.ItemDelegate {
	return regular.ItemDelegate{
		Styles: &m.styles.Styles,
	}
}

func (m *CopyDialog) updateStyles(isDark bool) {
	s := newCopyStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *CopyDialog) Init() tea.Cmd {
	return nil
}

func (m *CopyDialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog
		case key.Matches(msg, m.keyMap.enter):
			return m.selectItem()
		default:
			var cmd tea.Cmd
			m.content, cmd = m.content.Update(msg)
			return cmd
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitColumnCopy:
		return m.SetState(msg)
	}
	return nil
}

func (m *CopyDialog) SetState(msg messages.InitColumnCopy) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.AllColumns = msg.AllColumns
	m.state.ColValues = msg.ColValues
	return m.updateContent()
}

func (m *CopyDialog) updateContent() tea.Cmd {
	items := make([]list.Item, 0, len(m.state.AllColumns))
	for i := range m.state.AllColumns {
		items = append(items, regular.ListItem(m.state.AllColumns[i]))
	}
	cmd := m.content.SetItems(items)
	m.updateSize()
	return cmd
}

func (m *CopyDialog) selectItem() tea.Cmd {
	idx := m.content.Index()
	if idx > len(m.state.ColValues) {
		panic("dialog state not up to date")
	}

	if err := clipboard.WriteAll(m.state.ColValues[idx]); err != nil {
		return func() tea.Msg {
			return messages.ToggleNotificationDialog{Error: fmt.Errorf("failed to copy: %w", err)}
		}
	}

	return tea.Batch(m.toggleDialog, m.notifySuccess)
}

func (m *CopyDialog) notifySuccess() tea.Msg {
	return messages.ToggleNotificationDialog{Msg: "Copied!", Duration: 1 * time.Second}
}

func (m *CopyDialog) toggleDialog() tea.Msg {
	return messages.ToggleColumnCopy{}
}

func (m *CopyDialog) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *CopyDialog) updateSize() {
	items := m.content.Items()

	// set height of the list within the dialog
	padding := 4
	m.content.SetHeight(min(len(m.content.Items())+padding, m.window.height))

	// determine the width of the list within the dialog
	width := m.defaultDialogWidth
	for _, itm := range items {
		width = max(width, len(itm.(regular.ListItem)))
	}
	// set width of the list within the dialog
	m.content.SetWidth(width)

	// set height & width of dialog itself
	m.styles.dialog = m.styles.dialog.
		Height(m.content.Height() + 2).
		Width(width + 2)

}

func (m *CopyDialog) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.title.Render(m.content.Title),
			m.styles.content.Render(m.content.View()),
			m.styles.help.Render(m.JoinedHelp()),
		),
	)
}

func (m *CopyDialog) JoinedHelp() string {
	if !m.content.Help.ShowAll {
		helpV := m.content.Help.ShortHelpView
		helpLine := m.styles.helpLine
		return lipgloss.JoinVertical(lipgloss.Center,
			helpLine.Render(helpV(m.content.ShortHelp())),
		)
	}

	listBindings := m.content.FullHelp()
	firstCol := listBindings[0]
	listBindings[0] = firstCol
	return m.content.Help.FullHelpView(listBindings)
}
