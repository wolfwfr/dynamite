package viewoptions

type settingType int

const (
	setNone settingType = iota
	setSearch
	setSort
	setVis
)

// Setter & DoableSetter provide a unified method for updating ViewOptions
// state, which ensures that any update is subject to a shared requirements
// check.
type Setter struct {
	v ViewOptions // copy, no in-place editing
}

type DoableSetter struct {
	p *Setter                    // parent
	t settingType                // type
	f func() (ViewOptions, bool) // func
}

// SearchResults returns the SearchResults Setter
func (s *Setter) SearchResults() *SetSearchResults {
	return &SetSearchResults{p: s}
}

// ColumnSorting returns the ColumnSorting Setter
func (s *Setter) ColumnSorting() *SetColumnSorting {
	return &SetColumnSorting{p: s}
}

// ColumnVisibility returns the ColumnVisibility Setter
func (s *Setter) ColumnVisibility() *SetColumnVisibility {
	return &SetColumnVisibility{p: s}
}

// Do guards against incompatible option combinations and if compatible executes
// the configured setter. It returns the new state of ViewOptions and a boolean
// indicating whether the changes were accepted.
func (s *DoableSetter) Do() (ViewOptions, bool) {
	c := s.p.v.Check()

	// guard against incompatible option combinations
	if s.t == setNone ||
		s.t == setSearch && !c.SearchAllowed ||
		s.t == setSort && !c.ColumnSortingAllowed ||
		s.t == setVis && !c.ColumnVisibilityAllowed {
		return s.p.v, false
	}

	return s.f()
}
