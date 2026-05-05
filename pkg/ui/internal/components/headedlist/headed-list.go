// package headedlist collects generic resources to be used and composed in a
// list that requires items to be headed
package headedlist

import (
	"fmt"
	"io"
	"math"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Item struct {
	Name string
	Meta map[string]any
}

func (i Item) FilterValue() string { return "" }

type Styles struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Header       lipgloss.Style
}

// HeaderDelegate accepts an item from the list and returns a header if the item
// requires one. If not, the function is expected to return an empty string
type HeaderDelegate func(Item, int) string

type ItemDelegate struct {
	Styles *Styles

	HeadedItems []HeaderDelegate
	// FirstStarred *string
	// FirstNormal  string
}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := i.Name

	var header string
	for _, f := range d.HeadedItems {
		if h := f(i, index); h != "" {
			header = d.Styles.Header.Render(h) + "\n"
			break
		}
	}

	fn := d.Styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.Styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fmt.Sprintf("%s%s", header, fn(str)))
}

// HeaderPadding is a convenience helper that returns a best effort centralised
// header with padding on each side
func HeaderPadding(h string, l int) string {
	ll := len(h)
	if ll >= l {
		return h
	}
	pd := int(math.Round(float64(l-ll) / 2))
	s := strings.Builder{}
	for range pd {
		fmt.Fprint(&s, " ")
	}
	fmt.Fprint(&s, h)
	for range l - ll - pd {
		fmt.Fprint(&s, " ")
	}
	return s.String()
}
