package main

import (
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type paneID int

const (
	tablePaneID paneID = iota
	detailsPaneID
)

type model struct {
	window struct {
		width  int
		height int
	}
	tablePane     *tablePane
	detailsPane   *detailsPane
	focused       paneID
	awaitingInput bool // disables quit by 'q'
}

var (
	borderStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Top).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#415278"))
	selectedStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Top).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2381CF"))
)

func (m model) renderBorder(paneID paneID, content string) string {
	if m.focused == paneID {
		return selectedStyle.Render(content)
	}
	return borderStyle.Render(content)
}

func (m model) applySize() {
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	selectedStyle = selectedStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

	m.tablePane.applySize(m.window.height-2-3, m.window.width/2-4)
}

func newModel() model {
	m := model{
		tablePane:   newTablePane(columns, rows),
		detailsPane: &detailsPane{},
	}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg := msg.String(); msg {
		case "tab", "shift+tab":
			m = m.moveFocus()
		case "/":
			m.awaitingInput = true
		case "esc":
			m.awaitingInput = false
		case "ctrl+c", "q":
			if msg != "q" || !m.awaitingInput {
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
	}

	var cmd tea.Cmd

	m.applySize()

	switch m.focused {
	case tablePaneID:
		m.tablePane, cmd = m.tablePane.Update(msg)
	case detailsPaneID:
		m.detailsPane, cmd = m.detailsPane.Update(msg)
	}

	return m, cmd
}

func (m model) moveFocus() model {
	m.focused++
	if m.focused > detailsPaneID {
		m.focused = tablePaneID
	}
	return m
}

func (m model) View() tea.View {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderBorder(tablePaneID, m.tablePane.View()),
		m.renderBorder(detailsPaneID, m.detailsPane.View()),
	))
	v := tea.NewView(s.String())
	v.AltScreen = true // fullscreen
	return v
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
