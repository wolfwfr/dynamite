package itemstable

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// TableRowFieldDelegate provides the delegate function for the items-table that
// applies styling to each field based on the styling defined for the associated
// item, whether the row is selected, and search-matching.
//
// It returns cached responses when possible.
func (t *ItemsTable) TableRowFieldDelegate(row table.Row, col table.Column, colIdx, rowIdx, colW, padL, padR int, selected bool) string {
	fullWidth := colW + padL + padR

	// obtain field in question
	field := row.Fields[colIdx].(EnrichedField)

	// fill up with padding if empty
	if field.Style == nil {
		st := lipgloss.NewStyle().PaddingRight(fullWidth)
		st = u.Ternary(st.Background(t.styles.SelectedBackground), st, selected)
		return st.Render("")
	}

	style := *field.Style

	itemfiltering := t.viewOptions.GetSearchResultsOptions()

	// attempt to obtain cached value to prevent rerendering
	cachekey := fmt.Sprintf("%d-%d-%d", rowIdx, colIdx, colW)
	cachCond := !selected && (!itemfiltering.Enabled || itemfiltering.ColumnIndex != colIdx)
	cc, ok := t.renderCache[cachekey]
	if ok && cachCond {
		return cc
	}

	// add padding
	style = style.SetRightPaddingLast(padR)
	style = style.SetLeftPaddingFirst(padL)

	// truncate row value to fit within specified column width
	truncated := ansi.Truncate(field.RawValue, colW, "…")
	if len([]rune(truncated)) < len([]rune(field.RawValue)) {
		st, _ := style.GetAt(len([]rune(truncated)) - 1)
		style = style.Override(len([]rune(truncated))-1, st.PaddingRight(padR))
	}
	field.RawValue = truncated

	// apply background styling for selected row
	if selected {
		// fill up any remaining space
		if len([]rune(field.RawValue)) < fullWidth {
			st, _ := style.GetAt(len([]rune(field.RawValue)) - 1)
			style = style.Override(len([]rune(field.RawValue))-1, st.PaddingRight(fullWidth-len([]rune(field.RawValue))))
		}
		style = style.SetBackgroundAll(t.styles.SelectedBackground)
	}

	// ensure that row-index is correctly interpreted as pointing to 'actual'
	// rows or virtual rows. When the table content contains virtual rows, it
	// will always only render those virtual rows.
	applyingVirtualRows := len(t.table.VirtualRows()) > 0

	// override background styling for search matches
	if applyingVirtualRows && itemfiltering.Enabled && itemfiltering.ColumnIndex == colIdx {
		for _, idx := range itemfiltering.MatchedRunes[rowIdx] {
			runeStyle, _ := style.GetAt(idx)
			c := t.styles.SearchMatchBackground
			if selected {
				c = lipgloss.Blend1D(10, c, t.styles.SelectedBackground)[3]
			}
			style = style.Override(idx, runeStyle.Background(c))
		}
	}

	enforceWidth := lipgloss.NewStyle().Width(fullWidth).MaxWidth(fullWidth).Inline(true).Render
	res := enforceWidth(style.Render(field.RawValue))

	// cache when appropriate for improved performance
	if cachCond {
		t.renderCache[cachekey] = res
	}

	return res
}
