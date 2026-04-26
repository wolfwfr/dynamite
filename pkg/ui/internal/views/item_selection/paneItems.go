package itemselection

import (
	"context"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/table"
)

type ItemSelectionPane struct {
	// top-level context
	ctx context.Context

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

	content         *table.Model
	items           types.Items
	filteredItems   []int // indices referring to items
	lastPreviewItem int   // index

	selectedTable string
}

func NewItemSelectionPane(ctx context.Context, config *appconfig.Config) *ItemSelectionPane {
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
	p := &ItemSelectionPane{
		ctx:     ctx,
		config:  config,
		stdTO:   5 * time.Second,
		content: t,
	}
	p.search = search.NewSearchBox(
		search.SearchCallbacks{
			ToSearch: func() []string {
				return table.Rows(p.content.Rows()).ToStrings()
			},
			EmptyInput: func() tea.Cmd {
				p.content.ResetVirtualRows()
				return nil
			},
			Results: func(results []search.FilteredItem) {
				p.filteredItems = make([]int, len(results))
				rows := p.content.Rows()
				filtered := make([]table.Row, len(results))
				for i, match := range results {
					filtered[i] = rows[match.Index]
					p.filteredItems[i] = match.Index
				}
				p.content.SetVirtualRows(filtered)
			},
			Reset: func(searchHeight int) {
				p.filteredItems = make([]int, 0)
				p.content.ResetVirtualRows()
				p.content.SetHeight(p.content.Height() + searchHeight)
			},
			ViewBoxOpens: func(searchHeight int) {
				p.content.SetHeight(p.content.Height() - searchHeight)
			},
		},
	)
	return p
}

func (m *ItemSelectionPane) cleanSlate() {
	m.err = nil
}

func (m *ItemSelectionPane) Init() tea.Cmd {
	return nil
}

func (m *ItemSelectionPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	cmds := []tea.Cmd{}
	_, isSelect := msg.(messages.SelectTable)
	if search.IsSearchBoxMessage(msg) || (!isSelect && m.search.IsFocused()) {
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
		switch s := msg.String(); s {
		case "/":
			cmds = append(cmds, m.search.OpenSearchBox())
		case "esc":
			if m.search.IsEnabled() {
				m.search.Reset()
			} else {
				return m.escape()
			}
		case "W":
			m.content.SetDynamicColumnWidth(!m.content.DynamicColumnWidth())
		case "Z":
			return m.Zoom()
		}
	case messages.SelectTable:
		return m.selectTable(msg.TableName, msg.TableDetails)
	}
	cmds = append(cmds, m.content.Update(msg))
	return tea.Batch(cmds...)
}

// force is used on new pane initialization because lastPreviewItem could be 0
func (m *ItemSelectionPane) MaybePreviewItem(force bool) tea.Cmd {
	idx := m.content.Cursor()
	if len(m.filteredItems) > 0 { // cursor refers to filtered items
		idx = m.filteredItems[idx]
	}
	if idx == m.lastPreviewItem && !force {
		return nil
	}
	m.lastPreviewItem = idx
	return func() tea.Msg {
		return messages.PreviewItem{
			Item: m.items.YAML[idx],
		}
	}
}

func (m *ItemSelectionPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleItemSelectionPane{}
	}
}

// selectTable processes the select-table message, which indicates that the
// item-selection-view is opened because a table has been selected. It will
// default to scanning the first page of items.
func (m *ItemSelectionPane) selectTable(tableName string, details types.DescribeTableResponse) tea.Cmd {
	// resetting state
	m.cleanSlate()
	m.content.ResetVirtualRows()

	m.selectedTable = tableName
	// TODO: spinner & async
	ctx, cc := context.WithTimeout(m.ctx, m.stdTO)
	defer cc()
	scan, err := dynamodb.ScanTable(m.config.Client, ctx, tableName, types.ScanParameters{
		KeyDetails: details.AttributeDefinitions,
		KeySchema:  details.KeySchema,
		Limit:      m.window.height,
	})
	if err != nil {
		m.err = err
		return nil
	}

	m.items = scan.Items

	if len(scan.Items.TableKeys) > 0 {
		// set columns
		_, rang := primaryKeysFromSchema(details.KeySchema)
		completekeys := compileCompleteKeys(scan.Items.TableKeys, rang != nil)
		cols := make([]table.Column, len(completekeys))
		for i, k := range completekeys {
			cols[i] = table.Column{Title: k, Width: clamp(len(k), 16, 32)}
		}

		// set rows
		rows := make([]table.Row, len(scan.Items.TableKeys))
		for i, k := range scan.Items.TableKeys {
			row := make([]string, len(completekeys))
			var x int
			for j, key := range completekeys {
				if key == k[x].Key { // matching key
					row[j] = k[x].Value
					x = min(len(k)-1, x+1)
				} else { // no matching key
					row[j] = ""
				}
			}
			rows[i] = row
		}
		m.content.SetContent(cols, rows)
	}
	return m.MaybePreviewItem(true)
}

// compileCompleteKeys takes a table of key-value pairs, observes all keys and
// compiles a complete, in-order list of all unique key observed.
// This ensures that when individual table rows have keys missing, the final
// result still contains these keys when they are present in other rows in the
// specified table.
// TODO: accept existing key-slice for pagination compatibility
func compileCompleteKeys(table [][]types.KeyValue, hasRangeKey bool) []string {
	res := make([]string, 0)
	seen := map[string]struct{}{}
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
	searchBoxH := m.search.GetHeight()
	if !m.search.IsEnabled() {
		searchBoxH = 0
	}
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height - searchBoxH)
	m.content.SetWidth(width)
	m.search.SetWidth(width)
}

func (m *ItemSelectionPane) escape() tea.Cmd {
	return func() tea.Msg {
		return messages.SwitchView{
			OldView: messages.Item_selection,
			NewView: messages.Table_selection,
		}
	}
}

func (m *ItemSelectionPane) View() string {
	if m.err != nil { // TODO: formatting
		return m.err.Error()
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.content.View(),
		m.search.View(),
	)
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
