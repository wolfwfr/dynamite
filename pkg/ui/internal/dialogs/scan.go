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

var ScanDialogStyle = commonstyles.DialogStyle

type indexType int

const (
	table indexType = iota
	gsi
	lsi
)

const (
	tableIndexName string = "table"
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
}

type scanListStyles struct {
	title        lipgloss.Style
	content      lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	header       lipgloss.Style
	help         lipgloss.Style
	helpLine     lipgloss.Style
	keyInfo      lipgloss.Style
}

func newscanStyles(darkBG bool) scanListStyles {
	var s scanListStyles
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.header = lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0"))
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	s.keyInfo = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#636363")).Padding(1, 2, 1, 2)
	return s
}

type indexItem struct {
	name       string
	indexType  indexType
	sliceIndex int
}

func (i indexItem) FilterValue() string { return "" }

type scanItemDelegate struct {
	styles   *scanListStyles
	firstGSI *string
	firstLSI *string
}

func (d scanItemDelegate) Height() int                             { return 1 }
func (d scanItemDelegate) Spacing() int                            { return 0 }
func (d scanItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d scanItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(indexItem)
	if !ok {
		return
	}

	str := i.name

	// NOTE: not pretty but good enough for now
	headerFmt := func(s string) string {
		return d.styles.header.Render(fmt.Sprintf("\n\n%s\n%s", headerPadding(s, 30), "______________________________\n")) + "\n"
	}

	var header string
	if index == 0 { // is table index
		header = headerFmt("Table Index")
	} else if d.firstGSI != nil && i.name == *d.firstGSI {
		header = headerFmt("Global Secondary Indices")
	} else if d.firstLSI != nil && i.name == *d.firstLSI {
		header = headerFmt("Local Secondary Indices")
	}

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fmt.Sprintf("%s%s", header, fn(str)))
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
	r.dialog.width = r.defaultDialogWidth
	r.dialog.height = r.defaultDialogHeight

	r.window.width = 150
	r.window.height = 100

	l := list.New([]list.Item{}, scanItemDelegate{}, r.dialog.width, r.dialog.height)
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
	r.updateStyles(true) // default to dark styles.
	r.updateSize()

	return r
}

func (m *ScanDialog) newDelegate(s *scanListStyles) scanItemDelegate {
	var firstGSI *string
	var firstLSI *string
	if len(m.state.GSI) > 0 {
		firstGSI = &m.state.GSI[0].Name
	}
	if len(m.state.LSI) > 0 {
		firstLSI = &m.state.LSI[0].Name
	}
	return scanItemDelegate{
		styles:   s,
		firstGSI: firstGSI,
		firstLSI: firstLSI,
	}
}

func (m *ScanDialog) updateStyles(isDark bool) {
	s := newscanStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help
	s.keyInfo = s.keyInfo.Width(m.dialog.width - 10)

	m.styles = s
	m.content.SetDelegate(m.newDelegate(&s))
}

func (m *ScanDialog) Init() tea.Cmd {
	return nil
}

func (m *ScanDialog) Width() int {
	return m.dialog.width
}

func (m *ScanDialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.enter):
			return m.selectIndex()
		default:
			var cmd tea.Cmd
			m.content, cmd = m.content.Update(msg)
			return cmd
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitScanParameters:
		return m.SetState(msg)
	}
	return nil
}

func (m *ScanDialog) SetState(msg messages.InitScanParameters) tea.Cmd {
	m.state.TableARN = msg.TableARN
	m.state.TableIndex = msg.TableIndex
	m.state.GSI = msg.GSI
	m.state.LSI = msg.LSI

	m.updateStyles(true) // to set delegate
	return m.updateContent()
}

func (m *ScanDialog) updateContent() tea.Cmd {
	items := make([]list.Item, 0, 1+len(m.state.GSI)+len(m.state.LSI))
	items = append(items, indexItem{
		name:      tableIndexName,
		indexType: table,
	})
	for i, g := range m.state.GSI {
		items = append(items, indexItem{
			name:       g.Name,
			indexType:  gsi,
			sliceIndex: i,
		})
	}
	for i, l := range m.state.LSI {
		items = append(items, indexItem{
			name:       l.Name,
			indexType:  lsi,
			sliceIndex: i,
		})
	}
	cmd := m.content.SetItems(items)
	m.updateSize()
	return cmd
}

func (m *ScanDialog) selectIndex() tea.Cmd {
	itm := m.content.SelectedItem()
	selection := itm.(indexItem).name
	if selection == m.selected {
		return m.toggleDialog() // no change
	}
	return tea.Batch(m.changeIndex(), m.toggleDialog())
}

func (m *ScanDialog) changeIndex() tea.Cmd {
	m.selected = m.content.SelectedItem().(indexItem).name
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
}

func (m *ScanDialog) updateSize() {
	items := m.content.Items()

	// set height of the list within the dialog
	padding := 4
	m.content.SetHeight(min(len(m.content.Items())+padding, m.window.height))

	// determine the width of the list within the dialog
	width := m.defaultDialogWidth
	for _, itm := range items {
		width = max(width, len(itm.(indexItem).name))
	}
	// set width of the list within the dialog
	m.content.SetWidth(width)

	// set dialog size
	m.dialog.height = m.content.Height() + 2
	m.dialog.width = width + 2

	m.updateStyles(true)

	// set height & width of dialog itself
	ScanDialogStyle = ScanDialogStyle.
		Height(m.dialog.height).
		Width(m.dialog.width)

}

func (m *ScanDialog) View() string {
	title := m.styles.title.Render(m.content.Title)
	content := m.styles.content.Render(m.content.View())
	help := m.styles.help.Render(
		m.styles.helpLine.Render(m.content.Help.View(m.content)),
	)
	keyInfo := m.styles.keyInfo.Render(m.renderIndexInfo())
	return ScanDialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			content,
			keyInfo,
			help,
		),
	)
}

func (m *ScanDialog) renderIndexInfo() string {
	var hash, hashType, rangType string
	var rang *string

	sel, ok := m.content.SelectedItem().(indexItem)
	if !ok {
		return ""
	}
	switch sel.indexType {
	case table:
		i := m.state.TableIndex
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = i.RangeKey
		rangType = i.RangeKeyType
	case gsi:
		i := m.state.GSI[sel.sliceIndex]
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = i.RangeKey
		rangType = i.RangeKeyType
	case lsi:
		i := m.state.LSI[sel.sliceIndex]
		hash = i.HashKey
		hashType = i.HashKeyType
		rang = &i.RangeKey
		rangType = i.RangeKeyType
	}

	str := fmt.Sprintf("Hash Key  (%s): %s\n", hashType, hash)
	if rang != nil {
		str = fmt.Sprintf("%sRange Key (%s): %s", str, rangType, *rang)
	}
	return str
}
