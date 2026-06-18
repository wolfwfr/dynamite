package itemstable

import (
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	view "github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// assembleColumns returns a set of table columns that incorporates modulations
// based on the item-selection-pane state, such as the state of column
// visibility and sorting.
func assembleColumns(state view.ViewOptions, allColumnTitles []string) []table.Column {
	var (
		cols    = make([]table.Column, len(allColumnTitles))
		visopts = state.GetColumnVisibilityOptions()
	)

	for i, title := range allColumnTitles {
		col := table.Column{Title: title, Width: u.Clamp(len(title), 16, 32)}

		// visibility
		_, isInvisible := visopts.InVisible[title]
		col.InVisible = visopts.Enabled && isInvisible

		// suffix
		col.Suffix = state.GetColumnSuffix(title)

		// insert
		cols[i] = col
	}
	return cols
}

//	func (t *ItemsTable) getColumnSuffix(colTitle string) string {
//		var (
//			sortEnabled = t.state.ColumnSorting.Enabled
//			sortingOn   = t.state.ColumnSorting.SortingOn
//			sortasc     = t.state.ColumnSorting.Ascending
//		)
//		if sortEnabled && sortingOn == colTitle {
//			return fmt.Sprintf(" (%s)", u.Ternary("↑", "↓", sortasc))
//		}
//		return ""
//	}
func findColumnByTitle(cols []table.Column, title string) int {
	idx := -1
	for i, c := range cols {
		if c.Title == title {
			idx = i
			break
		}
	}
	return idx
}
