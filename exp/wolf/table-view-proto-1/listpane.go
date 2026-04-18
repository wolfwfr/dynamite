package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type listPane struct {
	list list.Model
}

type item string

func (i item) FilterValue() string { return "" }

// styles is a collection object for all the styles that apply or can apply to
// the list-pane and its contents.
type styles struct {
	item         lipgloss.Style
	selectedItem lipgloss.Style
}

type itemDelagate struct {
	styles *styles
}

// factory function for list-pane
func newListPane(items []string) *listPane {
	m := listPane{
		list: list.New(toListItems(items), itemDelagate{}, 10, 10),
	}
	m.applyStyles()
	return &m
}

func newStyles() styles {
	s := styles{}
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	return s
}

// applyStyles obtains a new styles object and applies the styles to the
// internal list bubble, as well as transferring the styles to the item-delagate.
func (m *listPane) applyStyles() {
	s := newStyles()
	// set list styles here when appliccable
	m.list.SetDelegate(itemDelagate{styles: &s})
}

func (d itemDelagate) Height() int                             { return 1 }
func (d itemDelagate) Spacing() int                            { return 0 }
func (d itemDelagate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelagate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	itm, ok := listItem.(item)
	if !ok {
		return
	}

	itemS := fmt.Sprintf("%s", itm)

	// set render wrappers for item rendering
	renderF := d.styles.item.Render
	if index == m.Index() {
		renderF = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}
	// can do formatting & styling
	fmt.Fprint(w, renderF(itemS))
}

func toListItems(items []string) []list.Item {
	res := make([]list.Item, len(items))
	for i := range items {
		res[i] = item(items[i])
	}
	return res
}

func (m *listPane) applySize(height, width int) {
	m.list.SetHeight(height)
	m.list.SetWidth(width)
}

func (m *listPane) Init() tea.Cmd {
	return nil
}

func (m *listPane) Update(msg tea.Msg) (*listPane, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *listPane) View() string {
	// s := "This is going to be a list pane"
	// return s
	return m.list.View()
}
