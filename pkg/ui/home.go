package ui

import (
	"context"
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/aws"
	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	itemselection "github.com/wolfwfr/dynamite/pkg/ui/internal/views/item_selection"
	tableselection "github.com/wolfwfr/dynamite/pkg/ui/internal/views/table_selection"
)

type View int

const (
	table_selection View = iota
	item_selection
)

type Model struct {
	// ActiveView determines tea.Msg forwarding
	activeView View

	// awaitingInput enables/disables letter-based-keymaps
	// TODO: consider handling all keymaps, including global, in views
	awaitingInput bool

	KeyMap *KeyMap

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
	return Model{
		ctx:    ctx,
		config: &cfg,

		KeyMap: DefaultKeyMap(),

		activeView:     table_selection,
		tableSelection: tableselection.NewTableSelectionView(ctx, &cfg),
		itemselection:  itemselection.NewItemSelectionView(ctx, &cfg),

		Help: help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.config.Client == nil {
		cfg, err := aws.LoadAWSConfig(m.ctx, m.config.Region, m.config.Profile)
		if err != nil {
			// TODO: handling
		}
		m.config.Client = dynamodb.NewClient(cfg)
	}
	var cmds []tea.Cmd
	cmds = append(cmds, m.tableSelection.Init())
	cmds = append(cmds, m.itemselection.Init())
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch s := msg.String(); s {
		case "ctrl+c", "q":
			if s != "q" || !m.awaitingInput {
				return m, tea.Quit
			}
		}
	}

	switch msg := msg.(type) {
	case messages.SwitchView:
		return m.handleSwitchView(msg)
	case tea.WindowSizeMsg:
		m.Help.SetWidth(msg.Width)
	}

	return m.forward(msg)
}

func (m Model) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		cmds := []tea.Cmd{}
		cmds = append(cmds, m.tableSelection.Update(msg))
		cmds = append(cmds, m.itemselection.Update(msg))
		return m, tea.Batch(cmds...)
	}

	if msg, ok := msg.(messages.SelectTable); ok {
		return m, m.itemselection.Update(msg)
	}

	switch m.activeView {
	case table_selection:
		return m.handleTableSelectionMode(msg)
	case item_selection:
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
	return m, nil
}

func (m Model) View() tea.View {
	var str string
	var help string
	switch m.activeView {
	case table_selection:
		str = m.tableSelection.View()
		help = m.Help.ShortHelpView(m.augmentHelp(m.tableSelection.ShortHelp()))
	case item_selection:
		str = m.itemselection.View()
		help = m.Help.ShortHelpView(m.augmentHelp(m.itemselection.ShortHelp()))
	}
	str = lipgloss.JoinVertical(lipgloss.Top, str, help)
	v := tea.NewView(str)
	v.AltScreen = true // fullscreen
	return v
}

func (m Model) augmentHelp(help []key.Binding) []key.Binding {
	h := m.KeyMap.ShortHelp()
	res := make([]key.Binding, len(help)+len(h))
	copy(res[:len(help)], help)
	copy(res[len(help):], h)
	return res
}
