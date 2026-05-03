package dialogs

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

var columnsDialogStyle = commonstyles.DialogStyle

type columnsKeyMap struct {
	close      key.Binding
	enter      key.Binding
	enableAll  key.Binding
	disableAll key.Binding
}

func (h columnsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter, h.enableAll, h.disableAll}
}

type columnsListStyles struct {
	title        lipgloss.Style
	content      lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	help         lipgloss.Style
	helpLine     lipgloss.Style
}

func newColumnStyles(darkBG bool) columnsListStyles {
	var s columnsListStyles
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

type checkboxItem struct {
	checked bool
	name    string
}

func (i checkboxItem) FilterValue() string { return "" }

// the Columns dialog enables the user to enable and disable visibility of
// individual columns
type Columns struct {
	keyMap columnsKeyMap

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

	styles columnsListStyles

	state struct {
		TableARN   string
		AllColumns []string // matching by index
		Visible    []bool   // matching by index
	}

	content list.Model
}

func NewColumnVisibilityDialog(close key.Binding) *Columns {
	c := &Columns{
		keyMap: columnsKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
			enableAll: key.NewBinding(
				key.WithKeys("E"),
				key.WithHelp("shift+e", "enable all"),
			),
			disableAll: key.NewBinding(
				key.WithKeys("D"),
				key.WithHelp("shift+d", "disable all"),
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
		l := list.New([]list.Item{}, columnsItemDelegate{}, c.dialog.width, c.dialog.height)
		l.Title = "Column Visibility"
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

type columnsItemDelegate struct {
	styles *columnsListStyles
}

func (d columnsItemDelegate) Height() int                             { return 1 }
func (d columnsItemDelegate) Spacing() int                            { return 0 }
func (d columnsItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d columnsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(checkboxItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s %s", u.Ternary("[x]", "[ ]", i.checked), i.name)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m *Columns) updateStyles(isDark bool) {
	s := newColumnStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *Columns) newDelegate(s *columnsListStyles) columnsItemDelegate {
	return columnsItemDelegate{
		styles: s,
	}
}

func (m *Columns) Init() tea.Cmd {
	return nil
}

func (m *Columns) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.enter):
			return m.selectItem()
		case key.Matches(msg, m.keyMap.enableAll):
			return m.EnableAll()
		case key.Matches(msg, m.keyMap.disableAll):
			return m.DisableAll()
		default:
			var cmd tea.Cmd
			m.content, cmd = m.content.Update(msg)
			return cmd
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitColumnVisibility:
		return m.SetState(msg)
	}
	return nil
}

func (m *Columns) SetState(msg messages.InitColumnVisibility) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.AllColumns = msg.AllColumns
	m.state.Visible = msg.Visible
	return m.updateContent()
}

func (m *Columns) EnableAll() tea.Cmd {
	for i := range m.state.Visible {
		m.state.Visible[i] = true
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *Columns) DisableAll() tea.Cmd {
	for i := range m.state.Visible {
		m.state.Visible[i] = false
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *Columns) updateContent() tea.Cmd {
	items := make([]list.Item, 0, len(m.state.AllColumns))
	for i := range m.state.AllColumns {
		items = append(items, checkboxItem{
			checked: m.state.Visible[i],
			name:    m.state.AllColumns[i],
		})
	}
	cmd := m.content.SetItems(items)
	m.updateSize()
	return cmd
}

func (m *Columns) selectItem() tea.Cmd {
	idx := m.content.Index()
	itm := m.content.SelectedItem()
	if idx > len(m.state.AllColumns) {
		panic("dialog state not up to date")
	}
	m.state.Visible[idx] = !m.state.Visible[idx]
	if typedItem, ok := itm.(checkboxItem); ok {
		typedItem.checked = m.state.Visible[idx]
		itm = typedItem
	}
	listUpdate := m.content.SetItem(idx, itm) // cmd for filtering
	columnUpdate := m.UpdateMessage()
	return tea.Batch(listUpdate, columnUpdate)
}

func (m *Columns) UpdateMessage() tea.Cmd {
	return func() tea.Msg {
		msg := messages.ColumnVisibilityUpdate{}
		msg.TableARN = m.state.TableARN
		msg.AllColumns = m.state.AllColumns
		msg.Visible = m.state.Visible
		return msg
	}
}

func (m *Columns) toggleDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleColumnVisibility{}
	}
}

func (m *Columns) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *Columns) updateSize() {
	items := m.content.Items()

	// set height of the list within the dialog
	padding := 4
	m.content.SetHeight(min(len(m.content.Items())+padding, m.window.height))

	// determine the width of the list within the dialog
	width := m.defaultDialogWidth
	for _, itm := range items {
		width = max(width, len(itm.(checkboxItem).name))
	}
	// set width of the list within the dialog
	m.content.SetWidth(width)

	// set height & width of dialog itself
	columnsDialogStyle = columnsDialogStyle.
		Height(m.content.Height() + 2).
		Width(width + 2)
}

func (m *Columns) View() string {
	return columnsDialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.title.Render(m.content.Title),
			m.styles.content.Render(m.content.View()),
			m.styles.help.Render(m.JoinedHelp()),
		),
	)
}

func (m *Columns) JoinedHelp() string {
	if !m.content.Help.ShowAll {
		helpV := m.content.Help.ShortHelpView
		helpLine := m.styles.helpLine
		return lipgloss.JoinVertical(lipgloss.Center,
			helpLine.Render(helpV(m.content.ShortHelp())),
			helpLine.Render(helpV([]key.Binding{m.keyMap.enableAll, m.keyMap.disableAll})),
		)
	}

	listBindings := m.content.FullHelp()
	firstCol := listBindings[0]
	firstCol = append(firstCol, m.keyMap.enableAll, m.keyMap.disableAll)
	listBindings[0] = firstCol
	return m.content.Help.FullHelpView(listBindings)
}
