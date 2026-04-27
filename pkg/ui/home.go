package ui

import (
	"context"
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/help"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/dialogs"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	itemselection "github.com/wolfwfr/dynamite/pkg/ui/internal/views/item_selection"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/keymaps"
	tableselection "github.com/wolfwfr/dynamite/pkg/ui/internal/views/table_selection"
)

type View int
type Dialog int

const (
	table_selection View = iota
	item_selection

	help_dialog    Dialog = iota
	regions_dialog Dialog = iota
)

var regionBlock = lipgloss.NewStyle().
	Background(lipgloss.Color("#80380E")).
	Align(lipgloss.Left, lipgloss.Top).
	PaddingLeft(1).
	PaddingRight(1).
	Height(1)

type Model struct {
	// ActiveView determines tea.Msg forwarding
	activeView View

	window struct {
		width  int
		height int
	}

	// dialogs
	dialogs struct {
		open   bool
		help   *dialogs.Help
		region *dialogs.Regions
		active Dialog
	}

	// top-level context
	ctx context.Context

	// shared config
	config *appconfig.Config

	// views
	tableSelection *tableselection.TableSelection
	itemselection  *itemselection.ItemSelection

	// help
	Help help.Model
}

func NewModel(ctx context.Context, cfg appconfig.Config) Model {
	m := Model{
		ctx:    ctx,
		config: &cfg,

		activeView: table_selection,
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

	tableViewInherit := make([]keymaps.AdditionalKey, len(inheritedKeys)+1)
	copy(tableViewInherit[:len(inheritedKeys)], inheritedKeys)
	copy(tableViewInherit[len(inheritedKeys):], []keymaps.AdditionalKey{
		{
			Binding: km.Regions,
			Call:    m.SignalOpenRegionsDialog(),
		},
	})

	m.tableSelection = tableselection.NewTableSelectionView(ctx, &cfg, tableselection.WithAdditionalKeys(keymaps.AdditionalKeys(tableViewInherit)))
	m.itemselection = itemselection.NewItemSelectionView(ctx, &cfg, itemselection.WithAdditionalKeys(keymaps.AdditionalKeys(inheritedKeys)))

	m.dialogs.help = dialogs.NewHelp(m.tableSelection, m.itemselection)
	m.dialogs.region = dialogs.NewRegionsDialog(m.config.AvailableRegions, m.config.StarredRegions, m.config.Region)

	return m

}

func (m Model) Init() tea.Cmd {
	cfg, err := aws.LoadAWSConfig(m.ctx, m.config.Region, m.config.Profile)
	if err != nil {
		// TODO: handling
	}
	m.config.Client = dynamodb.NewClient(cfg)
	var cmds []tea.Cmd
	cmds = append(cmds, m.tableSelection.Init())
	cmds = append(cmds, m.itemselection.Init())
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchView:
		return m.handleSwitchView(msg)
	case tea.WindowSizeMsg:
		m = m.applySize(msg.Height, msg.Width).(Model)
	case messages.ToggleHelp:
		return m.ToggleHelpDialog()
	case messages.ToggleRegions:
		return m.ToggleRegionsDialog()
	case messages.SwitchRegion:
		return m.switchRegion(msg.OldRegion, msg.NewRegion)
	}

	return m.forward(msg)
}

func (m Model) switchRegion(oldr, newr string) (tea.Model, tea.Cmd) {
	m.config.Region = newr
	return m, m.Init()
}

func (m Model) applySize(height, width int) tea.Model {
	m.Help.SetWidth(width)
	m.window.height = height
	m.window.width = width
	return m
}

func (m Model) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		cmds := []tea.Cmd{}
		cmds = append(cmds, m.tableSelection.Update(msg))
		cmds = append(cmds, m.itemselection.Update(msg))
		cmds = append(cmds, m.dialogs.help.Update(msg))
		cmds = append(cmds, m.dialogs.region.Update(msg))
		return m, tea.Batch(cmds...)
	}

	if msg, ok := msg.(messages.SelectTable); ok {
		return m, m.itemselection.Update(msg)
	}

	if msg, ok := msg.(messages.PreviewItem); ok {
		return m, m.itemselection.Update(msg)
	}

	switch {
	case m.dialogs.open:
		switch m.dialogs.active {
		case help_dialog:
			return m, m.dialogs.help.Update(msg)
		case regions_dialog:
			return m, m.dialogs.region.Update(msg)
		}
	case m.activeView == table_selection:
		return m.handleTableSelectionMode(msg)
	case m.activeView == item_selection:
		return m.handleItemSelectionMode(msg)
	default:
		log.Fatalf("could not identify active view '%d'", int(m.activeView))
	}
	return m, nil
}

func (m Model) handleTableSelectionMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.tableSelection.Update(msg)
}

func (m Model) handleItemSelectionMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.itemselection.Update(msg)
}

func (m Model) handleSwitchView(msg messages.SwitchView) (tea.Model, tea.Cmd) {
	switch msg.NewView {
	case messages.Table_selection:
		m.activeView = table_selection
	case messages.Item_selection:
		m.activeView = item_selection
	}
	return m, m.dialogs.help.Update(msg)
}

func (m Model) ToggleHelpDialog() (tea.Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != help_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = help_dialog
	}
	return m, nil
}

func (m Model) ToggleRegionsDialog() (tea.Model, tea.Cmd) {
	if m.dialogs.open && m.dialogs.active != regions_dialog {
		return m, nil
	}
	m.dialogs.open = !m.dialogs.open
	if m.dialogs.open {
		m.dialogs.active = regions_dialog
	}
	return m, nil
}

type dialog interface {
	View() string
	Width() int
	Height() int
}

func (m Model) View() tea.View {
	var str string
	var help string
	switch m.activeView {
	case table_selection:
		str = m.tableSelection.View()
		help = m.Help.ShortHelpView(m.tableSelection.ShortHelp())
	case item_selection:
		str = m.itemselection.View()
		help = m.Help.ShortHelpView(m.itemselection.ShortHelp())
	}

	region := regionBlock.Render(m.config.Region)
	gutter := lipgloss.JoinHorizontal(lipgloss.Left, region, " ", help)

	str = lipgloss.JoinVertical(lipgloss.Top, str, gutter)

	// dialog compositing
	mainLayer := lipgloss.NewLayer(str)
	c := lipgloss.NewCompositor(mainLayer)
	c.AddLayers(mainLayer)
	if m.dialogs.open {
		var dialog dialog
		switch m.dialogs.active {
		case help_dialog:
			dialog = m.dialogs.help
		case regions_dialog:
			dialog = m.dialogs.region
		}
		dialogLayer := lipgloss.NewLayer(dialog.View()).
			X(m.window.width/2 - dialog.Width()/2).
			Y(m.window.height/2 - dialog.Height()/2)
		c.AddLayers(dialogLayer)
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

func (m Model) SignalOpenRegionsDialog() tea.Cmd {
	return func() tea.Msg {
		return messages.ToggleRegions{}
	}
}
