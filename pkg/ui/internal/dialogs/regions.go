package dialogs

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

var regionsDialogStyle = commonstyles.DialogStyle

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
	selected  string

	keyMap regionsKeyMap

	defaultDialogHeight int
	defaultDialogWidth  int

	width  int
	height int

	help help.Model

	content list.Model
}

type regionListStyles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newStyles(darkBG bool) regionListStyles {
	var s regionListStyles
	s.title = lipgloss.NewStyle().MarginLeft(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct {
	styles *regionListStyles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewRegionsDialog(available, starred []string, current string) *Regions {
	r := &Regions{
		available: available,
		starred:   starred,
		selected:  current,

		keyMap: regionsKeyMap{
			close: key.NewBinding(
				key.WithKeys("?", "esc", "q"),
				key.WithHelp("?/esc/q", "close"),
			),
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		},

		defaultDialogHeight: 50,
		defaultDialogWidth:  50,

		help: help.New(),
	}
	r.width = r.defaultDialogWidth
	r.height = r.defaultDialogHeight

	// TODO: separate section for starred regions
	items := make([]list.Item, len(available))
	for i, a := range available {
		items[i] = item(a)
	}

	l := list.New(items, itemDelegate{}, r.width, r.height)
	l.SetFilteringEnabled(false)

	r.content = l
	r.updateStyles(true) // default to dark styles.

	return r
}

func (m *Regions) updateStyles(isDark bool) {
	s := newStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.PaginationStyle = s.pagination
	m.content.Styles.HelpStyle = s.help
	m.content.SetDelegate(itemDelegate{styles: &s})
}

func (m *Regions) Init() tea.Cmd {
	return nil
}

func (m *Regions) Width() int {
	return m.width
}

func (m *Regions) Height() int {
	return m.height
}

func (m *Regions) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.toggleDialog()
		case key.Matches(msg, m.keyMap.enter):
			return m.selectRegion()
		default:
			var cmd tea.Cmd
			m.content, cmd = m.content.Update(msg)
			return cmd
		}
	case tea.WindowSizeMsg:
		m.applySize(msg.Height, msg.Width)
		return nil
	}
	return nil
}

func (m *Regions) selectRegion() tea.Cmd {
	itm := m.content.SelectedItem()
	selection := string(itm.(item))
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
	m.width = m.defaultDialogWidth
	m.height = m.defaultDialogHeight
	regionsDialogStyle = regionsDialogStyle.
		Height(m.height).
		Width(m.width)
}

func (m *Regions) View() string {
	title := "AWS Regions"
	content := m.content.View()
	help := m.help.ShortHelpView((m.keyMap.ShortHelp()))
	help = ""
	return regionsDialogStyle.Render(title + content + help)
	// regionsDialogStyle.Render(title + nl + fullhelp + nl + m.Help.ShortHelpView((m.keyMap.ShortHelp())))
}
