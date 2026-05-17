package modulator

import (
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
)

// Reset completely resets all internal state parameters and returns empty table
// contents.
func (m *Modulator) Reset() ([]table.Column, []table.Row, []table.Row) {
	m.Items = types.Items{}
	m.KeysComplete = []string{}

	m.resetColumnSortingState()
	m.resetColumnVisibilityState()
	m.resetSearchState()

	return []table.Column{}, []table.Row{}, []table.Row{}
}

// ResetColumnVisibility resets column-visibility related state parameters and
// returns updates for the table contents (columns, rows, and virtual rows).
//
// Nil returns signify that no update is required
func (m *Modulator) ResetColumnVisibility() ([]table.Column, []table.Row, []table.Row) {
	m.resetColumnVisibilityState()
	return m.assembleColumns(m.KeysComplete), nil, nil
}

// ResetColumnSorting resets column-sorting related state parameters and returns
// updates for the table contents (columns, rows, and virtual rows).
//
// Nil returns signify that no update is required
func (m *Modulator) ResetColumnSorting() ([]table.Column, []table.Row, []table.Row) {
	m.resetColumnSortingState()
	return m.assembleColumns(m.KeysComplete), parseRows(m.KeysComplete, m.Items.TableKeys), nil
}

// ResetSearch resets search related state parameters and returns updates for
// the table contents (columns, rows, and virtual rows).
//
// Nil returns signify that no update is required
func (m *Modulator) ResetSearch() ([]table.Column, []table.Row, []table.Row) {
	m.resetSearchState()
	return nil, m.sortRows(parseRows(m.KeysComplete, m.Items.TableKeys)), nil
}

// resetColumnVisibilityState resets state relating to column-visibility functionality
func (m *Modulator) resetColumnVisibilityState() {
	m.ColumnVisibility.Enabled = false
	m.ColumnVisibility.InVisible = make(map[string]struct{}, 0)
}

// resetColumnSortingState resets internal state relating to column-sorting functionality
func (m *Modulator) resetColumnSortingState() {
	m.ColumnSorting.SortedItems = make([]int, 0)
	m.ColumnSorting.Ascending = true
	m.ColumnSorting.SortingOn = ""
	m.ColumnSorting.Enabled = false
}

// resetSearchState resets internal state relating to search functionality
func (m *Modulator) resetSearchState() {
	m.Itemfiltering.Enabled = false
	m.Itemfiltering.MatchedItems = make([]int, 0)
	m.Itemfiltering.MatchedRunes = make([][]int, 0)
}
