package itemstable

import (
	"slices"
	"strconv"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// sortRows applies sorting to the provided rows.
func (t *ItemsTable) sortRows(cols []table.Column, rows []table.Row) []table.Row {
	if allowed := t.viewOptions.Check(); !allowed.ColumnSortingAllowed {
		return rows
	}

	sortopts := t.viewOptions.GetColumnSortingOptions()
	res := sortRows(cols, rows, sortopts)

	return res
}

// sortRows compiles various pieces of information from table items and table
// state to apply sorting when sorting is enabled. It dynamically resolves the
// literal type of the column to sort on, distinguishing between integers,
// floats, and strings. It returns the sorted rows and a slice of indices, relating
func sortRows(cols []table.Column, rows []table.Row, sortopts viewoptions.ColumnSorting) []table.Row {
	var (
		sortEnabled   = sortopts.Enabled
		sortingOn     = sortopts.SortingOn
		sortAscending = sortopts.Ascending
	)

	if !sortEnabled || sortingOn == "" || len(rows) == 0 {
		return rows
	}

	colsS := make([]string, len(cols))
	for i, c := range cols {
		colsS[i] = c.Title
	}
	idx := u.Find(colsS, sortingOn)
	if idx < 0 {
		return rows
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

	return sorted
}
