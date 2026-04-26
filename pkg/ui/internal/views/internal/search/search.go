package search

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	SearchBoxStyle = lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4F4F4F")).
		PaddingLeft(2).
		Height(2)
)

type SearchCallbacks struct {
	ToSearch     func() []string
	EmptyInput   func() tea.Cmd
	Results      func([]FilteredItem)
	Reset        func(searchHeight int)
	ViewBoxOpens func(searchHeight int)
}

type SearchBox struct {
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
		if s.input.Value() == "" {
			cmds = append(cmds, s.Callbacks.EmptyInput())
			break
		}
		filtered := make([]string, len(msg))
		for i, match := range msg {
			filtered[i] = match.Item.Content
		}
		s.Callbacks.Results(msg)
	default:
		s.input, cmd = s.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

// Search applies the current search query to all searchable inputs returns a
// cmd that the tea-framework can asynchronously process. Upon completion, it
// returns a tea.Msg containing the items remaining post-filter.
func (s *SearchBox) Search(query string) tea.Cmd {
	rawItems := s.Callbacks.ToSearch()
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
		return FilterMatchesMsg(filtered)
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
	s.Callbacks.ViewBoxOpens(s.height)
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
func (s *SearchBox) Reset() {
	if !s.enabled {
		return
	}
	s.input.Reset()
	s.enabled = false
	s.focused = false

	s.Callbacks.Reset(s.height)
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
