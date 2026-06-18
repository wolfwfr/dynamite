package viewoptions

import (
	"fmt"

	u "github.com/wolfwfr/dynamite/pkg/util"
)

// ViewOptions defines the parameters that determine the view onto the loaded
// items. These are managed on top of the table-component layer.
//
// Its fields are private to prevent immediate access. This allows the struct to
// guard against invalid option combinations.
//
// Getters are available for read access to the current state.
//
// Setters are available but maintain the power to disregard the provided
// changes, if they are not compatible with the state of other options. Setters
// return a boolean that indicates whether the changes were accepted or not.
//
// A Check method is available that informs the client on which changes will
// be accepted.
type ViewOptions struct {
	// TODO: rename to search to distinguish from dynamodb-filter functionality
	searchResults    SearchResults
	columnVisibility ColumnVisibility
	columnSorting    ColumnSorting
}

func NewViewOptions() ViewOptions {
	v := ViewOptions{}
	v.columnVisibility.InVisible = make(map[string]struct{})
	return v
}

func (v *ViewOptions) Set() *Setter {
	return &Setter{v: *v}
}

// TODO: relocate
func (v *ViewOptions) GetColumnSuffix(colTitle string) string {
	var (
		sortEnabled = v.columnSorting.Enabled
		sortingOn   = v.columnSorting.SortingOn
		sortasc     = v.columnSorting.Ascending
	)
	if sortEnabled && sortingOn == colTitle {
		return fmt.Sprintf(" (%s)", u.Ternary("↑", "↓", sortasc))
	}
	return ""
}
