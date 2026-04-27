package tableselection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/keymaps"
)

type tableSelectionPane struct {
	// top-level context
	ctx context.Context

	spinner struct {
		active bool
		model  spinner.Model
		text   string
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

	content *table.Model

	tables           []string
	filteredTables   []int // indices referring to tables
	lastTableDetails int   // index
	details          apitypes.DescribeTableResponse
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
		ctx:           ctx,
		cancelDetails: func() {}, // noop on init
		cancelTables:  func() {}, // noop on init
		debounceDur:   50 * time.Millisecond,
		config:        config,
		stdTO:         5 * time.Second,
		KeyMap:        DefaultTablePaneKeyMap(),
		// TODO: add table feature to hide header
	}

	{ // contents table
		t := table.New(
			table.WithColumns([]table.Column{{Title: "table-name", Width: 64}}),
			table.WithFocused(true),
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
					p.filteredTables = make([]int, 0)
					p.content.ResetVirtualRows()
					return nil
				},
				Results: func(results []search.FilteredItem) {
					p.filteredTables = make([]int, len(results))
					rows := p.content.Rows()
					filtered := make([]table.Row, len(results))
					for i, match := range results {
						filtered[i] = rows[match.Index]
						p.filteredTables[i] = match.Index
					}
					p.content.SetVirtualRows(filtered)
				},
				Reset: func(searchHeight int) {
					p.filteredTables = make([]int, 0)
					p.content.ResetVirtualRows()
					p.updateSize()
				},
				SearchBoxOpens: func(searchHeight int) {
					p.updateSize()
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

func (m *tableSelectionPane) cleanSlate() {
	m.err = nil
}

func (m *tableSelectionPane) Init() tea.Cmd {
	m.content.ResetVirtualRows()
	m.content.SetCursor(0)
	m.cleanSlate()
	m.lastPageKey = nil
	m.tables = []string{}

	// cancel any lingering calls
	m.cancelTables()
	m.cancelDetails()
	// TODO: spinner
	return m.pageNext(true)
}

func (m *tableSelectionPane) activateSpinner() {
	m.spinner.active = true
	m.updateSize()
}

func (m *tableSelectionPane) deactivateSpinner() {
	m.spinner.active = false
	m.updateSize()
}

func (m *tableSelectionPane) pageNext(init bool) tea.Cmd {
	m.activateSpinner()
	if !init && m.lastPageKey == nil { // done paginating
		m.deactivateSpinner()
		return nil
	}
	client := m.config.Client
	region := m.config.Region
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	m.cancelTables = cc
	page := func() tea.Msg {
		defer cc()
		limit := min(100, m.config.MaxTables-len(m.tables)) // 100 is max
		if limit == 0 {
			return nil
		}
		out, err := dynamodb.ListTables(client, ctx, apitypes.ListTablesRequest{
			LastEvaluatedTableName: m.lastPageKey,
			Limit:                  toPtr(int32(limit)),
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
	spinner := m.spinner.model.Tick
	return tea.Batch(page, spinner)
}

func (m *tableSelectionPane) processPage(msg messages.TablePageReady, preview bool) tea.Cmd {
	if msg.Region != m.config.Region { // expired
		return nil
	}
	if msg.Err != nil {
		m.err = msg.Err // TODO: keymap for retry or continue on displaying error
		return nil
	}
	init := len(m.tables) == 0
	newTables := msg.Tables
	m.tables = append(m.tables, newTables...)
	m.lastPageKey = msg.PaginationKey

	// parse and set rows of the new tables
	rows := make([]table.Row, len(newTables))
	for i := range newTables {
		rows[i] = table.Row([]string{newTables[i]})
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
	if len(m.tables) < m.config.MaxTables {
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
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	}
	cmds = append(cmds, m.content.Update(msg))
	return tea.Batch(cmds...)
}

// force is used on new pane initialization because lastPreviewItem could be 0
func (m *tableSelectionPane) MaybePreviewItem(force bool) tea.Cmd {
	idx := m.content.Cursor()
	if len(m.filteredTables) > 0 { // cursor refers to filtered items
		idx = m.filteredTables[idx]
	}
	if idx == m.lastTableDetails && !force {
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
			return nil // debounce
		}

		ctx, cc := context.WithTimeout(ctx, m.stdTO)
		defer cc()
		details, err := dynamodb.DescribeTable(m.config.Client, ctx, table)
		if err != nil {
			return nil
		}

		return messages.TableDetails{
			Details: *details,
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
	r := []string(m.content.SelectedRow())
	if len(r) == 0 {
		return nil // nothing to select
	}
	if m.details.TableName != nil && *m.details.TableName != r[0] {
		m.cleanSlate()
		ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
		defer cc()
		details, err := dynamodb.DescribeTable(m.config.Client, ctx, r[0])
		if err != nil {
			m.err = err
			return nil
		}
		m.details = *details
	}

	selectTable := func() tea.Msg {
		return messages.SelectTable{
			TableName:    r[0],
			TableDetails: m.details,
		}
	}
	return tea.Batch(switchView, selectTable)
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

	searchBoxH := ternary(m.search.GetHeight(), 0, m.search.IsEnabled())
	m.content.SetHeight(h - searchBoxH - ternary(1, 0, m.spinner.active))
	m.content.SetWidth(w)
	m.search.SetWidth(w)
}

func (m *tableSelectionPane) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	rendering := []string{m.content.View(), m.search.View()}
	if m.spinner.active {
		rendering = slices.Insert(rendering, 1, fmt.Sprintf("%s %s", m.spinner.model.View(), m.spinner.text))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendering...)
}
