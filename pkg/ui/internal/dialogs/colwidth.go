package dialogs

import (
	"context"
	"log/slog"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/logging"
	checkbox "github.com/wolfwfr/dynamite/pkg/ui/internal/components/checkbox_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/theme"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type widthKeyMap struct {
	close      key.Binding
	enter      key.Binding
	reset      key.Binding
	enableAll  key.Binding
	disableAll key.Binding
}

func (h widthKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter, h.reset}
}

type widthListStyles struct {
	checkbox.Styles
	dialog   lipgloss.Style
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
}

func newWidthStyles(darkBG bool) widthListStyles {
	var s widthListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(theme.DialogFocusColour)

	s.dialog = theme.DialogStyle
	s.title = lipgloss.NewStyle().Foreground(theme.TitleFG).Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

// the WidthDialog dialog enables the user to select a column to adjust the width
type WidthDialog struct {
	ctx context.Context

	logger *slog.Logger

	keyMap widthKeyMap

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

	styles widthListStyles

	state struct {
		TableARN   string
		AllColumns []string // matching by index
		DynWidth   []bool   // matching by index
	}

	content list.Model
}

func NewWidthDialog(ctx context.Context, logger *slog.Logger, close key.Binding) *WidthDialog {
	c := &WidthDialog{
		ctx:    ctx,
		logger: logger.With(slog.String(logging.DialogKey, "column-width")),
		keyMap: widthKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
			enableAll: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "enable all"),
			),
			disableAll: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "disable all"),
			),
			reset: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl+r", "reset"),
			),
		},
		defaultDialogHeight: 46,
		defaultDialogWidth:  66,
	}

	c.styles = newWidthStyles(true)

	c.dialog.width = c.defaultDialogWidth
	c.dialog.height = c.defaultDialogHeight

	c.window.width = 150
	c.window.height = 100

	{ // list
		l := list.New([]list.Item{}, checkbox.ItemDelegate{}, c.dialog.width, c.dialog.height)
		l.Title = "Column Dynamic Width"
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

func (m *WidthDialog) updateStyles(isDark bool) {
	s := newWidthStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *WidthDialog) newDelegate(s *widthListStyles) checkbox.ItemDelegate {
	return checkbox.ItemDelegate{
		Styles: &s.Styles,
	}
}

func (m *WidthDialog) Init() tea.Cmd {
	return nil
}

func (m *WidthDialog) Update(msg tea.Msg) tea.Cmd {
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
			return m.DisableAll()
		case key.Matches(msg, m.keyMap.enableAll):
			return m.EnableAll()
		case key.Matches(msg, m.keyMap.disableAll):
			return m.DisableAll()
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitColumnWidth:
		return m.SetState(msg)
	}

	// default
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	m.updateSize()
	return cmd
}

func (m *WidthDialog) SetState(msg messages.InitColumnWidth) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.AllColumns = msg.AllColumns
	m.state.DynWidth = msg.DynWidth
	return m.updateContent()
}

func (m *WidthDialog) EnableAll() tea.Cmd {
	for i := range m.state.DynWidth {
		m.state.DynWidth[i] = true
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *WidthDialog) DisableAll() tea.Cmd {
	for i := range m.state.DynWidth {
		m.state.DynWidth[i] = false
	}
	return tea.Batch(m.updateContent(), m.UpdateMessage())
}

func (m *WidthDialog) updateContent() tea.Cmd {
	items := make([]list.Item, 0, len(m.state.AllColumns))
	for i := range m.state.AllColumns {
		items = append(items, checkbox.Item{
			Checked: m.state.DynWidth[i],
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

func (m *WidthDialog) selectItem() tea.Cmd {
	itm, ok := m.content.SelectedItem().(checkbox.Item)
	if !ok {
		return nil
	}
	idx := itm.Meta["idx"].(int)
	if idx >= len(m.state.AllColumns) {
		m.logger.Error("content returned index that exceeds maximum",
			slog.Int("selected_item_index", idx),
			slog.Int("n_columns", len(m.state.AllColumns)),
		)
		panic("dialog state not up to date")
	}
	m.state.DynWidth[idx] = !m.state.DynWidth[idx]
	itm.Checked = m.state.DynWidth[idx]
	listUpdate := m.content.SetItem(idx, itm) // cmd for filtering
	columnUpdate := m.UpdateMessage()
	return tea.Batch(listUpdate, columnUpdate)
}

func (m *WidthDialog) UpdateMessage() tea.Cmd {
	return func() tea.Msg {
		msg := messages.ColumnWidthUpdate{}
		msg.TableARN = m.state.TableARN
		msg.AllColumns = m.state.AllColumns
		msg.DynWidth = m.state.DynWidth
		return msg
	}
}

func (m *WidthDialog) toggleDialog() tea.Cmd {
	m.content.FilterInput.Reset()             // reset filter input value
	m.content.SetFilterState(list.Unfiltered) // set filter to inactive (hide & unfocus)
	return func() tea.Msg {
		return messages.ToggleColumnWidthDialog{}
	}
}

func (m *WidthDialog) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *WidthDialog) updateSize() {
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

func (m *WidthDialog) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderTitle(),
			m.renderContent(),
			m.renderFilter(),
			m.renderHelp(),
		),
	)
}

func (m *WidthDialog) renderContent() string {
	return m.styles.content.Render(m.content.View())
}

func (m *WidthDialog) renderFilter() string {
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		return m.content.FilterInput.View()
	}
	return lipgloss.NewStyle().Render("") // placeholder for filter
}

func (m *WidthDialog) renderTitle() string {
	return m.styles.title.Render(m.content.Title)
}

func (m *WidthDialog) renderHelp() string {
	return m.styles.help.Render(m.JoinedHelp())
}

func (m *WidthDialog) JoinedHelp() string {
	if !m.content.Help.ShowAll {
		helpV := m.content.Help.ShortHelpView
		helpLine := m.styles.helpLine
		return lipgloss.JoinVertical(lipgloss.Center,
			helpLine.Render(helpV(m.content.ShortHelp())),
			helpLine.Render(helpV([]key.Binding{m.keyMap.reset, m.keyMap.enableAll, m.keyMap.disableAll})),
		)
	}

	listBindings := m.content.FullHelp()
	firstCol := listBindings[0]
	firstCol = append(firstCol, m.keyMap.reset, m.keyMap.enableAll, m.keyMap.disableAll)
	listBindings[0] = firstCol
	return m.content.Help.FullHelpView(listBindings)
}
