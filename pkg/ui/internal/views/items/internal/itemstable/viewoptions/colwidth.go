package viewoptions

// column-dyn-width collects settings related to column dynamic width
type ColumnDynWidth struct {
	Enabled  bool // TODO: is this necessary?
	DynWidth map[string]struct{}
}

// SetColumnDynamicWidth is part of the setter-builder
type SetColumnDynamicWidth struct {
	p *Setter // parent
}

// SetAll returns a doable-setter wrapping a function to update the
// column-visibility settings
func (s *SetColumnDynamicWidth) SetAll(c ColumnDynWidth) *DoableSetter {
	d := &DoableSetter{}
	d.p = s.p
	d.t = setWidth
	d.f = func() (ViewOptions, bool) {
		v := s.p.v
		v.columnDynWidth = c
		return v, true
	}
	return d
}

// GetColumnDynWidthOptions returns the current state of the column-visibility
// options
func (v *ViewOptions) GetColumnDynWidthOptions() ColumnDynWidth {
	return v.columnDynWidth
}

// resetColumnDynWidthState resets state relating to column-visibility functionality
func (v *ViewOptions) ResetColumnDynWidthState() {
	v.columnDynWidth.Enabled = false
	v.columnDynWidth.DynWidth = make(map[string]struct{}, 0)
}
