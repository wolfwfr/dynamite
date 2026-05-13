package dialogs

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	headed "github.com/wolfwfr/dynamite/pkg/ui/internal/components/headed_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type regionsKeyMap struct {
	close key.Binding
	enter key.Binding
}

func (h regionsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter}
}

// the Regions dialog enables the user to select an AWS-region
type Regions struct {
	available []string
	starred   []string
	unstarred []string
	selected  string

	keyMap regionsKeyMap

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

	styles regionListStyles

	content list.Model

	// collapseHeaders is set to true when the full list including headers does
	// not fit in the available height
	collapseHeaders bool
}

type regionListStyles struct {
	headed.Styles
	dialog   lipgloss.Style
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style

	starFullHeader  string
	normFullHeader  string
	starShortHeader string
	normShortHeader string

	headerFmt func(string) string
}

func newRegionStyles(darkBG bool) regionListStyles {
	var s regionListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.Header = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour)

	s.dialog = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)

	s.starFullHeader = "*  Starred  *"
	s.normFullHeader = "Normal"
	s.starShortHeader = " (starred)"
	s.normShortHeader = ""

	s.headerFmt = func(s string) string {
		return fmt.Sprintf("\n%s\n%s", headed.HeaderPadding(s, 17), "_________________\n")
	}

	return s
}

func NewRegionsDialog(available, starred []string, current string, close key.Binding) *Regions {
	r := &Regions{
		available: available,
		starred:   starred,
		selected:  current,

		styles: newRegionStyles(true),

		keyMap: regionsKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		},

		defaultDialogHeight: 60,
		defaultDialogWidth:  55,
	}
	r.dialog.width = r.defaultDialogWidth
	r.dialog.height = r.defaultDialogHeight

	r.window.width = 150
	r.window.height = 100

	var sorted []list.Item
	sorted, r.unstarred = compileSortedList(available, starred)

	l := list.New(sorted, headed.ItemDelegate{}, r.dialog.width, r.dialog.height)
	l.Title = "AWS Regions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	// replace '?' with 'm'
	l.KeyMap.ShowFullHelp.SetKeys("m")
	l.KeyMap.ShowFullHelp.SetHelp("m", "more")
	l.KeyMap.CloseFullHelp.SetKeys("m")
	l.KeyMap.CloseFullHelp.SetHelp("m", "close help")
	l.KeyMap.Quit.SetKeys(r.keyMap.close.Keys()...)
	l.KeyMap.Quit.SetHelp(r.keyMap.close.Help().Key, r.keyMap.close.Help().Desc)

	r.content = l
	r.updateSize()
	r.updateStyles(true) // default to dark styles.

	return r
}

func compileSortedList(available, starred []string) (full []list.Item, unstarred []string) {
	seen := map[string]struct{}{}
	items := make([]list.Item, 0, len(available))
	unstarred = make([]string, 0, max(0, len(available)-len(starred)))

	for _, s := range starred {
		items = append(items, headed.Item{Name: s})
		seen[s] = struct{}{}
	}

	for _, a := range available {
		if _, ok := seen[a]; ok {
			continue
		}
		items = append(items, headed.Item{Name: a})
		unstarred = append(unstarred, a)
	}

	return items, unstarred
}

func (m *Regions) newDelegate(s *regionListStyles) headed.ItemDelegate {
	d := headed.ItemDelegate{
		Styles:   &s.Styles,
		Collapse: m.collapseHeaders,
	}

	headerFmt := m.styles.headerFmt
	if len(m.starred) > 0 {
		firstStarred := m.starred[0]
		f := func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(m.styles.starFullHeader), "", i.Name == firstStarred)
		}
		if m.collapseHeaders {
			f = func(i headed.Item, _ int) string {
				return u.Ternary(m.styles.starShortHeader, "", slices.Contains(m.starred, i.Name))
			}
		}
		d.HeadedItems = append(d.HeadedItems, f)
	}

	if len(m.unstarred) > 0 {
		firstNormal := m.unstarred[0]
		f := func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(m.styles.normFullHeader), "", i.Name == firstNormal)
		}
		if m.collapseHeaders {
			f = func(i headed.Item, _ int) string { return m.styles.normShortHeader }
		}
		d.HeadedItems = append(d.HeadedItems, f)
	}

	return d
}

// NOTE: updateStyles is best executed after updateSize, to first determine the
// requisite `collapseHeaders` property.
func (m *Regions) updateStyles(isDark bool) {
	s := newRegionStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *Regions) Init() tea.Cmd {
	return nil
}

func (m *Regions) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.enter):
			return m.selectRegion()
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	}
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	m.updateSize()
	m.updateStyles(true)
	return cmd
}

func (m *Regions) selectRegion() tea.Cmd {
	itm := m.content.SelectedItem()
	selection := itm.(headed.Item).Name
	if selection == m.selected {
		return m.toggleDialog() // no change
	}
	return tea.Batch(m.switchRegion(m.selected, selection), m.toggleDialog())
}

func (m *Regions) switchRegion(oldr, newr string) tea.Cmd {
	m.selected = newr
	return func() tea.Msg {
		return messages.SwitchRegion{
			OldRegion: oldr,
			NewRegion: newr,
		}
	}
}

func (m *Regions) toggleDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleRegions{}
	}
}

func (m *Regions) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
	m.updateStyles(true)
}

func (m *Regions) updateSize() {
	var (
		// dialog
		maxDialogHeight = m.window.height
		maxDialogWidth  = m.window.width

		// dialog elements
		titleH   = lipgloss.Height(m.renderTitle())
		contentH = 0
		filterH  = lipgloss.Height(m.renderFilter())
		helpH    = lipgloss.Height(m.renderHelp())

		// borders
		bordersW = getBorderWidth(m.styles.dialog)
		bordersH = getBorderHeight(m.styles.dialog)

		// padding
		contentPH = getPadHeight(m.styles.content)
		contentPW = getPadWidth(m.styles.content)
		helpPW    = getPadWidth(m.styles.help)
	)

	m.dialog.height = min(m.defaultDialogHeight, maxDialogHeight)
	m.dialog.width = min(m.defaultDialogWidth, maxDialogWidth)

	{ // update list height
		var (
			maxContentH       = maxDialogHeight - (bordersH + titleH + filterH + helpH + contentPH)
			paginatorH        = lipgloss.Height(m.content.Styles.PaginationStyle.Render(m.content.Paginator.View())) + 1 // margin is set dynamically in list, cannot access; ergo '+1'
			numHeaders        = u.Ternary(2, 0, len(m.starred) > 0)
			headerH           = lipgloss.Height(m.styles.Header.Render(m.styles.headerFmt("test-header")))
			totalHeaderH      = numHeaders * headerH
			collapsedContentH = len(m.content.Items()) + paginatorH
			idealContentH     = collapsedContentH + totalHeaderH
			listAwareH        = idealContentH - totalHeaderH // delagate specifies each item height equals '1', list is not aware of header-height and shouldn't be
		)

		m.collapseHeaders = maxContentH < idealContentH
		contentH = u.Ternary(min(maxContentH, collapsedContentH), listAwareH, m.collapseHeaders)
		m.content.SetHeight(contentH)
	}

	{ // update list width
		var (
			contentW   = bordersW + max(contentPW, helpPW) // help is now coupled to content (see render)
			maxHeaderW = u.Ternary(max(len(m.styles.starShortHeader), len(m.styles.normShortHeader)), max(len(m.styles.starFullHeader), len(m.styles.normFullHeader)), m.collapseHeaders)
		)

		// determine the width of the list within the dialog
		items := m.content.Items()
		for _, itm := range items {
			m.dialog.width = u.Clamp(m.dialog.width, len(itm.(headed.Item).Name)+contentW+maxHeaderW, m.window.width)
		}

		// set width of the list within the dialog
		m.content.SetWidth(m.dialog.width - contentW)
	}

	m.dialog.height = min(bordersH+titleH+contentH+contentPH+filterH+helpH, m.window.height)

	// update dialog style size
	m.styles.dialog = m.styles.dialog.
		Height(m.dialog.height).
		Width(m.dialog.width)
}

func (m *Regions) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderTitle(),
			m.renderContent(),
			// m.renderFilter(),
			m.renderHelp(),
		),
	)
}

func (m *Regions) renderContent() string {
	return m.styles.content.Render(m.content.View())
}

func (m *Regions) renderFilter() string {
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		return m.content.FilterInput.View()
	}
	return lipgloss.NewStyle().Render("") // placeholder for filter
}

func (m *Regions) renderTitle() string {
	return m.styles.title.Render(m.content.Title)
}

func (m *Regions) renderHelp() string {
	return m.styles.help.Render(m.JoinedHelp())
}

func (m *Regions) JoinedHelp() string {
	if !m.content.Help.ShowAll {
		helpV := m.content.Help.ShortHelpView
		helpLine := m.styles.helpLine
		return lipgloss.JoinVertical(lipgloss.Center,
			helpLine.Render(helpV(m.content.ShortHelp())),
			helpLine.Render(helpV([]key.Binding{m.keyMap.enter})),
		)
	}

	listBindings := m.content.FullHelp()
	firstCol := listBindings[0]
	firstCol = append(firstCol, m.keyMap.enter)
	listBindings[0] = firstCol
	return m.content.Help.FullHelpView(listBindings)
}
