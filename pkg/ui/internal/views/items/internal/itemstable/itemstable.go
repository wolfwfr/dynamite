package itemstable

import (
	"context"
	"log/slog"
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

func NewItemsTable(ctx context.Context, l *slog.Logger) *ItemsTable {
	m := ItemsTable{}

	// m.state.ColumnVisibility.InVisible = map[string]struct{}{}
	m.viewOptions = viewoptions.NewViewOptions()

	{ // contents table
		t := table.New(
			table.WithFocused(true),
			table.WithFieldDelegate(m.TableRowFieldDelegate),
		)
		m.table = t
	}

	m.ctx = ctx
	m.logger = l.With(slog.String(logging.ComponentKey, "items-table"))

	m.renderCache = map[string]string{}

	m.updateStyles()

	return &m
}

func (t *ItemsTable) updateStyles() {
	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(theme.TableHeaderFg).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.TableBorderFg).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(theme.TableSelectedFg).
		Background(theme.TableSelectedBg).
		Bold(false)
	t.table.SetStyles(s)

	st := TableStyles{
		SelectedBackground:    theme.TableSelectedBg,
		SearchMatchBackground: theme.SearchHighlight,
	}

	t.styles = st
}

func (t *ItemsTable) Init() tea.Cmd {
	t.logger.Info("initialising...")
	t.renderCache = map[string]string{}
	t.logger.Info("initialisation complete")
	return nil
}

func (t *ItemsTable) Update(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(tea.BackgroundColorMsg); ok {
		t.updateStyles()
		return nil
	}
	return t.table.Update(msg)
}

func (t *ItemsTable) PaginationEligible() bool {
	return !t.viewOptions.GetSearchResultsOptions().Enabled && t.table.ViewAtEnd()
}

func (t *ItemsTable) GetAllowedOptions() viewoptions.Check {
	return t.viewOptions.Check()
}

func (t *ItemsTable) UpdateSize(height, width int) {
	t.table.SetHeight(height)
	t.table.SetWidth(width)
	t.refreshCache()
}

// TODO: consider; leaky abstraction?
func (t *ItemsTable) GetViewOptionsState() viewoptions.ViewOptions {
	return t.viewOptions
}

func (t *ItemsTable) GetColumns() []table.Column {
	return t.table.Columns()
}

func (t *ItemsTable) GetColumnTypes() []ColumnAttributes {
	res := make([]ColumnAttributes, len(t.ColumnAttributes))
	copy(res, t.ColumnAttributes)
	return res
}

func (t *ItemsTable) GetRows() []table.Row {
	return t.table.Rows()
}

func (t *ItemsTable) GetVirtualRows() []table.Row {
	return t.table.VirtualRows()
}

func (t *ItemsTable) GetVisualRows() []table.Row {
	return t.table.VisualRows()
}

func (t *ItemsTable) GetKeyMap() *table.KeyMap {
	return t.table.KeyMap
}

// AddItems processes dynamo-db items and appends them to the table contents,
// applying all active modulations and updating the table as required.
func (t *ItemsTable) AddItems(items apitypes.Items, hasRangeKey bool) {
	t.logger.Debug("adding items",
		slog.Int("incoming_len", len(items)),
		slog.Int("existing_len", len(t.Items)),
		slog.Bool("with_range_key", hasRangeKey),
	)
	t.appendItems(items)
	if len(items) <= 0 {
		return
	}

	// set columns
	columnTitles := compileUniqueKeys(items, t.ColumnAttributes, hasRangeKey)
	defer func() { t.ColumnAttributes = columnTitles }()

	var (
		cols []table.Column = nil
		rows []table.Row    = nil
		virt []table.Row    = nil

		noColumnUpdate = slices.Equal(t.ColumnAttributes, columnTitles)
		columnUpdate   = !noColumnUpdate
		appendOnly     = noColumnUpdate && !t.viewOptions.GetColumnSortingOptions().Enabled
	)

	switch {
	case columnUpdate: // update columns & ALL rows
		cols = assembleColumns(t.viewOptions, columnTitles)
		rows = parseRows(columnTitles, t.Items, t.CompileTransforms())
	case appendOnly: // update with new rows (append)
		rows = parseRows(columnTitles, t.Items, t.CompileTransforms())
	default: // update ALL rows but no columns
		rows = parseRows(columnTitles, t.Items, t.CompileTransforms())
	}

	t.updateTable(cols, rows, virt)
}

func (t *ItemsTable) appendItems(newItems apitypes.Items) {
	t.Items = mergeSlices(t.Items, newItems)
}

func (t *ItemsTable) View() string {
	return t.table.View()
}

func (t *ItemsTable) GetSelectedRow() *table.Row {
	return t.table.SelectedRow()
}

func (t *ItemsTable) GetSelectedItem() (*apitypes.Item, int) {
	var (
		items = t.Items
		row   = t.table.SelectedRow()
	)

	if len(items) == 0 || row == nil {
		return nil, -1
	}

	idx := row.Metadata[ItemIndexMetaKey].(int)

	return &items[idx], idx
}

// updateTable processes the common response format from modulated-content
// mutations (Sets & Resets), which return updates to columns, rows, and virtual
// rows. It appropriately refreshes the internal render-cache when necessary.
func (t *ItemsTable) updateTable(columns []table.Column, rows []table.Row, virt []table.Row) {
	// return on noop
	if columns == nil && rows == nil && virt == nil {
		return
	}

	// always refresh the cache after updates
	defer func() {
		t.refreshCache()
	}()

	// always apply sorting
	rows = t.sortRows(u.Ternary(columns, t.table.Columns(), columns != nil), rows)

	switch {
	case columns == nil && rows == nil: // no update
	case columns != nil && rows != nil: // update both
		t.table.SetContent(columns, rows)
	case columns == nil: // update only rows
		t.table.SetRows(rows)
	default: // update only columns
		t.table.SetColumns(columns)
	}

	if virt == nil {
		return
	}
	t.table.SetVirtualRows(virt)
}
