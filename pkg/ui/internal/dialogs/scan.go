package dialogs

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	headed "github.com/wolfwfr/dynamite/pkg/ui/internal/components/headed_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type indexType int

const (
	table indexType = iota
	gsi
	lsi
)

const (
	tableIndexName string = "table"
	metaKey        string = "meta"
)

type scanKeyMap struct {
	close key.Binding
	enter key.Binding
}

func (h scanKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter}
}

// the ScanDialog dialog enables the user to select an index to scan
type ScanDialog struct {
	selected string

	keyMap scanKeyMap

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
		TableIndex messages.TableIndex
		GSI        []messages.GlobalSecondaryIndex
		LSI        []messages.LocalSecondaryIndex
	}

	styles scanListStyles

	content list.Model

	// collapseHeaders is set to true when the full list including headers does
	// not fit in the available height
	collapseHeaders bool
}

type scanListStyles struct {
	headed.Styles

	dialog   lipgloss.Style
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
	keyInfo  lipgloss.Style

	tableFullHeader  string
	gsiFullHeader    string
	lsiFullHeader    string
	tableShortHeader string
	gsiShortHeader   string
	lsiShortHeader   string

	headerFmt func(s string) string
}

func newscanStyles(darkBG bool) scanListStyles {
	var s scanListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.Header = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour)

	s.dialog = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	s.keyInfo = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogUnfocusColour).Padding(1, 2, 1, 2)

	s.tableFullHeader = "Table Index"
	s.gsiFullHeader = "Global Secondary Indices"
	s.lsiFullHeader = "Local Secondary Indices"
	s.tableShortHeader = " (table)"
	s.gsiShortHeader = " (GSI)"
	s.lsiShortHeader = " (LSI)"

	s.headerFmt = func(s string) string {
		return fmt.Sprintf("\n\n%s\n%s", headed.HeaderPadding(s, 30), "______________________________\n")
	}

	return s
}

type indexItemMeta struct {
	indexType  indexType
	sliceIndex int
}

func NewScanDialog(close key.Binding) *ScanDialog {
	r := &ScanDialog{
		keyMap: scanKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
		},

		defaultDialogHeight: 46,
		defaultDialogWidth:  55,
	}

	r.styles = newscanStyles(true)

	r.dialog.width = r.defaultDialogWidth
	r.dialog.height = r.defaultDialogHeight

	r.window.width = 150
	r.window.height = 100

	l := list.New([]list.Item{}, headed.ItemDelegate{}, r.dialog.width, r.dialog.height)
	l.Title = "Scan Parameters"
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

func (m *ScanDialog) newDelegate(s *scanListStyles) headed.ItemDelegate {
	headerFmt := m.styles.headerFmt
	d := headed.ItemDelegate{
		Styles:   &s.Styles,
		Collapse: m.collapseHeaders,
		HeadedItems: []headed.HeaderDelegate{
			func(i headed.Item, ix int) string { return u.Ternary(headerFmt(m.styles.tableFullHeader), "", ix == 0) },
		},
	}

	if m.collapseHeaders {
		d.HeadedItems[0] = func(i headed.Item, ix int) string { return u.Ternary(m.styles.tableShortHeader, "", ix == 0) }
	}

	if len(m.state.GSI) > 0 {
		firstGSI := m.state.GSI[0].Name
		f := func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(m.styles.gsiFullHeader), "", i.Name == firstGSI)
		}
		if m.collapseHeaders {
			f = func(i headed.Item, _ int) string {
				return u.Ternary(m.styles.gsiShortHeader, "", u.ContainsBy(m.state.GSI, func(e messages.GlobalSecondaryIndex) bool {
					return e.Name == i.Name
				}))
			}
		}
		d.HeadedItems = append(d.HeadedItems, f)
	}
	if len(m.state.LSI) > 0 {
		firstLSI := m.state.LSI[0].Name
		f := func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(m.styles.lsiFullHeader), "", i.Name == firstLSI)
		}
		if m.collapseHeaders {
			f = func(i headed.Item, _ int) string {
				return u.Ternary(m.styles.lsiShortHeader, "", u.ContainsBy(m.state.LSI, func(e messages.LocalSecondaryIndex) bool {
					return e.Name == i.Name
				}))
			}
		}
		d.HeadedItems = append(d.HeadedItems, f)
	}

	return d
}

func (m *ScanDialog) updateStyles(isDark bool) {
	s := newscanStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help
	s.keyInfo = s.keyInfo.Width(m.dialog.width - 10)

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *ScanDialog) Init() tea.Cmd {
	return nil
}

func (m *ScanDialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.enter):
			return m.selectIndex()
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitScanParameters:
		return m.SetState(msg)
	}
	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	m.updateSize()
	m.updateStyles(true)
	return cmd
}

func (m *ScanDialog) ResetState() {
	m.state.TableARN = ""
	m.state.TableIndex = messages.TableIndex{}
	m.state.GSI = nil
	m.state.LSI = nil
	m.selected = ""
	m.content.SetItems([]list.Item{})
	m.content.Select(0)
}

func (m *ScanDialog) SetState(msg messages.InitScanParameters) tea.Cmd {
	m.ResetState()

	m.state.TableARN = msg.TableARN
	m.state.TableIndex = msg.TableIndex
	m.state.GSI = msg.GSI
	m.state.LSI = msg.LSI
	if msg.CurrentIndex != nil {
		m.selected = *msg.CurrentIndex
	}

	m.updateStyles(true) // to set delegate
	return m.updateContent()
}

func (m *ScanDialog) updateContent() tea.Cmd {
	var idx int
	items := make([]list.Item, 0, 1+len(m.state.GSI)+len(m.state.LSI))
	items = append(items, headed.Item{
		Name: tableIndexName,
		Meta: map[string]any{metaKey: indexItemMeta{
			indexType: table,
		}},
	})
	for i, g := range m.state.GSI {
		items = append(items, headed.Item{
			Name: g.Name,
			Meta: map[string]any{metaKey: indexItemMeta{
				indexType:  lsi,
				sliceIndex: i,
			}},
		})
		if g.Name == m.selected {
			idx = len(items) - 1
		}
	}
	for i, l := range m.state.LSI {
		items = append(items, headed.Item{
			Name: l.Name,
			Meta: map[string]any{metaKey: indexItemMeta{
				indexType:  lsi,
				sliceIndex: i,
			}},
		})
		if l.Name == m.selected {
			idx = len(items) - 1
		}
	}
	cmd := m.content.SetItems(items)
	m.content.Select(idx)
	m.updateSize()
	return cmd
}

func (m *ScanDialog) selectIndex() tea.Cmd {
	itm := m.content.SelectedItem()
	selection := itm.(headed.Item).Name
	if selection == m.selected {
		return m.toggleDialog() // no change
	}
	return tea.Batch(m.changeIndex(), m.toggleDialog())
}

func (m *ScanDialog) changeIndex() tea.Cmd {
	m.selected = m.content.SelectedItem().(headed.Item).Name
	return func() tea.Msg {
		return messages.ScanIndexChanged{
			TableARN:  m.state.TableARN,
			IndexName: u.Ternary(m.selected, "", m.selected != tableIndexName),
		}
	}
}

func (m *ScanDialog) toggleDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleScanParameters{}
	}
}

func (m *ScanDialog) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
	m.updateStyles(true)
}

func (m *ScanDialog) updateSize() {
	var (
		// dialog
		maxDialogHeight = m.window.height
		maxDialogWidth  = m.window.width

		// dialog elements
		titleH   = lipgloss.Height(m.renderTitle())
		contentH = 0
		filterH  = lipgloss.Height(m.renderFilter())
		idxInfoH = lipgloss.Height(m.renderIndexInfo())
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
			maxContentH       = maxDialogHeight - (bordersH + titleH + filterH + idxInfoH + helpH + contentPH)
			paginatorH        = lipgloss.Height(m.content.Styles.PaginationStyle.Render(m.content.Paginator.View())) + 1 // margin is set dynamically in list, cannot access; ergo '+1'
			gsilen            = len(m.state.GSI)
			lsilen            = len(m.state.LSI)
			numHeaders        = u.Ternary(1, u.Ternary(2, 3, gsilen > 0 && lsilen == 0), gsilen+lsilen == 0)
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
			maxHeaderW = u.Ternary(max(len(m.styles.tableShortHeader), len(m.styles.gsiShortHeader), len(m.styles.lsiShortHeader)),
				max(len(m.styles.tableFullHeader), len(m.styles.gsiFullHeader), len(m.styles.lsiFullHeader)), m.collapseHeaders)
		)

		// determine the width of the list within the dialog
		items := m.content.Items()
		for _, itm := range items {
			m.dialog.width = u.Clamp(m.dialog.width, len(itm.(headed.Item).Name)+contentW+maxHeaderW, m.window.width)
		}

		// set width of the list within the dialog
		m.content.SetWidth(m.dialog.width - contentW)
	}

	m.dialog.height = min(bordersH+titleH+contentH+contentPH+filterH+idxInfoH+helpH, m.window.height)

	// update dialog style size
	m.styles.dialog = m.styles.dialog.
		Height(m.dialog.height).
		Width(m.dialog.width)
}

func (m *ScanDialog) View() string {
	return m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderTitle(),
			m.renderContent(),
			m.renderIndexInfo(),
			m.renderHelp(),
		),
	)
}

func (m *ScanDialog) renderContent() string {
	return m.styles.content.Render(m.content.View())
}

func (m *ScanDialog) renderFilter() string {
	if m.content.FilterState() != list.Unfiltered {
		m.content.FilterInput.SetWidth(len(m.content.FilterInput.Value()) + 2) // ensure filter stays centered and stable during cursor blinking
		return m.content.FilterInput.View()
	}
	return lipgloss.NewStyle().Render("") // placeholder for filter
}

func (m *ScanDialog) renderTitle() string {
	return m.styles.title.Render(m.content.Title)
}

func (m *ScanDialog) renderHelp() string {
	return m.styles.help.Render(m.JoinedHelp())
}

func (m *ScanDialog) JoinedHelp() string {
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

func (m *ScanDialog) renderIndexInfo() string {
	var hash, hashType, rangType string
	var rang *string

	sel, ok := m.content.SelectedItem().(headed.Item)
	if !ok {
		return ""
	}
	meta, ok := sel.Meta[metaKey].(indexItemMeta)
	if !ok {
		return ""
	}
	switch meta.indexType {
	case table:
		i := m.state.TableIndex
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = i.RangeKey
		rangType = i.RangeKeyType
	case gsi:
		i := m.state.GSI[meta.sliceIndex]
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = i.RangeKey
		rangType = i.RangeKeyType
	case lsi:
		i := m.state.LSI[meta.sliceIndex]
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = &i.RangeKey
		rangType = i.RangeKeyType
	}

	str := fmt.Sprintf("Hash Key  (%s): %s\n", hashType, hash)
	if rang != nil {
		str = fmt.Sprintf("%sRange Key (%s): %s", str, rangType, *rang)
	}
	return m.styles.keyInfo.Render(str)
}
