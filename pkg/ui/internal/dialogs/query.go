package dialogs

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

var queryDialogStyle = commonstyles.DialogStyle
var queryOperatorDialogStyle = commonstyles.DialogStyle.Border(lipgloss.RoundedBorder(), true, true, true, false).Padding(3, 3, 0, 0)

type queryKeyMap struct {
	close key.Binding
	enter key.Binding
	tab   key.Binding
	shtab key.Binding
}

func (h queryKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter, h.tab}
}

type queryDialogFocus int
type rangeOrder string

const (
	queryIndexSelection queryDialogFocus = iota
	queryHashKeyInput
	queryOperatorField
	queryRangeKeyInput1
	queryRangeKeyInput2
	queryOrderSelection
	queryApplyButton
)

const (
	rangeAscending  rangeOrder = "Ascending"
	rangeDescending rangeOrder = "Descending"
)

// the Queryialog dialog enables the user to select an index to query and
// provide inputs for the table or index keys
type Queryialog struct {
	focus queryDialogFocus

	keyMap queryKeyMap

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
		// init state is the state at initialisation and after the user commits
		// (or applies) their changes.
		init struct {
			selectedIndex string

			hashKeyValue     string
			rangeKeyValue    *string
			rangeKeyValue2   *string
			rangeKeyOperator messages.QueryOperator
			orderDescending  bool // default to ascending
		}
		// table state is set excusively on initialisation
		table struct {
			TableARN   string
			TableIndex messages.TableIndex
			GSI        []messages.GlobalSecondaryIndex
			LSI        []messages.LocalSecondaryIndex
		}
		// resolved state is state that is resolved from the user input
		resolved struct {
			// resolved from selected index
			HashKey     string
			HashKeyType string

			// resolved from selected index
			RangeKey     *string
			RangeKeyType string
		}
	}

	styles queryListStyles

	content struct {
		indexSelection      list.Model
		operatorSelection   list.Model
		rangeOrderSelection list.Model
		hashKeyInput        textinput.Model
		rangeKeyInput1      textinput.Model
		rangeKeyInput2      textinput.Model
	}
}

// TODO: turn into shared, generic list item for all dialogs
type queryListItemStyles struct {
	item         lipgloss.Style
	selectedItem lipgloss.Style
}

type queryListItem string

func (i queryListItem) FilterValue() string { return "" }

type queryItemDelegate struct {
	styles *queryListItemStyles
}

func (d queryItemDelegate) Height() int                             { return 1 }
func (d queryItemDelegate) Spacing() int                            { return 0 }
func (d queryItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d queryItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(queryListItem)
	if !ok {
		return
	}

	str := string(i)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type queryListStyles struct {
	indexItemStyles
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style

	// box at width of content
	narrowBox        lipgloss.Style
	narrowBoxFocused lipgloss.Style

	// box at full width of dialog
	wideBox        lipgloss.Style
	wideBoxFocused lipgloss.Style

	// titles
	hashKeyInputTitle  lipgloss.Style
	rangeKeyInputTitle lipgloss.Style
	rangeKeyOrderTitle lipgloss.Style

	applyButton        lipgloss.Style
	applyButtonFocused lipgloss.Style
}

func newQueryStyles(darkBG bool) queryListStyles {
	focusedColour := lipgloss.Color("#F58427")
	unFocusedColour := lipgloss.Color("#636363")
	headerColour := lipgloss.Color("#B0B0B0")
	var s queryListStyles
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.header = lipgloss.NewStyle().Foreground(headerColour)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)

	// narrow boxes
	s.narrowBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(unFocusedColour).Padding(0, 1, 0, 1)
	s.narrowBoxFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(focusedColour).Padding(0, 1, 0, 1)

	// wide boxes
	s.wideBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(unFocusedColour)
	s.wideBoxFocused = s.wideBox.BorderForeground(focusedColour)

	// inputs fields
	s.hashKeyInputTitle = lipgloss.NewStyle().PaddingLeft(1).Foreground(headerColour)
	s.rangeKeyInputTitle = lipgloss.NewStyle().PaddingLeft(1).Foreground(headerColour).Padding(1, 0, 0, 0)
	s.rangeKeyOrderTitle = lipgloss.NewStyle().PaddingLeft(1).Foreground(headerColour).Padding(1, 0, 0, 0)

	// query button
	s.applyButton = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(unFocusedColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)
	s.applyButtonFocused = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(focusedColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)

	return s
}

func NewQueryDialog(close key.Binding) *Queryialog {
	d := &Queryialog{
		keyMap: queryKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
			tab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "move focus forward"),
			),
			shtab: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "move focus backward"),
			),
		},

		defaultDialogHeight: 46,
		defaultDialogWidth:  55,
	}
	d.dialog.width = d.defaultDialogWidth
	d.dialog.height = d.defaultDialogHeight

	d.window.width = 150
	d.window.height = 100

	{ // index selection
		l := list.New([]list.Item{}, indexItemDelegate{}, d.dialog.width, d.dialog.height)
		l.Title = "Query Parameters"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		// replace '?' with 'm'
		l.KeyMap.ShowFullHelp.SetKeys("m")
		l.KeyMap.ShowFullHelp.SetHelp("m", "more")
		l.KeyMap.CloseFullHelp.SetKeys("m")
		l.KeyMap.CloseFullHelp.SetHelp("m", "close help")
		l.KeyMap.Quit.SetKeys(d.keyMap.close.Keys()...)
		l.KeyMap.Quit.SetHelp(d.keyMap.close.Help().Key, d.keyMap.close.Help().Desc)

		d.content.indexSelection = l
	}

	{ // operator selection
		l := list.New([]list.Item{}, queryItemDelegate{}, d.dialog.width, d.dialog.height)
		l.Title = "Range Key Operator"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		// replace '?' with 'm'
		l.KeyMap.ShowFullHelp.SetKeys("m")
		l.KeyMap.ShowFullHelp.SetHelp("m", "more")
		l.KeyMap.CloseFullHelp.SetKeys("m")
		l.KeyMap.CloseFullHelp.SetHelp("m", "close help")
		l.KeyMap.Quit.SetKeys(d.keyMap.close.Keys()...)
		l.KeyMap.Quit.SetHelp(d.keyMap.close.Help().Key, d.keyMap.close.Help().Desc)

		d.content.operatorSelection = l
	}

	{ // range order selection
		l := list.New([]list.Item{}, queryItemDelegate{}, d.dialog.width, d.dialog.height)
		l.Title = "List Order Selection"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		d.content.rangeOrderSelection = l
	}

	{ // hash key input
		hashKeyInput := textinput.New()
		d.content.hashKeyInput = hashKeyInput
	}
	{ // range key input
		rangeKeyInput := textinput.New()
		d.content.rangeKeyInput1 = rangeKeyInput
	}
	{ // range key input 2
		rangeKeyInput := textinput.New()
		d.content.rangeKeyInput2 = rangeKeyInput
	}

	d.updateStyles(true) // default to dark styles.
	d.updateSize()

	return d
}

func (m *Queryialog) newIndexDelegate(s *queryListStyles) indexItemDelegate {
	var firstGSI *string
	var firstLSI *string
	if len(m.state.table.GSI) > 0 {
		firstGSI = &m.state.table.GSI[0].Name
	}
	if len(m.state.table.LSI) > 0 {
		firstLSI = &m.state.table.LSI[0].Name
	}
	return indexItemDelegate{
		styles:   &s.indexItemStyles,
		firstGSI: firstGSI,
		firstLSI: firstLSI,
		focus:    m.focus == queryIndexSelection,
	}
}

func (m *Queryialog) newQueryItemDelegate(s *queryListStyles) queryItemDelegate {
	return queryItemDelegate{
		styles: &queryListItemStyles{
			item:         m.styles.indexItemStyles.item,         // use same styling
			selectedItem: m.styles.indexItemStyles.selectedItem, // use same styling
		},
	}
}

func (m *Queryialog) updateStyles(isDark bool) {
	s := newQueryStyles(isDark)
	m.content.indexSelection.Styles.Title = s.title
	m.content.indexSelection.Styles.HelpStyle = s.help

	subwidth := m.dialog.width - 10

	s.wideBox = s.wideBox.Width(subwidth)
	s.wideBoxFocused = s.wideBoxFocused.Width(subwidth)

	s.hashKeyInputTitle = s.hashKeyInputTitle.Width(subwidth)
	s.rangeKeyInputTitle = s.rangeKeyInputTitle.Width(subwidth)
	s.rangeKeyOrderTitle = s.rangeKeyOrderTitle.Width(subwidth)

	m.styles = s

	m.content.indexSelection.SetDelegate(m.newIndexDelegate(&s))
	m.content.operatorSelection.SetDelegate(m.newQueryItemDelegate(&s))
	m.content.rangeOrderSelection.SetDelegate(m.newQueryItemDelegate(&s))
}

func (m *Queryialog) Init() tea.Cmd {
	return nil
}

func (m *Queryialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.tab):
			return m.MoveFocus(1)
		case key.Matches(msg, m.keyMap.shtab):
			return m.MoveFocus(-1)
		default:
			return m.handleNavigation(msg)
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitQueryParameters:
		return m.SetState(msg)
	default:
		return m.handleNavigation(msg)
	}
	return nil
}

func (m *Queryialog) handleNavigation(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.focus {
	case queryIndexSelection:
		m.content.indexSelection, cmd = m.content.indexSelection.Update(msg)
	case queryHashKeyInput:
		m.content.hashKeyInput, cmd = m.content.hashKeyInput.Update(msg)
	case queryOperatorField:
		m.content.operatorSelection, cmd = m.content.operatorSelection.Update(msg)
	case queryRangeKeyInput1:
		m.content.rangeKeyInput1, cmd = m.content.rangeKeyInput1.Update(msg)
	case queryRangeKeyInput2:
		m.content.rangeKeyInput2, cmd = m.content.rangeKeyInput2.Update(msg)
	case queryOrderSelection:
		m.content.rangeOrderSelection, cmd = m.content.rangeOrderSelection.Update(msg)
	case queryApplyButton:
		if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keyMap.enter) {
			cmd = m.applyParameters()
		}
	}

	m.resolveIndexInfo()
	cmd = tea.Batch(cmd, m.updateContent())

	return cmd
}

func (m *Queryialog) MoveFocus(i int) tea.Cmd {
	switch m.focus {
	case queryIndexSelection:
	// nothing to do
	case queryHashKeyInput:
		m.content.hashKeyInput.Blur()
	case queryOperatorField:
		// nothing to do
	case queryRangeKeyInput1:
		m.content.rangeKeyInput1.Blur()
	case queryRangeKeyInput2:
		m.content.rangeKeyInput2.Blur()
	case queryOrderSelection:
		// nothing to do
	case queryApplyButton:
		// nothing to do
	}

	m.focus += queryDialogFocus(i)
	if m.focus > queryApplyButton {
		m.focus = 0
	} else if m.focus < 0 {
		m.focus = queryApplyButton
	}

	// range-input-2 only applies when 'between' operator is selected
	if m.focus == queryRangeKeyInput2 && messages.QueryOperator(m.content.operatorSelection.SelectedItem().(queryListItem)) != messages.Between {
		m.focus += queryDialogFocus(i)
	}

	switch m.focus {
	case queryIndexSelection:
	// nothing to do
	case queryHashKeyInput:
		m.content.hashKeyInput.Focus()
	case queryOperatorField:
		// nothing to do
	case queryRangeKeyInput1:
		m.content.rangeKeyInput1.Focus()
	case queryRangeKeyInput2:
		m.content.rangeKeyInput2.Focus()
	case queryOrderSelection:
		// nothing to do
	case queryApplyButton:
		// nothing to do
	}
	m.updateStyles(true)
	return nil
}

func (m *Queryialog) ResetState() {
	m.state.init.hashKeyValue = ""
	m.state.init.rangeKeyOperator = messages.Equals
	m.state.init.rangeKeyValue = nil
	m.state.init.rangeKeyValue2 = nil
	m.state.init.selectedIndex = ""
	m.state.init.orderDescending = false

	m.state.table.TableARN = ""
	m.state.table.TableIndex = messages.TableIndex{}
	m.state.table.GSI = nil
	m.state.table.LSI = nil
	m.state.resolved.HashKey = ""
	m.state.resolved.HashKeyType = ""
	m.state.resolved.RangeKey = nil
	m.state.resolved.RangeKeyType = ""
	m.focus = queryIndexSelection

	m.content.indexSelection.SetItems([]list.Item{})
	m.content.indexSelection.Select(0)
	m.content.operatorSelection.SetItems([]list.Item{})
	m.content.operatorSelection.Select(0)
	m.content.hashKeyInput.Reset()
	m.content.rangeKeyInput1.Reset()
	m.content.rangeKeyInput2.Reset()
}

func (m *Queryialog) SetState(msg messages.InitQueryParameters) tea.Cmd {
	m.ResetState()

	// init table state
	m.state.table.TableARN = msg.TableARN
	m.state.table.TableIndex = msg.TableIndex
	m.state.table.GSI = msg.GSI
	m.state.table.LSI = msg.LSI

	// init resolved state for updating contents later
	m.resolveIndexInfo()

	// init the initial state
	m.state.init.selectedIndex = u.IfNotNil(msg.CurrentIndex, tableIndexName)
	m.state.init.hashKeyValue = msg.HashKeyValue
	m.state.init.rangeKeyValue = msg.RangeKeyValue1
	m.state.init.rangeKeyValue2 = msg.RangeKeyValue2
	m.state.init.rangeKeyOperator = msg.RangeKeyOperator
	m.state.init.orderDescending = msg.RangeOrderDescending

	// update list item delegates
	m.updateStyles(true)

	// initialise the contents
	cmd := m.InitContent()

	return cmd
}

// InitContent relies on resolved & table state to initialise the contents
func (m *Queryialog) InitContent() tea.Cmd {
	var cmds []tea.Cmd

	{ // set indexes
		var idx int
		items := make([]list.Item, 0, 1+len(m.state.table.GSI)+len(m.state.table.LSI))
		items = append(items, indexItem{
			name:      tableIndexName,
			indexType: table,
		})
		for i, g := range m.state.table.GSI {
			items = append(items, indexItem{
				name:       g.Name,
				indexType:  gsi,
				sliceIndex: i,
			})
			if g.Name == m.state.init.selectedIndex {
				idx = len(items) - 1
			}
		}
		for i, l := range m.state.table.LSI {
			items = append(items, indexItem{
				name:       l.Name,
				indexType:  lsi,
				sliceIndex: i,
			})
			if l.Name == m.state.init.selectedIndex {
				idx = len(items) - 1
			}
		}
		cmds = append(cmds, m.content.indexSelection.SetItems(items))
		m.content.indexSelection.Select(idx)
	}

	{ // set query operators
		operators := make([]list.Item, 6, 7)
		operators[0] = queryListItem(messages.Equals)
		operators[1] = queryListItem(messages.Greater)
		operators[2] = queryListItem(messages.GreaterEqual)
		operators[3] = queryListItem(messages.Less)
		operators[4] = queryListItem(messages.LessEqual)
		operators[5] = queryListItem(messages.Between)
		if m.state.resolved.RangeKeyType != "N" {
			operators = append(operators, queryListItem(messages.BeginsWith))
		}
		var idx int
		for i, item := range operators {
			if messages.QueryOperator(item.(queryListItem)) == m.state.init.rangeKeyOperator {
				idx = i
				break
			}
		}
		m.content.operatorSelection.Select(idx)
		cmds = append(cmds, m.content.operatorSelection.SetItems(operators))
	}

	{ // set range order options
		orderOptions := []list.Item{
			queryListItem(rangeAscending),
			queryListItem(rangeDescending),
		}
		idx := u.Ternary(1, 0, m.state.init.orderDescending)
		m.content.rangeOrderSelection.Select(idx)
		cmds = append(cmds, m.content.rangeOrderSelection.SetItems(orderOptions))
	}

	{ // set input fields
		m.content.hashKeyInput.SetValue(m.state.init.hashKeyValue)
		m.content.rangeKeyInput1.SetValue(u.IfNotNil(m.state.init.rangeKeyValue, ""))
		m.content.rangeKeyInput2.SetValue(u.IfNotNil(m.state.init.rangeKeyValue2, ""))
	}

	m.updateSize()
	return tea.Batch(cmds...)
}

// updateContent updates the content based on the resolved state.
func (m *Queryialog) updateContent() tea.Cmd {
	operators := make([]list.Item, 6, 7)
	operators[0] = queryListItem(messages.Equals)
	operators[1] = queryListItem(messages.Greater)
	operators[2] = queryListItem(messages.GreaterEqual)
	operators[3] = queryListItem(messages.Less)
	operators[4] = queryListItem(messages.LessEqual)
	operators[5] = queryListItem(messages.Between)
	if m.state.resolved.RangeKeyType != "N" {
		operators = append(operators, queryListItem(messages.BeginsWith))
	}

	if m.content.operatorSelection.Index() > len(operators)-1 {
		m.content.operatorSelection.Select(0)
	}
	return m.content.operatorSelection.SetItems(operators)
}

func (m *Queryialog) applyParameters() tea.Cmd {
	indexSelection := m.content.indexSelection.SelectedItem().(indexItem).name
	hashKeySelection := m.content.hashKeyInput.Value()
	rangeKeySelection := m.content.rangeKeyInput1.Value()
	rangeKeySelection2 := m.content.rangeKeyInput2.Value()
	rangeKeyOpSelection := messages.QueryOperator(m.content.operatorSelection.SelectedItem().(queryListItem))
	orderDescending := string(m.content.rangeOrderSelection.SelectedItem().(queryListItem)) == string(rangeDescending)

	if true &&
		indexSelection == m.state.init.selectedIndex &&
		hashKeySelection == m.state.init.hashKeyValue &&
		rangeKeySelection == u.IfNotNil(m.state.init.rangeKeyValue, "") &&
		rangeKeySelection2 == u.IfNotNil(m.state.init.rangeKeyValue2, "") &&
		rangeKeyOpSelection == m.state.init.rangeKeyOperator &&
		orderDescending == m.state.init.orderDescending {
		return m.toggleDialog() // no changes
	}

	return tea.Batch(m.queryParametersUpdate(), m.toggleDialog())
}

func (m *Queryialog) queryParametersUpdate() tea.Cmd {
	// update the init state when committing changes
	m.state.init.selectedIndex = m.content.indexSelection.SelectedItem().(indexItem).name
	m.state.init.hashKeyValue = m.content.hashKeyInput.Value()
	rangeKeyV := m.content.rangeKeyInput1.Value()
	rangeKeyV2 := m.content.rangeKeyInput2.Value()
	m.state.init.rangeKeyValue = u.Ternary(&rangeKeyV, nil, rangeKeyV != "")
	m.state.init.rangeKeyValue2 = u.Ternary(&rangeKeyV2, nil, rangeKeyV2 != "")
	m.state.init.rangeKeyOperator = messages.QueryOperator(m.content.operatorSelection.SelectedItem().(queryListItem))
	m.state.init.orderDescending = string(m.content.rangeOrderSelection.SelectedItem().(queryListItem)) == string(rangeDescending)

	return func() tea.Msg {
		return messages.QueryParametersChanged{
			TableARN:             m.state.table.TableARN,
			IndexName:            u.Ternary(m.state.init.selectedIndex, "", m.state.init.selectedIndex != tableIndexName),
			HashKeyValue:         m.state.init.hashKeyValue,
			RangeKeyValue1:       m.state.init.rangeKeyValue,
			RangeKeyValue2:       m.state.init.rangeKeyValue2,
			RangeKeyOperator:     m.state.init.rangeKeyOperator,
			RangeOrderDescending: m.state.init.orderDescending,
		}
	}
}

func (m *Queryialog) toggleDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleQueryParameters{}
	}
}

// TODO: set max heights
func (m *Queryialog) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *Queryialog) updateSize() {
	items := m.content.indexSelection.Items()

	// set height of the index list
	padding := 4
	m.content.indexSelection.SetHeight(min(len(m.content.indexSelection.Items())+padding, m.window.height))

	// set height of the operator list
	padding = 3
	m.content.operatorSelection.SetHeight(min(len(m.content.operatorSelection.Items())+padding, m.window.height))

	// set height of range order options list
	padding = 3
	m.content.rangeOrderSelection.SetHeight(min(len(m.content.rangeOrderSelection.Items())+padding, m.window.height))

	// determine the width of the list within the dialog
	width := m.defaultDialogWidth
	for _, itm := range items {
		width = max(width, len(itm.(indexItem).name))
	}
	// set width of the list within the dialog
	m.content.indexSelection.SetWidth(width)

	// set dialog size
	m.dialog.height = m.content.indexSelection.Height() + 2
	m.dialog.width = width + 2

	m.updateStyles(true)

	// set height & width of dialog itself
	queryDialogStyle = queryDialogStyle.
		Height(m.dialog.height).
		Width(m.dialog.width)
}

func (m *Queryialog) View() string {
	title := m.styles.title.Render(m.content.indexSelection.Title)
	indexSelection := m.styles.content.Render(m.content.indexSelection.View())

	hashKeyInput := m.renderHashKey()

	help := m.styles.help.Render(
		m.styles.helpLine.Render(m.content.indexSelection.Help.View(m.content.indexSelection)),
	)

	apply := m.renderApplyButton()
	rendering := []string{
		title,
		indexSelection,
		hashKeyInput,
		apply,
		help,
	}

	// only render range-key parameters when range-key applies
	if m.state.resolved.RangeKey != nil {
		rangeKeyFields := m.renderJoinedRangeKeyFields()
		rendering = slices.Insert(rendering, 3, rangeKeyFields)
	}
	mainDialog := queryDialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			rendering...,
		),
	)

	mainLayer := lipgloss.NewLayer(mainDialog)
	c := lipgloss.NewCompositor(mainLayer)
	c.AddLayers(mainLayer)

	var subLayerContent string
	switch m.focus {
	case queryOperatorField:
		subLayerContent = m.renderOperatorSelection()
	case queryOrderSelection:
		subLayerContent = m.renderRangeOrderSelection()
	}
	if subLayerContent != "" {
		l := lipgloss.NewLayer(subLayerContent).
			X(mainLayer.GetX() + lipgloss.Width(mainDialog)).
			Y(mainLayer.GetY() + lipgloss.Height(mainDialog) - lipgloss.Height(subLayerContent))
		c.AddLayers(l)
	}

	return c.Render()
}

func (m *Queryialog) renderOperatorSelection() string {
	return queryOperatorDialogStyle.Render(m.styles.content.Render(m.content.operatorSelection.View()))
}

func (m *Queryialog) renderRangeOrderSelection() string {
	return queryOperatorDialogStyle.Render(m.styles.content.Render(m.content.rangeOrderSelection.View()))
}

func (m *Queryialog) renderHashKey() string {
	hashKeyInputStyle := u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.focus == queryHashKeyInput)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.hashKeyInputTitle.Render(fmt.Sprintf("Hash Key (%s): %s", m.state.resolved.HashKeyType, m.state.resolved.HashKey)),
		hashKeyInputStyle.Render(m.content.hashKeyInput.View()),
	)
}

func (m *Queryialog) renderJoinedRangeKeyFields() string {
	rangeKeyOperatorStyle := u.Ternary(m.styles.narrowBoxFocused, m.styles.narrowBox, m.focus == queryOperatorField)
	rangeKeyInputStyle1 := u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.focus == queryRangeKeyInput1)
	rangeKeyInputStyle2 := u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.focus == queryRangeKeyInput2)
	rangeOrderStyle := u.Ternary(m.styles.narrowBoxFocused, m.styles.narrowBox, m.focus == queryOrderSelection)
	op := m.content.operatorSelection.SelectedItem().(queryListItem)
	or := m.content.rangeOrderSelection.SelectedItem().(queryListItem)

	rendering := []string{
		m.styles.rangeKeyInputTitle.Render(fmt.Sprintf("Range Key (%s): %s", m.state.resolved.RangeKeyType, *m.state.resolved.RangeKey)),
		rangeKeyOperatorStyle.Render(string(op)),
		rangeKeyInputStyle1.Render(m.content.rangeKeyInput1.View()),
		m.styles.rangeKeyOrderTitle.Render("Range Order"),
		rangeOrderStyle.Render(string(or)),
	}
	if messages.QueryOperator(m.content.operatorSelection.SelectedItem().(queryListItem)) == messages.Between {
		rendering = slices.Insert(rendering, 3, rangeKeyInputStyle2.Render(m.content.rangeKeyInput2.View()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendering...)
}

func (m *Queryialog) renderApplyButton() string {
	applyButtonStyle := u.Ternary(m.styles.applyButtonFocused, m.styles.applyButton, m.focus == queryApplyButton)

	return applyButtonStyle.Render("Query!")
}

func (m *Queryialog) resolveIndexInfo() {
	sel, ok := m.content.indexSelection.SelectedItem().(indexItem)
	if !ok { // on empty list
		return
	}

	switch sel.indexType {
	case table:
		i := m.state.table.TableIndex
		m.state.resolved.HashKey = i.HashKey
		m.state.resolved.HashKeyType = i.HashKeyType
		m.state.resolved.RangeKey = i.RangeKey
		m.state.resolved.RangeKeyType = i.RangeKeyType
	case gsi:
		i := m.state.table.GSI[sel.sliceIndex]
		m.state.resolved.HashKey = i.HashKey
		m.state.resolved.HashKeyType = i.HashKeyType
		m.state.resolved.RangeKey = i.RangeKey
		m.state.resolved.RangeKeyType = i.RangeKeyType
	case lsi:
		i := m.state.table.LSI[sel.sliceIndex]
		m.state.resolved.HashKey = i.HashKey
		m.state.resolved.HashKeyType = i.HashKeyType
		m.state.resolved.RangeKey = &i.RangeKey
		m.state.resolved.RangeKeyType = i.RangeKeyType
	}
}
