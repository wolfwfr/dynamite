package itemstable

import (
	"fmt"
	"strings"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	view "github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// assembleColumns returns a set of table columns that incorporates modulations
// based on the item-selection-pane state, such as the state of column
// visibility and sorting.
func assembleColumns(state view.ViewOptions, columnAttributes []ColumnAttributes) []table.Column {
	var (
		cols      = make([]table.Column, len(columnAttributes))
		visopts   = state.GetColumnVisibilityOptions()
		widthopts = state.GetColumnDynWidthOptions()
	)

	for i, attrs := range columnAttributes {
		title := attrs.Title

		col := table.Column{Title: title}

		// suffix
		col.Suffix = getColumnSuffix(state, title)

		// update width, respecting suffix
		col.Width = u.Clamp(len(title)+len(col.Suffix), 16, 32)

		// visibility
		_, isInvisible := visopts.InVisible[title]
		col.InVisible = visopts.Enabled && isInvisible

		// column dynamic width
		_, dyn := widthopts.DynWidth[title]
		col.UseDynamicWidth = widthopts.Enabled && dyn

		// insert
		cols[i] = col
	}
	return cols
}

func getColumnSuffix(state view.ViewOptions, colTitle string) string {
	sortopts := state.GetColumnSortingOptions()
	tranopts := state.GetColumnTransformOptions()
	var (
		sortEnabled = sortopts.Enabled
		sortingOn   = sortopts.SortingOn
		sortasc     = sortopts.Ascending
		transformed = tranopts.Transformed
	)
	var suffix strings.Builder
	if sortEnabled && sortingOn == colTitle {
		suffix.WriteString(fmt.Sprintf(" (%s)", u.Ternary("↑", "↓", sortasc)))
	}
	if _, ok := transformed[colTitle]; ok {
		suffix.WriteString(" (⇒)")
	}
	return suffix.String()
}

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
