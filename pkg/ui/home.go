package ui

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/help"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/dialogs"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	itemsview "github.com/wolfwfr/dynamite/pkg/ui/internal/views/items"
	tablesview "github.com/wolfwfr/dynamite/pkg/ui/internal/views/tables"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type View int
type Dialog int

const (
	tables_view View = iota
	items_view
)

const (
	help_dialog Dialog = iota
	regions_dialog
	columns_dialog
	column_sorting_dialog
	scan_param_dialog
)

var (
	queryColour = "#046645"
	scanColour  = "#0E3080"
	adminColour = "#0E5680"
)

var regionBlock = lipgloss.NewStyle().
	Background(lipgloss.Color("#80380E")).
	Align(lipgloss.Left, lipgloss.Top).
	PaddingLeft(1).
	PaddingRight(1).
	Height(1)

var queryModeBlock = lipgloss.NewStyle().
	Background(lipgloss.Color(scanColour)).
	Align(lipgloss.Left, lipgloss.Top).
	PaddingLeft(1).
	PaddingRight(1).
	Height(1)

type Model struct {
	// ActiveView determines tea.Msg forwarding
	activeView View

	QueryMode messages.ItemsQueryMode

	window struct {
		width  int
		height int
	}

	// dialogs
	dialogs struct {
		open             bool
		help             *dialogs.Help
		region           *dialogs.Regions
		columnVisibility *dialogs.Columns
		columnSorting    *dialogs.ColumnSorting
		scanParams       *dialogs.ScanDialog
		active           Dialog
	}

	// top-level context
	ctx context.Context

	// shared config
	config *appconfig.Config

	// views
	tableSelection *tablesview.TableSelection
	itemselection  *itemsview.ItemSelection

	// help
	Help help.Model
}

func NewModel(ctx context.Context, cfg appconfig.Config) Model {
	m := Model{
		ctx:    ctx,
		config: &cfg,

		activeView: tables_view,
		Help:       help.New(),
	}

	km := DefaultKeyMap()

	inheritedKeys := []keymaps.AdditionalKey{
		{
			Binding: km.Quit,
			Call:    tea.Quit,
		}, {
			Binding: km.Help,
			Call:    m.SignalOpenHelpDialog(),
		},
	}

	{ // views
		m.tableSelection = tablesview.NewTableSelectionView(ctx, &cfg, tablesview.WithAdditionalKeys(keymaps.AdditionalKeys(inheritedKeys)))
		m.itemselection = itemsview.NewItemSelectionView(ctx, &cfg, itemsview.WithAdditionalKeys(keymaps.AdditionalKeys(inheritedKeys)))
	}

	{ // table view bound dialogs
		tableViewDialogKeys := m.tableSelection.DialogKeyMaps()
		m.dialogs.help = dialogs.NewHelp(m.tableSelection, m.itemselection, DialogCloseKeymapFrom(km.Help))
		m.dialogs.region = dialogs.NewRegionsDialog(m.config.AvailableRegions, m.config.StarredRegions, m.config.Region, DialogCloseKeymapFrom(tableViewDialogKeys.RegionDialog))
	}

	{ // table view bound dialogs
		itemViewDialogKeys := m.itemselection.DialogKeyMaps()
		m.dialogs.columnVisibility = dialogs.NewColumnVisibilityDialog(DialogCloseKeymapFrom(itemViewDialogKeys.ColumnVisibility))
		m.dialogs.columnSorting = dialogs.NewColumnSortingDialog(DialogCloseKeymapFrom(itemViewDialogKeys.ColumnSorting))
		m.dialogs.scanParams = dialogs.NewScanDialog(DialogCloseKeymapFrom(itemViewDialogKeys.ScanParams))
	}

	return m

}

func (m Model) Init() tea.Cmd {
	cfg, err := aws.LoadAWSConfig(m.ctx, m.config.Region, m.config.Profile)
	if err != nil {
		// TODO: handling
	}
	m.config.Client = dynamodb.NewClient(cfg, m.config.URL)
	var cmds []tea.Cmd
	cmds = append(cmds, m.tableSelection.Init())
	cmds = append(cmds, m.itemselection.Init())
	return tea.Batch(cmds...)
}

// update handles the message and proceeds to forward it to the model's children
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.SwitchView:
		m, cmd = m.handleSwitchView(msg)
	case tea.WindowSizeMsg:
		m = m.applySize(msg.Height, msg.Width).(Model)
	case messages.ToggleHelp:
		m, cmd = m.ToggleHelpDialog()
	case messages.ToggleRegions:
		m, cmd = m.ToggleRegionsDialog()
	case messages.ToggleColumnVisibility:
		m, cmd = m.ToggleColumnsDialog()
	case messages.ToggleColumnSorting:
		m, cmd = m.ToggleColumnSortingDialog()
	case messages.ToggleScanParameters:
		m, cmd = m.ToggleScanParametersDialog()
	case messages.SwitchRegion:
		m, cmd = m.switchRegion(msg.OldRegion, msg.NewRegion)
	case messages.SwitchQueryMode:
		m, cmd = m.SwitchQueryMode(msg)
	}

	var fwdCmd tea.Cmd
	m, fwdCmd = m.forward(msg)
	return m, tea.Batch(cmd, fwdCmd)
}

// forward takes a message and decides to broadcast or to forward only to active
// children
func (m Model) forward(msg tea.Msg) (Model, tea.Cmd) {
	if _, isKeyPress := msg.(tea.KeyPressMsg); isKeyPress {
		return m.routeToActiveOnly(msg)
	}
	return m.broadcast(msg)
}

// broadcast takes a message and forwards it to all children
func (m Model) broadcast(msg tea.Msg) (Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	// views
	cmds = append(cmds, m.tableSelection.Update(msg))
	cmds = append(cmds, m.itemselection.Update(msg))

	// dialogs
	cmds = append(cmds, m.dialogs.help.Update(msg))
	cmds = append(cmds, m.dialogs.region.Update(msg))
	cmds = append(cmds, m.dialogs.columnVisibility.Update(msg))
	cmds = append(cmds, m.dialogs.columnSorting.Update(msg))
	cmds = append(cmds, m.dialogs.scanParams.Update(msg))

	return m, tea.Batch(cmds...)
}

// routeToActiveOnly takes a message and only routes it to a single child, the
// active child with highest precedence (dialogs take precedence over views)
func (m Model) routeToActiveOnly(msg tea.Msg) (Model, tea.Cmd) {
	// exclusively forward keypresses to dialogs if open
	if m.dialogs.open {
		switch m.dialogs.active {
		case help_dialog:
			return m, m.dialogs.help.Update(msg)
		case regions_dialog:
			return m, m.dialogs.region.Update(msg)
		case columns_dialog:
			return m, m.dialogs.columnVisibility.Update(msg)
		case column_sorting_dialog:
			return m, m.dialogs.columnSorting.Update(msg)
		case scan_param_dialog:
			return m, m.dialogs.scanParams.Update(msg)
		}
	}

	switch m.activeView {
	case tables_view:
		return m, m.tableSelection.Update(msg)
	case items_view:
		return m, m.itemselection.Update(msg)
	default:
		log.Fatalf("could not identify active view '%d'", int(m.activeView))
	}

	return m, nil
}

func (m Model) SwitchQueryMode(msg messages.SwitchQueryMode) (Model, tea.Cmd) {
	m.QueryMode = msg.NewMode
	switch m.QueryMode {
	case messages.ScanMode:
		queryModeBlock = queryModeBlock.Background(lipgloss.Color(scanColour))
	case messages.QueryMode:
		queryModeBlock = queryModeBlock.Background(lipgloss.Color(queryColour))
	}
	return m, nil
}

func (m Model) switchRegion(oldr, newr string) (Model, tea.Cmd) {
	m.config.Region = newr
	return m, m.Init()
}

func (m Model) applySize(height, width int) tea.Model {
	m.Help.SetWidth(width)
	m.window.height = height
	m.window.width = width
	return m
}

func (m Model) handleSwitchView(msg messages.SwitchView) (Model, tea.Cmd) {
	switch msg.NewView {
	case messages.Table_selection:
		m.activeView = tables_view
	case messages.Item_selection:
		m.activeView = items_view
	}
	return m, m.dialogs.help.Update(msg)
}

func (m Model) ToggleHelpDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != help_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = help_dialog
	}
	return m, nil
}

func (m Model) ToggleRegionsDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != regions_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = regions_dialog
	}
	return m, nil
}

func (m Model) ToggleColumnsDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != columns_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = columns_dialog
	}
	return m, nil
}

func (m Model) ToggleColumnSortingDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != column_sorting_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = column_sorting_dialog
	}
	return m, nil
}

func (m Model) ToggleScanParametersDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != scan_param_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = scan_param_dialog
	}
	return m, nil
}

type dialog interface {
	View() string
	Width() int
}

func (m Model) View() tea.View {
	var page string
	var help string
	switch m.activeView {
	case tables_view:
		page = m.tableSelection.View()
		help = m.Help.ShortHelpView(m.tableSelection.ShortHelp())
	case items_view:
		page = m.itemselection.View()
		help = m.Help.ShortHelpView(m.itemselection.ShortHelp())
	}

	// assemble gutter
	region := regionBlock.Render(m.config.Region)
	queryMode := u.Ternary("QUERY", "SCAN", m.QueryMode == messages.QueryMode)
	query := u.Ternary(fmt.Sprintf(" %s", queryModeBlock.Render(queryMode)), "", m.activeView == items_view)
	gutter := lipgloss.JoinHorizontal(lipgloss.Left, region, query, " ", help)

	page = lipgloss.JoinVertical(lipgloss.Top, page, gutter)

	// dialog compositing
	mainLayer := lipgloss.NewLayer(page)
	c := lipgloss.NewCompositor(mainLayer)
	c.AddLayers(mainLayer)
	if m.dialogs.open {
		var dialog dialog
		switch m.dialogs.active {
		case help_dialog:
			dialog = m.dialogs.help
		case regions_dialog:
			dialog = m.dialogs.region
		case columns_dialog:
			dialog = m.dialogs.columnVisibility
		case column_sorting_dialog:
			dialog = m.dialogs.columnSorting
		case scan_param_dialog:
			dialog = m.dialogs.scanParams
		}
		renderedDialog := dialog.View()
		dialogLayer := lipgloss.NewLayer(renderedDialog).
			X(m.window.width/2 - dialog.Width()/2).
			Y(m.window.height/2 - heightFromView(renderedDialog)/2)
		c.AddLayers(dialogLayer)
	}

	v := tea.NewView(c.Render())
	v.AltScreen = true // fullscreen
	return v
}

func heightFromView(v string) int {
	return strings.Count(v, "\n")
}

func (m Model) SignalOpenHelpDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleHelp{}
	}
}
