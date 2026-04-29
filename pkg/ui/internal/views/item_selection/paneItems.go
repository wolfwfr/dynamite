package itemselection

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/keymaps"
)

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
	queryMode   messages.ItemsQueryMode
	chosenIndex *string

	scanLimit  int
	queryLimit int

	items           types.Items
	filteredItems   []int // indices referring to items
	filtering       bool
	lastPreviewItem int // index
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
			table.WithDynamicColumnWidth(false), // TODO: configurable
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
					p.filtering = false
					p.filteredItems = make([]int, 0)
					p.content.ResetVirtualRows()
					return p.MaybePreviewItem(true)
				},
				Results: func(results []search.FilteredItem) tea.Cmd {
					p.filtering = true
					p.filteredItems = make([]int, len(results))
					rows := p.content.Rows()
					filtered := make([]table.Row, len(results))
					for i, match := range results {
						filtered[i] = rows[match.Index]
						p.filteredItems[i] = match.Index
					}
					p.content.SetVirtualRows(filtered)
					return nil
				},
				Reset: func(searchHeight int) tea.Cmd {
					p.filtering = false
					p.filteredItems = make([]int, 0)
					p.content.ResetVirtualRows()
					p.content.SetHeight(p.content.Height() + searchHeight)
					return p.MaybePreviewItem(true)
				},
				SearchBoxOpens: func(searchHeight int) tea.Cmd {
					p.content.SetHeight(p.content.Height() - searchHeight)
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
	m.content.ResetVirtualRows()
	m.content.SetCursor(0)
	m.initialised = false

	// cancel any lingering calls
	m.pageCancel()

	return nil
}

func (m *ItemSelectionPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	cmds := []tea.Cmd{}
	_, isSelect := msg.(messages.SelectTable)
	_, isToggleFmt := msg.(messages.ToggleJSONYAML)
	_, isTick := msg.(spinner.TickMsg)

	excludeSearch := isSelect || isToggleFmt || isTick

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
		case key.Matches(msg, m.KeyMap.Copy):
			return m.copy()
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	case messages.SelectTable:
		return m.selectTable(msg.TableName, msg.TableDetails)
	case messages.ToggleJSONYAML:
		return m.ToggleJSONYAMLFormat()
	case messages.ScanPageReady:
		return m.ProcessScanPage(msg)
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
	if !m.filtering && m.content.ViewAtEnd() {
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
	idx := m.chosenIndex
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
	if m.initialised && len(m.items.Raw) == 0 || m.filtering && len(m.filteredItems) == 0 {
		return func() tea.Msg {
			return messages.PreviewItem{
				Item: "",
			}
		}

	}
	idx := m.content.Cursor()
	if len(m.filteredItems) > 0 { // cursor refers to filtered items
		idx = m.filteredItems[idx]
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

	if m.selectedTable.TableArn != msg.Table.TableArn || m.chosenIndex != msg.Index { // expired
		return nil
	}

	m.appendItems(scan.Items)
	m.pageKey = scan.LastEvaluatedKey

	if len(scan.Items.TableKeys) > 0 {
		// set columns
		_, rang := primaryKeysFromSchema(keysFromIndex(m.chosenIndex, details))
		completeKeys := compileCompleteKeys(scan.Items.TableKeys, m.keysComplete, rang != nil)
		defer func() { m.keysComplete = completeKeys }()

		if slices.Equal(m.keysComplete, completeKeys) {
			// prep new rows & append
			rows := parseRows(completeKeys, scan.Items.TableKeys)
			m.content.AppendRows(rows)
		} else {
			// prep cols, prep ALL rows, set content
			cols := make([]table.Column, len(completeKeys))
			for i, k := range completeKeys {
				cols[i] = table.Column{Title: k, Width: clamp(len(k), 16, 32)}
			}
			rows := parseRows(completeKeys, m.items.TableKeys)
			m.content.SetContent(cols, rows)
		}
	}
	m.paging = false
	m.initialised = true
	return m.MaybePreviewItem(true)
}

// selectTable processes the select-table message, which indicates that the
// item-selection-view is opened because a table has been selected. It will
// default to scanning the first page of items.
func (m *ItemSelectionPane) selectTable(tableName string, details types.DescribeTableResponse) tea.Cmd {
	if session, remembered := m.sessions[*details.TableArn]; remembered {
		// restore session parameters
		m.queryMode = session.queryMode
		m.chosenIndex = session.chosenIndex
	} else {
		// defaults on newly opened table
		m.queryMode = messages.ScanMode
		m.chosenIndex = nil
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
	m.window.height = h
	m.window.width = w
	m.content.SetHeight(h - searchBoxH - ternary(1, 0, m.spinner.active))
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
	m.filteredItems = []int{}
	m.lastPreviewItem = 0
}

func (m *ItemSelectionPane) escape() tea.Cmd {
	m.pageCancel()
	m.resetQueryParameters()
	m.content.SetContent([]table.Column{}, []table.Row{})
	switchView := func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Item_selection,
			NewView: messages.Table_selection,
		}
	}
	resetPreview := func() tea.Msg {
		return messages.PreviewItem{
			Item: "",
		}
	}

	return tea.Batch(resetPreview, switchView)
}

func (m *ItemSelectionPane) copy() tea.Cmd {
	return func() tea.Msg {
		return messages.CopyItem{}
	}
}

func (m *ItemSelectionPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	content := m.content.View()
	content = ternary(content, m.noContentMessage(), !emptyContent(content))
	rendering := []string{content, m.search.View()}
	if m.spinner.active {
		rendering = slices.Insert(rendering, 1, fmt.Sprintf("%s %s", m.spinner.model.View(), m.spinner.text))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendering...)
	// return lipgloss.JoinVertical(lipgloss.Left,
	// 	ternary(content, m.noContentMessage(), !emptyContent(content)),
	// 	m.search.View(),
	// )
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
