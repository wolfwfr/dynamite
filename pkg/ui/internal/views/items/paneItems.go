package itemselection

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/x/ansi"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

var tableInfoBox = lipgloss.NewStyle().
	Height(2).
	Padding(0, 1, 1, 1).
	Foreground(lipgloss.Color("#878787"))

type previewFormat int
type queryMode int

const (
	YAMLformat previewFormat = iota
	JSONformat
)

type SessionData struct {
	queryMode   messages.ItemsQueryMode
	chosenIndex *string
}

type ItemSelectionPane struct {
	// top-level context
	ctx context.Context

	spinner struct {
		active bool
		model  spinner.Model
		text   string
	}

	// standard timeout
	stdTO time.Duration

	// shared config
	config *appconfig.Config

	// error
	err error

	// view window
	window struct {
		width  int
		height int
	}

	// fuzzy finding
	search *search.SearchBox

	// key map
	KeyMap *ItemPaneKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	content *table.Model

	// sessions (per table ARN)
	sessions map[string]SessionData

	// query & scan parameters
	queryMode messages.ItemsQueryMode

	tableIndex struct {
		activeIndex    *string
		indexItemCount int64
	}

	scanLimit  int
	queryLimit int

	items types.Items

	// item filtering collects settings related to item filtering
	itemfiltering struct {
		items   []int // indices referring to items
		enabled bool
	}

	// column visibility collects settings related to column visibillity
	columnVisibility struct {
		enabled   bool
		inVisible map[string]struct{}
	}

	// column sorting collects settings related to column sorting
	columnSorting struct {
		SortingOn string
		Ascending bool // if false, descending
		Enabled   bool
	}

	lastPreviewItem int                   // index
	lastPreviewMsg  *messages.PreviewItem // prevents preview message looping
	pageKey         map[string]dynamotypes.AttributeValue
	pageCancel      func()
	paging          bool

	// keysComplete represents a unique set of dynamo-db item keys that
	// exhaustively cover all keys in the currently paged set of items
	keysComplete []string

	selectedTable types.DescribeTableResponse

	previewFormat previewFormat

	// specifies whether the first page has been loaded
	initialised bool
}

type itemsPaneOption func(p *ItemSelectionPane)

// withItemsPaneKeys
func withItemsPaneKeys(keys keymaps.AdditionalKeys) itemsPaneOption {
	return func(t *ItemSelectionPane) {
		t.AddKeyMap = keys
	}
}

func newItemSelectionPane(ctx context.Context, config *appconfig.Config, opts ...itemsPaneOption) *ItemSelectionPane {
	p := &ItemSelectionPane{
		ctx:           ctx,
		config:        config,
		stdTO:         5 * time.Second,
		KeyMap:        DefaultItemPaneKeyMap(),
		sessions:      map[string]SessionData{},
		queryMode:     messages.ScanMode,
		previewFormat: JSONformat,
		scanLimit:     10,
		queryLimit:    10,
		pageCancel:    func() {}, // init as noop
	}

	{ // contents table
		t := table.New(
			table.WithFocused(true),
			table.WithDynamicColumnWidth(false),
		)
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		p.content = t
	}

	{ // spinner
		sp := spinner.New()
		sp.Spinner = spinner.Dot
		sp.Style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			PaddingLeft(1)
		p.spinner.model = sp
		p.spinner.text = "obtaining next page..."
	}

	{ // search box
		p.search = search.NewSearchBox(
			search.SearchCallbacks{
				ToSearch: func() []string {
					return table.Rows(p.content.Rows()).ToStrings()
				},
				EmptyInput: func() tea.Cmd {
					p.itemfiltering.enabled = false
					p.itemfiltering.items = make([]int, 0)
					p.content.ResetVirtualRows()
					p.KeyMap.ColSort.SetEnabled(true)
					return p.MaybePreviewItem(true)
				},
				Results: func(results []search.FilteredItem) tea.Cmd {
					p.itemfiltering.enabled = true
					p.itemfiltering.items = make([]int, len(results))
					rows := p.content.Rows()
					filtered := make([]table.Row, len(results))
					for i, match := range results {
						filtered[i] = rows[match.Index]
						p.itemfiltering.items[i] = match.Index
					}
					p.content.SetVirtualRows(filtered)
					return nil
				},
				Reset: func(searchHeight int) tea.Cmd {
					p.itemfiltering.enabled = false
					p.itemfiltering.items = make([]int, 0)
					p.content.ResetVirtualRows()
					p.updateSize()
					p.KeyMap.ColSort.SetEnabled(true)
					return p.MaybePreviewItem(true)
				},
				SearchBoxOpens: func(searchHeight int) tea.Cmd {
					p.KeyMap.ColSort.SetEnabled(false)
					p.updateSize()
					return nil
				},
			},
		)
	}

	for _, o := range opts {
		o(p)
	}

	if !keymaps.UniqueKeyMaps(p.KeyMap.ShortHelp(), p.AddKeyMap.Bindings()) {
		panic("overlapping keymaps!")
	}

	return p
}

func (m *ItemSelectionPane) cleanSlate() {
	m.err = nil
}

func (m *ItemSelectionPane) activateSpinner() tea.Cmd {
	m.spinner.active = true
	m.updateSize()
	return m.spinner.model.Tick
}

func (m *ItemSelectionPane) deactivateSpinner() {
	m.spinner.active = false
	m.updateSize()
}

func (m *ItemSelectionPane) Init() tea.Cmd {
	m.softReset()
	return nil
}

// softReset initalises stateful parameters except for sessions
func (m *ItemSelectionPane) softReset() {
	// cancel any lingering calls
	m.pageCancel()

	// clean up content
	m.content.ResetVirtualRows()
	m.content.SetCursor(0)
	m.initialised = false

	m.resetQueryParameters() // must come first to reinitialise items in state (which may be used for updating content in other functions)
	m.resetColumnVisibility()
	m.resetColumnSorting()
}

func (m *ItemSelectionPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	cmds := []tea.Cmd{}
	_, isSelect := msg.(messages.SelectTable)
	_, isToggleFmt := msg.(messages.ToggleJSONYAML)
	_, isTick := msg.(spinner.TickMsg)
	_, isColVis := msg.(messages.ColumnVisibilityUpdate)
	_, isColSort := msg.(messages.ColumnSortingUpdate)
	_, isColSortRes := msg.(messages.ColumnSortingReset)
	_, isPreview := msg.(messages.PreviewItem)

	excludeSearch := isSelect || isToggleFmt || isTick || isColVis || isColSort || isColSortRes || isPreview

	if search.IsSearchBoxMessage(msg) || (!excludeSearch && m.search.IsFocused()) {
		cmds = append(cmds, m.search.Update(msg))
	} else {
		cmds = append(cmds, m.handleNavigation(msg))
	}
	cmds = append(cmds, m.MaybePreviewItem(false))
	return tea.Batch(cmds...)
}

// handleNavigation handles events when search is not active.
func (m *ItemSelectionPane) handleNavigation(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Search):
			cmds = append(cmds, m.search.OpenSearchBox())
		case key.Matches(msg, m.KeyMap.Esc):
			if m.search.IsEnabled() {
				return m.search.Reset()
			} else {
				return m.escape()
			}
		case key.Matches(msg, m.KeyMap.ChCols):
			m.content.SetDynamicColumnWidth(!m.content.DynamicColumnWidth())
		case key.Matches(msg, m.KeyMap.Zoom):
			return m.Zoom()
		case key.Matches(msg, m.KeyMap.ToggleFmt):
			return m.ToggleJSONYAMLFormat()
		case key.Matches(msg, m.KeyMap.Query):
			return m.enableQueryMode()
		case key.Matches(msg, m.KeyMap.Scan):
			return m.enableScanMode()
		case key.Matches(msg, m.KeyMap.ScanParameters):
			return m.ToggleScanParametersDialog()
		case key.Matches(msg, m.KeyMap.Copy):
			return m.copy()
		case key.Matches(msg, m.KeyMap.ColVis):
			return m.toggleColumnVsibilityDialog(msg)
		case key.Matches(msg, m.KeyMap.ColSort):
			return m.toggleColumnSortingDialog(msg)
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	case messages.PreviewItem:
		m.lastPreviewMsg = &msg
		return nil
	case messages.SelectTable:
		return m.selectTable(msg.TableName, msg.TableDetails)
	case messages.ToggleJSONYAML:
		return m.ToggleJSONYAMLFormat()
	case messages.ColumnVisibilityUpdate:
		return m.UpdateColumnVisibility(msg)
	case messages.ColumnSortingUpdate:
		return m.UpdateColumnSorting(msg)
	case messages.ScanIndexChanged:
		return m.ChangeScanIndex(msg)
	case messages.ScanPageReady:
		return m.ProcessScanPage(msg)
	case messages.ColumnSortingReset:
		m.resetColumnSorting()
		return nil
	case spinner.TickMsg:
		if !m.spinner.active {
			return nil
		}
		var cmd tea.Cmd
		m.spinner.model, cmd = m.spinner.model.Update(msg)
		return cmd
	}
	cmds = append(cmds, m.content.Update(msg))
	// paginate when not filtering and at end of content
	if !m.itemfiltering.enabled && m.content.ViewAtEnd() {
		cmds = append(cmds, m.PageNext(false))
	}
	return tea.Batch(cmds...)
}

func (m *ItemSelectionPane) PageNext(init bool) tea.Cmd {
	// don't page when at end of paging and not the initialising call
	if (len(m.pageKey) == 0 && !init) || m.paging {
		return nil
	}
	m.paging = true
	spinnerCmd := m.activateSpinner()
	mode := m.queryMode
	table := m.selectedTable
	key := m.pageKey
	idx := m.tableIndex.activeIndex
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	m.pageCancel = cc
	pageCmd := func() tea.Msg {
		defer cc()
		switch mode {
		case messages.QueryMode:
			panic("not supported yet")
		case messages.ScanMode:
			scan, err := dynamodb.ScanTable(m.config.Client, ctx, *table.TableName, types.ScanParameters{
				KeyDetails:       m.selectedTable.AttributeDefinitions,
				IndexName:        idx,
				KeySchema:        keysFromIndex(idx, table),
				Limit:            m.scanLimit,
				LastEvaluatedKey: key,
			})
			return messages.ScanPageReady{
				Table:    table,
				Index:    idx,
				Response: scan,
				Err:      err,
			}
		}
		return nil
	}
	return tea.Batch(pageCmd, spinnerCmd)
}

func (m *ItemSelectionPane) ToggleJSONYAMLFormat() tea.Cmd {
	m.previewFormat += 1
	if m.previewFormat > JSONformat {
		m.previewFormat = YAMLformat
	}
	return m.MaybePreviewItem(true)
}

// force is used on new pane initialization because lastPreviewItem could be 0
func (m *ItemSelectionPane) MaybePreviewItem(force bool) tea.Cmd {
	// render empty preview when no items or no filter results
	if m.initialised && len(m.items.Raw) == 0 || m.itemfiltering.enabled && len(m.itemfiltering.items) == 0 {
		if m.lastPreviewMsg != nil && m.lastPreviewMsg.Item == "" { // prevent looping
			return nil
		}
		return func() tea.Msg {
			return messages.PreviewItem{
				Item: "",
			}
		}

	}
	idx := m.content.Cursor()
	if len(m.itemfiltering.items) > 0 { // cursor refers to filtered items
		idx = m.itemfiltering.items[idx]
	}
	// if preview was already instructed to preview this item, skip
	if idx == m.lastPreviewItem && !force {
		return nil
	}
	m.lastPreviewItem = idx
	var item string
	switch m.previewFormat {
	case JSONformat:
		item = m.items.JSON[idx]
	case YAMLformat:
		item = m.items.YAML[idx]
	}
	return func() tea.Msg {
		return messages.PreviewItem{
			Item: item,
		}
	}
}

func (m *ItemSelectionPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleItemSelectionPane{}
	}
}

func (m *ItemSelectionPane) ProcessScanPage(msg messages.ScanPageReady) tea.Cmd {
	defer func() { m.deactivateSpinner() }()

	if msg.Err != nil {
		m.err = msg.Err // TODO: allow user to exit error message state
	}

	scan := msg.Response
	details := m.selectedTable

	if m.selectedTable.TableArn != msg.Table.TableArn || m.tableIndex.activeIndex != msg.Index { // expired
		return nil
	}

	m.appendItems(scan.Items)
	m.pageKey = scan.LastEvaluatedKey

	if len(scan.Items.TableKeys) > 0 {
		// set columns
		_, rang := primaryKeysFromSchema(keysFromIndex(m.tableIndex.activeIndex, details))
		completeKeys := compileCompleteKeys(scan.Items.TableKeys, m.keysComplete, rang != nil)
		defer func() { m.keysComplete = completeKeys }()

		noColumnUpdate := slices.Equal(m.keysComplete, completeKeys)
		columnUpdate := !noColumnUpdate
		appendOnly := noColumnUpdate && !m.columnSorting.Enabled

		switch {
		case columnUpdate: // update columns & ALL rows
			cols := m.assembleColumns(completeKeys)
			rows := parseRows(completeKeys, m.items.TableKeys)
			m.content.SetContent(cols, m.sortRows(rows))
		case appendOnly: // update with  new rows (append)
			rows := parseRows(completeKeys, scan.Items.TableKeys)
			m.content.AppendRows(m.sortRows(rows))
		default: // update ALL rows but no columns
			rows := parseRows(completeKeys, m.items.TableKeys)
			m.content.SetRows(m.sortRows(rows))
		}
	}
	m.paging = false
	m.initialised = true
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) sortRows(rows []table.Row) []table.Row {
	if !m.columnSorting.Enabled || m.columnSorting.SortingOn == "" || len(rows) == 0 {
		return rows
	}
	cols := m.content.Columns()
	colsS := make([]string, len(cols))
	for i, c := range cols {
		colsS[i] = c.Title
	}
	idx := u.Find(colsS, m.columnSorting.SortingOn)
	if idx < 0 {
		return rows
	}

	// determine field-type
	var field string
	for _, r := range rows {
		if r[idx] != "" {
			field = r[idx]
			break
		}
	}
	_, errInt := strconv.ParseInt(field, 10, 64)
	_, errFloat := strconv.ParseFloat(field, 64)

	// choose the appropriate sorting function
	var sortFunc func(a, b table.Row) int
	switch {
	// NOTE: assumes that float fields always contain decimal point
	case errFloat == nil:
		sortFunc = func(a, b table.Row) int {
			aI, _ := strconv.ParseFloat(a[idx], 64)
			bI, _ := strconv.ParseFloat(b[idx], 64)
			check := ternary(aI < bI, aI > bI, m.columnSorting.Ascending)
			return ternary(-1, 1, check)
		}
	case errInt == nil:
		sortFunc = func(a, b table.Row) int {
			aI, _ := strconv.ParseInt(a[idx], 10, 64)
			bI, _ := strconv.ParseInt(b[idx], 10, 64)
			check := ternary(aI < bI, aI > bI, m.columnSorting.Ascending)
			return ternary(-1, 1, check)
		}
	default:
		sortFunc = func(a, b table.Row) int {
			s := []string{a[idx], b[idx]}
			slices.Sort(s)
			check := ternary(s[0] == a[idx], s[1] == a[idx], m.columnSorting.Ascending)
			return ternary(-1, 1, check)
		}
	}

	// apply sorting function on slice backed by new array
	sorted := make([]table.Row, len(rows))
	copy(sorted, rows)
	slices.SortFunc(sorted, sortFunc)

	return sorted
}

// selectTable processes the select-table message, which indicates that the
// item-selection-view is opened because a table has been selected. It will
// default to scanning the first page of items.
func (m *ItemSelectionPane) selectTable(tableName string, details types.DescribeTableResponse) tea.Cmd {
	if session, remembered := m.sessions[*details.TableArn]; remembered {
		// restore session parameters
		m.queryMode = session.queryMode
		m.tableIndex.activeIndex = session.chosenIndex
	} else {
		// defaults on newly opened table
		m.queryMode = messages.ScanMode
		m.tableIndex.activeIndex = nil
	}
	// resetting state
	m.cleanSlate()
	m.content.ResetVirtualRows()
	m.selectedTable = details

	return m.PageNext(true)
}

// compileCompleteKeys takes a table of key-value pairs, observes all keys and
// compiles a complete, in-order list of all unique key observed.
// This ensures that when individual table rows have keys missing, the final
// result still contains these keys when they are present in other rows in the
// specified table.
func compileCompleteKeys(table [][]types.KeyValue, existing []string, hasRangeKey bool) []string {
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

	sortLenOffset := ternary(2, 1, hasRangeKey)
	toSort := make([]string, len(res)-sortLenOffset)
	copy(toSort, res[sortLenOffset:])
	slices.Sort(toSort)
	copy(res[sortLenOffset:], toSort)

	return res
}

func (m *ItemSelectionPane) applySize(height, width int) {
	m.window.height = height
	m.window.width = width
	m.updateSize()
}

// updateSize updates dimensions of the pane's contents based on the current
// window dimensions.
func (m *ItemSelectionPane) updateSize() {
	h, w := m.window.height, m.window.width

	searchBoxH := ternary(m.search.GetHeight(), 0, m.search.IsEnabled())
	tableInfoH := tableInfoBox.GetHeight()
	m.window.height = h
	m.window.width = w
	m.content.SetHeight(h - searchBoxH - tableInfoH - ternary(1, 0, m.spinner.active))
	m.content.SetWidth(w)
	m.search.SetWidth(w)
	m.queryLimit = h
	m.scanLimit = h
}

// TODO: merge Init, reset, & cleanslate
func (m *ItemSelectionPane) resetQueryParameters() {
	m.initialised = false
	m.paging = false
	m.keysComplete = []string{}
	m.queryMode = messages.ScanMode
	m.pageKey = nil
	m.items = types.Items{}
	m.itemfiltering.items = []int{}
	m.lastPreviewItem = 0
	m.lastPreviewMsg = nil
	m.tableIndex.activeIndex = nil
	m.tableIndex.indexItemCount = -1
}

func (m *ItemSelectionPane) resetColumnVisibility() {
	m.columnVisibility.enabled = false
	m.columnVisibility.inVisible = make(map[string]struct{}, 0)
}

func (m *ItemSelectionPane) handleResetColumnSortingMessage(msg messages.ColumnSortingReset) {
	if msg.TableARN != u.IfNotNil(m.selectedTable.TableArn, "") { // expired
		return
	}
	m.resetColumnSorting()
}

// resetColumnSorting re-initialises column-sorting associated state parameters
// and restores the columns and rows based on the items stored in state.
func (m *ItemSelectionPane) resetColumnSorting() {
	m.columnSorting.Ascending = true
	m.columnSorting.SortingOn = ""
	m.columnSorting.Enabled = false

	// reassemble cols
	cols := m.assembleColumns(m.keysComplete)

	// reassemble rows
	rows := parseRows(m.keysComplete, m.items.TableKeys)

	// set content
	m.content.SetContent(cols, rows)
}

func (m *ItemSelectionPane) escape() tea.Cmd {
	// cancel pending calls
	m.pageCancel()

	// store session data
	if arn := m.selectedTable.TableArn; arn != nil {
		m.sessions[*arn] = SessionData{
			queryMode:   m.queryMode,
			chosenIndex: m.tableIndex.activeIndex,
		}
	}

	// clean up state
	m.softReset()

	// switch to previous view
	switchView := func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Item_selection,
			NewView: messages.Table_selection,
		}
	}

	// clean up preview window
	resetPreview := func() tea.Msg {
		return messages.PreviewItem{
			Item: "",
		}
	}

	return tea.Batch(resetPreview, switchView)
}

func (m *ItemSelectionPane) UpdateColumnVisibility(msg messages.ColumnVisibilityUpdate) tea.Cmd {
	if msg.TableARN != u.IfNotNil(m.selectedTable.TableArn, "") { // expired
		return nil
	}
	cols := m.content.Columns()
	if len(cols) != len(msg.AllColumns) {
		// TODO: better handling of new columns appearing in view
		m.resetColumnVisibility()
		return nil
	}
	m.columnVisibility.enabled = true
	for i, c := range msg.AllColumns {
		if !msg.Visible[i] {
			m.columnVisibility.inVisible[c] = struct{}{}
		} else {
			delete(m.columnVisibility.inVisible, c)
		}
	}
	for i, c := range cols {
		_, isInvisible := m.columnVisibility.inVisible[c.Title]
		cols[i].InVisible = isInvisible
	}
	m.content.SetColumns(cols)

	if len(m.columnVisibility.inVisible) == 0 {
		m.columnVisibility.enabled = false
		return nil
	}

	return nil
}

// toggle column visibility dialog & provide current state (in case dialog opens)
func (m *ItemSelectionPane) toggleColumnVsibilityDialog(msg tea.Msg) tea.Cmd {
	cols := m.content.Columns()
	vis := m.columnVisibility.inVisible

	colsS := make([]string, 0, len(cols))
	visB := make([]bool, 0, len(cols))
	for _, c := range cols {
		colsS = append(colsS, c.Title)
		_, isInVisible := vis[c.Title]
		visB = append(visB, !isInVisible)
	}
	arn := u.IfNotNil(m.selectedTable.TableArn, "")
	toggle := func() tea.Msg {
		return messages.ToggleColumnVisibility{}
	}
	state := func() tea.Msg {
		msg := messages.InitColumnVisibility{}
		msg.TableARN = arn
		msg.AllColumns = colsS
		msg.Visible = visB
		return msg
	}
	return tea.Batch(toggle, state)
}

func (m *ItemSelectionPane) UpdateColumnSorting(msg messages.ColumnSortingUpdate) tea.Cmd {
	if msg.TableARN != u.IfNotNil(m.selectedTable.TableArn, "") { // expired
		return nil
	}
	cols := m.content.Columns()
	if len(cols) != len(msg.AllColumns) {
		// TODO: better handling of new columns appearing in view
		m.resetColumnSorting()
		return nil
	}

	// update panel state
	m.columnSorting.Enabled = true
	m.columnSorting.Ascending = msg.Ascending
	m.columnSorting.SortingOn = msg.SortingOn

	// prepare table column update
	for i, c := range cols {
		c.Suffix = m.getColumnSuffix(c.Title)
		cols[i] = c
	}

	// update table columns
	m.content.SetColumns(cols)

	// sort table rows
	rows := m.sortRows(m.content.Rows())

	// update table rows
	m.content.SetRows(rows)

	return nil
}

// toggle column sorting dialog & provide current state (in case dialog opens)
func (m *ItemSelectionPane) toggleColumnSortingDialog(msg tea.Msg) tea.Cmd {
	cols := m.content.Columns()
	colsS := make([]string, 0, len(cols))
	for _, c := range cols {
		colsS = append(colsS, c.Title)
	}
	sorting := m.columnSorting.SortingOn
	ascending := m.columnSorting.Ascending
	arn := u.IfNotNil(m.selectedTable.TableArn, "")
	toggle := func() tea.Msg {
		return messages.ToggleColumnSorting{}
	}
	state := func() tea.Msg {
		msg := messages.InitColumnSorting{}
		msg.TableARN = arn
		msg.AllColumns = colsS
		msg.SortingOn = sorting
		msg.Ascending = ascending
		return msg
	}
	return tea.Batch(toggle, state)
}

func (m *ItemSelectionPane) ChangeScanIndex(msg messages.ScanIndexChanged) tea.Cmd {
	if u.IfNotNil(m.selectedTable.TableArn, "") != msg.TableARN { // expired
		return nil
	}
	m.tableIndex.activeIndex = u.Ternary(&msg.IndexName, nil, msg.IndexName != "")
	m.tableIndex.indexItemCount = u.IfNotNil(m.selectedTable.ItemCount, 0)
	if m.tableIndex.activeIndex != nil {
		m.tableIndex.indexItemCount = indexCountFromTable(*m.tableIndex.activeIndex, m.selectedTable)
	}
	m.Init()
	return m.PageNext(true)
}

func (m *ItemSelectionPane) copy() tea.Cmd {
	return func() tea.Msg {
		return messages.CopyItem{}
	}
}

type dialog interface {
	View() string
	Width() int
}

func (m *ItemSelectionPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	info := m.renderTableInfo()
	content := m.content.View()
	content = ternary(content, m.noContentMessage(), !emptyContent(content))
	rendering := []string{info, content, m.search.View()}
	if m.spinner.active {
		rendering = slices.Insert(rendering, 2, fmt.Sprintf("%s %s", m.spinner.model.View(), m.spinner.text))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendering...)
}

func emptyContent(content string) bool {
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "\t", "")
	return content == ""
}

func (m *ItemSelectionPane) noContentMessage() string {
	if m.paging {
		return ""
	}
	s := strings.Builder{}
	fmt.Fprintf(&s, "==================================================\n")
	fmt.Fprintf(&s, "                    NO CONTENT                    \n")
	fmt.Fprintf(&s, "==================================================\n")
	return s.String()
}

func (m *ItemSelectionPane) renderTableInfo() string {
	width := m.window.width / 2
	leftAligned := lipgloss.NewStyle().Width(width).Align(lipgloss.Left)
	rightAligned := lipgloss.NewStyle().Width(width).Align(lipgloss.Right)

	// table name
	name := u.IfNotNil(m.selectedTable.TableName, "")

	// determine item count & index name
	count := m.tableIndex.indexItemCount
	indexName := u.IfNotNil(m.tableIndex.activeIndex, "")

	right := fmt.Sprintf("Count: %d/%d", len(m.content.VisualRows()), count)
	right = ansi.Truncate(right, width, "…")

	left := fmt.Sprintf("Table: %s%s", name, u.Ternary(" / Index: "+indexName, "", indexName != ""))
	left = ansi.Truncate(left, u.Ternary(width-2, width, strings.HasPrefix(right, "…")), "…")

	return tableInfoBox.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		leftAligned.Render(left),
		rightAligned.Render(right),
	))
}

func (m *ItemSelectionPane) appendItems(newItems types.Items) {
	// JSON
	j := make([]string, len(m.items.JSON)+len(newItems.JSON))
	copy(j[:len(m.items.JSON)], m.items.JSON)
	copy(j[len(m.items.JSON):], newItems.JSON)
	m.items.JSON = j
	// YAML
	y := make([]string, len(m.items.YAML)+len(newItems.YAML))
	copy(y[:len(m.items.YAML)], m.items.YAML)
	copy(y[len(m.items.YAML):], newItems.YAML)
	m.items.YAML = y
	// RAW
	r := make([]map[string]dynamotypes.AttributeValue, len(m.items.Raw)+len(newItems.Raw))
	copy(r[:len(m.items.Raw)], m.items.Raw)
	copy(r[len(m.items.Raw):], newItems.Raw)
	m.items.Raw = r
	// KEYS
	k := make([][]types.KeyValue, len(m.items.TableKeys)+len(newItems.TableKeys))
	copy(k[:len(m.items.TableKeys)], m.items.TableKeys)
	copy(k[len(m.items.TableKeys):], newItems.TableKeys)
	m.items.TableKeys = k
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

func ternary[T any](first T, second T, cond bool) T {
	if cond {
		return first
	}
	return second
}

func primaryKeysFromSchema(s []dynamotypes.KeySchemaElement) (hash string, rang *string) {
	for _, e := range s {
		if e.KeyType == dynamotypes.KeyTypeHash {
			hash = *e.AttributeName
		} else {
			rang = e.AttributeName
		}
	}
	return
}

func keysFromIndex(idx *string, details types.DescribeTableResponse) []dynamotypes.KeySchemaElement {
	if idx == nil {
		return details.KeySchema
	}
	for _, g := range details.GlobalSecondaryIndexes {
		if *g.IndexName == *idx {
			return g.KeySchema
		}
	}
	for _, l := range details.LocalSecondaryIndexes {
		if *l.IndexName == *idx {
			return l.KeySchema
		}
	}
	return details.KeySchema
}

// assembleColumns returns a set of table columns that incorporates modulations
// based on the item-selection-pane state, such as the state of column
// visibility and sorting.
func (m *ItemSelectionPane) assembleColumns(allColumnTitles []string) []table.Column {
	cols := make([]table.Column, len(allColumnTitles))

	for i, title := range allColumnTitles {
		col := table.Column{Title: title, Width: clamp(len(title), 16, 32)}

		// visibility
		_, isInvisible := m.columnVisibility.inVisible[title]
		col.InVisible = m.columnVisibility.enabled && isInvisible

		// suffix
		col.Suffix = m.getColumnSuffix(title)

		// insert
		cols[i] = col
	}
	return cols
}

func (m *ItemSelectionPane) getColumnSuffix(colTitle string) string {
	if m.columnSorting.Enabled && m.columnSorting.SortingOn == colTitle {
		return fmt.Sprintf(" (%s)", ternary("↑", "↓", m.columnSorting.Ascending))
	}
	return ""
}

func parseRows(cols []string, tableKeys [][]types.KeyValue) []table.Row {
	rows := make([]table.Row, len(tableKeys))
	for i, k := range tableKeys {
		row := make([]string, len(cols))
		var x int
		for j, key := range cols {
			if key == k[x].Key { // matching key
				row[j] = k[x].Value
				x = min(len(k)-1, x+1)
			} else { // no matching key
				row[j] = ""
			}
		}
		rows[i] = row
	}
	return rows
}

func indexCountFromTable(indexName string, tableDetails types.DescribeTableResponse) int64 {
	for _, g := range tableDetails.GlobalSecondaryIndexes {
		if *g.IndexName == indexName {
			return *g.ItemCount
		}
	}

	for _, l := range tableDetails.LocalSecondaryIndexes {
		if *l.IndexName == indexName {
			return *l.ItemCount
		}
	}
	return -1
}
