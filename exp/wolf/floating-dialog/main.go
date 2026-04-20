package main

import (
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	window struct {
		width  int
		height int
	}
	detailsPane tea.Model
	dialogOpen  bool
}

var (
	borderStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2381CF"))

	dialogStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F58427")).
			Height(20).
			Width(100)
)

func (m model) applySize() {
	borderStyle = borderStyle.
		Height(m.window.height - 2).
		Width(m.window.width / 2)

}

func newModel() model {
	m := model{}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "o":
			m = m.OpenDialog()

		case "c":
			m = m.CloseDialog()
		}
	case tea.WindowSizeMsg:
		m.window.height = msg.Height
		m.window.width = msg.Width
	}

	m.applySize()

	return m, nil
}

func (m model) OpenDialog() model {
	m.dialogOpen = true
	return m
}
func (m model) CloseDialog() model {
	m.dialogOpen = false
	return m
}

func (m model) View() tea.View {
	s := strings.Builder{}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, borderStyle.Render("This is a list pane"), borderStyle.Render("This is the future details-pane")))
	mainPage := s.String()

	mainLayer := lipgloss.NewLayer(mainPage)
	c := lipgloss.NewCompositor(mainLayer)
	c.AddLayers(mainLayer)
	if m.dialogOpen {
		dialogLayer := lipgloss.NewLayer(dialogStyle.Render("I'm a dialog"))
		c.AddLayers(dialogLayer)
	}

	v := tea.NewView(c.Render())
	v.AltScreen = true // fullscreen
	return v
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
