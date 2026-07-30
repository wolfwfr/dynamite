package viewoptions

type Check struct {
	SearchAllowed           bool
	ColumnSortingAllowed    bool
	ColumnVisibilityAllowed bool
	ColumnTransformAllowed  bool
	ColumnDynWidthAllowed   bool
}

func (v *ViewOptions) Check() Check {
	c := Check{}
	c.SearchAllowed = !v.columnSorting.Enabled
	// TODO: allow compatibility
	c.ColumnSortingAllowed = !v.searchResults.Enabled
	c.ColumnVisibilityAllowed = true
	c.ColumnTransformAllowed = true
	c.ColumnDynWidthAllowed = true
	return c
}
