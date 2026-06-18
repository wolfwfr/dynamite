package itemstable

import (
	"slices"
	"strconv"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// sortRowsAndUpdate applies sorting to the provided rows and updates the
// table's state with the new order of references to the backing items.
func (t *ItemsTable) sortRowsAndUpdate(cols []table.Column, rows []table.Row) []table.Row {
	res, sortedItems := t.sortRowsV2(cols, rows)

	sortopts := t.viewOptions.GetColumnSortingOptions()
	sortopts.SortedItems = sortedItems

	t.viewOptions, _ = t.viewOptions.Set().ColumnSorting().SetAll(sortopts).Do()

	return res
}

// sortRows compiles various pieces of information from table items and table
// state to apply sorting when sorting is enabled. It dynamically resolves the
// literal type of the column to sort on, distinguishing between integers,
// floats, and strings. It returns the sorted rows and a slice of indices, relating
// TODO: remove following test pass of v2
func (t *ItemsTable) sortRows(cols []table.Column, rows []table.Row) ([]table.Row, []int) {
	var (
		sorting       = t.viewOptions.GetColumnSortingOptions()
		sortEnabled   = sorting.Enabled
		sortingOn     = sorting.SortingOn
		sortAscending = sorting.Ascending
	)

	allowed := t.viewOptions.Check()

	if !sortEnabled || sortingOn == "" || len(rows) == 0 || !allowed.ColumnSortingAllowed {
		return rows, sorting.SortedItems
	}

	colsS := make([]string, len(cols))
	for i, c := range cols {
		colsS[i] = c.Title
	}
	idx := u.Find(colsS, sortingOn)
	if idx < 0 {
		return rows, sorting.SortedItems
	}

	// determine field-type
	var field string
	for _, r := range rows {
		if r.Fields[idx].Value() != "" {
			field = r.Fields[idx].Value()
			break
		}
	}
	_, errInt := strconv.ParseInt(field, 10, 64)
	_, errFloat := strconv.ParseFloat(field, 64)

	// choose the appropriate sorting function
	var sortFunc func(a, b sortingRow) int
	switch {
	// NOTE: assumes that float fields always contain decimal point
	case errFloat == nil:
		sortFunc = func(a, b sortingRow) int {
			aI, _ := strconv.ParseFloat(a.r.Fields[idx].Value(), 64)
			bI, _ := strconv.ParseFloat(b.r.Fields[idx].Value(), 64)
			check := u.Ternary(aI < bI, aI > bI, sortAscending)
			return u.Ternary(-1, 1, check)
		}
	case errInt == nil:
		sortFunc = func(a, b sortingRow) int {
			aI, _ := strconv.ParseInt(a.r.Fields[idx].Value(), 10, 64)
			bI, _ := strconv.ParseInt(b.r.Fields[idx].Value(), 10, 64)
			check := u.Ternary(aI < bI, aI > bI, sortAscending)
			return u.Ternary(-1, 1, check)
		}
	default:
		sortFunc = func(a, b sortingRow) int {
			s := []string{a.r.Fields[idx].Value(), b.r.Fields[idx].Value()}
			slices.Sort(s)
			check := u.Ternary(s[0] == a.r.Fields[idx].Value(), s[1] == a.r.Fields[idx].Value(), sortAscending)
			return u.Ternary(-1, 1, check)
		}
	}

	// apply sorting function on slice backed by new array
	sorted := make([]sortingRow, len(rows))
	for i, r := range rows {
		sorted[i] = sortingRow{
			r: r,
			i: r.Metadata[ItemIndexMetaKey].(int),
		}
	}

	// sort
	slices.SortFunc(sorted, sortFunc)

	// reset sorted-item-mapping
	sortedItems := make([]int, len(sorted))

	res := make([]table.Row, len(sorted))
	for i := range sorted {
		sortedItems[i] = sorted[i].i
		res[i] = sorted[i].r
	}

	return res, sortedItems

}

// sortRows compiles various pieces of information from table items and table
// state to apply sorting when sorting is enabled. It dynamically resolves the
// literal type of the column to sort on, distinguishing between integers,
// floats, and strings. It returns the sorted rows and a slice of indices, relating
func (t *ItemsTable) sortRowsV2(cols []table.Column, rows []table.Row) ([]table.Row, []int) {
	var (
		sorting       = t.viewOptions.GetColumnSortingOptions()
		sortEnabled   = sorting.Enabled
		sortingOn     = sorting.SortingOn
		sortAscending = sorting.Ascending
	)

	allowed := t.viewOptions.Check()

	if !sortEnabled || sortingOn == "" || len(rows) == 0 || !allowed.ColumnSortingAllowed {
		return rows, sorting.SortedItems
	}

	colsS := make([]string, len(cols))
	for i, c := range cols {
		colsS[i] = c.Title
	}
	idx := u.Find(colsS, sortingOn)
	if idx < 0 {
		return rows, sorting.SortedItems
	}

	// determine field-type
	var field string
	for _, r := range rows {
		if r.Fields[idx].Value() != "" {
			field = r.Fields[idx].Value()
			break
		}
	}
	_, errInt := strconv.ParseInt(field, 10, 64)
	_, errFloat := strconv.ParseFloat(field, 64)

	// choose the appropriate sorting function
	var sortFunc func(a, b table.Row) int
	switch {
	// NOTE: assumes that float fields always contain decimal point
	case errFloat == nil:
		sortFunc = func(a, b table.Row) int {
			aI, _ := strconv.ParseFloat(a.Fields[idx].Value(), 64)
			bI, _ := strconv.ParseFloat(b.Fields[idx].Value(), 64)
			check := u.Ternary(aI < bI, aI > bI, sortAscending)
			return u.Ternary(-1, 1, check)
		}
	case errInt == nil:
		sortFunc = func(a, b table.Row) int {
			aI, _ := strconv.ParseInt(a.Fields[idx].Value(), 10, 64)
			bI, _ := strconv.ParseInt(b.Fields[idx].Value(), 10, 64)
			check := u.Ternary(aI < bI, aI > bI, sortAscending)
			return u.Ternary(-1, 1, check)
		}
	default:
		sortFunc = func(a, b table.Row) int {
			s := []string{a.Fields[idx].Value(), b.Fields[idx].Value()}
			slices.Sort(s)
			check := u.Ternary(s[0] == a.Fields[idx].Value(), s[1] == a.Fields[idx].Value(), sortAscending)
			return u.Ternary(-1, 1, check)
		}
	}

	// apply sorting function on slice backed by new array
	sorted := make([]table.Row, len(rows))
	copy(sorted, rows)
	slices.SortFunc(sorted, sortFunc)

	// assemble references to backing items
	sortedItems := make([]int, len(sorted))
	for i := range sorted {
		sortedItems[i] = sorted[i].Metadata[ItemIndexMetaKey].(int)
	}

	return sorted, sortedItems
}
