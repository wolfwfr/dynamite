package dialogs

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	checkbox "github.com/wolfwfr/dynamite/pkg/ui/internal/components/checkbox_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
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
	checkbox.Styles
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
}

func newColumnStyles(darkBG bool) columnsListStyles {
	var s columnsListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)

	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

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
		defaultDialogWidth:  66,
	}
	c.dialog.width = c.defaultDialogWidth
	c.dialog.height = c.defaultDialogHeight

	c.window.width = 150
	c.window.height = 100

	{ // list
		l := list.New([]list.Item{}, checkbox.ItemDelegate{}, c.dialog.width, c.dialog.height)
		l.Title = "Column Visibility " // space for even length (helps with keeping centered alignment stable)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(true)
		l.SetShowFilter(false)
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

func (m *Columns) updateStyles(isDark bool) {
	s := newColumnStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *Columns) newDelegate(s *columnsListStyles) checkbox.ItemDelegate {
	return checkbox.ItemDelegate{
		Styles: &s.Styles,
	}
}

func (m *Columns) Init() tea.Cmd {
	return nil
}

func (m *Columns) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.content.FilterState() == list.Filtering ||
			m.content.IsFiltered() && key.Matches(msg, m.content.KeyMap.ClearFilter) {
			break // only perform filtering
		}
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

	// default
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return cmd
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
		items = append(items, checkbox.Item{
			Checked: m.state.Visible[i],
			Name:    m.state.AllColumns[i],
			Meta: map[string]any{
				"idx": i,
			},
		})
	}
	cmd := m.content.SetItems(items)
	m.updateSize()
	return cmd
}

func (m *Columns) selectItem() tea.Cmd {
	itm, ok := m.content.SelectedItem().(checkbox.Item)
	if !ok {
		return nil
	}
	idx := itm.Meta["idx"].(int)
	if idx > len(m.state.AllColumns) {
		panic("dialog state not up to date")
	}
	m.state.Visible[idx] = !m.state.Visible[idx]
	itm.Checked = m.state.Visible[idx]
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
	m.content.FilterInput.Reset()             // reset filter input value
	m.content.SetFilterState(list.Unfiltered) // set filter to inactive (hide & unfocus)
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
		width = max(width, len(itm.(checkbox.Item).Name))
	}
	// set width of the list within the dialog
	m.content.SetWidth(width)

	// set height & width of dialog itself
	columnsDialogStyle = columnsDialogStyle.
		Height(m.content.Height() + 2).
		Width(width + 2)
}

func (m *Columns) View() string {
	toRender := []string{
		m.styles.title.Render(m.content.Title),
		m.styles.content.Render(m.content.View()),
		lipgloss.NewStyle().Render(""), // placeholder for filter
		m.styles.help.Render(m.JoinedHelp()),
	}
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		toRender[2] = m.content.FilterInput.View()
	}
	return columnsDialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			toRender...,
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
