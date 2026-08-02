package tableselection

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
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

	"github.com/atotto/clipboard"

	"github.com/wolfwfr/dynamite/lib/styles"
	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/adapters/dynamodb"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type TableStyles struct {
	SelectedBackground    color.Color
	SearchMatchBackground color.Color
	MatchedNames          []lipgloss.Style
	DefaultStyle          lipgloss.Style
}

type tableSelectionPane struct {
	// top-level context
	ctx context.Context

	// logger
	logger *slog.Logger

	// styles
	styles struct {
		Table TableStyles
	}

	// spinner
	spinner struct {
		active    bool
		model     spinner.Model
		text      string
		textStyle lipgloss.Style
	}
	// cancel last call context (debounce)
	cancelDetails func()
	debounceDur   time.Duration

	// paging tables
	cancelTables func()
	lastPageKey  *string

	// standard timeout
	stdTO time.Duration

	// shared config
	config *appconfig.Config

	// dynamo-db adapter for UI purposes
	dynamodbClient dynamodbClient

	// errorText
	err error

	// pane's view window
	window struct {
		width  int
		height int
	}

	// fuzzy finding
	search *search.SearchBox

	// key map
	KeyMap *TablePaneKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	// the underlying table
	content *table.Model

	// the table-names retrieved from dynamodb
	tables []string

	// filtering parameters
	tablefiltering struct {
		matchedTables []int   // indices referring to tables
		matchedRunes  [][]int //matches by index to tablefiltering.mathedTables
		enabled       bool
	}

	// index to most recently received table details
	lastTableDetails int // index

	// table details
	details *apitypes.DescribeTableResponse
}

type tablePaneOption func(p *tableSelectionPane)

// withTablePaneKeys
func withTablePaneKeys(keys keymaps.AdditionalKeys) tablePaneOption {
	return func(t *tableSelectionPane) {
		t.AddKeyMap = keys
	}
}

func newTableSelectionPane(ctx context.Context, config *appconfig.Config, opts ...tablePaneOption) *tableSelectionPane {
	p := &tableSelectionPane{
		ctx:            ctx,
		logger:         config.Logger.With(slog.String(logging.ViewKey, Log_TablesView), slog.String(logging.PaneKey, "tables-pane")),
		cancelDetails:  func() {}, // noop on init
		cancelTables:   func() {}, // noop on init
		debounceDur:    50 * time.Millisecond,
		config:         config,
		dynamodbClient: dynamodb.NewAdapter(config.Logger),
		stdTO:          30 * time.Second,
		KeyMap:         DefaultTablePaneKeyMap(),
	}

	{ // contents table
		t := table.New(
			table.WithColumns([]table.Column{{Title: "table-name", Width: 64, UseDynamicWidth: true}}),
			table.WithFocused(true),
			table.WithFieldDelegate(p.TableRowFieldDelegate),
		)
		p.content = t
	}

	{ // spinner
		sp := spinner.New()
		sp.Spinner = spinner.Dot
		p.spinner.model = sp
		p.spinner.text = "obtaining next page..."
	}

	{ // search box
		p.search = search.NewSearchBox(
			search.SearchCallbacks{
				ToSearch: func(string) []string {
					return table.Rows(p.content.Rows()).ToStrings()
				},
				EmptyInput: func() tea.Cmd {
					p.tablefiltering.enabled = false
					p.tablefiltering.matchedTables = make([]int, 0)
					p.tablefiltering.matchedRunes = make([][]int, 0)
					p.content.ResetVirtualRows()
					return p.MaybePreviewItem(true)
				},
				Results: func(_ string, results []search.FilteredItem) tea.Cmd {
					p.tablefiltering.enabled = true
					p.tablefiltering.matchedTables = make([]int, len(results))
					p.tablefiltering.matchedRunes = make([][]int, len(results))
					rows := p.content.Rows()
					filtered := make([]table.Row, len(results))
					for i, match := range results {
						filtered[i] = rows[match.Index]
						p.tablefiltering.matchedTables[i] = match.Index
						p.tablefiltering.matchedRunes[i] = match.Matches
					}
					p.content.SetVirtualRows(filtered)
					return p.MaybePreviewItem(true)
				},
				Reset: func(searchHeight int) tea.Cmd {
					p.tablefiltering.enabled = false
					p.tablefiltering.matchedTables = make([]int, 0)
					p.tablefiltering.matchedRunes = make([][]int, 0)
					p.content.ResetVirtualRows()
					p.updateSize()
					return p.MaybePreviewItem(true)
				},
				SearchBoxOpens: func(searchHeight int) tea.Cmd {
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

	p.updateStyles()

	return p
}

func (m *tableSelectionPane) updateStyles() {
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
	m.content.SetStyles(s)

	st := TableStyles{
		SelectedBackground:    theme.TableSelectedBg,
		SearchMatchBackground: theme.SearchHighlight,
		MatchedNames: []lipgloss.Style{
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault1),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault2),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault3),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault4),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault5),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault6),
			lipgloss.NewStyle().Foreground(theme.TableHighlightDefault7),
		},
		DefaultStyle: lipgloss.NewStyle().Foreground(theme.TableHighlightDefault1),
	}

	m.styles.Table = st

	m.spinner.model.Style = lipgloss.NewStyle().
		Foreground(theme.SpinnerSymbolFg).
		Background(theme.SpinnerSymbolBg).
		PaddingLeft(1)

	m.spinner.textStyle = lipgloss.NewStyle().
		Foreground(theme.SpinnerTextFg).
		Background(theme.SpinnerTextBg)
}

func (m *tableSelectionPane) cleanSlate() {
	m.err = nil
}

func (m *tableSelectionPane) Init() tea.Cmd {
	m.logger.Info("initialising...")

	m.search.Reset()
	m.content.SetRows([]table.Row{})
	m.content.ResetVirtualRows()
	m.content.SetCursor(0)
	m.cleanSlate()
	m.lastPageKey = nil
	m.tables = []string{}
	m.initialiseRegex(m.config.Tables.HighlightRegexp)

	// cancel any lingering calls
	m.cancelTables()
	m.cancelDetails()
	cmd := m.pageNext(true)

	m.logger.Info("initilasation complete")

	return cmd
}

func (m *tableSelectionPane) activateSpinner() tea.Cmd {
	m.spinner.active = true
	m.updateSize()
	return m.spinner.model.Tick
}

func (m *tableSelectionPane) deactivateSpinner() {
	m.spinner.active = false
	m.updateSize()
}

func (m *tableSelectionPane) pageNext(init bool) tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace, "attempting paging", slog.Bool("init_flag", init))
	spinnerCmd := m.activateSpinner()
	if !init && m.lastPageKey == nil { // done paginating
		m.logger.Log(m.ctx, logging.LevelTrace, "not eligible for paging; aborting",
			slog.Bool("page_key_not_nil", m.lastPageKey != nil),
			slog.String("page_key", u.IfNotNil(m.lastPageKey, "")),
		)
		m.deactivateSpinner()
		return nil
	}
	client := m.config.Client
	region := m.config.Region
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	m.cancelTables = cc
	pagesize := u.Ternary(m.config.Tables.Pagesize, max(100, m.window.height), m.config.Tables.Pagesize > 0)
	page := func() tea.Msg {
		defer cc()
		limit := min(pagesize, m.config.Tables.MaxTables-len(m.tables)) // 100 is max
		if limit == 0 {
			m.logger.Warn("limit is 0; aborting paging")
			return nil
		}
		out, err := m.dynamodbClient.ListTables(client, ctx, apitypes.ListTablesRequest{
			LastEvaluatedTableName: m.lastPageKey,
			Limit:                  u.ToPtr(int32(limit)),
		})

		msg := messages.TablePageReady{
			Err:    err,
			Region: region,
		}
		if out != nil {
			msg.Tables = out.TableNames
			msg.PaginationKey = out.LastEvaluatedTableName
		}
		return msg
	}
	return tea.Batch(page, spinnerCmd)
}

func (m *tableSelectionPane) processPage(msg messages.TablePageReady, preview bool) tea.Cmd {
	m.logger.Debug("received a new page",
		slog.Int("page_size", len(msg.Tables)),
	)
	if msg.Region != m.config.Region { // expired
		m.logger.Info("ignoring message due to outdated region",
			slog.String("current_region", m.config.Region),
			slog.String("message_region", msg.Region),
		)
		return nil
	}
	if msg.Err != nil {
		m.logger.Error(
			"received error on new page",
			slog.Any("error", msg.Err),
		)
		m.err = msg.Err
		return nil
	}
	init := len(m.tables) == 0
	newTables := msg.Tables
	m.tables = append(m.tables, newTables...)
	m.lastPageKey = msg.PaginationKey

	// parse and set rows of the new tables
	rows := make([]table.Row, len(newTables))
	for i := range newTables {
		rows[i] = table.Row{Fields: []table.Field{
			enrichedField{value: newTables[i]},
		}}
	}
	if init {
		m.content.SetRows(rows)
	} else {
		m.content.AppendRows(rows)
	}

	// return commands
	cmds := []tea.Cmd{}
	if preview {
		cmds = append(cmds, m.MaybePreviewItem(true))
	}
	if len(m.tables) < m.config.Tables.MaxTables {
		cmds = append(cmds, m.pageNext(false))
	}
	return tea.Batch(cmds...)
}

func (m *tableSelectionPane) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case messages.TableDetails:
		m.details = msg.Details
		return nil
	case messages.SwitchView:
		if msg.NewView != messages.Table_selection {
			return nil
		}
		return m.MaybePreviewItem(true)
	case messages.TablePageReady:
		return m.processPage(msg, len(m.tables) == 0)
	case spinner.TickMsg:
		if !m.spinner.active {
			return nil
		}
		var cmd tea.Cmd
		m.spinner.model, cmd = m.spinner.model.Update(msg)
		return cmd
	}

	if search.IsSearchBoxMessage(msg) || m.search.IsFocused() {
		cmds = append(cmds, m.search.Update(msg))
	} else if _, ok := msg.(tea.BackgroundColorMsg); ok {
		m.updateStyles()
		return nil
	} else {
		cmds = append(cmds, m.handleNavigation(msg))
	}

	cmds = append(cmds, m.MaybePreviewItem(false))
	return tea.Batch(cmds...)
}

// handleNavigation handles events when search is not active.
func (m *tableSelectionPane) handleNavigation(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Search):
			cmds = append(cmds, m.search.OpenSearchBox())
		case key.Matches(msg, m.KeyMap.Select):
			return m.selectTable()
		case key.Matches(msg, m.KeyMap.Zoom):
			return m.Zoom()
		case key.Matches(msg, m.KeyMap.Esc):
			m.search.Reset()
		case key.Matches(msg, m.KeyMap.Reload):
			return m.Init()
		case key.Matches(msg, m.KeyMap.Browser):
			return m.openInBrowser()
		case key.Matches(msg, m.KeyMap.Copy):
			return m.copy()
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	}
	cmds = append(cmds, m.content.Update(msg))
	return tea.Batch(cmds...)
}

type enrichedField struct {
	value string
}

// Value implements the matching table.Field interface function
func (f enrichedField) Value() string {
	return f.value
}

func (m *tableSelectionPane) TableRowFieldDelegate(row table.Row, col table.Column, colIdx, rowIdx, colW, padL, padR int, selected, inview bool, _, _ int) string {
	fullWidth := colW + padL + padR

	// obtain field in question
	field := row.Fields[colIdx].(enrichedField)

	enforceWidth := lipgloss.NewStyle().Width(fullWidth).MaxWidth(fullWidth).Inline(true).Render

	segments, styling := compileMatchedStyles(field.value, m.styles.Table.MatchedNames, m.styles.Table.DefaultStyle)
	var style styles.LineStyle
	for i := range segments {
		style = style.AppendStringLG(segments[i], styling[i])
	}

	// add padding
	style = style.SetRightPaddingLast(padR)
	style = style.SetLeftPaddingFirst(padL)

	// apply background styling for selected row
	if selected {
		// fill up any remaining space
		if len([]rune(field.value)) < fullWidth {
			st, _ := style.GetAt(len([]rune(field.value)) - 1)
			style = style.Override(len([]rune(field.value))-1, st.PaddingRight(fullWidth-len([]rune(field.value))))
		}
		style = style.SetBackgroundAll(m.styles.Table.SelectedBackground)
	}

	// override background styling for search matches
	if m.tablefiltering.enabled {
		for _, idx := range m.tablefiltering.matchedRunes[rowIdx] {
			runeStyle, _ := style.GetAt(idx)
			c := m.styles.Table.SearchMatchBackground
			if selected {
				c = lipgloss.Blend1D(10, c, m.styles.Table.SelectedBackground)[3]
			}
			style = style.Override(idx, runeStyle.Background(c))
		}
	}

	return enforceWidth(style.Render(field.value))
}

func (m *tableSelectionPane) copy() tea.Cmd {
	m.logger.Debug("executing copy operation")
	r := m.content.VisualRows()
	c := max(0, m.content.Cursor())
	if c >= len(r) {
		return nil
	}

	if err := clipboard.WriteAll(r[c].String()); err != nil {
		m.logger.Error("copy to clipboard returned an error", slog.Any("error", err))
		return func() tea.Msg {
			return messages.ToggleNotificationDialog{Error: fmt.Errorf("failed to copy: %w", err)}
		}
	}
	return notifyCopySuccess
}

// force is used on new pane initialization because lastPreviewItem could be 0
func (m *tableSelectionPane) MaybePreviewItem(force bool) tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace,
		"received request to preview table details",
		slog.Bool("force", force),
	)
	if len(m.tables) == 0 || (m.tablefiltering.enabled && len(m.tablefiltering.matchedTables) == 0) {
		m.logger.Debug("nothing to view; emptying details pane",
			slog.Int("len_tables", len(m.tables)),
			slog.Bool("search_enabled", m.tablefiltering.enabled),
			slog.Int("seach_matches", len(m.tablefiltering.matchedTables)),
		)
		return func() tea.Msg {
			return messages.TableDetails{
				Details: nil,
			}
		}
	}

	idx := m.content.Cursor()
	if len(m.tablefiltering.matchedTables) > 0 { // cursor refers to filtered items
		idx = m.tablefiltering.matchedTables[idx]
	}
	if idx == m.lastTableDetails && !force {
		m.logger.Debug("preview request is a duplicate; skipping",
			slog.Int("table_index", idx),
			slog.Int("last_previewed_index", m.lastTableDetails),
			slog.Bool("force", force),
		)
		return nil
	}
	m.lastTableDetails = idx
	table := m.tables[idx]

	// prepare debounce cancellation
	m.cancelDetails()
	ctx, cc := context.WithCancel(m.ctx)
	m.cancelDetails = cc

	return func() tea.Msg {
		time.Sleep(m.debounceDur)
		if ctx.Err() != nil { // context canceled
			m.logger.Log(m.ctx, logging.LevelTrace, "debounce encountered on describing table")
			return nil // debounce
		}

		ctx, cc := context.WithTimeout(ctx, m.stdTO)
		defer cc()
		details, err := m.dynamodbClient.DescribeTable(m.config.Client, ctx, table)
		if err != nil {
			// logging at debug because almost certainly a deliberately canceled call
			m.logger.Debug("encountered error describing table", slog.Any("error", err))
			return nil
		}

		return messages.TableDetails{
			Details: details,
		}
	}
}

func (m *tableSelectionPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleTableSelectionPane{}
	}
}

func (m *tableSelectionPane) selectTable() tea.Cmd {
	m.cancelDetails()
	m.cancelTables()
	switchView := func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Table_selection,
			NewView: messages.Item_selection,
		}
	}
	rowP := m.content.SelectedRow()
	if rowP == nil {
		m.logger.Debug("table selected, but row returned nil; aborting")
		return nil
	}
	row := *rowP
	if len(row.Fields) == 0 {
		m.logger.Debug("table selected, but row returned no fields; aborting")
		return nil // nothing to select
	}

	selectTable := func() tea.Msg {
		return messages.SelectTable{
			TableName: row.Fields[0].Value(),
		}
	}
	return tea.Batch(switchView, selectTable)
}

// openInBrowser opens the selected table in the system's default browser
func (m *tableSelectionPane) openInBrowser() tea.Cmd {
	selection := m.content.SelectedRow()
	if selection == nil || len(selection.Fields) == 0 {
		return nil
	}

	var (
		region = m.config.Region
		// TODO: think about config workaround for when AWS would change URL
		weburl    = fmt.Sprintf("https://%s.console.aws.amazon.com/dynamodbv2/home", region)
		tableName = selection.Fields[0].Value()
		cmd       string
		args      []string
	)

	paramkeys := []string{
		"region",
		"name",
	}

	paramVals := []string{
		fmt.Sprintf("%s#table?", region),
		url.PathEscape(tableName),
	}

	// manually parsing query parameters, because of the strange double query
	// parameter section in the dynamo-db url
	weburl = fmt.Sprintf("%s%s", weburl, u.Ternary("?", "", len(paramkeys) > 0))
	for i := range paramkeys {
		sep := u.Ternary("&", "", i > 1)
		weburl = fmt.Sprintf("%s%s%s=%s", weburl, sep, paramkeys[i], paramVals[i])
	}

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, weburl)

	m.logger.Debug("opening in browser",
		slog.String("url", weburl),
	)

	if err := exec.Command(cmd, args...).Start(); err != nil {
		m.logger.Debug("encountered error opening url in browser",
			slog.String("url", weburl),
			slog.String("cmd", cmd),
			slog.String("args", fmt.Sprintf("%q", args)),
			slog.Any("error", err),
		)
		return notifyError(err)
	}

	return nil
}

func (m *tableSelectionPane) applySize(height, width int) {
	m.window.height = height
	m.window.width = width
	m.updateSize()
}

// updateSize updates dimensions of the pane's contents based on the current
// window dimensions.
func (m *tableSelectionPane) updateSize() {
	h, w := m.window.height, m.window.width

	searchBoxH := u.Ternary(m.search.GetHeight(), 0, m.search.IsEnabled())
	// TODO: fix the '1' content prints one empty row beyond its allowed height
	m.content.SetHeight(h - 1 - searchBoxH - u.Ternary(1, 0, m.spinner.active))
	m.content.SetWidth(w)
	m.search.SetWidth(w)
}

func (m *tableSelectionPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	content := u.Ternary(m.content.View(), m.noContentMessage(), len(m.content.Rows()) > 0)
	rendering := []string{content, m.search.View()}
	if m.spinner.active {
		rendering = slices.Insert(rendering, 1, fmt.Sprintf("%s %s", m.spinner.model.View(), m.spinner.text))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendering...)
}

func (m *tableSelectionPane) noContentMessage() string {
	if m.spinner.active {
		return ""
	}
	s := strings.Builder{}
	fmt.Fprintf(&s, "==================================================\n")
	fmt.Fprintf(&s, "             NO TABLES IN THIS REGION             \n")
	fmt.Fprintf(&s, "==================================================\n")
	return s.String()
}

func notifyError(err error) tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleNotificationDialog{Error: err}
	}
}

func notifyCopySuccess() tea.Msg {
	return messages.ToggleNotificationDialog{Msg: "Copied!", Duration: 1 * time.Second}
}
