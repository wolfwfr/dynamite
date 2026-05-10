// package regularlist collects generic resources to be used and composed in a
// list that requires only basic functionality for its items
package regularlist

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
}

type ListItem struct {
	Value string
	Meta  map[string]any
}

func (i ListItem) FilterValue() string { return i.Value }

type ItemDelegate struct {
	Styles *Styles
}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ListItem)
	if !ok {
		return
	}

	str := i.Value

	fn := d.Styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.Styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
