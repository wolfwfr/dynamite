package modulator

import (
	"fmt"
	"slices"
	"strconv"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

const ItemIndexMetaKey = "item_index"

// NewModulator constructs a new Modulator
func NewModulator(table *table.Model) *Modulator {
	m := &Modulator{}

	m.ColumnVisibility.InVisible = map[string]struct{}{}

	m.table = table

	return m
}

// SetColumnSorting updates the internal state appropriately and returns updates
// for the table contents (columns, rows, and virtual rows).
//
// SetColumnSorting returns both sorted rows as well as sorted virtual rows.
//
// Nil returns signify that no update is required
func (m *Modulator) SetColumnSorting(cols []string, sortingOn string, ascending bool) ([]table.Column, []table.Row, []table.Row) {
	tablecols := m.table.Columns()
	if len(tablecols) != len(cols) {
		// TODO: better handling of new columns appearing in view
		return m.ResetColumnSorting()
	}

	// update panel state
	m.ColumnSorting.Enabled = true
	m.ColumnSorting.Ascending = ascending
	m.ColumnSorting.SortingOn = sortingOn

	// prepare table column update
	for i, c := range tablecols {
		c.Suffix = m.getColumnSuffix(c.Title)
		tablecols[i] = c
	}

	return tablecols, m.sortRows(m.table.Rows()), m.sortRows(m.table.VirtualRows())
}

// SetColumnVisibility updates the internal state appropriately and returns
// updates for the table contents (columns, rows, and virtual rows).
//
// Nil returns signify that no update is required
func (m *Modulator) SetColumnVisibility(cols []string, visible []bool) ([]table.Column, []table.Row, []table.Row) {
	tablecols := m.table.Columns()
	if len(tablecols) != len(cols) {
		// TODO: better handling of new columns appearing in view
		return m.ResetColumnVisibility()
	}
	m.ColumnVisibility.Enabled = true
	for i, c := range cols {
		if !visible[i] {
			m.ColumnVisibility.InVisible[c] = struct{}{}
		} else {
			delete(m.ColumnVisibility.InVisible, c)
		}
	}

	if len(m.ColumnVisibility.InVisible) == 0 {
		m.ColumnVisibility.Enabled = false
		return nil, nil, nil
	}

	for i, c := range tablecols {
		_, isInvisible := m.ColumnVisibility.InVisible[c.Title]
		tablecols[i].InVisible = isInvisible
	}
	return tablecols, nil, nil
}

// SetSearchResults updates the internal state appropriately and returns updates
// for the table contents (columns, rows, and virtual rows).
//
// Note that it is not compatible with column-sorting and will return the
// results of a column-sorting reset in addition to updates to virtual rows,
// when column-sorting was enabled.
//
// Nil returns signify that no update is required
func (m *Modulator) SetSearchResults(col string, results []search.FilteredItem) ([]table.Column, []table.Row, []table.Row) {
	var cols []table.Column = nil
	var rows []table.Row = nil
	// column-sorting & search are currently not compatible, because search
	// applies its own sorting
	// TODO: allow compatibility
	if m.ColumnSorting.Enabled {
		cols, rows, _ = m.ResetColumnSorting()
	}

	m.Itemfiltering.Enabled = true
	m.Itemfiltering.MatchedItems = make([]int, len(results))
	m.Itemfiltering.MatchedRunes = make([][]int, len(results))
	matchedRows := m.table.Rows()
	colIdx := findColumnByTitle(m.table.Columns(), col)
	m.Itemfiltering.ColumnIndex = colIdx
	filtered := make([]table.Row, len(results))
	for i, match := range results {
		filtered[i] = matchedRows[match.Index]
		m.Itemfiltering.MatchedItems[i] = match.Index
		m.Itemfiltering.MatchedRunes[i] = match.Matches
	}
	// TODO: remove VirtualRows from table
	return cols, rows, filtered
}

// AddItems processes dynamo-db items and appends them to the table contents,
// applying all active modulations and updating the table as required.
func (m *Modulator) AddItems(items apitypes.Items, hasRangeKey bool) ([]table.Column, []table.Row, []table.Row) {
	m.appendItems(items)
	if len(items.TableKeys) > 0 {
		// set columns
		completeKeys := compileCompleteKeys(items.TableKeys, m.KeysComplete, hasRangeKey)
		defer func() { m.KeysComplete = completeKeys }()

		noColumnUpdate := slices.Equal(m.KeysComplete, completeKeys)
		columnUpdate := !noColumnUpdate
		appendOnly := noColumnUpdate && !m.ColumnSorting.Enabled

		switch {
		case columnUpdate: // update columns & ALL rows
			cols := m.assembleColumns(completeKeys)
			rows := parseRows(completeKeys, m.Items.TableKeys)
			return cols, m.sortRows(rows), nil
		case appendOnly: // update with  new rows (append)
			rows := parseRows(completeKeys, items.TableKeys)
			return nil, m.sortRows(rows), nil
		default: // update ALL rows but no columns
			rows := parseRows(completeKeys, m.Items.TableKeys)
			return nil, m.sortRows(rows), nil
		}
	}
	return nil, nil, nil
}

func (m *Modulator) appendItems(newItems apitypes.Items) {
	// JSON
	m.Items.JSON = mergeSlices(m.Items.JSON, newItems.JSON)
	// JSON-styled
	m.Items.JSONStyled = mergeSlices(m.Items.JSONStyled, newItems.JSONStyled)
	// YAML
	m.Items.YAML = mergeSlices(m.Items.YAML, newItems.YAML)
	// YAML-styled
	m.Items.YAMLStyled = mergeSlices(m.Items.YAMLStyled, newItems.YAMLStyled)
	// RAW
	m.Items.Raw = mergeSlices(m.Items.Raw, newItems.Raw)
	// KEYS
	m.Items.TableKeys = mergeSlices(m.Items.TableKeys, newItems.TableKeys)
}

func mergeSlices[S ~[]E, E any](s1, s2 S) S {
	n := make([]E, len(s1)+len(s2))
	copy(n[:len(s1)], s1)
	copy(n[len(s1):], s2)
	return n
}

// assembleColumns returns a set of table columns that incorporates modulations
// based on the item-selection-pane state, such as the state of column
// visibility and sorting.
func (m *Modulator) assembleColumns(allColumnTitles []string) []table.Column {
	cols := make([]table.Column, len(allColumnTitles))

	for i, title := range allColumnTitles {
		col := table.Column{Title: title, Width: u.Clamp(len(title), 16, 32)}

		// visibility
		_, isInvisible := m.ColumnVisibility.InVisible[title]
		col.InVisible = m.ColumnVisibility.Enabled && isInvisible

		// suffix
		col.Suffix = m.getColumnSuffix(title)

		// insert
		cols[i] = col
	}
	return cols
}

func (m *Modulator) getColumnSuffix(colTitle string) string {
	if m.ColumnSorting.Enabled && m.ColumnSorting.SortingOn == colTitle {
		return fmt.Sprintf(" (%s)", u.Ternary("↑", "↓", m.ColumnSorting.Ascending))
	}
	return ""
}

func (m *Modulator) sortRows(rows []table.Row) []table.Row {
	if !m.ColumnSorting.Enabled || m.ColumnSorting.SortingOn == "" || len(rows) == 0 {
		return rows
	}
	cols := m.table.Columns()
	colsS := make([]string, len(cols))
	for i, c := range cols {
		colsS[i] = c.Title
	}
	idx := u.Find(colsS, m.ColumnSorting.SortingOn)
	if idx < 0 {
		return rows
	}

	// determine field-type
	var field string
	for _, r := range rows {
		if r.Fields[idx].Value() != "" {
			field = r.Fields[idx].Value()
			break
		}
	}
	_, errInt := strconv.ParseInt(field, 10, 64)
	_, errFloat := strconv.ParseFloat(field, 64)

	// choose the appropriate sorting function
	var sortFunc func(a, b sortingRow) int
	switch {
	// NOTE: assumes that float fields always contain decimal point
	case errFloat == nil:
		sortFunc = func(a, b sortingRow) int {
			aI, _ := strconv.ParseFloat(a.r.Fields[idx].Value(), 64)
			bI, _ := strconv.ParseFloat(b.r.Fields[idx].Value(), 64)
			check := u.Ternary(aI < bI, aI > bI, m.ColumnSorting.Ascending)
			return u.Ternary(-1, 1, check)
		}
	case errInt == nil:
		sortFunc = func(a, b sortingRow) int {
			aI, _ := strconv.ParseInt(a.r.Fields[idx].Value(), 10, 64)
			bI, _ := strconv.ParseInt(b.r.Fields[idx].Value(), 10, 64)
			check := u.Ternary(aI < bI, aI > bI, m.ColumnSorting.Ascending)
			return u.Ternary(-1, 1, check)
		}
	default:
		sortFunc = func(a, b sortingRow) int {
			s := []string{a.r.Fields[idx].Value(), b.r.Fields[idx].Value()}
			slices.Sort(s)
			check := u.Ternary(s[0] == a.r.Fields[idx].Value(), s[1] == a.r.Fields[idx].Value(), m.ColumnSorting.Ascending)
			return u.Ternary(-1, 1, check)
		}
	}

	// apply sorting function on slice backed by new array
	sorted := make([]sortingRow, len(rows))
	for i, r := range rows {
		sorted[i] = sortingRow{
			r: r,
			i: r.Metadata[ItemIndexMetaKey].(int),
		}
	}

	// sort
	slices.SortFunc(sorted, sortFunc)

	// reset sorted-item-mapping
	m.ColumnSorting.SortedItems = make([]int, len(sorted))

	res := make([]table.Row, len(sorted))
	for i := range sorted {
		m.ColumnSorting.SortedItems[i] = sorted[i].i
		res[i] = sorted[i].r
	}

	return res
}

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

// compileCompleteKeys takes a table of key-value pairs, observes all keys and
// compiles a complete, in-order list of all unique key observed.
// This ensures that when individual table rows have keys missing, the final
// result still contains these keys when they are present in other rows in the
// specified table.
func compileCompleteKeys(table [][]apitypes.KeyValue, existing []string, hasRangeKey bool) []string {
	res := make([]string, 0)
	seen := map[string]struct{}{}
	if len(existing) > 0 {
		res = existing
	}
	for _, e := range existing {
		seen[e] = struct{}{}
	}
	for _, row := range table {
		for _, col := range row {
			key := col.Key
			if _, ok := seen[key]; !ok {
				res = append(res, key)
				seen[key] = struct{}{}
			}
		}
	}

	sortLenOffset := u.Ternary(2, 1, hasRangeKey)
	toSort := make([]string, len(res)-sortLenOffset)
	copy(toSort, res[sortLenOffset:])
	slices.Sort(toSort)
	copy(res[sortLenOffset:], toSort)

	return res
}
