package viewoptions

type Check struct {
	SearchAllowed           bool
	ColumnSortingAllowed    bool
	ColumnVisibilityAllowed bool
	ColumnTransformAllowed  bool
}

func (v *ViewOptions) Check() Check {
	c := Check{}
	c.SearchAllowed = !v.columnSorting.Enabled
	// TODO: allow compatibility
	c.ColumnSortingAllowed = !v.searchResults.Enabled
	c.ColumnVisibilityAllowed = true
	c.ColumnTransformAllowed = true
	return c
}
