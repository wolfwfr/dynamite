package dialogs

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	headed "github.com/wolfwfr/dynamite/pkg/ui/internal/components/headed_list"
	regular "github.com/wolfwfr/dynamite/pkg/ui/internal/components/regular_list"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// TODO: add up & down arrow keys for navigating rows
type filterKeyMap struct {
	right key.Binding
	up    key.Binding
	down  key.Binding
	left  key.Binding
	enter key.Binding
	exec  key.Binding
	close key.Binding
	reset key.Binding
}

type filterDialogFocus int

const (
	filterAttrNameInput filterDialogFocus = iota
	filterAttrTypeField
	filterOperatorField
	filterAttrValueInput1
	filterAttrValueInput2
	removeButton
	addButton
	applyButton
)

type filterContent struct {
	operatorSelection list.Model
	attrTypeSelection list.Model
	attrNameInput     textinput.Model
	attrValueInput1   textinput.Model
	attrValueInput2   textinput.Model
}

type FilterStateInit struct {
	AttrName       string
	AttrType       types.ScalarAttributeType
	AttrValue1     *string
	AttrValue2     *string
	FilterOperator messages.FilterOperator
}

// the FilterDialog dialog enables the user to specify a filter for the scan or
// query
type FilterDialog struct {
	focus      filterDialogFocus
	contentIdx int

	keyMap filterKeyMap

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
		init []FilterStateInit
		// table state is set excusively on initialisation
		table struct {
			TableARN string
		}
	}

	styles filterListStyles

	help help.Model

	content []filterContent
}

type filterListStyles struct {
	headed.Styles
	dialog         lipgloss.Style
	title          lipgloss.Style
	content        lipgloss.Style
	contentLine    lipgloss.Style
	help           lipgloss.Style
	helpLine       lipgloss.Style
	attrTypeDialog lipgloss.Style
	operatordialog lipgloss.Style

	// box at width of content
	narrowBox        lipgloss.Style
	narrowBoxFocused lipgloss.Style

	// box at full width of dialog
	wideBox        lipgloss.Style
	wideBoxFocused lipgloss.Style

	// ignored fields
	ignored lipgloss.Style

	// titles
	AttrNameInputTitle  lipgloss.Style
	AttrValueInputTitle lipgloss.Style
	AttrTypeTitle       lipgloss.Style
	OperatorTitle       lipgloss.Style

	// remove filter button
	removeButton        lipgloss.Style
	removeButtonFocused lipgloss.Style

	// add filter button
	addButton        lipgloss.Style
	addButtonFocused lipgloss.Style

	// apply filter button
	applyButton        lipgloss.Style
	applyButtonFocused lipgloss.Style
}

func newFilterStyles(darkBG bool) filterListStyles {
	var s filterListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.Header = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour)

	s.dialog = commonstyles.DialogStyle
	s.operatordialog = commonstyles.DialogStyle.Border(lipgloss.RoundedBorder()).Padding(3, 3, 0, 0)
	s.attrTypeDialog = commonstyles.DialogStyle.Border(lipgloss.RoundedBorder()).Padding(3, 3, 0, 0)
	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.contentLine = lipgloss.NewStyle().PaddingTop(1)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().Padding(7, 0, 1, 0)

	// narrow boxes
	s.narrowBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogUnfocusColour).Padding(0, 1, 0, 1)
	s.narrowBoxFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogFocusColour).Padding(0, 1, 0, 1)

	// wide boxes
	s.wideBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogUnfocusColour)
	s.wideBoxFocused = s.wideBox.BorderForeground(commonstyles.DialogFocusColour)

	// ignored fields
	s.ignored = lipgloss.NewStyle().Foreground(commonstyles.DialogUnfocusColour).Padding(1, 1, 0, 1)

	// inputs fields
	s.AttrNameInputTitle = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour).Padding(0, 0, 0, 1)
	s.AttrValueInputTitle = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour).Padding(0, 0, 0, 1)
	s.AttrTypeTitle = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour).Padding(0, 0, 0, 1)
	s.OperatorTitle = lipgloss.NewStyle().Foreground(commonstyles.SubtleColour).Padding(0, 0, 0, 1)

	// remove button
	s.removeButton = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogUnfocusColour).Padding(0, 2, 0, 2).Margin(0, 0, 1, 0)
	s.removeButtonFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogFocusColour).Padding(0, 2, 0, 2).Margin(0, 0, 1, 0)

	// add button
	s.addButton = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogUnfocusColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)
	s.addButtonFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogFocusColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)

	// apply button
	s.applyButton = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(commonstyles.DialogUnfocusColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)
	s.applyButtonFocused = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(commonstyles.DialogFocusColour).Padding(0, 2, 0, 2).Margin(1, 0, 1, 0)

	return s
}

func NewFilterDialog(close key.Binding) *FilterDialog {
	d := &FilterDialog{
		keyMap: filterKeyMap{
			close: close,
			up: key.NewBinding(
				key.WithKeys("up"),
				key.WithHelp("↑", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("down"),
				key.WithHelp("↓", "down"),
			),
			enter: key.NewBinding(
				key.WithKeys("space", "enter"),
				key.WithHelp("space/enter", "select"),
			),
			exec: key.NewBinding(
				key.WithKeys("alt+enter"),
				key.WithHelp("alt+enter", "apply!"),
			),
			right: key.NewBinding(
				key.WithKeys("tab", "right"),
				key.WithHelp("tab/→", "right"),
			),
			left: key.NewBinding(
				key.WithKeys("shift+tab", "left"),
				key.WithHelp("shift+tab/←", "left"),
			),
			reset: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl+r", "reset"),
			),
		},

		defaultDialogHeight: 30,
		defaultDialogWidth:  156,

		help: help.New(),
	}

	d.focus = -1

	d.styles = newFilterStyles(true)

	d.dialog.width = d.defaultDialogWidth
	d.dialog.height = d.defaultDialogHeight

	d.window.width = 156
	d.window.height = 100

	d.content = make([]filterContent, 1)

	d.InitContent()

	d.updateSize()
	d.updateStyles(true) // default to dark styles.

	return d
}

func (m *FilterDialog) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		m.keyMap.right,
		m.keyMap.left,
		m.keyMap.up,
		m.keyMap.down,
		m.keyMap.enter,
		m.keyMap.exec,
		m.keyMap.reset,
	}
	listHelp := func(l list.Model) []key.Binding {
		return append(bindings, []key.Binding{
			l.KeyMap.CursorUp,
			l.KeyMap.CursorDown,
			l.KeyMap.NextPage,
			l.KeyMap.PrevPage,
			l.KeyMap.GoToStart,
			l.KeyMap.GoToEnd,
			m.keyMap.close,
		}...)
	}
	inputHelp := func(i textinput.Model) []key.Binding {
		return append(bindings, []key.Binding{
			m.keyMap.close,
		}...)
	}

	switch m.focus {
	case filterAttrNameInput:
		return inputHelp(m.content[m.contentIdx].attrNameInput)
	case filterAttrTypeField:
		return listHelp(m.content[m.contentIdx].attrTypeSelection)
	case filterOperatorField:
		return listHelp(m.content[m.contentIdx].operatorSelection)
	case filterAttrValueInput1:
		return inputHelp(m.content[m.contentIdx].attrValueInput1)
	case filterAttrValueInput2:
		return inputHelp(m.content[m.contentIdx].attrValueInput2)
	case applyButton:
		return bindings
	}
	return bindings
}

func (m *FilterDialog) newFilterItemDelegate(s *filterListStyles) regular.ItemDelegate {
	return regular.ItemDelegate{
		Styles: &regular.Styles{
			Item:         m.styles.Styles.Item,         // use same styling
			SelectedItem: m.styles.Styles.SelectedItem, // use same styling
		},
	}
}

func (m *FilterDialog) updateStyles(isDark bool) {
	s := newFilterStyles(isDark)

	subwidth := m.dialog.width/2 - 28

	s.wideBox = s.wideBox.Width(subwidth)
	s.wideBoxFocused = s.wideBoxFocused.Width(subwidth)

	s.AttrNameInputTitle = s.AttrNameInputTitle.Width(subwidth)
	s.AttrValueInputTitle = s.AttrValueInputTitle.Width(subwidth)
	s.AttrTypeTitle = s.AttrTypeTitle.Width(10)
	s.OperatorTitle = s.OperatorTitle.Width(subwidth)

	// dialog-style is actively resized; retain
	s.dialog = m.styles.dialog

	m.styles = s

	for i := range m.content {
		m.content[i].attrTypeSelection.SetDelegate(m.newFilterItemDelegate(&s))
		m.content[i].operatorSelection.SetDelegate(m.newFilterItemDelegate(&s))
		m.content[i].attrNameInput.SetWidth(subwidth - 2 - len(m.content[i].attrNameInput.Prompt) - 1)
		m.content[i].attrValueInput1.SetWidth(subwidth - 2 - len(m.content[i].attrValueInput1.Prompt) - 1)
		m.content[i].attrValueInput2.SetWidth(subwidth - 2 - len(m.content[i].attrValueInput2.Prompt) - 1)
	}
}

func (m *FilterDialog) Init() tea.Cmd {
	return nil
}

func (m *FilterDialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.exec):
			return m.applyParameters()
		case key.Matches(msg, m.keyMap.close):
			if m.safeToClose(msg) {
				return m.toggleDialog()
			}
		case key.Matches(msg, m.keyMap.right):
			return m.MoveFocus(1, horizontal)
		case key.Matches(msg, m.keyMap.left):
			return m.MoveFocus(-1, horizontal)
		case key.Matches(msg, m.keyMap.up):
			return m.MoveFocus(-1, vertical)
		case key.Matches(msg, m.keyMap.down):
			return m.MoveFocus(1, vertical)
		case key.Matches(msg, m.keyMap.reset):
			m.ResetState()
			return nil
		default:
			return m.handleNavigation(msg)
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	case messages.InitFilterParameters:
		return m.SetState(msg)
	}
	return m.handleNavigation(msg)
}

// safeToClose returns false when the key-press is a single alphanumeric
// character & an input-field is focused. This prevents closing the dialog
// accidentally when typing a key mapped to 'close' into a text-box.
func (m *FilterDialog) safeToClose(msg tea.KeyPressMsg) bool {
	bts := []byte(msg.String())
	if (m.focus == filterAttrNameInput || m.focus == filterAttrValueInput1 || m.focus == filterAttrValueInput2) && alphanum.Match(bts) && singleChar.Match(bts) {
		return false
	}
	return true
}

func (m *FilterDialog) handleNavigation(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.focus {
	case filterAttrNameInput:
		m.content[m.contentIdx].attrNameInput, cmd = m.content[m.contentIdx].attrNameInput.Update(msg)
	case filterAttrTypeField:
		m.content[m.contentIdx].attrTypeSelection, cmd = m.content[m.contentIdx].attrTypeSelection.Update(msg)
	case filterOperatorField:
		m.content[m.contentIdx].operatorSelection, cmd = m.content[m.contentIdx].operatorSelection.Update(msg)
	case filterAttrValueInput1:
		m.content[m.contentIdx].attrValueInput1, cmd = m.content[m.contentIdx].attrValueInput1.Update(msg)
	case filterAttrValueInput2:
		m.content[m.contentIdx].attrValueInput2, cmd = m.content[m.contentIdx].attrValueInput2.Update(msg)
	case removeButton:
		if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keyMap.enter) {
			cmd = m.RemoveFilterLine(m.contentIdx)
		}
	case addButton:
		if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keyMap.enter) {
			cmd = m.AddFilterLine()
		}
	case applyButton:
		if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keyMap.enter) {
			cmd = m.applyParameters()
		}
	}

	if m.focus == filterOperatorField {
		m.updateInputPlaceholders(m.contentIdx)
	}

	return cmd
}

func (m *FilterDialog) updateInputPlaceholders(i int) {
	if i < 0 || i >= len(m.content) {
		return
	}
	op := m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value
	m.content[i].attrNameInput.Placeholder = m.mapOperatorInput1Name(op)
	m.content[i].attrValueInput1.Placeholder = m.mapOperatorInput2Name(op)
	m.content[i].attrValueInput2.Placeholder = m.mapOperatorInput2Name(op)
}

func (m *FilterDialog) initInputPlaceholders() {
	for i := range m.content {
		m.updateInputPlaceholders(i)
	}
}

func (m *FilterDialog) AddFilterLine() tea.Cmd {
	lines := make([]filterContent, len(m.content)+1)
	copy(lines, m.content)
	m.content = lines
	return m.initContentLine(len(m.content) - 1)
}

func (m *FilterDialog) RemoveFilterLine(i int) tea.Cmd {
	m.content = slices.Delete(m.content, i, i+1)
	m.focus = 0
	m.contentIdx = max(0, m.contentIdx-1)
	if len(m.content) == 0 {
		m.ResetState()
	}
	return nil
}

type direction int

const (
	horizontal direction = iota
	vertical
)

func (m *FilterDialog) MoveFocus(i int, dir direction) tea.Cmd {
	switch m.focus {
	case filterAttrNameInput:
		m.content[m.contentIdx].attrNameInput.Blur()
	case filterAttrTypeField:
		// nothing to do
	case filterOperatorField:
		// nothing to do
	case filterAttrValueInput1:
		m.content[m.contentIdx].attrValueInput1.Blur()
	case filterAttrValueInput2:
		m.content[m.contentIdx].attrValueInput2.Blur()
	case removeButton, addButton, applyButton:
		// nothing to do
	}

	if dir == vertical {
		m.recalculateFocusVertically(i)
	} else {
		m.recalculateFocusHorizontally(i)
	}

	// default to false
	m.keyMap.enter.SetEnabled(false)

	switch m.focus {
	case filterAttrNameInput:
		m.content[m.contentIdx].attrNameInput.Focus()
	case filterAttrTypeField:
		// nothing to do
	case filterOperatorField:
		// nothing to do
	case filterAttrValueInput1:
		m.content[m.contentIdx].attrValueInput1.Focus()
	case filterAttrValueInput2:
		m.content[m.contentIdx].attrValueInput2.Focus()
	case removeButton, addButton, applyButton:
		m.keyMap.enter.SetEnabled(true)
	}
	m.updateSize()
	m.updateStyles(true)
	return nil
}

// recalculateFocusHorizontally moves the focus variables in the horizontal
// direction, but does not execute side-effects based on the actual movement of
// the focus.
func (m *FilterDialog) recalculateFocusHorizontally(i int) {
	// TODO: refactor & simplify
	for i != 0 {
		// move from apply button
		if m.focus == applyButton {
			m.contentIdx = 0
			m.focus = 0
			i -= 1
			if i < 0 {
				m.focus = addButton
				m.contentIdx = len(m.content) - 1
				i += 2 // +1 for having moved focus by 1 & +1 to compensate for earlier -1
			}
			continue
		}

		// move from add button
		if m.focus == addButton {
			m.contentIdx = 0
			m.focus = applyButton
			i -= 1
			if i < 0 {
				m.focus = removeButton
				m.contentIdx = len(m.content) - 1
				i += 2 // +1 for having moved focus by 1 & +1 to compensate for earlier -1
			}
			continue
		}

		// move away from current line when receding from first element
		if m.focus == filterAttrNameInput && i < 0 {
			atFirstLine := m.contentIdx == 0
			m.contentIdx -= 1
			m.focus = u.Ternary(applyButton, removeButton, atFirstLine)
			i += 1
			continue
		}

		// move away from current line when progressing beyond last element
		if m.focus == removeButton && i > 0 {
			atLastLine := m.contentIdx == len(m.content)-1
			m.contentIdx += 1
			m.focus = u.Ternary(addButton, 0, atLastLine)
			i -= 1
			continue
		}

		diff := u.Ternary(+1, -1, i >= 0)
		next := m.focus + filterDialogFocus(diff)
		// NOTE: relies on the assumption that the first and last line element
		// will never be hidden
		for !m.hasContentLineField(next, m.contentIdx) {
			next += filterDialogFocus(diff)
		}
		m.focus = next
		i -= diff
	}
}

// recalculateFocusVertically moves the focus variables in the vertical
// direction, but does not execute side-effects based on the actual movement of
// the focus.
// TODO: refactor
func (m *FilterDialog) recalculateFocusVertically(i int) {
	incr := u.Ternary(+1, -1, i > 0)
	incrF := filterDialogFocus(incr)
	for i != 0 {
		switch m.focus {
		case addButton:
			m.focus += incrF
			m.contentIdx = u.Ternary(m.contentIdx, len(m.content)-1, i > 0)
		case applyButton:
			m.focus = u.Ternary(filterAttrNameInput, addButton, i > 0)
			m.contentIdx = u.Ternary(0, m.contentIdx, i > 0)
		default:
			if m.contentIdx < 0 || m.contentIdx >= len(m.content) {
				return // TODO: log; bug
			}
			if m.contentIdx == 0 && i < 0 {
				m.contentIdx = -1
				m.focus = applyButton
			} else if m.focus == filterAttrValueInput2 && i < 0 {
				m.focus -= 1
			} else if m.focus == filterAttrValueInput1 && m.hasContentLineField(filterAttrValueInput2, m.contentIdx) && i > 0 {
				m.focus += 1
			} else if m.contentIdx == len(m.content)-1 && i > 0 {
				m.contentIdx = -1
				m.focus = addButton
			} else {
				m.contentIdx += incr
				var d filterDialogFocus = 0
				var found = m.hasContentLineField(m.focus, m.contentIdx)
				for !found {
					d++
					c1 := u.Clamp(m.focus-d, filterAttrNameInput, removeButton)
					c2 := u.Clamp(m.focus+d, filterAttrNameInput, removeButton)
					if m.hasContentLineField(c1, m.contentIdx) {
						found = true
						m.focus = c1
					} else if m.hasContentLineField(c2, m.contentIdx) {
						found = true
						m.focus = c2
					}
				}
			}
		}
		i -= incr
	}
}

func (m *FilterDialog) hasContentLineField(f filterDialogFocus, i int) bool {
	if i < 0 || i >= len(m.content) {
		return false
	}
	op := m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value

	switch f {
	case filterAttrNameInput:
		return true
	case filterAttrTypeField:
		return op != string(messages.Exists_F) && op != string(messages.NotExists_F)
	case filterOperatorField:
		return true
	case filterAttrValueInput1:
		return op != string(messages.Exists_F) && op != string(messages.NotExists_F)
	case filterAttrValueInput2:
		return op == string(messages.Between_F)
	case removeButton:
		return true
	}
	return false
}

func (m *FilterDialog) ResetState() tea.Cmd {
	var cmds []tea.Cmd
	m.content = make([]filterContent, 1)
	for i := range m.content {
		cmds = append(cmds, m.initContentLine(i))
	}
	m.focus = 0
	m.contentIdx = 0
	return tea.Batch(cmds...)
}

func (m *FilterDialog) SetState(msg messages.InitFilterParameters) tea.Cmd {
	m.ResetState()

	// init table state
	m.state.table.TableARN = msg.TableARN

	// init the initial state
	m.state.init = make([]FilterStateInit, len(msg.State))
	for i := range msg.State {
		m.state.init[i].AttrName = msg.State[i].AttrPath
		m.state.init[i].AttrType = msg.State[i].AttrType
		m.state.init[i].AttrValue1 = msg.State[i].AttrValue1
		m.state.init[i].AttrValue2 = msg.State[i].AttrValue2
		m.state.init[i].FilterOperator = msg.State[i].FilterOperator
	}

	// update list item delegates
	m.updateSize()
	m.updateStyles(true)

	// initialise the contents
	cmd := m.InitContent()

	return cmd
}

// InitContent relies on resolved & table state to initialise the contents
func (m *FilterDialog) InitContent() tea.Cmd {
	var cmds []tea.Cmd

	m.content = make([]filterContent, max(1, len(m.state.init)))
	for i := range m.content {
		cmds = append(cmds, m.initContentLine(i))
	}

	{ // set filter attr-scalarTypes
		scalarTypes := m.scalarTypes()
		typeIdx := map[types.ScalarAttributeType]int{}
		for i, o := range scalarTypes {
			typeIdx[o.(regular.ListItem).Meta[scalarTypeMatchKey].(types.ScalarAttributeType)] = i
		}
		for i := range m.content {
			if len(m.state.init) > i {
				m.content[i].attrTypeSelection.Select(typeIdx[m.state.init[i].AttrType])
			}
		}
	}

	{ // set filter operators
		operators := m.operators()
		operatorIdx := map[string]int{}
		for i, o := range operators {
			operatorIdx[o.(regular.ListItem).Value] = i
		}
		for i := range m.content {
			if len(m.state.init) > i {
				m.content[i].operatorSelection.Select(operatorIdx[string(m.state.init[i].FilterOperator)])
			}
		}
	}

	{ // set input fields
		for i := range m.content {
			if len(m.state.init) > i {
				m.content[i].attrNameInput.SetValue(m.state.init[i].AttrName)
				m.content[i].attrValueInput1.SetValue(u.IfNotNil(m.state.init[i].AttrValue1, ""))
				m.content[i].attrValueInput2.SetValue(u.IfNotNil(m.state.init[i].AttrValue2, ""))
			}
		}
	}

	m.MoveFocus(0, horizontal) // ensures focused element can execute side-effects
	m.updateSize()
	m.initInputPlaceholders()
	return tea.Batch(cmds...)
}

const scalarTypeMatchKey string = "match"

func (m *FilterDialog) scalarTypes() []list.Item {
	return []list.Item{
		regular.ListItem{Value: "string", Meta: map[string]any{scalarTypeMatchKey: types.ScalarAttributeTypeS}},
		regular.ListItem{Value: "number", Meta: map[string]any{scalarTypeMatchKey: types.ScalarAttributeTypeN}},
		regular.ListItem{Value: "bool  ", Meta: map[string]any{scalarTypeMatchKey: types.ScalarAttributeTypeB}},
	}
}

func (m *FilterDialog) operators() []list.Item {
	return []list.Item{
		regular.ListItem{Value: string(messages.Equals_F)},
		regular.ListItem{Value: string(messages.NotEquals_F)},
		regular.ListItem{Value: string(messages.Greater_F)},
		regular.ListItem{Value: string(messages.GreaterEqual_F)},
		regular.ListItem{Value: string(messages.Less_F)},
		regular.ListItem{Value: string(messages.LessEqual_F)},
		regular.ListItem{Value: string(messages.Between_F)},
		regular.ListItem{Value: string(messages.Exists_F)},
		regular.ListItem{Value: string(messages.NotExists_F)},
		regular.ListItem{Value: string(messages.Contains_F)},
		regular.ListItem{Value: string(messages.NotContains_F)},
		regular.ListItem{Value: string(messages.BeginsWith_F)},
	}
}

func (m *FilterDialog) initContentLine(i int) tea.Cmd {
	cmds := make([]tea.Cmd, 2)
	{ // scalar type selection
		l := list.New([]list.Item{}, regular.ItemDelegate{}, m.dialog.width, m.dialog.height)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		m.content[i].attrTypeSelection = l
		cmds[0] = m.content[i].attrTypeSelection.SetItems(m.scalarTypes())
	}

	{ // operator selection
		l := list.New([]list.Item{}, regular.ItemDelegate{}, m.dialog.width, m.dialog.height)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowTitle(false)

		m.content[i].operatorSelection = l
		cmds[1] = m.content[i].operatorSelection.SetItems(m.operators())
	}

	{ // attribute name input
		attrNameInput := textinput.New()
		attrNameInput.SetWidth(30)
		m.content[i].attrNameInput = attrNameInput
	}
	{ // attribute value input 1
		attrValueInput1 := textinput.New()
		attrValueInput1.SetWidth(30)
		m.content[i].attrValueInput1 = attrValueInput1
	}
	{ // attribute value input 2
		attrValueInput2 := textinput.New()
		attrValueInput2.SetWidth(30)
		m.content[i].attrValueInput2 = attrValueInput2
	}

	m.updateInputPlaceholders(i)

	return tea.Batch(cmds...)
}

func (m *FilterDialog) applyParameters() tea.Cmd {
	if len(m.content) != len(m.state.init) {
		return tea.Batch(m.filterParametersUpdate(), m.toggleDialog())
	}
	for i := range m.content {
		if true &&
			m.content[i].attrNameInput.Value() != m.state.init[i].AttrName ||
			m.content[i].attrTypeSelection.SelectedItem().(regular.ListItem).Meta[scalarTypeMatchKey].(types.ScalarAttributeType) != m.state.init[i].AttrType ||
			m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value != string(m.state.init[i].FilterOperator) ||
			m.content[i].attrValueInput1.Value() != u.IfNotNil(m.state.init[i].AttrValue1, "") ||
			m.content[i].attrValueInput2.Value() != u.IfNotNil(m.state.init[i].AttrValue2, "") {
			return tea.Batch(m.filterParametersUpdate(), m.toggleDialog())
		}
	}

	return m.toggleDialog() // no changes
}

func (m *FilterDialog) filterParametersUpdate() tea.Cmd {
	state := make([]messages.FilterState, len(m.content))
	for i := range m.content {
		state[i].AttrPath = m.content[i].attrNameInput.Value()
		state[i].AttrType = m.content[i].attrTypeSelection.SelectedItem().(regular.ListItem).Meta[scalarTypeMatchKey].(types.ScalarAttributeType)
		attrValue1 := m.content[i].attrValueInput1.Value()
		attrValue2 := m.content[i].attrValueInput2.Value()
		state[i].AttrValue1 = u.Ternary(&attrValue1, nil, attrValue1 != "")
		state[i].AttrValue2 = u.Ternary(&attrValue2, nil, attrValue2 != "")
		state[i].FilterOperator = messages.FilterOperator(m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value)
	}

	return func() tea.Msg {
		return messages.FilterParametersChanged{
			TableARN: m.state.table.TableARN,
			State:    state,
		}
	}
}

func (m *FilterDialog) toggleDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleFilterParameters{}
	}
}

// TODO: set max heights
func (m *FilterDialog) applySize(height, width int) {
	m.window.width = width
	m.window.height = height
	m.updateSize()
}

func (m *FilterDialog) updateSize() {
	// set height of the scalar-type list
	padding := 3
	for i := range m.content {
		m.content[i].attrTypeSelection.SetHeight(min(len(m.content[i].attrTypeSelection.Items())+padding, m.window.height))
	}
	// set height of the operator list
	for i := range m.content {
		m.content[i].operatorSelection.SetHeight(min(len(m.content[i].operatorSelection.Items())+padding, m.window.height))
	}

	// determine the halfwidth of the list within the dialog
	width := m.defaultDialogWidth

	// set dialog size
	m.dialog.height = m.defaultDialogHeight
	m.dialog.width = width + 2

	borderW := 2

	// set help size
	m.help.SetWidth(width - borderW)

	m.updateStyles(true)

	// set height & width of dialog itself
	m.styles.dialog = m.styles.dialog.
		Height(m.dialog.height).
		MaxHeight(m.window.height).
		Width(m.dialog.width)
}

// TODO: add visual indicator of field-type (S, N, B) specified by attrName
func (m *FilterDialog) View() string {
	help := m.renderHelp()

	addButton := m.renderAddButton()
	applyButton := m.renderApplyButton()

	mainDialog := m.styles.dialog.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Right,
				lipgloss.JoinVertical(lipgloss.Center,
					m.renderTitle(),
					lipgloss.JoinHorizontal(lipgloss.Top,
						lipgloss.NewStyle().PaddingRight(0).Render(m.renderContent()),
					),
				),
				addButton,
			),
			applyButton,
			help,
		),
	)

	mainLayer := lipgloss.NewLayer(mainDialog)
	c := lipgloss.NewCompositor(mainLayer)
	c.AddLayers(mainLayer)

	var subLayerContent string
	switch m.focus {
	case filterAttrTypeField:
		subLayerContent = m.renderScalarTypeSelection()
	case filterOperatorField:
		subLayerContent = m.renderOperatorSelection()
	}
	if subLayerContent != "" {
		l := lipgloss.NewLayer(subLayerContent).
			X(mainLayer.GetX() + lipgloss.Width(mainDialog) - lipgloss.Width(subLayerContent) - 4).
			Y(mainLayer.GetY() + lipgloss.Height(mainDialog) - lipgloss.Height(subLayerContent) - 4)
		c.AddLayers(l)
	}

	return c.Render()
}

// noop, unused required to satisfy help.Keymap interface
func (m *FilterDialog) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func (m *FilterDialog) renderTitle() string {
	return m.styles.title.Render("Filter Parameters")
}

func (m *FilterDialog) renderContent() string {
	longestOp := 0
	opPadding := 4 // 2x border & 2x padding of 1
	for i := range m.content {
		longestOp = max(longestOp, len(m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value))
	}

	lines := make([]string, len(m.content)+1)

	field1Titles := make(map[string]struct{})
	field2Titles := make(map[string]struct{})

	for i := range m.content {
		var (
			attrNameStyle   = m.selectContentLineFieldStyle(filterAttrNameInput, i)
			attrValue1Style = m.selectContentLineFieldStyle(filterAttrValueInput1, i)
			attrValue2Style = m.selectContentLineFieldStyle(filterAttrValueInput2, i)
			scalarTypeStyle = m.selectContentLineFieldStyle(filterAttrTypeField, i)
			opStyle         = m.selectContentLineFieldStyle(filterOperatorField, i)
		)

		opStyle = opStyle.Width(longestOp + opPadding)
		op := m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value

		field1Titles[m.mapOperatorInput1Name(op)] = struct{}{}
		field2Titles[m.mapOperatorInput2Name(op)] = struct{}{}

		rendering := make([]string, 4)
		rendering[0] = attrNameStyle.Render(m.content[i].attrNameInput.View())
		rendering[1] = scalarTypeStyle.Render(m.content[i].attrTypeSelection.SelectedItem().(regular.ListItem).Value)
		rendering[2] = opStyle.Render(op)
		rendering[3] = attrValue1Style.Render(m.content[i].attrValueInput1.View())

		l := i + 1
		lines[l] = lipgloss.JoinHorizontal(lipgloss.Top,
			rendering...,
		)
		if m.hasContentLineField(filterAttrValueInput2, i) {
			lines[l] = lipgloss.JoinVertical(lipgloss.Right,
				lines[l],
				attrValue2Style.Render(m.content[i].attrValueInput2.View()),
			)
		}
		lines[l] = lipgloss.JoinHorizontal(lipgloss.Top, lines[l], m.renderRemoveButton(i))
	}

	field1Title := make([]string, 0)
	field2Title := make([]string, 0)

	for n := range field1Titles {
		field1Title = append(field1Title, n)
	}
	for n := range field2Titles {
		field2Title = append(field2Title, n)
	}

	slices.Sort(field1Title)
	slices.Sort(field2Title)

	// title bar
	lines[0] = lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.AttrNameInputTitle.Render(strings.Join(field1Title, " / ")),
		m.styles.AttrTypeTitle.Render("Type"),
		m.styles.OperatorTitle.Width(longestOp+opPadding).Render(""),
		m.styles.AttrValueInputTitle.Render(strings.Join(field2Title, " / ")),
	)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *FilterDialog) selectContentLineFieldStyle(f filterDialogFocus, i int) lipgloss.Style {
	var (
		fieldLen int
		style    lipgloss.Style

		boxAddW  = 2 // border width
		nBoxPadW = m.styles.narrowBox.GetPaddingLeft() + m.styles.narrowBox.GetPaddingRight()
		wBoxPadW = m.styles.wideBox.GetPaddingLeft() + m.styles.wideBox.GetPaddingRight()
		hasField = m.hasContentLineField(f, i)
	)

	if i < 0 || i >= len(m.content) {
		return style
	}

	switch f {
	case filterAttrNameInput:
		style = u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.contentIdx == i && m.focus == f)
		fieldLen = lipgloss.Width(m.content[i].attrNameInput.View()) + wBoxPadW
	case filterAttrTypeField:
		style = u.Ternary(m.styles.narrowBoxFocused, m.styles.narrowBox, m.contentIdx == i && m.focus == f)
		fieldLen = len(m.content[i].attrTypeSelection.SelectedItem().(regular.ListItem).Value) + nBoxPadW
	case filterOperatorField:
		style = u.Ternary(m.styles.narrowBoxFocused, m.styles.narrowBox, m.contentIdx == i && m.focus == f)
		fieldLen = len(m.content[i].operatorSelection.SelectedItem().(regular.ListItem).Value) + nBoxPadW
	case filterAttrValueInput1:
		style = u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.contentIdx == i && m.focus == f)
		fieldLen = lipgloss.Width(m.content[i].attrValueInput1.View()) + wBoxPadW
	case filterAttrValueInput2:
		style = u.Ternary(m.styles.wideBoxFocused, m.styles.wideBox, m.contentIdx == i && m.focus == f)
		fieldLen = lipgloss.Width(m.content[i].attrValueInput2.View()) + wBoxPadW
	}

	if !hasField {
		return m.styles.ignored.Width(fieldLen + boxAddW).Transform(func(string) string { return u.RepeatString("─", fieldLen) })
	}
	return style
}

func (m *FilterDialog) mapOperatorInput1Name(op string) string {
	name := []string{
		string(messages.Equals_F),
		string(messages.NotEquals_F),
		string(messages.Greater_F),
		string(messages.GreaterEqual_F),
		string(messages.Less_F),
		string(messages.LessEqual_F),
		string(messages.Between_F),
	}
	path := []string{
		string(messages.Exists_F),
		string(messages.NotExists_F),
		string(messages.Contains_F),
		string(messages.NotContains_F),
		string(messages.BeginsWith_F),
	}
	if slices.Contains(name, op) {
		return "Attribute Name"
	}
	if slices.Contains(path, op) {
		return "Path"
	}
	return ""
}

func (m *FilterDialog) mapOperatorInput2Name(op string) string {
	if op == "" {
		return ""
	}
	if op == string(messages.BeginsWith_F) {
		return "Substring"
	}
	return "Attribute Value"
}

func (m *FilterDialog) renderHelp() string {
	return m.styles.help.Render(m.styles.helpLine.Render(m.help.View(m)))
}

func (m *FilterDialog) renderScalarTypeSelection() string {
	return m.styles.attrTypeDialog.Render(m.styles.content.Render(m.content[m.contentIdx].attrTypeSelection.View()))
}

func (m *FilterDialog) renderOperatorSelection() string {
	return m.styles.operatordialog.Render(m.styles.content.Render(m.content[m.contentIdx].operatorSelection.View()))
}

func (m *FilterDialog) renderRemoveButton(i int) string {
	removeButtonStyle := u.Ternary(m.styles.removeButtonFocused, m.styles.removeButton, m.contentIdx == i && m.focus == removeButton)

	return removeButtonStyle.Render("Remove")
}

func (m *FilterDialog) renderAddButton() string {
	addButtonStyle := u.Ternary(m.styles.addButtonFocused, m.styles.addButton, m.focus == addButton)

	return addButtonStyle.Render("Add new filter")
}

func (m *FilterDialog) renderApplyButton() string {
	applyButtonStyle := u.Ternary(m.styles.applyButtonFocused, m.styles.applyButton, m.focus == applyButton)

	return applyButtonStyle.Render("Apply filter(s)!")
}
