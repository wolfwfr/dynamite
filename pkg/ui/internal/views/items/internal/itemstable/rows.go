package itemstable

import (
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
)

func parseRows(cols []string, tableKeys [][]apitypes.KeyValue) []table.Row {
	rows := make([]table.Row, len(tableKeys))
	for i, k := range tableKeys {
		raw := make([]string, len(cols))
		styled := make([]string, len(cols))
		fields := make([]table.Field, len(cols))
		var x int
		for j, key := range cols {
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
	return rows
}
