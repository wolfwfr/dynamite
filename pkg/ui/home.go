package ui

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/dialogs"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/theme"
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
	column_transform_dialog
	column_width_dialog
	scan_param_dialog
	query_param_dialog
	filter_param_dialog
	copy_dialog
	mfa_dialog
)

var regionBlock = lipgloss.NewStyle().
	Foreground(theme.BoxFg).
	Background(theme.RegionBoxBg).
	Align(lipgloss.Left, lipgloss.Top).
	Padding(0, 1, 0, 1).
	Margin(0, 1, 0, 0).
	Height(1)

var queryModeBlock = lipgloss.NewStyle().
	Foreground(theme.BoxFg).
	Background(theme.QueryModeBoxScanBg).
	Align(lipgloss.Left, lipgloss.Top).
	Padding(0, 1, 0, 1).
	Margin(0, 1, 0, 0).
	Height(1)

var filterModeBlock = lipgloss.NewStyle().
	Foreground(theme.BoxFg).
	Background(theme.FilterBoxBg).
	Align(lipgloss.Left, lipgloss.Top).
	Padding(0, 1, 0, 1).
	Margin(0, 1, 0, 0).
	Height(1)

var pageSuspendBlock = lipgloss.NewStyle().
	Strikethrough(true).
	Foreground(theme.BoxFg).
	Background(theme.PageSuspendBoxBg).
	Align(lipgloss.Left, lipgloss.Top).
	Padding(0, 1, 0, 1).
	Margin(0, 1, 0, 0).
	Height(1)

var helpBlock = lipgloss.NewStyle().
	Foreground(theme.BoxFg).
	Background(theme.HelpBoxBg).
	Align(lipgloss.Left, lipgloss.Top).
	Padding(0, 1, 0, 1).
	Margin(0, 1, 0, 0).
	Height(1)

type Model struct {
	// ActiveView determines tea.Msg forwarding
	activeView View

	// current query-mode
	QueryMode messages.ItemsQueryMode

	// badge flags
	FiltersEnabled  bool
	PagingSuspended bool

	// key bindings
	KeyMap KeyMap

	// window attributes
	window struct {
		width  int
		height int
	}

	// dialogs
	dialogs struct {
		open             bool
		help             *dialogs.Help
		region           *dialogs.Regions
		columnVisibility *dialogs.ColumnVis
		columnSorting    *dialogs.ColumnSorting
		columnTransform  *dialogs.TransformDialog
		columnWidth      *dialogs.WidthDialog
		scanParams       *dialogs.ScanDialog
		queryParams      *dialogs.Queryialog
		filterParams     *dialogs.FilterDialog
		copy             *dialogs.CopyDialog
		mfa              *dialogs.MFA
		active           Dialog

		notification []*dialogs.NotificationDialog
	}

	// top-level context
	ctx context.Context

	// logger
	logger *slog.Logger

	// shared config
	config *appconfig.Config

	// views
	tableSelection *tablesview.TableSelection
	itemselection  *itemsview.ItemSelection

	// help
	Help help.Model

	// additional options
	options options
}

type options struct {
	InitialError error
}

type Option func(*options)

func WithInitialErrorNotification(err error) Option {
	return func(o *options) {
		o.InitialError = err
	}
}

func NewModel(ctx context.Context, cfg appconfig.Config, opts ...Option) Model {
	m := Model{
		ctx:    ctx,
		logger: cfg.Logger.With(slog.String(logging.ComponentKey, "UI_HOME")),
		config: &cfg,

		activeView: tables_view,
		Help:       help.New(),
	}

	for _, o := range opts {
		o(&m.options)
	}

	m.KeyMap = DefaultKeyMap()

	inheritedKeys := []keymaps.AdditionalKey{
		{
			Binding:   m.KeyMap.ForceQuit,
			Call:      tea.Quit,
			ShortHelp: true,
		}, {
			Binding:   m.KeyMap.Help,
			Call:      m.SignalOpenHelpDialog(),
			ShortHelp: false,
		},
	}

	{ // mfa dialog
		m.dialogs.mfa = dialogs.NewMFADialog(ctx, cfg.Logger, cfg.MFACredentialC)
	}

	{ // views
		m.tableSelection = tablesview.NewTableSelectionView(ctx, &cfg, tablesview.WithAdditionalKeys(keymaps.AdditionalKeys(inheritedKeys)))
		m.itemselection = itemsview.NewItemSelectionView(ctx, &cfg, itemsview.WithAdditionalKeys(keymaps.AdditionalKeys(inheritedKeys)))
	}

	{ // table view bound dialogs
		tableViewDialogKeys := m.tableSelection.DialogKeyMaps()
		m.dialogs.help = dialogs.NewHelp(m.tableSelection, m.itemselection, DialogCloseKeymapFrom(m.KeyMap.Help))
		m.dialogs.region = dialogs.NewRegionsDialog(ctx, cfg.Logger, m.config.AvailableRegions, m.config.StarredRegions, m.config.Region, DialogCloseKeymapFrom(tableViewDialogKeys.RegionDialog))
	}

	{ // items view bound dialogs
		itemViewDialogKeys := m.itemselection.DialogKeyMaps()
		m.dialogs.columnVisibility = dialogs.NewColumnVisibilityDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.ColumnVisibility))
		m.dialogs.columnSorting = dialogs.NewColumnSortingDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.ColumnSorting))
		m.dialogs.columnTransform = dialogs.NewTransformDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.ColumnTransform))
		m.dialogs.columnWidth = dialogs.NewWidthDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.ColumnWidth))
		m.dialogs.scanParams = dialogs.NewScanDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.ScanParams))
		m.dialogs.queryParams = dialogs.NewQueryDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.QueryParams))
		m.dialogs.filterParams = dialogs.NewFilterDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.FilterParams))
		m.dialogs.copy = dialogs.NewCopyDialog(ctx, cfg.Logger, DialogCloseKeymapFrom(itemViewDialogKeys.Copy))
	}

	return m
}

func (m Model) Init() tea.Cmd {
	m.logger.Info("initialising...")
	var cmds []tea.Cmd
	cmds = append(cmds, tea.RequestBackgroundColor)

	// notify user of any initialisation errors
	if err := m.options.InitialError; err != nil {
		cmds = append(cmds, errorMsg("", err))
		m.options.InitialError = nil // reset
	}

	// load a new aws client
	cfg, err := aws.LoadAWSConfig(m.ctx, m.config.Region, m.config.Profile, m.config.MFACredentialCB)
	if err != nil {
		m.logger.Error("encountered error initialising AWS config", slog.Any("error", err))
		cmds = append(cmds, errorMsg("", err))
		return tea.Batch(cmds...)
	}

	// custom initialisation
	if m.config.Initialisation.Table != "" {
		cmds = append(cmds, func() tea.Msg {
			return messages.SwitchView{
				OldView: messages.Table_selection,
				NewView: messages.Item_selection,
			}
		})
	}

	// set and reinitialise
	m.config.Client = aws.NewDynamoDBClient(cfg, m.config.URL)
	cmds = append(cmds, m.tableSelection.Init())
	cmds = append(cmds, m.itemselection.Init())

	// reset initialisation parameters after setup in case model.Init methods
	// are manually triggered by user
	m.config.Initialisation = appconfig.Initialisation{}

	m.logger.Info("initialisation complete")
	return tea.Batch(cmds...)
}

// errorMsg returns a command to open a new error notification dialog
func errorMsg(msg string, err error) tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleNotificationDialog{Msg: msg, Error: err}
	}
}

// update handles the message and proceeds to forward it to the model's children
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.KeyMap.ForceQuit) {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m, cmd = m.updateTheme(msg)
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
	case messages.ToggleColumnTransform:
		m, cmd = m.ToggleColumnTransformDialog()
	case messages.ToggleColumnWidthDialog:
		m, cmd = m.ToggleColumnWidthDialog()
	case messages.ToggleScanParameters:
		m, cmd = m.ToggleScanParametersDialog()
	case messages.ToggleQueryParameters:
		m, cmd = m.ToggleQueryParametersDialog()
	case messages.ToggleFilterParameters:
		m, cmd = m.ToggleFilterParametersDialog()
	case messages.ToggleColumnCopy:
		m, cmd = m.ToggleCopyDialog()
	case messages.ToggleNotificationDialog:
		m, cmd = m.ToggleNotificationDialog(msg)
	case messages.NotificationExpired:
		m, cmd = m.HandleExpiredError(msg)
	case messages.NotificationTick:
		m, cmd = m.handleErrorTick(msg)
	case messages.SwitchRegion:
		m, cmd = m.switchRegion(msg.OldRegion, msg.NewRegion)
	case messages.SwitchQueryMode:
		m, cmd = m.SwitchQueryMode(msg)
	case messages.TableFiltersEnabled:
		m, cmd = m.UpdateTableFilterMode(msg)
	case messages.TablePaginationSuspended:
		m, cmd = m.UpdatePaginationSuspended(msg)
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
	cmds = append(cmds, m.dialogs.columnTransform.Update(msg))
	cmds = append(cmds, m.dialogs.columnWidth.Update(msg))
	cmds = append(cmds, m.dialogs.scanParams.Update(msg))
	cmds = append(cmds, m.dialogs.queryParams.Update(msg))
	cmds = append(cmds, m.dialogs.filterParams.Update(msg))
	cmds = append(cmds, m.dialogs.copy.Update(msg))
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
		case column_transform_dialog:
			return m, m.dialogs.columnTransform.Update(msg)
		case column_width_dialog:
			return m, m.dialogs.columnWidth.Update(msg)
		case scan_param_dialog:
			return m, m.dialogs.scanParams.Update(msg)
		case query_param_dialog:
			return m, m.dialogs.queryParams.Update(msg)
		case filter_param_dialog:
			return m, m.dialogs.filterParams.Update(msg)
		case copy_dialog:
			return m, m.dialogs.copy.Update(msg)
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
		m.logger.Error("could not identify active view", slog.Int("view_id", int(m.activeView)))
	}

	return m, nil
}

func (m Model) updateTheme(msg tea.BackgroundColorMsg) (Model, tea.Cmd) {
	theme.UpdateTheme(msg.IsDark(), m.config.ThemeOverrides)
	regionBlock = regionBlock.Foreground(theme.BoxFg).Background(theme.RegionBoxBg)
	queryModeBlock = queryModeBlock.Foreground(theme.BoxFg).Background(theme.QueryModeBoxScanBg)
	filterModeBlock = filterModeBlock.Foreground(theme.BoxFg).Background(theme.FilterBoxBg)
	pageSuspendBlock = pageSuspendBlock.Foreground(theme.BoxFg).Background(theme.RegionBoxBg)
	helpBlock = helpBlock.Foreground(theme.BoxFg).Background(theme.HelpBoxBg)
	return m, nil
}

func (m Model) SwitchQueryMode(msg messages.SwitchQueryMode) (Model, tea.Cmd) {
	m.QueryMode = msg.NewMode
	switch m.QueryMode {
	case messages.ScanMode:
		queryModeBlock = queryModeBlock.Background(theme.QueryModeBoxScanBg)
	case messages.QueryMode:
		queryModeBlock = queryModeBlock.Background(theme.QueryModeBoxQeuryBg)
	}
	return m, nil
}

func (m Model) UpdateTableFilterMode(msg messages.TableFiltersEnabled) (Model, tea.Cmd) {
	m.FiltersEnabled = msg.Enabled
	return m, nil
}

func (m Model) UpdatePaginationSuspended(msg messages.TablePaginationSuspended) (Model, tea.Cmd) {
	m.PagingSuspended = msg.Suspended
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

func (m Model) handleErrorTick(msg messages.NotificationTick) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, d := range m.dialogs.notification {
		cmds = append(cmds, d.Update(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m Model) HandleExpiredError(msg messages.NotificationExpired) (Model, tea.Cmd) {
	if idx := u.FindBy(m.dialogs.notification, func(d *dialogs.NotificationDialog) bool {
		return d != nil && d.ID() == msg.ID
	}); idx >= 0 {
		m.dialogs.notification = slices.Delete(m.dialogs.notification, idx, idx+1)
	}
	return m, nil
}

func (m Model) ToggleNotificationDialog(msg messages.ToggleNotificationDialog) (Model, tea.Cmd) {
	options := []dialogs.Option{}
	if msg.Duration != 0 {
		options = append(options, dialogs.WithDuration(msg.Duration))
	}
	d := dialogs.NewNotificationDialog(msg.Msg, msg.Error, options...)
	m.dialogs.notification = append(m.dialogs.notification, d)
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

func (m Model) ToggleColumnWidthDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != column_width_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = column_width_dialog
	}
	return m, nil
}

func (m Model) ToggleColumnTransformDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != column_transform_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = column_transform_dialog
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

func (m Model) ToggleFilterParametersDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != filter_param_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = filter_param_dialog
	}
	return m, nil

}

func (m Model) ToggleCopyDialog() (Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != copy_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = copy_dialog
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
	var keys []key.Binding
	switch m.activeView {
	case tables_view:
		page = m.tableSelection.View()
		keys = m.tableSelection.ShortHelp()
	case items_view:
		page = m.itemselection.View()
		keys = m.itemselection.ShortHelp()
	}

	// assemble gutter
	region := regionBlock.Render(m.config.Region)
	queryMode := u.Ternary("QUERY", "SCAN", m.QueryMode == messages.QueryMode)
	// TODO: refactor blocks to be managed more directly by view
	query := u.Ternary(queryModeBlock.Render(queryMode), "", m.activeView == items_view)
	filter := u.Ternary(filterModeBlock.Render("FILTER"), "", m.FiltersEnabled && m.activeView == items_view)
	pageSus := u.Ternary(pageSuspendBlock.Render("PAGING"), "", m.PagingSuspended && m.activeView == items_view)
	helpBlock := helpBlock.Render(strings.Join(m.KeyMap.Help.Keys(), ","), " help")
	pregutterLeft := lipgloss.JoinHorizontal(lipgloss.Left, region, query, filter, pageSus, " ")
	pregutterRight := lipgloss.JoinHorizontal(lipgloss.Right, " ", helpBlock)

	m.Help.SetWidth(m.window.width - lipgloss.Width(pregutterLeft) - lipgloss.Width(pregutterRight))
	help := m.Help.ShortHelpView(keys)
	var fill string
	if remainingSpace := m.Help.Width() - lipgloss.Width(help); remainingSpace > 0 {
		fill = u.RepeatString(" ", remainingSpace)
	}

	gutter := lipgloss.JoinHorizontal(lipgloss.Left, pregutterLeft, help, fill, pregutterRight)

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
		case column_transform_dialog:
			dialog = m.dialogs.columnTransform
		case column_width_dialog:
			dialog = m.dialogs.columnWidth
		case scan_param_dialog:
			dialog = m.dialogs.scanParams
		case query_param_dialog:
			dialog = m.dialogs.queryParams
		case filter_param_dialog:
			dialog = m.dialogs.filterParams
		case copy_dialog:
			dialog = m.dialogs.copy
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
	for _, d := range m.dialogs.notification {
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
