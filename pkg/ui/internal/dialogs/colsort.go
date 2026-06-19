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

type columnSortingKeyMap struct {
	close key.Binding
	enter key.Binding
	reset key.Binding
}

func (h columnSortingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter}
}

// the ColumnSorting dialog enables the user to select a column for sorting ASC
// or DESC
type ColumnSorting struct {
	selected string

	keyMap columnSortingKeyMap

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
		SortingOn  string
		Ascending  bool // if false, descending
	}

	styles sortingListStyles

	content list.Model
}

type sortingListStyles struct {
	dialog       lipgloss.Style
	title        lipgloss.Style
	content      lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	help         lipgloss.Style
	helpLine     lipgloss.Style
}

func newColumnSortingStyles(darkBG bool) sortingListStyles {
	var s sortingListStyles
	s.dialog = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

type sortingItem struct {
	checked   bool
	ascending bool
	name      string
	idx       int
}

func (i sortingItem) FilterValue() string { return i.name }

type sortingItemDelegate struct {
	styles *sortingListStyles
}

func (d sortingItemDelegate) Height() int                             { return 1 }
func (d sortingItemDelegate) Spacing() int                            { return 0 }
func (d sortingItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d sortingItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sortingItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s %s", i.name, u.Ternary(u.Ternary("[ASC] ", "[DESC]", i.ascending), "      ", i.checked))

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewColumnSortingDialog(close key.Binding) *ColumnSorting {
	c := &ColumnSorting{
		keyMap: columnSortingKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
		},

		defaultDialogHeight: 46,
		defaultDialogWidth:  66,
	}

	c.styles = newColumnSortingStyles(true)

	c.dialog.width = c.defaultDialogWidth
	c.dialog.height = c.defaultDialogHeight

	c.window.width = 150
	c.window.height = 100

	{ // list
		l := list.New([]list.Item{}, sortingItemDelegate{}, c.dialog.width, c.dialog.height)
		l.Title = "Column Order"
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

func (m *ColumnSorting) newDelegate(s *sortingListStyles) sortingItemDelegate {
	return sortingItemDelegate{
		styles: s,
	}
}

func (m *ColumnSorting) updateStyles(isDark bool) {
	s := newColumnSortingStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *ColumnSorting) Init() tea.Cmd {
	return nil
}

func (m *ColumnSorting) Update(msg tea.Msg) tea.Cmd {
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
		case key.Matches(msg, m.keyMap.reset):
			return m.reset()
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitColumnSorting:
		return m.SetState(msg)
	}
	// default
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	m.updateSize()
	return cmd
}

func (m *ColumnSorting) SetState(msg messages.InitColumnSorting) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.AllColumns = msg.AllColumns
	m.state.Ascending = msg.Ascending
	m.state.SortingOn = msg.SortingOn
	return m.updateContent()
}

func (m *ColumnSorting) updateContent() tea.Cmd {
	items := make([]list.Item, 0, len(m.state.AllColumns))
	for i := range m.state.AllColumns {
		items = append(items, sortingItem{
			checked:   m.state.AllColumns[i] == m.state.SortingOn,
			name:      m.state.AllColumns[i],
			ascending: m.state.Ascending,
			idx:       i,
		})
	}
	cmd := m.content.SetItems(items)
	m.updateSize()
	return cmd
}

func (m *ColumnSorting) reset() tea.Cmd {
	m.state.SortingOn = ""
	m.state.Ascending = true
	m.updateContent()
	return func() tea.Msg {
		return messages.ColumnSortingReset{
			TableARN: m.state.TableARN,
		}
	}
}

func (m *ColumnSorting) selectItem() tea.Cmd {
	sel := m.content.SelectedItem().(sortingItem)
	idx := sel.idx
	items := m.content.Items()
	if idx > len(m.state.AllColumns) {
		panic("dialog state not up to date")
	}
	cmds := make([]tea.Cmd, 0)

	if sel.name != m.state.SortingOn {
		// when selecting new item, reset old item
		if m.state.SortingOn != "" {
			oldIdx := u.Find(m.state.AllColumns, m.state.SortingOn)
			itm := items[oldIdx].(sortingItem)
			itm.checked = false
			cmds = append(cmds, m.content.SetItem(oldIdx, itm))
		}

		// and initialise selected to ASC
		sel.checked = true
		sel.ascending = true
		cmds = append(cmds, m.content.SetItem(idx, sel))
	} else { // if already selected toggle sorting
		sel.ascending = !sel.ascending
		cmds = append(cmds, m.content.SetItem(idx, sel))
	}

	m.state.Ascending = sel.ascending
	m.state.SortingOn = sel.name

	cmds = append(cmds, m.UpdateMessage())
	return tea.Batch(cmds...)
}

func (m *ColumnSorting) UpdateMessage() tea.Cmd {
	return func() tea.Msg {
		msg := messages.ColumnSortingUpdate{}
		msg.TableARN = m.state.TableARN
		msg.AllColumns = m.state.AllColumns
		msg.SortingOn = m.state.SortingOn
		msg.Ascending = m.state.Ascending
		return msg
	}
}

func (m *ColumnSorting) toggleDialog() tea.Cmd {
	m.content.FilterInput.Reset()             // reset filter input value
	m.content.SetFilterState(list.Unfiltered) // set filter to inactive (hide & unfocus)
	return func() tea.Msg {
		return messages.ToggleColumnSorting{}
	}
}

func (m *ColumnSorting) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *ColumnSorting) updateSize() {
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
			m.dialog.width = u.Clamp(m.dialog.width, len(itm.(sortingItem).name)+contentW, m.window.width)
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

func (m *ColumnSorting) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderTitle(),
			m.renderContent(),
			m.renderFilter(),
			m.renderHelp(),
		),
	)
}

func (m *ColumnSorting) renderContent() string {
	return m.styles.content.Render(m.content.View())
}

func (m *ColumnSorting) renderFilter() string {
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		return m.content.FilterInput.View()
	}
	return lipgloss.NewStyle().Render("") // placeholder for filter
}

func (m *ColumnSorting) renderTitle() string {
	return m.styles.title.Render(m.content.Title)
}

func (m *ColumnSorting) renderHelp() string {
	return m.styles.help.Render(m.JoinedHelp())
}

func (m *ColumnSorting) JoinedHelp() string {
	if !m.content.Help.ShowAll {
		helpV := m.content.Help.ShortHelpView
		helpLine := m.styles.helpLine
		return lipgloss.JoinVertical(lipgloss.Center,
			helpLine.Render(helpV(m.content.ShortHelp())),
			helpLine.Render(helpV([]key.Binding{m.keyMap.reset})),
		)
	}

	listBindings := m.content.FullHelp()
	firstCol := listBindings[0]
	firstCol = append(firstCol, m.keyMap.reset)
	listBindings[0] = firstCol
	return m.content.Help.FullHelpView(listBindings)
}
