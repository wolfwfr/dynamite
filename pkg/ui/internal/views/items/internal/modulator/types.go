package modulator

import (
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

// TableMask masks the table's interface to only allow relevant
// read-operations for the content-modulator.
type TableMask interface {
	Columns() []table.Column
	Rows() []table.Row
	VirtualRows() []table.Row
}

// Modulator defines the table content modulation pipeline and dictates when
// table columns or rows are updated.
type Modulator struct {
	// access to the tables current contents
	table TableMask

	// dynamo-db-Items including JSON/YAML render & styling instructions
	Items types.Items

	// item filtering collects settings related to item filtering
	Itemfiltering struct {
		MatchedItems []int   // indices referring to items
		MatchedRunes [][]int //matches by index to itemfiltering.matchedItems
		ColumnIndex  int
		Enabled      bool
	}

	// column visibility collects settings related to column visibillity
	ColumnVisibility struct {
		Enabled   bool
		InVisible map[string]struct{}
	}

	// column sorting collects settings related to column sorting
	ColumnSorting struct {
		SortedItems []int // indices referring to items
		SortingOn   string
		Ascending   bool // if false, descending
		Enabled     bool
	}

	// KeysComplete represents a unique set of dynamo-db item keys that
	// exhaustively cover all keys in the currently paged set of items
	KeysComplete []string
}

// EnrichedField defines the field-type that populates a table-row.
type EnrichedField struct {
	RawValue string
	Style    *commonstyles.LineStyle
}

// Value implements the matching table.Field interface function
func (f EnrichedField) Value() string {
	return f.RawValue
}

// sortingRow is a wrapper around row that couples the row to the index of the
// original item
type sortingRow struct {
	r table.Row
	i int
}
