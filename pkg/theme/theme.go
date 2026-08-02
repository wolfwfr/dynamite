package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

func init() {
	UpdateTheme(DarkTheme, ThemeOverrides{})
}

var (
	DarkTheme bool = true // default
	c              = lipgloss.Color
)

// UpdateTheme updates colours in response to a `tea.BackgroundColorMsg`. It
// will not touch any colours that have been overridden by the user.
// TODO: tweak theme, light theme & table-name highlights in particular
// TODO: some UI elements do not yet employ colours from theme
func UpdateTheme(isDark bool, overrides ThemeOverrides) {
	DarkTheme = isDark

	// func(light_theme_colour, dark_theme_colour)
	choose := lipgloss.LightDark(DarkTheme)

	// primary palette
	SubtleColour1 = choose(c("#585858"), c("#B0B0B0"))
	SubtleColour2 = choose(c("#5E5E5E"), c("#878787"))
	SubtleColour3 = choose(c("#636363"), c("#636363"))
	SubtleColour4 = choose(c("#878787"), c("#5E5E5E"))
	SubtleColour5 = choose(c("#B0B0B0"), c("#585858"))

	AccentOrange = choose(c("#B8611A"), c("#F58427"))
	AccentBlue = choose(c("#17B2FF"), c("#2381CF"))
	AccentFadedBlue = choose(c("#95A0BA"), c("#415278"))
	AccentDarkBlue = choose(c("#8A9FBA"), c("#244673"))

	// NOTE: applying early here because primary palette can be referenced later on
	overrides.apply()

	// WARN: non-primary-palette colours can ONLY reference primary-palette
	// colours, or define unique ones. This is required to ensure overrides are
	// applied correctly.

	// dialog
	DialogFocusColour = AccentOrange
	DialogUnfocusColour = SubtleColour3
	DialogBorderColour = AccentOrange
	TitleFG = TerminalDefaultColour

	// spinners
	SpinnerTextFg = TerminalDefaultColour
	SpinnerTextBg = nil // transparent
	SpinnerSymbolFg = choose(c("#ff5faf"), c("#ff5faf"))
	SpinnerSymbolBg = nil // transparent

	// list
	ListFocusFg = AccentOrange

	// pane borders
	ViewFocusBorderColour = AccentBlue
	ViewUnFocusBorderColour = AccentFadedBlue

	// table
	TableSelectedBg = AccentDarkBlue
	TableSelectedFg = choose(c("#E6E6E6"), c("#E6E6E6")) // not in active use
	TableBorderFg = SubtleColour5
	TableHeaderFg = TerminalDefaultColour

	// search
	SearchHighlight = choose(c("#317566"), c("#317566"))

	// object parsing colours
	FieldNameFg = SubtleColour1
	NumberFg = AccentOrange
	BoolFg = choose(c("#D9AF2E"), c("#D9AF2E"))
	BytesFg = AccentOrange
	NULLFg = choose(c("#A18975"), c("#A18975"))
	StringFg = choose(c("#196B1C"), c("#a7bc85"))
	TokenFg = SubtleColour4
	ErrorFg = choose(c("#B51010"), c("#B51010"))
	TimestampFg = NumberFg

	// gutter boxes
	RegionBoxBg = choose(c("#80380E"), c("#80380E"))
	FilterBoxBg = choose(c("#681FA1"), c("#681FA1"))
	PageSuspendBoxBg = SubtleColour4
	QueryModeBoxQeuryBg = choose(c("#046645"), c("#046645"))
	QueryModeBoxScanBg = choose(c("#0E3080"), c("#0E3080"))
	QueryModeBoxAdminBg = choose(c("#0E5680"), c("#0E5680"))
	HelpBoxBg = choose(c("#042B19"), c("#042B19"))
	BoxFg = choose(c("#B0B0B0"), TerminalDefaultColour)
	HelpBoxFg = choose(SubtleColour2, SubtleColour2)

	// table matching
	TableHighlightDefault1 = SubtleColour1
	TableHighlightDefault2 = choose(c("#4F6E1E"), c("#a7bc85"))
	TableHighlightDefault3 = choose(c("#19664B"), c("#489C7F"))
	TableHighlightDefault4 = choose(c("#70570A"), c("#B59226"))
	TableHighlightDefault5 = choose(c("#145C49"), c("#668C82"))
	TableHighlightDefault6 = choose(c("#804504"), c("#D97604"))
	TableHighlightDefault7 = choose(c("#7C6755"), c("#A18975"))

	// apply again to ensure all overrides are applied
	overrides.apply()

	BorderStyle = BorderStyle.
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ViewUnFocusBorderColour)

	FocusedBorderStyle = FocusedBorderStyle.
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ViewFocusBorderColour)

	DialogStyle = DialogStyle.
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(DialogBorderColour)
}

// primary palette
var (
	TerminalDefaultColour = c("") // empty

	SubtleColour1 color.Color
	SubtleColour2 color.Color
	SubtleColour3 color.Color
	SubtleColour4 color.Color
	SubtleColour5 color.Color

	AccentOrange    color.Color
	AccentBlue      color.Color
	AccentFadedBlue color.Color
	AccentDarkBlue  color.Color
)

var (
	// dialog
	DialogFocusColour   color.Color
	DialogUnfocusColour color.Color
	DialogBorderColour  color.Color
	TitleFG             color.Color

	// spinners
	SpinnerTextFg   color.Color
	SpinnerTextBg   color.Color
	SpinnerSymbolFg color.Color
	SpinnerSymbolBg color.Color

	// list
	ListFocusFg color.Color

	// pane borders
	ViewFocusBorderColour   color.Color
	ViewUnFocusBorderColour color.Color

	// table
	TableSelectedBg color.Color
	TableSelectedFg color.Color
	TableBorderFg   color.Color
	TableHeaderFg   color.Color

	// search
	SearchHighlight color.Color

	// object parsing colours
	FieldNameFg color.Color
	NumberFg    color.Color
	BoolFg      color.Color
	BytesFg     color.Color
	NULLFg      color.Color
	StringFg    color.Color
	TokenFg     color.Color
	ErrorFg     color.Color
	TimestampFg color.Color

	// gutter boxes
	RegionBoxBg         color.Color
	FilterBoxBg         color.Color
	PageSuspendBoxBg    color.Color
	QueryModeBoxQeuryBg color.Color
	QueryModeBoxScanBg  color.Color
	QueryModeBoxAdminBg color.Color
	HelpBoxBg           color.Color
	BoxFg               color.Color
	HelpBoxFg           color.Color

	// table matching
	TableHighlightDefault1 color.Color
	TableHighlightDefault2 color.Color
	TableHighlightDefault3 color.Color
	TableHighlightDefault4 color.Color
	TableHighlightDefault5 color.Color
	TableHighlightDefault6 color.Color
	TableHighlightDefault7 color.Color
)

// styles
var (
	BorderStyle        lipgloss.Style
	FocusedBorderStyle lipgloss.Style
	DialogStyle        lipgloss.Style
)
