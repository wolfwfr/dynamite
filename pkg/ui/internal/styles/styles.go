package styles

import (
	"image/color"
	"maps"
	"strings"

	"charm.land/lipgloss/v2"
)

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

	SearchHighlight = lipgloss.Color("#317566")

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

// TODO: clean up and revise naming
//
// styles for JSON
// essentially a subset of lipgloss.Style
type JSONStyle struct {
	fgColor      color.Color
	bgColor      color.Color
	paddingLeft  int
	paddingRight int
	bold         bool
}

func (j JSONStyle) equals(o JSONStyle) bool {
	return j.fgColor == o.fgColor &&
		j.bgColor == o.bgColor &&
		j.paddingLeft == o.paddingLeft &&
		j.paddingRight == o.paddingRight &&
		j.bold == o.bold
}

func (j JSONStyle) FromLipgloss(st lipgloss.Style) JSONStyle {
	if st.GetForeground() != nil {
		j.fgColor = st.GetForeground()
	}
	if st.GetBackground() != nil {
		j.bgColor = st.GetBackground()
	}
	if st.GetPaddingLeft() > 0 {
		j.paddingLeft = st.GetPaddingLeft()
	}
	if st.GetPaddingRight() > 0 {
		j.paddingRight = st.GetPaddingRight()
	}
	if st.GetBold() {
		j.bold = st.GetBold()
	}
	return j
}

// toLipgloss parses the styling spec to the lipgloss.Style type
func (j JSONStyle) toLipgloss() lipgloss.Style {
	style := lipgloss.NewStyle()
	if j.fgColor != nil {
		style = style.Foreground(j.fgColor)
	}
	if j.bgColor != nil {
		style = style.Background(j.bgColor)
	}
	if j.paddingLeft != 0 {
		style = style.PaddingLeft(j.paddingLeft)
	}
	if j.paddingRight != 0 {
		style = style.PaddingRight(j.paddingRight)
	}
	if j.bold {
		style = style.Bold(true)
	}
	return style
}

func (j JSONStyle) Foreground(c color.Color) JSONStyle {
	j.fgColor = c
	return j
}

func (j JSONStyle) Background(c color.Color) JSONStyle {
	j.bgColor = c
	return j
}

func (j JSONStyle) PaddingLeft(n int) JSONStyle {
	j.paddingLeft = n
	return j
}

func (j JSONStyle) PaddingRight(n int) JSONStyle {
	j.paddingRight = n
	return j
}

func (j JSONStyle) Bold(b bool) JSONStyle {
	j.bold = b
	return j
}

func (j JSONStyle) GetForeground() color.Color {
	return j.fgColor
}

func (j JSONStyle) GetBackground() color.Color {
	return j.bgColor
}

func (j JSONStyle) GetPaddingLeft() int {
	return j.paddingLeft
}

func (j JSONStyle) GetPaddingRight() int {
	return j.paddingRight
}

func (j JSONStyle) GetBold() bool {
	return j.bold
}

// Per rune index styling for each line of a JSON object.
// This exhaustive tracking of styling allows per rune styling adjustments, such
// as when adding search highlighting (e.g. painting background of individual runes)
type JSONLineStyling struct {
	next   int
	styles map[int]JSONStyle
}

func (s JSONLineStyling) Len() int {
	return len(s.styles)
}

func (s JSONLineStyling) GetAt(n int) (JSONStyle, bool) {
	res, ok := s.styles[n]
	return res, ok
}

// copy prevents making changes to the original's reference types (map)
func (s JSONLineStyling) copy() JSONLineStyling {
	cp := make(map[int]JSONStyle, len(s.styles))
	maps.Copy(cp, s.styles)
	s.styles = cp
	return s
}

// Override overrides the given style when it existed at the given index. It
// returns a copy of the JSONLineStyling, so with GetAt one can check whether
// the override was successful.
func (s JSONLineStyling) Override(n int, st JSONStyle) JSONLineStyling {
	s = s.copy()
	if _, ok := s.styles[n]; !ok {
		return s
	}
	s.styles[n] = st
	return s
}

func (s JSONLineStyling) init() JSONLineStyling {
	if len(s.styles) == 0 {
		s.styles = map[int]JSONStyle{}
	}
	return s
}

func (s JSONLineStyling) TrimEnd(n int) JSONLineStyling {
	s = s.init()
	for i := n; i > 0 && len(s.styles) > 0; i-- {
		s.next -= 1
		delete(s.styles, s.next)
	}
	return s
}

func (s JSONLineStyling) UnsetPadding() JSONLineStyling {
	s = s.init()
	s = s.copy()
	for k, st := range s.styles {
		st.paddingLeft = 0
		s.styles[k] = st
	}
	return s
}

func (s JSONLineStyling) SetPaddingAll(n int) JSONLineStyling {
	s = s.init()
	s = s.copy()
	for k, st := range s.styles {
		st.paddingLeft = n
		s.styles[k] = st
	}
	return s
}

func (s JSONLineStyling) SetBackgroundAll(c color.Color) JSONLineStyling {
	s = s.init()
	s = s.copy()
	for k, st := range s.styles {
		st.bgColor = c
		s.styles[k] = st
	}
	return s
}

// SetLeftPaddingFirst adds left-padding only to the first rune in the line,
// essentially prefixing the line with the padding, instead of adding padding
// between each rune
func (s JSONLineStyling) SetLeftPaddingFirst(n int) JSONLineStyling {
	s = s.init()
	s = s.copy()
	st, ok := s.styles[0]
	if !ok {
		return s
	}
	st.paddingLeft = n
	s.styles[0] = st
	return s
}

// SetRightPaddingLast adds right-padding only to the last rune in the line,
// essentially suffixing the line with the padding, instead of adding padding
// between each rune
func (s JSONLineStyling) SetRightPaddingLast(n int) JSONLineStyling {
	s = s.init()
	s = s.copy()
	st, ok := s.styles[s.next-1]
	if !ok {
		return s
	}
	st.paddingRight = n
	s.styles[s.next-1] = st
	return s
}

//	func (s JSONLineStyling) AppendLineStyles(sts []JSONLineStyling) JSONLineStyling {
//		for _, st := range sts {
//			s = s.AppendLineStyle(st)
//		}
//		return s
//	}
func (s JSONLineStyling) AppendLineStyle(st JSONLineStyling) JSONLineStyling {
	s = s.init()
	s = s.copy()
	for i := 0; i < len(st.styles); i++ {
		s.styles[s.next] = st.styles[i]
		s.next += 1
	}
	return s
}

func (s JSONLineStyling) AppendRuneStyle(st JSONStyle) JSONLineStyling {
	s = s.init()
	s = s.copy()
	s.styles[s.next] = st
	s.next += 1
	return s
}

func (s JSONLineStyling) AppendRuneStyleLG(st lipgloss.Style) JSONLineStyling {
	s = s.init()
	s = s.copy()
	return s.AppendRuneStyle(JSONStyle{}.FromLipgloss(st))
}

type stylingOptions struct {
	stringInitialPadding int
}

type JSONStyleOption func(o *stylingOptions)

func WithStringInitialPadding(n int) JSONStyleOption {
	return func(o *stylingOptions) {
		o.stringInitialPadding = n
	}
}

func (s JSONLineStyling) AppendStringStyle(in string, style JSONStyle, opts ...JSONStyleOption) JSONLineStyling {
	options := &stylingOptions{}
	for _, o := range opts {
		o(options)
	}
	s = s.init()
	for i := range []rune(in) {
		st := style
		if i == 0 && options.stringInitialPadding > 0 {
			st.paddingLeft += options.stringInitialPadding
		}
		s = s.AppendRuneStyle(st)
	}
	return s
}

func (s JSONLineStyling) AppendStringStyleLG(in string, style lipgloss.Style, opts ...JSONStyleOption) JSONLineStyling {
	s = s.init()
	return s.AppendStringStyle(in, JSONStyle{}.FromLipgloss(style), opts...)
}

// Render renders the string according to the per rune index styling specified.
// Note that left-padding is included in the styling, inserted lines of JSON
// will have the spaces trimmed.
func (s JSONLineStyling) Render(in string) string {
	in = strings.TrimSpace(in)
	s = s.init()
	b := strings.Builder{}

	var stylingThis string
	runes := []rune(in)

	for i, rn := range runes {
		style, ok := s.styles[i]
		if !ok {
			b.WriteRune(rn)
			continue
		}

		bb := strings.Builder{}
		bb.WriteString(stylingThis)
		bb.WriteRune(rn)
		stylingThis = bb.String()

		if i < len(runes)-1 && style.equals(s.styles[i+1]) {
			continue
		}
		b.WriteString(style.toLipgloss().Render(stylingThis))
		stylingThis = ""
	}
	return b.String()
}

type JSONObjectStyle []JSONLineStyling

func (o JSONObjectStyle) Render(in string) string {
	lines := strings.Split(in, "\n")
	res := strings.Builder{}
	for i := 0; i < len(lines) && i < len(o); i++ {
		res.WriteString(o[i].Render(lines[i]))
		res.WriteString("\n")
	}
	return res.String()
}
