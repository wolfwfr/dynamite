package itemstable

import (
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
)

type transform func(rowIndex int, cols []ColumnAttributes, row table.Row) table.Row

func parseRows(cols []ColumnAttributes, tableKeys [][]apitypes.KeyValue, transforms []transform) []table.Row {
	rows := make([]table.Row, len(tableKeys))
	for i, k := range tableKeys {
		raw := make([]string, len(cols))
		styled := make([]string, len(cols))
		fields := make([]table.Field, len(cols))
		var x int
		for j, attrs := range cols {
			key := attrs.Title
			if key == k[x].Key { // matching key
				raw[j] = k[x].Value
				styled[j] = k[x].ValueStyling.Render(k[x].Value)
				fields[j] = EnrichedField{
					RawValue: k[x].Value,
					Style:    &k[x].ValueStyling,
				}
				x = min(len(k)-1, x+1)
			} else { // no matching key
				raw[j] = ""
				styled[j] = ""
				fields[j] = EnrichedField{
					RawValue: "",
					Style:    nil,
				}
			}
		}
		rows[i].Fields = fields
		rows[i].Metadata = map[string]any{ItemIndexMetaKey: i}
	}
	return transformRows(cols, rows, transforms)
}

func transformRows(cols []ColumnAttributes, rows []table.Row, transforms []transform) []table.Row {
	safe := make([]table.Row, len(rows))
	copy(safe, rows)
	for _, t := range transforms {
		for i := range safe {
			safe[i] = t(i, cols, safe[i])
		}
	}
	return safe
}
