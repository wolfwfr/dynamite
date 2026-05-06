package styles

import "charm.land/lipgloss/v2"

var (
	DialogFocusColour   = lipgloss.Color("#F58427")
	DialogUnfocusColour = lipgloss.Color("#636363")
	DialogBorderColour  = lipgloss.Color("#F58427")

	ViewFocusBorderColour   = lipgloss.Color("#2381CF")
	ViewUnFocusBorderColour = lipgloss.Color("#415278")

	SubtleColour  = lipgloss.Color("#B0B0B0")
	SubtleColour2 = lipgloss.Color("#878787")

	RegionBoxBg         = lipgloss.Color("#80380E")
	QueryModeBoxQeuryBg = lipgloss.Color("#046645")
	QueryModeBoxScanBg  = lipgloss.Color("#0E3080")
	QueryModeBoxAdminBg = lipgloss.Color("#0E5680")

	SearchFg = lipgloss.Color("#4F4F4F")

	BorderStyle = lipgloss.NewStyle().
			Align(lipgloss.Left, lipgloss.Top).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ViewUnFocusBorderColour)

	FocusedBorderStyle = lipgloss.NewStyle().
				Align(lipgloss.Left, lipgloss.Top).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ViewFocusBorderColour)

	DialogStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(DialogBorderColour)
)
