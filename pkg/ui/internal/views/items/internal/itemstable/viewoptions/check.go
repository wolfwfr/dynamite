package viewoptions

type Check struct {
	searchAllowed           bool
	ColumnSortingAllowed    bool
	ColumnVisibilityAllowed bool
}

func (v *ViewOptions) Check() Check {
	c := Check{}
	c.searchAllowed = !v.columnSorting.Enabled
	// TODO: allow compatibility
	c.ColumnSortingAllowed = !v.searchResults.Enabled
	c.ColumnVisibilityAllowed = true
	return c
}
