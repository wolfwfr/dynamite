package itemstable

import (
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
)

// NOTE: each view-options-update-handler below (i.e. each set-handler) is
// individually responsible for correctly merging the change with the entire
// existing state of view-options. This includes applying view-options in the
// correct order, when relevant. This distributed responsibility could have been
// isolated in a single component, like a pipeline, at the cost of complexity
// and/or performance (a pipeline might completely rebuild the table state upon
// each update). Given the limited number of supported features, and limited
// interoperability between features, the pipeline oriented design is considered
// an effort of overengineering. But it should be reconsidered when the number
// of view-option combinations and their interoperability increase.

// SetColumnSorting updates the column-sorting state. Changes to column sorting
// affect the column suffix and the (virtual) rows being displayed. The function
// returns a boolean that indicates whether the mutation was accepted and successfully
// applied.
func (t *ItemsTable) SetColumnSorting(cols []string, sortingOn string, ascending bool) bool {
	// guard against mismatched states
	tablecols := t.table.Columns()
	if len(tablecols) != len(cols) {
		// TODO: better handling of new columns appearing in view
		t.ResetColumnSorting()
		return false
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().ColumnSorting().SetAll(viewoptions.ColumnSorting{
		SortingOn: sortingOn,
		Ascending: ascending,
		Enabled:   true,
	}).Do(); !ok {
		return false
	}

	// prepare table column update
	for i, c := range tablecols {
		c.Suffix = t.viewOptions.GetColumnSuffix(c.Title)
		tablecols[i] = c
	}

	t.updateTable(tablecols, t.table.Rows(), t.table.VirtualRows())
	return true
}

// SetColumnSorting updates the column-visibility state. Changes to column
// visibility only affect the columns and do not affect table rows. The function
// returns a boolean that indicates whether the mutation was accepted and
// successfully applied.
func (t *ItemsTable) SetColumnVisibility(cols []string, visible []bool) bool {
	// guard against mismatched states
	tablecols := t.table.Columns()
	if len(tablecols) != len(cols) {
		// TODO: better handling of new columns appearing in view
		t.ResetColumnVisibility()
		return false
	}

	// map visible → invisible
	invisible := make(map[string]struct{})
	for i, c := range cols {
		if !visible[i] {
			invisible[c] = struct{}{}
		}
	}

	// ensure visibility is reset when
	if len(invisible) == 0 {
		t.ResetColumnVisibility()
		return false
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().ColumnVisibility().SetAll(viewoptions.ColumnVisibility{
		Enabled:   true,
		InVisible: invisible,
	}).Do(); !ok {
		return false
	}

	for i, c := range tablecols {
		_, isInvisible := invisible[c.Title]
		tablecols[i].InVisible = isInvisible
	}

	t.updateTable(tablecols, nil, nil)

	return true
}

// SetSearchEnable merely enables the search view-options, without setting any
// additional parameters or updating the table view.
func (t *ItemsTable) SetSearchEnable() bool {
	search := t.viewOptions.GetSearchResultsOptions()
	search.Enabled = true
	var ok bool
	t.viewOptions, ok = t.viewOptions.Set().SearchResults().SetAll(search).Do()
	return ok
}

// SetSearchResults updates the searchResults state. Changes to search results
// affects only the virtual rows being displayed. The function returns a boolean
// that indicates whether the mutation was accepted and successfully applied.
func (t *ItemsTable) SetSearchResults(col string, results []search.FilteredItem) bool {
	var (
		matchedItems = make([]int, len(results))
		matchedRunes = make([][]int, len(results))
		matchedRows  = t.table.Rows()
		colIdx       = findColumnByTitle(t.table.Columns(), col)
	)

	filtered := make([]table.Row, len(results))
	for i, match := range results {
		filtered[i] = matchedRows[match.Index]
		matchedItems[i] = match.Index
		matchedRunes[i] = match.Matches
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().SearchResults().SetAll(viewoptions.SearchResults{
		MatchedItems: matchedItems,
		MatchedRunes: matchedRunes,
		ColumnIndex:  colIdx,
		Enabled:      true, // TODO: not set enable here
	}).Do(); !ok {
		return false
	}

	t.updateTable(nil, nil, filtered)
	return true
}

func (t *ItemsTable) SetDynamicColumnWidth(b bool) {
	t.table.SetDynamicColumnWidth(b)
}
