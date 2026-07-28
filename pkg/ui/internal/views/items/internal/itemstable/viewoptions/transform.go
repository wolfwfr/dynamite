package viewoptions

// column transform collects settings related to column transforms
type ColumnTransform struct {
	Enabled     bool
	Transformed map[string]struct{}
}

// SetColumnTransform is part of the setter-builder
type SetColumnTransform struct {
	p *Setter // parent
}

// SetAll returns a doable-setter wrapping a function to update the
// column-transform settings
func (s *SetColumnTransform) SetAll(c ColumnTransform) *DoableSetter {
	d := &DoableSetter{}
	d.p = s.p
	d.t = setSort
	d.f = func() (ViewOptions, bool) {
		v := s.p.v
		v.columnTransform = c
		return v, true
	}
	return d
}

// GetColumnTransformOptions returns the current state of the column-transform options
func (v *ViewOptions) GetColumnTransformOptions() ColumnTransform {
	return v.columnTransform
}

// resetColumnTransformState resets internal state relating to column-transform functionality
func (v *ViewOptions) ResetColumnTransformState() {
	v.columnTransform.Enabled = false
	v.columnTransform.Transformed = make(map[string]struct{}, 0)
}
