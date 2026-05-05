// package checkboxlist collects generic resources to be used and composed in a
// list that requires checkboxable items
package checkboxlist

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	u "github.com/wolfwfr/dynamite/pkg/util"
)

type Item struct {
	Checked bool
	Name    string
}

func (i Item) FilterValue() string { return "" }

type Styles struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
}

type ItemDelegate struct {
	Styles *Styles
}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s %s", u.Ternary("[x]", "[ ]", i.Checked), i.Name)

	fn := d.Styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.Styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
