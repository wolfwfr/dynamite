package styles

import "charm.land/lipgloss/v2"

// TODO: enable configurability through config file
// TODO: prepare basic dark & light theme
var (
	DialogFocusColour   = lipgloss.Color("#F58427")
	DialogUnfocusColour = lipgloss.Color("#636363")
	DialogBorderColour  = lipgloss.Color("#F58427")

	ViewFocusBorderColour   = lipgloss.Color("#2381CF")
	ViewUnFocusBorderColour = lipgloss.Color("#415278")

	TableSelectedBg = lipgloss.Color("#244673")
	TableSelectedFg = lipgloss.Color("#E6E6E6")
	TableDefaultFg  = lipgloss.Color("240")

	SubtleColour  = lipgloss.Color("#B0B0B0")
	SubtleColour2 = lipgloss.Color("#878787")
	SubtleColour3 = lipgloss.Color("#5E5E5E")

	FieldNameColour = lipgloss.Color("#B0B0B0")
	NumberColour    = lipgloss.Color("#F58427")
	BoolColour      = lipgloss.Color("#F58427")
	BytesColour     = lipgloss.Color("#F58427")
	NULLColour      = lipgloss.Color("#F58427")
	StringColour    = lipgloss.Color("#a7bc85")
	TokenColour     = SubtleColour3
	ErrorColour     = lipgloss.Color("#B51010")

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
