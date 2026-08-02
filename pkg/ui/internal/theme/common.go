package theme

import (
	"charm.land/lipgloss/v2"
)

// TODO: enable configurability through config file
// TODO: prepare basic dark & light theme

// primary palette
var (
	TerminalDefaultColour = lipgloss.Color("") // empty

	SubtleColour1 = lipgloss.Color("#B0B0B0")
	SubtleColour2 = lipgloss.Color("#878787")
	SubtleColour3 = lipgloss.Color("#636363")
	SubtleColour4 = lipgloss.Color("#5E5E5E")
	SubtleColour5 = lipgloss.Color("#585858")

	AccentOrange    = lipgloss.Color("#F58427")
	AccentBlue      = lipgloss.Color("#2381CF")
	AccentFadedBlue = lipgloss.Color("#415278")
	AccentDarkBlue  = lipgloss.Color("#244673")
)

var (
	// dialog
	DialogFocusColour   = AccentOrange
	DialogUnfocusColour = SubtleColour3
	DialogBorderColour  = AccentOrange
	TitleFG             = TerminalDefaultColour

	// list
	ListFocusFg = AccentOrange

	// pane borders
	ViewFocusBorderColour   = AccentBlue
	ViewUnFocusBorderColour = AccentFadedBlue

	// table
	TableSelectedBg = AccentDarkBlue
	TableSelectedFg = lipgloss.Color("#E6E6E6") // not in active use
	TableBorderFg   = SubtleColour5
	TableHeaderFg   = TerminalDefaultColour

	// search
	SearchHighlight = lipgloss.Color("#317566")

	// object parsing colours
	FieldNameFg = SubtleColour1
	NumberFg    = AccentOrange
	BoolFg      = lipgloss.Color("#D9AF2E")
	BytesFg     = AccentOrange
	NULLFg      = lipgloss.Color("#A18975")
	StringFg    = lipgloss.Color("#a7bc85")
	TokenFg     = SubtleColour4
	ErrorFg     = lipgloss.Color("#B51010")
	TimestampFg = NumberFg

	// gutter boxes
	RegionBoxBg         = lipgloss.Color("#80380E")
	FilterBoxBg         = lipgloss.Color("#681FA1")
	PageSuspendBoxBg    = SubtleColour4
	QueryModeBoxQeuryBg = lipgloss.Color("#046645")
	QueryModeBoxScanBg  = lipgloss.Color("#0E3080")
	QueryModeBoxAdminBg = lipgloss.Color("#0E5680")
	HelpBoxBg           = lipgloss.Color("#042B19")
	BoxFg               = TerminalDefaultColour

	// table matching
	TableHighlightDefault1 = SubtleColour1
	TableHighlightDefault2 = lipgloss.Color("#a7bc85")
	TableHighlightDefault3 = lipgloss.Color("#489C7F")
	TableHighlightDefault4 = lipgloss.Color("#B59226")
	TableHighlightDefault5 = lipgloss.Color("#668C82")
	TableHighlightDefault6 = lipgloss.Color("#D97604")
	TableHighlightDefault7 = lipgloss.Color("#A18975")
)

var (
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
