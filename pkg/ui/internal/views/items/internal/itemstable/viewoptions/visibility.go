package viewoptions

// column visibility collects settings related to column visibillity
type ColumnVisibility struct {
	Enabled   bool
	InVisible map[string]struct{}
}

// SetColumnVisibility is part of the setter-builder
type SetColumnVisibility struct {
	p *Setter // parent
}

// SetAll returns a doable-setter wrapping a function to update the
// column-visibility settings
func (s *SetColumnVisibility) SetAll(c ColumnVisibility) *DoableSetter {
	d := &DoableSetter{}
	d.p = s.p
	d.t = setVis
	d.f = func() (ViewOptions, bool) {
		v := s.p.v
		v.columnVisibility = c
		return v, true
	}
	return d
}

// GetColumnVisibilityOptions returns the current state of the column-visibility
// options
func (v *ViewOptions) GetColumnVisibilityOptions() ColumnVisibility {
	return v.columnVisibility
}

// resetColumnVisibilityState resets state relating to column-visibility functionality
func (v *ViewOptions) ResetColumnVisibilityState() {
	v.columnVisibility.Enabled = false
	v.columnVisibility.InVisible = make(map[string]struct{}, 0)
}
