package itemstable

import (
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/lib/styles"
	"github.com/wolfwfr/dynamite/pkg/common"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/theme"
)

func (i *ItemsTable) CompileTransforms() []transform {
	opts := i.viewOptions.GetColumnTransformOptions()
	if len(opts.Transformed) == 0 {
		return make([]transform, 0)
	}

	res := make([]transform, 1)
	res[0] = func(rowIndex int, cols []ColumnAttributes, row table.Row) table.Row {
		for i, f := range row.Fields {
			if _, ok := opts.Transformed[cols[i].Title]; !ok {
				continue
			}
			if cols[i].Type != common.DynamoDBAttributeTypeS && cols[i].Type != common.DynamoDBAttributeTypeN {
				continue
			}
			v := f.Value()
			v = strings.TrimSuffix(strings.TrimPrefix(v, "\""), "\"")
			unix, err := strconv.Atoi(v)
			if err != nil {
				continue // TODO: debug log
			}
			// NOTE: best effort distinction between unix and smaller formats
			threshold := time.Date(1e4, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			maxIterations := 3
			iter := 1
			for int64(unix) > threshold && iter < maxIterations {
				iter++
				unix = unix / 1000
			}
			u := time.Unix(int64(unix), 0)
			tr := u.Format("2006-01-02 15:04:05 Z07:00") // TODO: support custom formats
			st := styles.LineStyle{}.AppendStringLG(tr, lipgloss.NewStyle().Foreground(theme.TimestampColour))
			row.Fields[i] = EnrichedField{
				RawValue: tr,
				Style:    &st,
			}
		}
		return row
	}
	return res
}
