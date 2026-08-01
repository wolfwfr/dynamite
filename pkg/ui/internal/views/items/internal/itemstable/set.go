package itemstable

import (
	"fmt"
	"log/slog"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
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
		t.logger.Warn("column length mismatch on set-column-sorting; rejecting",
			slog.Int("len_table_columns", len(tablecols)),
			slog.Int("len_incom_columns", len(cols)),
		)
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
		t.logger.Warn("set-column-sorting viewoptions update rejected",
			slog.String("sorting_on", sortingOn),
			slog.Bool("ascending", ascending),
		)
		return false
	}

	// NOTE: parsing rows anew assures a consistent input to update-table &
	// sort-rows, preventing sort-rows from regurgitating its own output on when
	// changing sorting without resets; leads to more consistent outputs.
	t.updateTable(assembleColumns(t.viewOptions, t.ColumnAttributes), parseRows(t.ColumnAttributes, t.Items.TableKeys, t.CompileTransforms()), nil)
	return true
}

// SetColumnVisibility updates the column-visibility state. Changes to column
// visibility only affect the columns and do not affect table rows. The function
// returns a boolean that indicates whether the mutation was accepted and
// successfully applied.
func (t *ItemsTable) SetColumnVisibility(cols []string, visible []bool) bool {
	// guard against mismatched states
	tablecols := t.table.Columns()
	if len(tablecols) != len(cols) {
		// TODO: better handling of new columns appearing in view
		t.logger.Warn("column length mismatch on set-column-visibility; rejecting",
			slog.Int("len_table_columns", len(tablecols)),
			slog.Int("len_incom_columns", len(cols)),
		)
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
		t.logger.Debug("no invisible columns; resetting")
		t.ResetColumnVisibility()
		return false
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().ColumnVisibility().SetAll(viewoptions.ColumnVisibility{
		Enabled:   true,
		InVisible: invisible,
	}).Do(); !ok {
		t.logger.Warn("set-column-visibility viewoptions update rejected",
			slog.Any("invisible_set", invisible),
		)
		return false
	}

	t.updateTable(assembleColumns(t.viewOptions, t.ColumnAttributes), nil, nil)

	return true
}

// SetColumnTransform updates the column-transform state. Changes to column
// transform affect the column suffix and the (virtual) rows being displayed.
// The function returns a boolean that indicates whether the mutation was
// accepted and successfully applied.
func (t *ItemsTable) SetColumnTransform(cols []string, transformed []bool) bool {
	// ensure transform is reset when
	// map transformed → transformedM
	transformedM := make(map[string]struct{})
	for i, c := range cols {
		if transformed[i] {
			transformedM[c] = struct{}{}
		}
	}

	if len(transformedM) == 0 {
		t.logger.Debug("nothing to transform; resetting")
		t.ResetColumnTransform()
		return false
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().ColumnTransform().SetAll(viewoptions.ColumnTransform{
		Enabled:     true,
		Transformed: transformedM,
	}).Do(); !ok {
		t.logger.Warn("set-column-transform viewoptions update rejected",
			slog.Any("transformed_set", transformed),
		)
		return false
	}

	t.updateTable(assembleColumns(t.viewOptions, t.ColumnAttributes), parseRows(t.ColumnAttributes, t.Items.TableKeys, t.CompileTransforms()), nil)
	t.RebuildSearchResults()

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
		t.logger.Warn("set-search viewoptions update rejected",
			slog.String("matched_item_indices", fmt.Sprintf("%d", matchedItems)),
			slog.Any("matched_item_runes", matchedRunes),
			slog.Int("column_index", colIdx),
		)
		return false
	}

	t.updateTable(nil, nil, filtered)
	return true
}

// RebuildSearchResults can be used when the underlying rowset has changed but
// did not affect search results or search order. Calling this function will
// re-execute the existing search state and store the resulting virtual rows.
// The function returns true when successfully executed or false when the
// underlying rows were found to be incompatible with the current state of
// search results.
func (t *ItemsTable) RebuildSearchResults() bool {
	t.logger.Debug("rebuilding search results")
	var (
		viewopts = t.viewOptions.GetSearchResultsOptions()
		rows     = t.table.Rows()
		filtered = make([]table.Row, len(viewopts.MatchedItems))
	)

	for i, matchedIndex := range viewopts.MatchedItems {
		if matchedIndex >= len(rows) {
			return false
		}
		filtered[i] = rows[matchedIndex]
	}

	t.updateTable(nil, nil, u.Ternary(filtered, nil, len(filtered) > 0))
	return true
}

func (t *ItemsTable) SetColumnDynamicWidth(cols []string, dynamicWidth []bool) bool {
	// guard against mismatched states
	tablecols := t.table.Columns()
	if len(tablecols) != len(cols) {
		t.logger.Warn("column length mismatch on set-column-width; rejecting",
			slog.Int("len_table_columns", len(tablecols)),
			slog.Int("len_incom_columns", len(cols)),
		)
		// TODO: better handling of new columns appearing in view
		t.ResetColumnDynWidth()
		return false
	}

	// map width → widthM
	widthM := make(map[string]struct{})
	for i, c := range cols {
		if dynamicWidth[i] {
			widthM[c] = struct{}{}
		}
	}

	// ensure dynamic-width is reset when no columns match
	if len(widthM) == 0 {
		t.logger.Debug("no widened columns; resetting")
		t.ResetColumnDynWidth()
		return false
	}

	// update internal state
	var ok bool
	if t.viewOptions, ok = t.viewOptions.Set().ColumnDynamicWidth().SetAll(viewoptions.ColumnDynWidth{
		Enabled:  true,
		DynWidth: widthM,
	}).Do(); !ok {
		t.logger.Warn("set-column-width viewoptions update rejected",
			slog.Any("dynamic_width_set", widthM),
		)
		return false
	}

	t.updateTable(assembleColumns(t.viewOptions, t.ColumnAttributes), nil, nil)

	return true
}
