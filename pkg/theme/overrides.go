package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// adding yaml flags for convenience, don't want to define everything again in
// configfile package
type ThemeOverrides struct {
	// primary palette
	TerminalDefaultColour string `yaml:"terminal_default_color"`

	SubtleColour1 string `yaml:"subtle_color1"`
	SubtleColour2 string `yaml:"subtle_color2"`
	SubtleColour3 string `yaml:"subtle_color3"`
	SubtleColour4 string `yaml:"subtle_color4"`
	SubtleColour5 string `yaml:"subtle_color5"`

	AccentOrange    string `yaml:"accent_orange"`
	AccentBlue      string `yaml:"accent_blue"`
	AccentFadedBlue string `yaml:"accent_faded_blue"`
	AccentDarkBlue  string `yaml:"accent_dark_blue"`

	// dialog
	DialogFocusColour   string `yaml:"dialog_focus_color"`
	DialogUnfocusColour string `yaml:"dialog_unfocus_color"`
	DialogBorderColour  string `yaml:"dialog_border_color"`
	TitleFG             string `yaml:"title_fg"`

	// spinners
	SpinnerTextFg   string `yaml:"spinner_text_fg"`
	SpinnerTextBg   string `yaml:"spinner_text_bg"`
	SpinnerSymbolFg string `yaml:"spinner_symbol_fg"`
	SpinnerSymbolBg string `yaml:"spinner_symbol_bg"`

	// list
	ListFocusFg string `yaml:"list_focus_fg"`

	// pane borders
	ViewFocusBorderColour   string `yaml:"view_focus_border_color"`
	ViewUnFocusBorderColour string `yaml:"view_un_focus_border_color"`

	// table
	TableSelectedBg string `yaml:"table_selected_bg"`
	TableSelectedFg string `yaml:"table_selected_fg"`
	TableBorderFg   string `yaml:"table_border_fg"`
	TableHeaderFg   string `yaml:"table_header_fg"`

	// search
	SearchHighlight string `yaml:"search_highlight"`

	// object parsing colours
	FieldNameFg string `yaml:"field_namefg"`
	NumberFg    string `yaml:"number_fg"`
	BoolFg      string `yaml:"bool_fg"`
	BytesFg     string `yaml:"bytes_fg"`
	NULLFg      string `yaml:"nullfg"`
	StringFg    string `yaml:"string_fg"`
	TokenFg     string `yaml:"token_fg"`
	ErrorFg     string `yaml:"error_fg"`
	TimestampFg string `yaml:"timestamp_fg"`

	// gutter boxes
	RegionBoxBg         string `yaml:"region_box_bg"`
	FilterBoxBg         string `yaml:"filter_box_bg"`
	PageSuspendBoxBg    string `yaml:"page_suspend_box_bg"`
	QueryModeBoxQeuryBg string `yaml:"query_mode_box_qeury_bg"`
	QueryModeBoxScanBg  string `yaml:"query_mode_box_scan_bg"`
	QueryModeBoxAdminBg string `yaml:"query_mode_box_admin_bg"`
	HelpBoxBg           string `yaml:"help_box_bg"`
	BoxFg               string `yaml:"box_fg"`
	HelpBoxFg           string `yaml:"help_box_fg"`

	// table matching
	TableHighlightDefault0 string `yaml:"table_highlight_default0"`
	TableHighlightDefault1 string `yaml:"table_highlight_default1"`
	TableHighlightDefault2 string `yaml:"table_highlight_default2"`
	TableHighlightDefault3 string `yaml:"table_highlight_default3"`
	TableHighlightDefault4 string `yaml:"table_highlight_default4"`
	TableHighlightDefault5 string `yaml:"table_highlight_default5"`
	TableHighlightDefault6 string `yaml:"table_highlight_default6"`
	TableHighlightDefault7 string `yaml:"table_highlight_default7"`
}

func (o ThemeOverrides) apply() {
	maybeOverride := func(override string, current color.Color) color.Color {
		switch override {
		case "":
			return current
		case "transparent", "nil":
			return lipgloss.NoColor{}
		default:
			return c(override)
		}
	}

	TerminalDefaultColour = maybeOverride(o.TerminalDefaultColour, TerminalDefaultColour)

	SubtleColour1 = maybeOverride(o.SubtleColour1, SubtleColour1)
	SubtleColour2 = maybeOverride(o.SubtleColour2, SubtleColour2)
	SubtleColour3 = maybeOverride(o.SubtleColour3, SubtleColour3)
	SubtleColour4 = maybeOverride(o.SubtleColour4, SubtleColour4)
	SubtleColour5 = maybeOverride(o.SubtleColour5, SubtleColour5)

	AccentOrange = maybeOverride(o.AccentOrange, AccentOrange)
	AccentBlue = maybeOverride(o.AccentBlue, AccentBlue)
	AccentFadedBlue = maybeOverride(o.AccentFadedBlue, AccentFadedBlue)
	AccentDarkBlue = maybeOverride(o.AccentDarkBlue, AccentDarkBlue)

	DialogFocusColour = maybeOverride(o.DialogFocusColour, DialogFocusColour)
	DialogUnfocusColour = maybeOverride(o.DialogUnfocusColour, DialogUnfocusColour)
	DialogBorderColour = maybeOverride(o.DialogBorderColour, DialogBorderColour)
	TitleFG = maybeOverride(o.TitleFG, TitleFG)

	SpinnerTextFg = maybeOverride(o.SpinnerTextFg, SpinnerTextFg)
	SpinnerTextBg = maybeOverride(o.SpinnerTextBg, SpinnerTextBg)
	SpinnerSymbolFg = maybeOverride(o.SpinnerSymbolFg, SpinnerSymbolFg)
	SpinnerSymbolBg = maybeOverride(o.SpinnerSymbolBg, SpinnerSymbolBg)

	ListFocusFg = maybeOverride(o.ListFocusFg, ListFocusFg)

	ViewFocusBorderColour = maybeOverride(o.ViewFocusBorderColour, ViewFocusBorderColour)
	ViewUnFocusBorderColour = maybeOverride(o.ViewUnFocusBorderColour, ViewUnFocusBorderColour)

	TableSelectedBg = maybeOverride(o.TableSelectedBg, TableSelectedBg)
	TableSelectedFg = maybeOverride(o.TableSelectedFg, TableSelectedFg)
	TableBorderFg = maybeOverride(o.TableBorderFg, TableBorderFg)
	TableHeaderFg = maybeOverride(o.TableHeaderFg, TableHeaderFg)

	SearchHighlight = maybeOverride(o.SearchHighlight, SearchHighlight)

	FieldNameFg = maybeOverride(o.FieldNameFg, FieldNameFg)
	NumberFg = maybeOverride(o.NumberFg, NumberFg)
	BoolFg = maybeOverride(o.BoolFg, BoolFg)
	BytesFg = maybeOverride(o.BytesFg, BytesFg)
	NULLFg = maybeOverride(o.NULLFg, NULLFg)
	StringFg = maybeOverride(o.StringFg, StringFg)
	TokenFg = maybeOverride(o.TokenFg, TokenFg)
	ErrorFg = maybeOverride(o.ErrorFg, ErrorFg)
	TimestampFg = maybeOverride(o.TimestampFg, TimestampFg)

	RegionBoxBg = maybeOverride(o.RegionBoxBg, RegionBoxBg)
	FilterBoxBg = maybeOverride(o.FilterBoxBg, FilterBoxBg)
	PageSuspendBoxBg = maybeOverride(o.PageSuspendBoxBg, PageSuspendBoxBg)
	QueryModeBoxQeuryBg = maybeOverride(o.QueryModeBoxQeuryBg, QueryModeBoxQeuryBg)
	QueryModeBoxScanBg = maybeOverride(o.QueryModeBoxScanBg, QueryModeBoxScanBg)
	QueryModeBoxAdminBg = maybeOverride(o.QueryModeBoxAdminBg, QueryModeBoxAdminBg)
	HelpBoxBg = maybeOverride(o.HelpBoxBg, HelpBoxBg)
	BoxFg = maybeOverride(o.BoxFg, BoxFg)
	HelpBoxFg = maybeOverride(o.HelpBoxFg, HelpBoxFg)

	TableHighlightDefault0 = maybeOverride(o.TableHighlightDefault0, TableHighlightDefault0)
	TableHighlightDefault1 = maybeOverride(o.TableHighlightDefault1, TableHighlightDefault1)
	TableHighlightDefault2 = maybeOverride(o.TableHighlightDefault2, TableHighlightDefault2)
	TableHighlightDefault3 = maybeOverride(o.TableHighlightDefault3, TableHighlightDefault3)
	TableHighlightDefault4 = maybeOverride(o.TableHighlightDefault4, TableHighlightDefault4)
	TableHighlightDefault5 = maybeOverride(o.TableHighlightDefault5, TableHighlightDefault5)
	TableHighlightDefault6 = maybeOverride(o.TableHighlightDefault6, TableHighlightDefault6)
	TableHighlightDefault7 = maybeOverride(o.TableHighlightDefault7, TableHighlightDefault7)
}
