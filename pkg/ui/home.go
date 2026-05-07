package ui

import (
	"context"
	"fmt"
	"log"
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/help"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/dialogs"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
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
	query_param_dialog
	mfa_dialog
)

var regionBlock = lipgloss.NewStyle().
	Background(commonstyles.RegionBoxBg).
	Align(lipgloss.Left, lipgloss.Top).
	PaddingLeft(1).
	PaddingRight(1).
	Height(1)

var queryModeBlock = lipgloss.NewStyle().
	Background(commonstyles.QueryModeBoxScanBg).
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
		queryParams      *dialogs.Queryialog
		mfa              *dialogs.MFA
		active           Dialog

		errors []*dialogs.ErrorDialog
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

	{ // mfa dialog
		m.dialogs.mfa = dialogs.NewMFADialog(cfg.MFACredentialC)
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
		m.dialogs.queryParams = dialogs.NewQueryDialog(DialogCloseKeymapFrom(itemViewDialogKeys.QueryParams))
	}

	return m
}

func (m Model) Init() tea.Cmd {
	cfg, err := aws.LoadAWSConfig(m.ctx, m.config.Region, m.config.Profile, m.config.MFACredentialCB)
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
	case appconfig.CredentialsRequest:
		m, cmd = m.OpenMFADialog()
	case messages.CloseMFADialog:
		m, cmd = m.CloseMFADialog()
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
	case messages.ToggleQueryParameters:
		m, cmd = m.ToggleQueryParametersDialog()
	case messages.ToggleErrorDialog:
		m, cmd = m.ToggleErrorDialog(msg)
	case messages.ErrorExpired:
		m, cmd = m.HandleExpiredError(msg)
	case messages.ErrorTick:
		m, cmd = m.handleErrorTick(msg)
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
	cmds = append(cmds, m.dialogs.queryParams.Update(msg))
	cmds = append(cmds, m.dialogs.mfa.Update(msg))

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
		case query_param_dialog:
			return m, m.dialogs.queryParams.Update(msg)
		case mfa_dialog:
			return m, m.dialogs.mfa.Update(msg)
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
		queryModeBlock = queryModeBlock.Background(commonstyles.QueryModeBoxScanBg)
	case messages.QueryMode:
		queryModeBlock = queryModeBlock.Background(commonstyles.QueryModeBoxQeuryBg)
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

func (m Model) handleErrorTick(msg messages.ErrorTick) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, d := range m.dialogs.errors {
		cmds = append(cmds, d.Update(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m Model) HandleExpiredError(msg messages.ErrorExpired) (Model, tea.Cmd) {
	if idx := u.FindBy(m.dialogs.errors, func(d *dialogs.ErrorDialog) bool {
		return d != nil && d.ID() == msg.ID
	}); idx >= 0 {
		m.dialogs.errors = slices.Delete(m.dialogs.errors, idx, idx+1)
	}
	return m, nil
}

func (m Model) ToggleErrorDialog(msg messages.ToggleErrorDialog) (Model, tea.Cmd) {
	d := dialogs.NewErrorDialog(msg.Error)
	m.dialogs.errors = append(m.dialogs.errors, d)
	return m, d.Tick() // initialise ticking
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

func (m Model) ToggleQueryParametersDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != query_param_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = query_param_dialog
	}
	return m, nil
}

// TODO: now assuming no dialog can be open prior to MFA call; ensure existing
// dialogs are closed first!
func (m Model) OpenMFADialog() (Model, tea.Cmd) {
	m.dialogs.open = true
	m.dialogs.active = mfa_dialog
	return m, m.dialogs.mfa.Update(messages.MFAFocus{}) // init focus
}

// TODO: now assuming no dialog can be open prior to MFA call; fallback to
// previous dialog if appliccable!
func (m Model) CloseMFADialog() (Model, tea.Cmd) {
	m.dialogs.open = false
	return m, nil
}

type dialog interface {
	View() string
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
		case query_param_dialog:
			dialog = m.dialogs.queryParams
		case mfa_dialog:
			dialog = m.dialogs.mfa
		}
		renderedDialog := dialog.View()
		dialogLayer := lipgloss.NewLayer(renderedDialog).
			X(m.window.width/2 - lipgloss.Width(renderedDialog)/2).
			Y(m.window.height/2 - lipgloss.Height(renderedDialog)/2)
		c.AddLayers(dialogLayer)
	}

	// error messages
	var errors []string
	for _, d := range m.dialogs.errors {
		errors = append(errors, d.View())
	}
	if len(errors) > 0 {
		errorContent := lipgloss.JoinVertical(lipgloss.Left, errors...)
		errorLayer := lipgloss.NewLayer(errorContent).X(1).Y(1)
		c.AddLayers(errorLayer)
	}

	v := tea.NewView(c.Render())
	v.AltScreen = true // fullscreen
	return v
}

func (m Model) SignalOpenHelpDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleHelp{}
	}
}
