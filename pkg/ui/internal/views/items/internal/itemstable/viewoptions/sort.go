package viewoptions

// column sorting collects settings related to column sorting
type ColumnSorting struct {
	// TODO: remove: only used in 'get-item' and this can fall back to the
	// row-metadata
	SortedItems []int // indices referring to items
	SortingOn   string
	Ascending   bool // if false, descending
	Enabled     bool
}

// SetColumnSorting is part of the setter-builder
type SetColumnSorting struct {
	p *Setter // parent
}

// SetAll returns a doable-setter wrapping a function to update the
// column-sorting settings
func (s *SetColumnSorting) SetAll(c ColumnSorting) *DoableSetter {
	d := &DoableSetter{}
	d.p = s.p
	d.t = setSort
	d.f = func() (ViewOptions, bool) {
		v := s.p.v
		v.columnSorting = c
		return v, true
	}
	return d
}

// GetColumnSortingOptions returns the current state of the column-sorting options
func (v *ViewOptions) GetColumnSortingOptions() ColumnSorting {
	return v.columnSorting
}

// resetColumnSortingState resets internal state relating to column-sorting functionality
func (v *ViewOptions) ResetColumnSortingState() {
	v.columnSorting.SortedItems = make([]int, 0)
	v.columnSorting.Ascending = true
	v.columnSorting.SortingOn = ""
	v.columnSorting.Enabled = false
}
