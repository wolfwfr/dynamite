package styles

import "charm.land/lipgloss/v2"

var (
	BorderStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Top).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#415278"))

	FocusedBorderStyle = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Top).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#2381CF"))

	DialogStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F58427"))
)
