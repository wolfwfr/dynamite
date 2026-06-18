package viewoptions

// search-results collects settings related to item searching
type SearchResults struct {
	MatchedItems []int   // indices referring to items
	MatchedRunes [][]int //matches by index to SearchResults.matchedItems
	ColumnIndex  int
	Enabled      bool
}

// SetSearchResults is part of the setter-builder
type SetSearchResults struct {
	p *Setter // parent
}

// SetAll returns a doable-setter wrapping a function to update the
// item-search settings
func (s *SetSearchResults) SetAll(i SearchResults) *DoableSetter {
	d := &DoableSetter{}
	d.p = s.p
	d.t = setSearch
	d.f = func() (ViewOptions, bool) {
		v := s.p.v
		v.searchResults = i
		return v, true
	}
	return d
}

// GetSearchResultsOptions returns the current state of the item-search options
func (v *ViewOptions) GetSearchResultsOptions() SearchResults {
	return v.searchResults
}

// resetSearchState resets internal state relating to search functionality
func (v *ViewOptions) ResetSearchState() {
	v.searchResults.Enabled = false
	v.searchResults.MatchedItems = make([]int, 0)
	v.searchResults.MatchedRunes = make([][]int, 0)
}
