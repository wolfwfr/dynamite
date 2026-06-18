package itemstable

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

func NewItemsTable() *ItemsTable {
	m := ItemsTable{}

	// m.state.ColumnVisibility.InVisible = map[string]struct{}{}
	m.viewOptions = viewoptions.NewViewOptions()

	{ // contents table
		t := table.New(
			table.WithFocused(true),
			table.WithDynamicColumnWidth(false),
			table.WithFieldDelegate(m.TableRowFieldDelegate),
		)
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(commonstyles.TableDefaultFg).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(commonstyles.TableSelectedFg).
			Background(commonstyles.TableSelectedBg).
			Bold(false)
		t.SetStyles(s)

		st := TableStyles{
			SelectedBackground:    commonstyles.TableSelectedBg,
			SearchMatchBackground: commonstyles.SearchHighlight,
		}

		m.table = t
		m.styles = st
	}

	return &m
}

func (t *ItemsTable) Init() tea.Cmd {
	t.renderCache = map[string]string{}
	return nil
}

func (t *ItemsTable) Update(msg tea.Msg) tea.Cmd {
	return t.table.Update(msg)
}

func (t *ItemsTable) GetAllowedOptions() viewoptions.Check {
	return t.viewOptions.Check()
}

// TODO: consider; leaky abstraction?
func (t *ItemsTable) GetViewOptionsState() viewoptions.ViewOptions {
	return t.viewOptions
}

// AddItems processes dynamo-db items and appends them to the table contents,
// applying all active modulations and updating the table as required.
func (t *ItemsTable) AddItems(items apitypes.Items, hasRangeKey bool) {
	t.appendItems(items)
	if len(items.TableKeys) <= 0 {
		return
	}

	// set columns
	completeKeys := compileCompleteKeys(items.TableKeys, t.KeysComplete, hasRangeKey)
	defer func() { t.KeysComplete = completeKeys }()

	var (
		cols []table.Column
		rows []table.Row
		virt []table.Row

		noColumnUpdate = slices.Equal(t.KeysComplete, completeKeys)
		columnUpdate   = !noColumnUpdate
		appendOnly     = noColumnUpdate && !t.viewOptions.GetColumnSortingOptions().Enabled
	)

	switch {
	case columnUpdate: // update columns & ALL rows
		cols = assembleColumns(t.viewOptions, completeKeys)
		rows = parseRows(completeKeys, t.Items.TableKeys)
	case appendOnly: // update with  new rows (append)
		rows = parseRows(completeKeys, items.TableKeys)
	default: // update ALL rows but no columns
		rows = parseRows(completeKeys, t.Items.TableKeys)
	}

	t.updateTable(cols, rows, virt)
}

func (t *ItemsTable) appendItems(newItems apitypes.Items) {
	t.Items = apitypes.Items{
		JSON:       mergeSlices(t.Items.JSON, newItems.JSON),
		JSONStyled: mergeSlices(t.Items.JSONStyled, newItems.JSONStyled),
		YAML:       mergeSlices(t.Items.YAML, newItems.YAML),
		YAMLStyled: mergeSlices(t.Items.YAMLStyled, newItems.YAMLStyled),
		Raw:        mergeSlices(t.Items.Raw, newItems.Raw),
		TableKeys:  mergeSlices(t.Items.TableKeys, newItems.TableKeys),
	}
}

func (t *ItemsTable) View() string {
	return t.table.View()
}

func (t *ItemsTable) GetSelectedItem() *Item {
	var (
		sorting = t.viewOptions.GetColumnSortingOptions()
		filter  = t.viewOptions.GetSearchResultsOptions()
		items   = t.Items
	)

	if len(items.Raw) == 0 || filter.Enabled && len(filter.MatchedItems) == 0 {
		return nil
	}

	idx := t.table.Cursor()

	switch {
	case len(sorting.SortedItems) > 0:
		idx = sorting.SortedItems[idx]
	case len(filter.MatchedItems) > 0:
		idx = filter.MatchedItems[idx]
	}

	return &Item{
		JSON:       items.JSON[idx],
		JSONStyled: items.JSONStyled[idx],
		YAML:       items.YAML[idx],
		YAMLStyled: items.YAMLStyled[idx],
		Raw:        items.Raw[idx],
		TableKeys:  items.TableKeys[idx],
	}
}

// updateTable processes the common response format from modulated-content
// mutations (Sets & Resets), which return updates to columns, rows, and virtual
// rows. It appropriately refreshes the internal render-cache when necessary.
func (t *ItemsTable) updateTable(columns []table.Column, rows []table.Row, virt []table.Row) {
	// always apply sorting
	rows = t.sortRowsAndUpdate(u.Ternary(columns, t.table.Columns(), columns != nil), rows)
	virt = t.sortRowsAndUpdate(u.Ternary(columns, t.table.Columns(), columns != nil), virt)

	switch {
	case columns == nil && rows == nil: // no update
	case columns != nil && rows != nil: // update both
		t.table.SetContent(columns, rows)
		t.refreshCache()
	case columns == nil: // update only rows
		t.table.SetRows(rows)
		t.refreshCache()
	default: // update only columns
		t.table.SetColumns(columns)
	}

	if virt == nil {
		return
	}
	t.table.SetVirtualRows(virt)
	t.refreshCache()
}
