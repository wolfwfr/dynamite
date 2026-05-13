package dialogs

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	checkbox "github.com/wolfwfr/dynamite/pkg/ui/internal/components/checkbox_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

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
	dialog   lipgloss.Style
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
}

func newColumnStyles(darkBG bool) columnsListStyles {
	var s columnsListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)

	s.dialog = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

// the ColumnVis dialog enables the user to enable and disable visibility of
// individual columns
type ColumnVis struct {
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

func NewColumnVisibilityDialog(close key.Binding) *ColumnVis {
	c := &ColumnVis{
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

func (m *ColumnVis) updateStyles(isDark bool) {
	s := newColumnStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *ColumnVis) newDelegate(s *columnsListStyles) checkbox.ItemDelegate {
	return checkbox.ItemDelegate{
		Styles: &s.Styles,
	}
}

func (m *ColumnVis) Init() tea.Cmd {
	return nil
}

func (m *ColumnVis) Update(msg tea.Msg) tea.Cmd {
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
	m.updateSize()
	return cmd
}

func (m *ColumnVis) SetState(msg messages.InitColumnVisibility) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.AllColumns = msg.AllColumns
	m.state.Visible = msg.Visible
	return m.updateContent()
}

func (m *ColumnVis) EnableAll() tea.Cmd {
	for i := range m.state.Visible {
		m.state.Visible[i] = true
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *ColumnVis) DisableAll() tea.Cmd {
	for i := range m.state.Visible {
		m.state.Visible[i] = false
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *ColumnVis) updateContent() tea.Cmd {
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

func (m *ColumnVis) selectItem() tea.Cmd {
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

func (m *ColumnVis) UpdateMessage() tea.Cmd {
	return func() tea.Msg {
		msg := messages.ColumnVisibilityUpdate{}
		msg.TableARN = m.state.TableARN
		msg.AllColumns = m.state.AllColumns
		msg.Visible = m.state.Visible
		return msg
	}
}

func (m *ColumnVis) toggleDialog() tea.Cmd {
	m.content.FilterInput.Reset()             // reset filter input value
	m.content.SetFilterState(list.Unfiltered) // set filter to inactive (hide & unfocus)
	return func() tea.Msg {
		return messages.ToggleColumnVisibility{}
	}
}

func (m *ColumnVis) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *ColumnVis) updateSize() {
	m.dialog.height = min(m.defaultDialogHeight, m.window.height)
	m.dialog.width = min(m.defaultDialogWidth, m.window.width)

	var (
		titleH   = lipgloss.Height(m.renderTitle())
		contentH = 0
		filterH  = lipgloss.Height(m.renderFilter())
		helpH    = lipgloss.Height(m.renderHelp())

		bordersW = m.styles.dialog.GetBorderLeftSize() + m.styles.dialog.GetBorderRightSize()
		bordersH = m.styles.dialog.GetBorderBottomSize() + m.styles.dialog.GetBorderTopSize()

		contentPH = m.styles.content.GetPaddingBottom() + m.styles.content.GetPaddingTop()
		contentPW = m.styles.content.GetPaddingLeft() + m.styles.content.GetPaddingRight()
		helpPW    = m.styles.help.GetPaddingLeft() + m.styles.help.GetPaddingRight()
	)

	{ // update list height
		maxContentH := m.dialog.height
		maxContentH -= (bordersH + titleH + filterH + helpH + contentPH)

		// leave room for inline paginator + paginator padding
		paginatorH := 2

		// set height of the list within the dialog
		contentH = min(maxContentH, len(m.content.Items())+paginatorH)
		m.content.SetHeight(contentH)
	}

	{ // update list width
		contentW := bordersW + max(contentPW, helpPW) // help is now coupled to content (see render)

		// determine the width of the list within the dialog
		items := m.content.Items()
		for _, itm := range items {
			m.dialog.width = u.Clamp(m.dialog.width, len(itm.(checkbox.Item).Name)+contentW, m.window.width)
		}

		// set width of the list within the dialog
		// TODO: help menu goes funky when at width between 55 and 57, uncertain why
		m.content.SetWidth(m.dialog.width - contentW)
	}

	m.dialog.height = min(bordersH+titleH+contentH+contentPH+filterH+helpH, m.window.height)

	// update dialog style size
	m.styles.dialog = m.styles.dialog.
		Height(m.dialog.height).
		Width(m.dialog.width)
}

func (m *ColumnVis) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderTitle(),
			m.renderContent(),
			m.renderFilter(),
			m.renderHelp(),
		),
	)
}

func (m *ColumnVis) renderContent() string {
	return m.styles.content.Render(m.content.View())
}

func (m *ColumnVis) renderFilter() string {
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		return m.content.FilterInput.View()
	}
	return lipgloss.NewStyle().Render("") // placeholder for filter
}

func (m *ColumnVis) renderTitle() string {
	return m.styles.title.Render(m.content.Title)
}

func (m *ColumnVis) renderHelp() string {
	return m.styles.help.Render(m.JoinedHelp())
}

func (m *ColumnVis) JoinedHelp() string {
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
