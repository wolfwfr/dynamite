package itemstable

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
)

func TestSortRows(t *testing.T) {
	cols := []table.Column{
		{Title: "id"},
		{Title: "int"},
		{Title: "float"},
		{Title: "bool"},
		{Title: "bytes"},
	}

	rows := []table.Row{
		{
			Fields: []table.Field{
				EnrichedField{RawValue: "id-1"},
				EnrichedField{RawValue: "10"},
				EnrichedField{RawValue: "10.5"},
				EnrichedField{RawValue: "true"},
			},
			Metadata: map[string]any{
				ItemIndexMetaKey: 0,
			},
		},
		{
			Fields: []table.Field{
				EnrichedField{RawValue: "id-0"},
				EnrichedField{RawValue: "8"},
				EnrichedField{RawValue: "8.5"},
				EnrichedField{RawValue: "false"},
			},
			Metadata: map[string]any{
				ItemIndexMetaKey: 1,
			},
		},
		{
			Fields: []table.Field{
				EnrichedField{RawValue: "id-2"},
				EnrichedField{RawValue: "12"},
				EnrichedField{RawValue: "12.5"},
				EnrichedField{RawValue: "true"},
			},
			Metadata: map[string]any{
				ItemIndexMetaKey: 2,
			},
		},
	}

	t.Run("sort-rows should", func(t *testing.T) {
		// convenience resources

		// testcases
		testcases := []struct {
			desc          string
			viewOptions   viewoptions.ColumnSorting
			expectSorted  []table.Row
			expectIndices []int
		}{
			{
				desc:          "sort on string value ascending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[0].Title, Ascending: true},
				expectSorted:  []table.Row{rows[1], rows[0], rows[2]},
				expectIndices: []int{1, 0, 2},
			},
			{
				desc:          "sort on string value descending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[0].Title, Ascending: false},
				expectSorted:  []table.Row{rows[2], rows[0], rows[1]},
				expectIndices: []int{2, 0, 1},
			},
			{
				desc:          "sort on int value ascending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[1].Title, Ascending: true},
				expectSorted:  []table.Row{rows[1], rows[0], rows[2]},
				expectIndices: []int{1, 0, 2},
			},
			{
				desc:          "sort on int value descending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[1].Title, Ascending: false},
				expectSorted:  []table.Row{rows[2], rows[0], rows[1]},
				expectIndices: []int{2, 0, 1},
			},
			{
				desc:          "sort on float value ascending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[2].Title, Ascending: true},
				expectSorted:  []table.Row{rows[1], rows[0], rows[2]},
				expectIndices: []int{1, 0, 2},
			},
			{
				desc:          "sort on float value descending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[2].Title, Ascending: false},
				expectSorted:  []table.Row{rows[2], rows[0], rows[1]},
				expectIndices: []int{2, 0, 1},
			},
			{
				desc:          "sort on bool value as string ascending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[3].Title, Ascending: true},
				expectSorted:  []table.Row{rows[1], rows[2], rows[0]},
				expectIndices: []int{1, 2, 0},
			},
			{
				desc:          "sort on bool value as string descending",
				viewOptions:   viewoptions.ColumnSorting{Enabled: true, SortingOn: cols[3].Title, Ascending: false},
				expectSorted:  []table.Row{rows[2], rows[0], rows[1]},
				expectIndices: []int{2, 0, 1},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				// test
				sorted, indices := sortRows(cols, rows, tc.viewOptions)

				// assert
				assert.EqualValues(t, tc.expectSorted, sorted)
				assert.EqualValues(t, tc.expectIndices, indices)
			})
		}
	})
}
