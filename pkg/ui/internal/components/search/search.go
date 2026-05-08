package search

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

var (
	SearchBoxStyle = lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(commonstyles.SearchFg).
		PaddingLeft(2).
		Height(2)
)

type SearchCallbacks struct {
	ToSearch       func(prefix string) []string
	EmptyInput     func() tea.Cmd
	Results        func(prefix string, items []FilteredItem) tea.Cmd
	Reset          func(searchHeight int) tea.Cmd
	SearchBoxOpens func(searchHeight int) tea.Cmd
}

type SearchBox struct {
	id string

	divider string

	height int
	width  int

	input textinput.Model

	enabled bool // enabled determines whether searchbox is visible
	focused bool // focused determines whether searchbox is actively receiving input and displays (blinking) cursor

	F FilterFunc

	Callbacks SearchCallbacks
}

func NewSearchBox(cb SearchCallbacks) *SearchBox {
	// default search-input style
	searchInput := textinput.New()
	searchInput.Prompt = "Search > "
	searchInput.CharLimit = 64
	searchInput.Placeholder = "type to search..."

	return &SearchBox{
		id: uuid.New().String(),

		height: 2,
		width:  64,
		input:  searchInput,

		F: DefaultFilter,

		Callbacks: cb,
	}
}

// IsFocused returns whether the search-box is focused
func (s *SearchBox) IsFocused() bool {
	return s.focused
}

// IsEnabled returns whether the search-box is enabled (visible)
func (s *SearchBox) IsEnabled() bool {
	return s.enabled
}

// GetHeight returns the current search box height
func (s *SearchBox) GetHeight() int {
	return s.height
}

// GetWidth returns the current search box width
func (s *SearchBox) GetWidth() int {
	return s.width
}

// SetWidth sets the height of the search-box
func (s *SearchBox) SetHeight(h int) {
	s.height = h
}

// SetWidth sets the width of the text input.
func (s *SearchBox) SetWidth(w int) {
	s.width = w
	s.input.SetWidth(w)
}

// SetPlaceHolder sets the placeholder
func (s *SearchBox) SetPlaceHolder(h string) {
	s.input.Placeholder = h
}

// SetDivider sets the divider dividing prefix and search.
//
// Search will only interpret what follows after the divider and what preceeds
// the divider is passed on to the caller when retrieving search items.
//
// E.g. (when divider equals `=`):
//
// `client_id=The best`
//
// will request search items for:
//
// `client_id`
//
// and search the items for:
//
// `The best`.
func (s *SearchBox) SetDivider(divider string) {
	s.divider = divider
}

func (s *SearchBox) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch txt := msg.String(); txt {
		case "esc", "enter":
			s.UnFocus()
			fallthrough
		default:
			newQuery, cmd := s.input.Update(msg)
			cmds = append(cmds, cmd)
			if newQuery.Value() != s.input.Value() { // if new query
				cmds = append(cmds, s.Search(newQuery.Value()))
			}
			s.input = newQuery
		}
	case FilterMatchesMsg:
		if msg.ID != s.id {
			return nil
		}
		if s.emptyInput() {
			cmds = append(cmds, s.Callbacks.EmptyInput())
			break
		}
		filtered := make([]string, len(msg.Items))
		for i, match := range msg.Items {
			filtered[i] = match.Item.Content
		}
		cmds = append(cmds, s.Callbacks.Results(msg.Prefix, msg.Items))
	default:
		s.input, cmd = s.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *SearchBox) emptyInput() bool {
	v := s.input.Value()
	if s.divider == "" && v == "" {
		return true
	} else if s.divider == "" {
		return false
	}
	idx := strings.Index(v, s.divider)
	return idx < 0 || idx == len(v)-1
}

// Search applies the current search query to all searchable inputs returns a
// cmd that the tea-framework can asynchronously process. Upon completion, it
// returns a tea.Msg containing the items remaining post-filter.
func (s *SearchBox) Search(q string) tea.Cmd {
	query := q
	var prefix string
	var rawItems []string

	// determine divider presence
	idx := strings.Index(q, s.divider)
	if s.divider != "" && (idx < 0 || idx == len(q)-1) {
		return s.Callbacks.EmptyInput()
	}

	// apply divider when appliccable
	if s.divider != "" {
		prefix = q[:idx]
		rawItems = s.Callbacks.ToSearch(prefix)
		query = q[idx+1:]
	} else {
		rawItems = s.Callbacks.ToSearch("")
	}

	f := s.F
	// OPTIM: cancel on next text input for performance
	return func() tea.Msg { // will execute async
		ranks := f(query, rawItems)
		filtered := make([]FilteredItem, len(ranks))
		for i, r := range ranks {
			item := FilteredItem{
				Index:   r.Index,
				Item:    Item{Content: rawItems[r.Index]},
				Matches: r.MatchedIndexes,
			}
			filtered[i] = item
		}
		return FilterMatchesMsg{
			Prefix: prefix,
			ID:     s.id,
			Items:  filtered,
		}
	}
}

// OpenSearchBox is the general entrypoint for enabling or re-enabling the
// search-box. If the search-box is already enabled but not focused, this
// function will only refocus the search-box. If the search-box was not yet
// enabled (i.e. hidden), the function will enable it and call the appropriate
// callback.
func (s *SearchBox) OpenSearchBox() tea.Cmd {
	cmds := []tea.Cmd{}
	s.focused = true
	cmds = append(cmds, s.input.Focus())
	if s.enabled {
		return tea.Batch(cmds...)
	}
	cmds = append(cmds, func() tea.Msg { return textinput.Blink })
	s.enabled = true
	cmds = append(cmds, s.Callbacks.SearchBoxOpens(s.height))
	return tea.Batch(cmds...)
}

// UnFocus removes focus from the search-box, it will no longer process text
// input and remove the cursor.
func (s *SearchBox) UnFocus() {
	s.focused = false
	s.input.Blur()
}

// Reset removes any text and completely disables the search-box. It will also
// call the appropriate callback if the search-box was not already disabled.
func (s *SearchBox) Reset() tea.Cmd {
	if !s.enabled {
		return nil
	}
	s.input.Reset()
	s.enabled = false
	s.focused = false

	return s.Callbacks.Reset(s.height)
}

func (s *SearchBox) View() string {
	if !s.enabled {
		return ""
	}
	return lipgloss.NewStyle().PaddingTop(1).Render(s.input.View())
}

func IsSearchBoxMessage(msg tea.Msg) bool {
	_, ok := msg.(FilterMatchesMsg)
	return ok
}
