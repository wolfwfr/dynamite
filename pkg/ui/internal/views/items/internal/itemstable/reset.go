package itemstable

import (
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
)

// Reset completely resets all internal state parameters and empties the table
func (t *ItemsTable) Reset() {
	t.Items = types.Items{}
	t.KeysComplete = []string{}

	t.viewOptions.ResetColumnSortingState()
	t.viewOptions.ResetColumnVisibilityState()
	t.viewOptions.ResetSearchState()

	t.table.SetCursor(0)

	t.updateTable([]table.Column{}, []table.Row{}, []table.Row{})
}

// ResetColumnVisibility resets column-visibility related state parameters and
// updates the table contents
func (t *ItemsTable) ResetColumnVisibility() {
	t.viewOptions.ResetColumnVisibilityState()
	t.updateTable(assembleColumns(t.viewOptions, t.KeysComplete), nil, nil)
}

// ResetColumnSorting resets column-sorting related state parameters and updates
// the table contents
func (t *ItemsTable) ResetColumnSorting() {
	t.viewOptions.ResetColumnSortingState()
	t.updateTable(assembleColumns(t.viewOptions, t.KeysComplete), parseRows(t.KeysComplete, t.Items.TableKeys), nil)
}

// ResetSearch resets search related state parameters and updates the table
// contents
func (t *ItemsTable) ResetSearch() {
	t.viewOptions.ResetSearchState()
	t.table.ResetVirtualRows()
	t.updateTable(nil, t.sortRows(t.table.Columns(), parseRows(t.KeysComplete, t.Items.TableKeys)), nil)
}

// clearCache completely removes any cached state
// Note that clearing of cache does not automatically imply that the table's
// rendered rows will be updated anew. Use refreshCache if this is your goal.
func (t *ItemsTable) clearCache() {
	t.renderCache = map[string]string{}
	t.table.ResetCache()
}

// refreshCache clears the cache and then forces a rerender of rows
func (t *ItemsTable) refreshCache() {
	t.clearCache()
	t.table.UpdateContent()
}
