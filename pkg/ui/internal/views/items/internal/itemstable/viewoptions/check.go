package viewoptions

type Check struct {
	SearchAllowed           bool
	ColumnSortingAllowed    bool
	ColumnVisibilityAllowed bool
}

func (v *ViewOptions) Check() Check {
	c := Check{}
	c.SearchAllowed = !v.columnSorting.Enabled
	// TODO: allow compatibility
	c.ColumnSortingAllowed = !v.searchResults.Enabled
	c.ColumnVisibilityAllowed = true
	return c
}
