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

	"github.com/wolfwfr/dynamite/pkg/util"
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

// HeaderDelegate accepts an item from the list and returns a header strings if
// the item requires one. If not, the function is expected to return an empty
// string. When `ItemDelegate.Collapse` equals `false`, this header is rendered above
// the item. If `ItemDelegate.Collapse` equals `true`, this header is rendered
// inline, in brackets following the item Name.
type HeaderDelegate func(Item, int) string

type ItemDelegate struct {
	Styles *Styles

	HeadedItems []HeaderDelegate

	// when collapse is false, headers will be rendered above the items. When it
	// is `true` the collapsed header will instead be rendered inline in
	// brackets following the item name.
	Collapse bool
}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	var (
		header string
		str    = i.Name
	)

	for _, f := range d.HeadedItems {
		if h := f(i, index); h != "" && !d.Collapse {
			header = d.Styles.Header.Render(h) + "\n"
			break
		} else if h != "" && d.Collapse {
			header = d.Styles.Header.Render(fmt.Sprintf("(%s)", h))
			break
		}
	}

	fn := d.Styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.Styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	if d.Collapse {
		fmt.Fprintf(w, "%s%s%s", fn(str), util.Ternary(" ", "", len(header) > 0), header)
		return
	}
	fmt.Fprintf(w, "%s%s", header, fn(str))
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
