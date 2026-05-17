package itemselection

// TODO: private/public field consistency (entire project)
import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/charmbracelet/x/ansi"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/modulator"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

var tableInfoBox = lipgloss.NewStyle().
	Height(2).
	Padding(0, 1, 1, 1).
	Foreground(commonstyles.SubtleColour2)

type previewFormat int

const (
	YAMLformat previewFormat = iota
	JSONformat
)

const itemIndexMetaKey = "item_index"

type SessionData struct {
	queryMode   messages.ItemsQueryMode
	queryParams struct {
		index                *string
		hashKeyValue         string
		rangeKeyValue1       *string
		rangeKeyValue2       *string
		rangeKeyOperator     messages.QueryOperator
		rangeOrderDescending bool
	}
	scanParams struct {
		index *string
	}
}

type TableStyles struct {
	SelectedBackground    color.Color
	SearchMatchBackground color.Color
}

type ItemSelectionPane struct {
	// top-level context
	ctx context.Context

	// styles
	styles struct {
		Table TableStyles
	}

	// spinner
	spinner struct {
		active bool
		model  spinner.Model
		text   string
	}

	// standard timeout
	stdTO time.Duration

	// shared config
	config *appconfig.Config

	// dynamo-db adapter for UI purposes
	dynamodbClient dynamodbClient

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

	// the underlying table
	content *table.Model

	// modulated-content manages state & modulation of mutable table contents,
	// e.g. column visibility, row field sorting, etc.
	modulatedContent *modulator.Modulator

	// sessions (per table ARN)
	sessions map[string]SessionData

	// query & scan parameters
	queryMode messages.ItemsQueryMode

	// limits for dynamo-db operations
	scanLimit  int
	queryLimit int

	// currently active dynamo-db index
	tableIndex struct {
		activeIndex    *string
		indexItemCount int64
	}

	// currently active scan parameters
	scanParameters struct {
		index *string
	}

	// currently active query parameters
	// TODO: name collision with reset function
	queryParameters struct {
		index                *string
		hashKeyValue         string
		rangeKeyValue1       *string
		rangeKeyValue2       *string
		rangeKeyOperator     messages.QueryOperator
		rangeOrderDescending bool
	}

	// render-cache caches row-fields rendered by the table's field-delegate
	renderCache map[string]string

	lastPreviewItem int                   // index
	lastPreviewMsg  *messages.PreviewItem // prevents preview message looping
	pageKey         map[string]dynamotypes.AttributeValue
	pageCancel      func()
	paging          bool
	pagingDisabled  bool

	// keysComplete represents a unique set of dynamo-db item keys that
	// exhaustively cover all keys in the currently paged set of items
	keysComplete []string

	// the currently active table
	selectedTable types.DescribeTableResponse

	// json/yaml format for preview
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
		ctx:            ctx,
		renderCache:    map[string]string{},
		config:         config,
		dynamodbClient: dynamodb.NewAdapter(),
		stdTO:          30 * time.Second,
		KeyMap:         DefaultItemPaneKeyMap(),
		sessions:       map[string]SessionData{},
		queryMode:      messages.ScanMode,
		previewFormat:  JSONformat,
		scanLimit:      10,
		queryLimit:     10,
		pageCancel:     func() {}, // init as noop
	}

	{ // contents table
		t := table.New(
			table.WithFocused(true),
			table.WithDynamicColumnWidth(false),
			table.WithFieldDelegate(p.TableRowFieldDelegate),
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

		p.content = t
		p.styles.Table = st
		p.modulatedContent = modulator.NewModulator(t)
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
				ToSearch: func(col string) []string {
					cols := p.content.Columns()
					idx := findColumnByTitle(cols, col)
					return extractColumnFromRows(p.content.Rows(), idx)
				},
				EmptyInput: func() tea.Cmd {
					p.content.ResetVirtualRows()
					p.applyContentMods(p.modulatedContent.ResetSearch())
					p.refreshCache()
					return p.MaybePreviewItem(true)
				},
				Results: func(col string, results []search.FilteredItem) tea.Cmd {
					p.applyContentMods(p.modulatedContent.SetSearchResults(col, results))
					return nil
				},
				Reset: func(searchHeight int) tea.Cmd {
					p.content.ResetVirtualRows()
					p.applyContentMods(p.modulatedContent.ResetSearch())
					p.updateSize()
					p.refreshCache()
					p.KeyMap.ColSort.SetEnabled(true)
					p.pagingDisabled = false
					return p.MaybePreviewItem(true)
				},
				SearchBoxOpens: func(searchHeight int) tea.Cmd {
					p.pagingDisabled = true
					p.applyContentMods(p.modulatedContent.ResetColumnSorting())
					p.updateSize()
					return nil
				},
			},
		)
		p.search.SetDivider("=")
		p.search.SetPlaceHolder("<column_name>=<search_input>")
	}

	for _, o := range opts {
		o(p)
	}

	if !keymaps.UniqueKeyMaps(p.KeyMap.ShortHelp(), p.AddKeyMap.Bindings()) {
		panic("overlapping keymaps!")
	}

	return p
}

// applyContentMods processes the common response format from modulated-content
// mutations (Sets & Resets), which return updates to columns, rows, and virtual
// rows. It appropriately refreshes the internal render-cache when necessary.
func (m *ItemSelectionPane) applyContentMods(columns []table.Column, rows []table.Row, virtual []table.Row) {
	switch {
	case columns == nil && rows == nil: // no update
	case columns != nil && rows != nil: // update both
		m.content.SetContent(columns, rows)
		m.refreshCache()
	case columns == nil: // update only rows
		m.content.SetRows(rows)
		m.refreshCache()
	default: // update only columns
		m.content.SetColumns(columns)
	}

	if virtual == nil {
		return
	}
	m.content.SetVirtualRows(virtual)
	m.refreshCache()
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
	return m.softReset()
}

// softReset initalises stateful parameters except for sessions and the selected
// table
func (m *ItemSelectionPane) softReset() tea.Cmd {
	m.err = nil
	// cancel any lingering calls
	m.pageCancel()

	m.resetContents()
	cmd := m.resetQueryParameters() // must come first to reinitialise items in state (which may be used for updating content in other functions)
	m.resetKeyMap()

	// no need to process cols/rows this time, just reset parameters
	m.modulatedContent.ResetColumnVisibility()
	m.modulatedContent.ResetColumnSorting()

	// clear cache last!
	m.clearCache()
	return cmd
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
		case key.Matches(msg, m.KeyMap.Reload):
			return m.Reload()
		case key.Matches(msg, m.KeyMap.ChCols):
			m.content.SetDynamicColumnWidth(!m.content.DynamicColumnWidth())
		case key.Matches(msg, m.KeyMap.Zoom):
			return m.Zoom()
		case key.Matches(msg, m.KeyMap.ToggleFmt):
			return m.ToggleJSONYAMLFormat()
		case key.Matches(msg, m.KeyMap.Query):
			return m.enableQueryMode(false)
		case key.Matches(msg, m.KeyMap.Scan):
			return m.enableScanMode(false)
		case key.Matches(msg, m.KeyMap.ScanParameters):
			return m.ToggleScanParametersDialog()
		case key.Matches(msg, m.KeyMap.QueryParameters):
			return m.ToggleQueryParametersDialog()
		case key.Matches(msg, m.KeyMap.Copy):
			return m.copy()
		case key.Matches(msg, m.KeyMap.Browser):
			return m.openInBrowser(m.resolveBrowserURL())
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
	case messages.QueryParametersChanged:
		return m.ChangeQueryParameters(msg)
	case messages.PageReady:
		return m.ProcessPage(msg)
	case messages.ColumnSortingReset:
		return m.resetColumnSorting(msg)
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
	if !m.pagingDisabled && m.content.ViewAtEnd() {
		cmds = append(cmds, m.PageNext(false))
	}
	return tea.Batch(cmds...)
}

func (m *ItemSelectionPane) TableRowFieldDelegate(row table.Row, col table.Column, colIdx, rowIdx, colW, padL, padR int, selected bool) string {
	fullWidth := colW + padL + padR

	// obtain field in question
	field := row.Fields[colIdx].(modulator.EnrichedField)

	// fill up with padding if empty
	if field.Style == nil {
		st := lipgloss.NewStyle().PaddingRight(fullWidth)
		st = u.Ternary(st.Background(m.styles.Table.SelectedBackground), st, selected)
		return st.Render("")
	}

	style := *field.Style

	itemfiltering := m.modulatedContent.Itemfiltering

	// attempt to obtain cached value to prevent rerendering
	cachekey := fmt.Sprintf("%d-%d-%d", rowIdx, colIdx, colW)
	cachCond := !selected && (!itemfiltering.Enabled || itemfiltering.ColumnIndex != colIdx)
	cc, ok := m.renderCache[cachekey]
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
		style = style.SetBackgroundAll(m.styles.Table.SelectedBackground)
	}

	// ensure that row-index is correctly interpreted as pointing to 'actual'
	// rows or virtual rows. When the table content contains virtual rows, it
	// will always only render those virtual rows.
	applyingVirtualRows := len(m.content.VirtualRows()) > 0

	// override background styling for search matches
	if applyingVirtualRows && itemfiltering.Enabled && itemfiltering.ColumnIndex == colIdx {
		for _, idx := range itemfiltering.MatchedRunes[rowIdx] {
			runeStyle, _ := style.GetAt(idx)
			c := m.styles.Table.SearchMatchBackground
			if selected {
				c = lipgloss.Blend1D(10, c, m.styles.Table.SelectedBackground)[3]
			}
			style = style.Override(idx, runeStyle.Background(c))
		}
	}

	enforceWidth := lipgloss.NewStyle().Width(fullWidth).MaxWidth(fullWidth).Inline(true).Render
	res := enforceWidth(style.Render(field.RawValue))

	// cache when appropriate for improved performance
	if cachCond {
		m.renderCache[cachekey] = res
	}

	return res
}

func (m *ItemSelectionPane) PageNext(init bool) tea.Cmd {
	// don't page when at end of paging and not the initialising call
	if (len(m.pageKey) == 0 && !init) || m.paging {
		return nil
	}
	m.paging = true
	mode := m.queryMode
	table := m.selectedTable
	key := m.pageKey
	idx := m.tableIndex.activeIndex
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	m.pageCancel = cc
	client := m.config.Client
	scanLimit := m.scanLimit
	queryLimit := m.queryLimit
	hash := m.queryParameters.hashKeyValue
	rang1 := m.queryParameters.rangeKeyValue1
	rang2 := m.queryParameters.rangeKeyValue2
	rangOp := m.queryParameters.rangeKeyOperator
	rangOr := m.queryParameters.rangeOrderDescending
	pageCmd := func() tea.Msg {
		defer cc()
		switch mode {
		case messages.QueryMode:
			if hash == "" { // prevent impossible query
				return messages.PageReady{
					TableARN: u.IfNotNil(table.TableArn, ""),
					Index:    idx,
					Response: nil,
					Err:      nil,
				}
			}
			result, err := m.dynamodbClient.QueryTable(client, ctx, *table.TableName, types.QueryParameters{
				KeyDetails:       table.AttributeDefinitions,
				IndexName:        idx,
				KeySchema:        keysFromIndex(idx, table),
				HashKeyValue:     hash,
				RangeKeyValue1:   rang1,
				RangeKeyValue2:   rang2,
				RangeKeyOperator: parseRangeKeyOperator(rangOp),
				Descending:       rangOr,
				Limit:            queryLimit,
				LastEvaluatedKey: key,
			})
			return messages.PageReady{
				TableARN: u.IfNotNil(table.TableArn, ""),
				Index:    idx,
				Response: queryPageToPage(result),
				Err:      err,
			}
		case messages.ScanMode:
			result, err := m.dynamodbClient.ScanTable(client, ctx, *table.TableName, types.ScanParameters{
				KeyDetails:       table.AttributeDefinitions,
				IndexName:        idx,
				KeySchema:        keysFromIndex(idx, table),
				Limit:            scanLimit,
				LastEvaluatedKey: key,
			})
			return messages.PageReady{
				TableARN: u.IfNotNil(table.TableArn, ""),
				Index:    idx,
				Response: scanPageToPage(result),
				Err:      err,
			}
		}
		return nil
	}
	return tea.Batch(pageCmd, m.activateSpinner())
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
	if !m.initialised {
		return nil
	}
	sorting := m.modulatedContent.ColumnSorting
	itemfiltering := m.modulatedContent.Itemfiltering
	items := m.modulatedContent.Items
	// render empty preview when no items or no filter results
	if m.initialised && len(items.Raw) == 0 || itemfiltering.Enabled && len(itemfiltering.MatchedItems) == 0 {
		if m.lastPreviewMsg != nil && m.lastPreviewMsg.StyledItem == "" { // prevent looping
			return nil
		}
		return func() tea.Msg {
			return messages.PreviewItem{}
		}
	}

	idx := m.content.Cursor()
	if len(sorting.SortedItems) > 0 {
		idx = sorting.SortedItems[idx]
	} else if len(itemfiltering.MatchedItems) > 0 { // cursor refers to filtered items
		idx = itemfiltering.MatchedItems[idx]
	}
	// if preview was already instructed to preview this item, skip
	if idx == m.lastPreviewItem && !force {
		return nil
	}
	m.lastPreviewItem = idx
	var styled string
	var raw string
	switch m.previewFormat {
	case JSONformat:
		raw = items.JSON[idx]
		styled = items.JSONStyled[idx].Render(raw)
	case YAMLformat:
		raw = items.YAML[idx]
		styled = items.YAMLStyled[idx].Render(raw)
	}
	return func() tea.Msg {
		return messages.PreviewItem{
			StyledItem: styled,
			RawItem:    raw,
		}
	}
}

func (m *ItemSelectionPane) Reload() tea.Cmd {
	m.resetContents()
	return m.PageNext(true)
}

func (m *ItemSelectionPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleItemSelectionPane{}
	}
}

func (m *ItemSelectionPane) ProcessPage(msg messages.PageReady) tea.Cmd {
	defer func() { m.deactivateSpinner() }()

	if msg.Err != nil {
		m.err = msg.Err
	}

	page := msg.Response
	details := m.selectedTable

	if u.IfNotNil(m.selectedTable.TableArn, "") != msg.TableARN || m.tableIndex.activeIndex != msg.Index { // expired
		return nil
	}

	if page == nil {
		return nil
	}

	m.pageKey = page.LastEvaluatedKey

	_, rang := primaryKeysFromSchema(keysFromIndex(m.tableIndex.activeIndex, details))

	m.applyContentMods(m.modulatedContent.AddItems(page.Items, rang != nil))

	// always refresh cache to respect potential new sorting
	m.refreshCache()

	m.paging = false
	m.initialised = true
	return m.MaybePreviewItem(true)
}

// selectTable processes the select-table message, which indicates that the
// item-selection-view is opened because a table has been selected. It will
// default to scanning the first page of items.
func (m *ItemSelectionPane) selectTable(tableName string, details types.DescribeTableResponse) tea.Cmd {
	m.selectedTable = details
	var cmd tea.Cmd
	if session, remembered := m.sessions[*details.TableArn]; remembered {
		// restore session parameters
		m.scanParameters.index = session.scanParams.index
		m.queryParameters.index = session.queryParams.index
		m.queryParameters.hashKeyValue = session.queryParams.hashKeyValue
		m.queryParameters.rangeKeyValue1 = session.queryParams.rangeKeyValue1
		m.queryParameters.rangeKeyValue2 = session.queryParams.rangeKeyValue2
		m.queryParameters.rangeKeyOperator = session.queryParams.rangeKeyOperator
		m.queryParameters.rangeOrderDescending = session.queryParams.rangeOrderDescending
		switch session.queryMode {
		case messages.ScanMode:
			m.tableIndex.activeIndex = session.scanParams.index
			cmd = m.enableScanMode(true)
		case messages.QueryMode:
			m.tableIndex.activeIndex = session.queryParams.index
			cmd = m.enableQueryMode(true)
		}
		if m.tableIndex.activeIndex == nil {
			m.tableIndex.indexItemCount = *details.ItemCount
		} else {
			m.tableIndex.indexItemCount = indexCountFromTable(*m.tableIndex.activeIndex, details)
		}
	} else {
		// defaults on newly opened table
		m.tableIndex.activeIndex = nil
		m.tableIndex.indexItemCount = *details.ItemCount
		cmd = m.enableScanMode(true)
	}
	// resetting state
	m.applyContentMods(m.modulatedContent.ResetSearch())

	return cmd
}

func (m *ItemSelectionPane) resolveBrowserURL() string {
	selection := m.content.SelectedRow()
	if selection == nil || len(selection.Fields) == 0 || m.selectedTable.TableName == nil {
		return ""
	}

	var (
		region = m.config.Region
		// TODO: think about config workaround for when AWS would change URL
		weburl    = fmt.Sprintf("https://%s.console.aws.amazon.com/dynamodbv2/home", region)
		tableName = *m.selectedTable.TableName
		fields    = selection.Fields
	)
	_, r := primaryKeysFromSchema(keysFromIndex(m.tableIndex.activeIndex, m.selectedTable))

	paramkeys := []string{
		"region",
		"itemMode",
		"pk",
		"table",
	}

	// NOTE: dynamo-db uses path-escaping for query-values (e.g. '%20' for
	// ' ' (path-escaping) instead of '+' (query-escaping))
	paramVals := []string{
		fmt.Sprintf("%s#edit-item?", region),
		"2", // 1:create, 2:edit, 3:duplicate
		url.PathEscape(strings.Trim(fields[0].Value(), "\"")),
		url.PathEscape(tableName),
	}

	if r != nil {
		paramkeys = append(paramkeys, "sk")
		paramVals = append(paramVals, url.PathEscape(strings.Trim(fields[1].Value(), "\"")))
	}

	// manually parsing query parameters, because of the strange double query
	// parameter section in the dynamo-db url
	weburl = fmt.Sprintf("%s%s", weburl, u.Ternary("?", "", len(paramkeys) > 0))
	for i := range paramkeys {
		sep := u.Ternary("&", "", i > 1)
		weburl = fmt.Sprintf("%s%s%s=%s", weburl, sep, paramkeys[i], paramVals[i])
	}

	return weburl
}

func (m *ItemSelectionPane) openInBrowser(url string) tea.Cmd {
	if url == "" {
		return nil
	}

	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		return notifyError(err)
	}

	return nil
}

func notifyError(err error) tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleNotificationDialog{Error: err}
	}
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
	tableInfoH := lipgloss.Height(m.renderTableInfo())
	m.window.height = h
	m.window.width = w
	// TODO: fix the '1'; content prints one empty row beyond its allowed height
	m.content.SetHeight(h - 1 - searchBoxH - tableInfoH - ternary(1, 0, m.spinner.active))
	m.content.SetWidth(w)
	m.search.SetWidth(w)
	m.queryLimit = h
	m.scanLimit = h
}

func (m *ItemSelectionPane) resetKeyMap() {
	m.KeyMap.QueryParameters.SetEnabled(false)
	m.KeyMap.Query.SetEnabled(true)
	m.KeyMap.ScanParameters.SetEnabled(true)
	m.KeyMap.Scan.SetEnabled(false)
}

// clearCache completely removes any cached state
// Note that clearing of cache does not automatically imply that the table's
// rendered rows will be updated anew. Use refreshCache if this is your goal.
func (m *ItemSelectionPane) clearCache() {
	m.renderCache = map[string]string{}
}

// refreshCache clears the cache and then forces a rerender of rows
func (m *ItemSelectionPane) refreshCache() {
	m.clearCache()
	m.content.UpdateContent()
}

// reset contents resets any table modifications and resets the table contents
// to empty. It also cancels and resets paging and resets preview tracking.
func (m *ItemSelectionPane) resetContents() {
	m.err = nil
	m.resetPaging()
	m.initialised = false
	m.lastPreviewItem = 0
	m.lastPreviewMsg = nil

	// reset search
	m.search.Reset()

	// reset modulation parameters; no need to process table-content effects
	m.modulatedContent.Reset()

	// remove any table contents
	m.content.ResetVirtualRows()
	m.content.SetContent([]table.Column{}, []table.Row{})
	// TODO: change to clearCache instead, cleanup & test
	m.refreshCache()
}

// resetPaging resets any paging related parameters and calcels any lingering
// paging calls
func (m *ItemSelectionPane) resetPaging() {
	m.pageCancel()
	m.paging = false
	m.pagingDisabled = false
	m.pageKey = nil
}

// resetQueryParameters resets any parameters required for sanning or querying a
// dynamodb table
func (m *ItemSelectionPane) resetQueryParameters() tea.Cmd {
	var cmd tea.Cmd
	if m.queryMode != messages.ScanMode {
		cmd = func() tea.Msg {
			return messages.SwitchQueryMode{
				OldMode: m.queryMode,
				NewMode: messages.ScanMode,
			}
		}
	}
	m.queryMode = messages.ScanMode
	m.tableIndex.activeIndex = nil
	m.tableIndex.indexItemCount = -1
	m.scanParameters.index = nil
	m.queryParameters.index = nil
	m.queryParameters.hashKeyValue = ""
	m.queryParameters.rangeKeyOperator = messages.Noop
	m.queryParameters.rangeKeyValue1 = nil
	m.queryParameters.rangeKeyValue2 = nil
	m.queryParameters.rangeOrderDescending = false
	return cmd
}

func (m *ItemSelectionPane) resetColumnSorting(msg messages.ColumnSortingReset) tea.Cmd {
	if msg.TableARN != u.IfNotNil(m.selectedTable.TableArn, "") { // expired
		return nil
	}
	m.applyContentMods(m.modulatedContent.ResetColumnSorting())
	return nil
}

func (m *ItemSelectionPane) escape() tea.Cmd {
	// cancel pending calls
	m.pageCancel()

	// store session data
	if arn := m.selectedTable.TableArn; arn != nil {
		d := SessionData{
			queryMode: m.queryMode,
		}
		d.queryParams.index = m.queryParameters.index
		d.queryParams.hashKeyValue = m.queryParameters.hashKeyValue
		d.queryParams.rangeKeyValue1 = m.queryParameters.rangeKeyValue1
		d.queryParams.rangeKeyValue2 = m.queryParameters.rangeKeyValue2
		d.queryParams.rangeKeyOperator = m.queryParameters.rangeKeyOperator
		d.queryParams.rangeOrderDescending = m.queryParameters.rangeOrderDescending
		d.scanParams.index = m.scanParameters.index
		m.sessions[*arn] = d
	}

	// clean up state
	reset := m.softReset()

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
			StyledItem: "",
		}
	}

	return tea.Batch(reset, resetPreview, switchView)
}

func (m *ItemSelectionPane) UpdateColumnVisibility(msg messages.ColumnVisibilityUpdate) tea.Cmd {
	if msg.TableARN != u.IfNotNil(m.selectedTable.TableArn, "") { // expired
		return nil
	}
	m.applyContentMods(m.modulatedContent.SetColumnVisibility(msg.AllColumns, msg.Visible))
	return nil
}

// toggle column visibility dialog & provide current state (in case dialog opens)
func (m *ItemSelectionPane) toggleColumnVsibilityDialog(tea.Msg) tea.Cmd {
	cols := m.content.Columns()
	vis := m.modulatedContent.ColumnVisibility.InVisible

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
	m.applyContentMods(m.modulatedContent.SetColumnSorting(msg.AllColumns, msg.SortingOn, msg.Ascending))
	return nil
}

// toggle column sorting dialog & provide current state (in case dialog opens)
func (m *ItemSelectionPane) toggleColumnSortingDialog(tea.Msg) tea.Cmd {
	cols := m.content.Columns()
	colsS := make([]string, 0, len(cols))
	for _, c := range cols {
		colsS = append(colsS, c.Title)
	}
	sorting := m.modulatedContent.ColumnSorting.SortingOn
	ascending := m.modulatedContent.ColumnSorting.Ascending
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
	if u.IfNotNil(m.selectedTable.TableArn, "") != msg.TableARN || m.queryMode != messages.ScanMode { // expired
		return nil
	}

	m.resetContents()
	m.clearCache()

	m.queryMode = messages.ScanMode

	idx := u.Ternary(&msg.IndexName, nil, msg.IndexName != "")

	m.scanParameters.index = idx
	m.tableIndex.activeIndex = idx

	m.tableIndex.indexItemCount = u.IfNotNil(m.selectedTable.ItemCount, 0)
	if m.tableIndex.activeIndex != nil {
		m.tableIndex.indexItemCount = indexCountFromTable(*m.tableIndex.activeIndex, m.selectedTable)
	}
	// ensure scan mode is enabled and force new page
	return m.enableScanMode(true)
}

func (m *ItemSelectionPane) ChangeQueryParameters(msg messages.QueryParametersChanged) tea.Cmd {
	if u.IfNotNil(m.selectedTable.TableArn, "") != msg.TableARN || m.queryMode != messages.QueryMode { // expired
		return nil
	}

	// cancel paging, and refresh table contents
	m.resetContents()

	idx := u.Ternary(&msg.IndexName, nil, msg.IndexName != "")

	m.queryParameters.index = idx
	m.tableIndex.activeIndex = idx

	m.tableIndex.activeIndex = u.Ternary(&msg.IndexName, nil, msg.IndexName != "")
	m.tableIndex.indexItemCount = u.IfNotNil(m.selectedTable.ItemCount, 0)
	if m.tableIndex.activeIndex != nil {
		m.tableIndex.indexItemCount = indexCountFromTable(*m.tableIndex.activeIndex, m.selectedTable)
	}
	m.queryParameters.hashKeyValue = msg.HashKeyValue
	m.queryParameters.rangeKeyValue1 = msg.RangeKeyValue1
	m.queryParameters.rangeKeyValue2 = msg.RangeKeyValue2
	m.queryParameters.rangeKeyOperator = msg.RangeKeyOperator
	m.queryParameters.rangeOrderDescending = msg.RangeOrderDescending

	m.resetContents()
	m.clearCache()
	// ensure query mode is enabled and force new page
	return m.enableQueryMode(true)
}

func (m *ItemSelectionPane) copy() tea.Cmd {
	copyDialog := func() tea.Msg {
		return messages.ToggleColumnCopy{}
	}

	cols := m.content.Columns()
	colStr := make([]string, len(cols))
	for i, c := range cols {
		colStr[i] = c.Title
	}

	rowP := m.content.SelectedRow()
	if rowP == nil {
		return nil
	}
	row := *rowP
	values := make([]string, len(row.Fields))
	for i := range row.Fields {
		// remove surrounding quotes if present, for string values
		values[i] = strings.Trim(row.Fields[i].Value(), "\"")
	}
	init := func() tea.Msg {
		return messages.InitColumnCopy{
			TableARN:   u.IfNotNil(m.selectedTable.TableArn, ""),
			AllColumns: colStr,
			ColValues:  values,
		}
	}
	return tea.Batch(copyDialog, init)
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
	width := m.window.width
	leftHalf := width / 2
	rightHalf := width - leftHalf
	// table name
	name := u.IfNotNil(m.selectedTable.TableName, "")

	// determine item count & index name
	count := m.tableIndex.indexItemCount
	indexName := u.IfNotNil(m.tableIndex.activeIndex, "")

	rowcount := int64(len(m.content.VisualRows()))
	right := fmt.Sprintf("Count: %d/%d", rowcount, max(count, rowcount))
	right = ansi.Truncate(right, rightHalf, "…")

	minGap := 15
	left := fmt.Sprintf("Table: %s%s", name, u.Ternary(" / Index: "+indexName, "", indexName != ""))
	left = ansi.Truncate(left, width-len(right)-minGap, "…")

	leftAligned := lipgloss.NewStyle().Width(width - len(right)).Align(lipgloss.Left)
	rightAligned := lipgloss.NewStyle().Width(len(right)).Align(lipgloss.Right)

	return tableInfoBox.Inline(true).Render(lipgloss.JoinHorizontal(lipgloss.Top,
		leftAligned.Render(left),
		rightAligned.Render(right),
	))
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

func extractColumnFromRows(rows []table.Row, idx int) []string {
	if idx < 0 {
		return nil
	}
	items := make([]string, len(rows))
	for i, r := range rows {
		items[i] = r.Fields[idx].Value()
	}
	return items
}
