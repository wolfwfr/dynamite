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
}

type regionListStyles struct {
	headed.Styles
	title    lipgloss.Style
	content  lipgloss.Style
	help     lipgloss.Style
	helpLine lipgloss.Style
}

func newRegionStyles(darkBG bool) regionListStyles {
	var s regionListStyles

	s.Item = lipgloss.NewStyle().PaddingLeft(4)
	s.SelectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(commonstyles.DialogFocusColour)
	s.Header = lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0"))

	s.title = lipgloss.NewStyle().Padding(1, 0, 2, 0)
	s.content = lipgloss.NewStyle().PaddingTop(1).PaddingBottom(2)
	s.help = list.DefaultStyles(darkBG).HelpStyle.Padding(1, 2, 0, 2)
	s.helpLine = lipgloss.NewStyle().PaddingBottom(1)
	return s
}

func NewRegionsDialog(available, starred []string, current string, close key.Binding) *Regions {
	r := &Regions{
		available: available,
		starred:   starred,
		selected:  current,

		keyMap: regionsKeyMap{
			close: close,
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
		},

		defaultDialogHeight: 46,
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
	r.updateStyles(true) // default to dark styles.
	r.updateSize()

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
		Styles: &s.Styles,
	}

	headerFmt := func(s string) string {
		return fmt.Sprintf("\n%s\n%s", headed.HeaderPadding(s, 16), "_________________\n")
	}

	if len(m.starred) > 0 {
		firstStarred := m.starred[0]
		d.HeadedItems = append(d.HeadedItems, func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(i.Name), "", i.Name == firstStarred)
		})
	}

	if len(m.unstarred) > 0 {
		firstNormal := m.unstarred[0]
		d.HeadedItems = append(d.HeadedItems, func(i headed.Item, _ int) string {
			return u.Ternary(headerFmt(i.Name), "", i.Name == firstNormal)
		})
	}

	return d
}

func (m *Regions) updateStyles(isDark bool) {
	s := newRegionStyles(isDark)
	m.content.Styles.Title = s.title
	m.content.Styles.HelpStyle = s.help

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
}

func (m *Regions) updateSize() {
	items := m.content.Items()

	// set height of the list within the dialog
	padding := 4
	m.content.SetHeight(min(len(m.content.Items())+padding, m.window.height))

	// determine the width of the list within the dialog
	width := m.defaultDialogWidth
	for _, itm := range items {
		width = max(width, len(itm.(headed.Item).Name))
	}
	// set width of the list within the dialog
	m.content.SetWidth(width)

	// set height & width of dialog itself
	regionsDialogStyle = regionsDialogStyle.
		Height(m.content.Height() + 2).
		Width(width + 2)

}

func (m *Regions) View() string {
	title := m.styles.title.Render(m.content.Title)
	content := m.styles.content.Render(m.content.View())
	help := m.styles.help.Render(
		m.styles.helpLine.Render(m.content.Help.View(m.content)),
	)
	return regionsDialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			content,
			help,
		),
	)
}
