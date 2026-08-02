package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// small set of text-styling used for per-rune rendering of text
// essentially a subset of lipgloss.Style
type textStyle struct {
	fgColor      color.Color
	bgColor      color.Color
	paddingLeft  int
	paddingRight int
	bold         bool
}

// equals (private) returns the equality of two text-styles by comparing all
// their constituents
func (t textStyle) equals(o textStyle) bool {
	return t.fgColor == o.fgColor &&
		t.bgColor == o.bgColor &&
		t.paddingLeft == o.paddingLeft &&
		t.paddingRight == o.paddingRight &&
		t.bold == o.bold
}

// FromLipgloss parses a lipgloss style to a text-style type
func (t textStyle) FromLipgloss(st lipgloss.Style) textStyle {
	if st.GetForeground() != nil {
		t.fgColor = st.GetForeground()
	}
	if st.GetBackground() != nil {
		t.bgColor = st.GetBackground()
	}
	if st.GetPaddingLeft() > 0 {
		t.paddingLeft = st.GetPaddingLeft()
	}
	if st.GetPaddingRight() > 0 {
		t.paddingRight = st.GetPaddingRight()
	}
	if st.GetBold() {
		t.bold = st.GetBold()
	}
	return t
}

// toLipgloss parses the styling spec to the lipgloss.Style type
func (t textStyle) toLipgloss() lipgloss.Style {
	style := lipgloss.NewStyle()
	if t.fgColor != nil {
		style = style.Foreground(t.fgColor)
	}
	if t.bgColor != nil {
		style = style.Background(t.bgColor)
	}
	if t.paddingLeft != 0 {
		style = style.PaddingLeft(t.paddingLeft)
	}
	if t.paddingRight != 0 {
		style = style.PaddingRight(t.paddingRight)
	}
	if t.bold {
		style = style.Bold(true)
	}
	return style
}

// Foreground sets the text-style's foreground colour
func (t textStyle) Foreground(c color.Color) textStyle {
	t.fgColor = c
	return t
}

// Background sets the text-style's background colour
func (t textStyle) Background(c color.Color) textStyle {
	t.bgColor = c
	return t
}

// PaddingLeft sets the text-style's left-padding
func (t textStyle) PaddingLeft(n int) textStyle {
	t.paddingLeft = n
	return t
}

// PaddingRight sets the text-style's right-padding
func (t textStyle) PaddingRight(n int) textStyle {
	t.paddingRight = n
	return t
}

// Bold sets the text-style's bold status
func (t textStyle) Bold(b bool) textStyle {
	t.bold = b
	return t
}

// GetForeground returns the text-style's current foreground colour
func (t textStyle) GetForeground() color.Color {
	return t.fgColor
}

// GetBackground returns the text-style's current background colour
func (t textStyle) GetBackground() color.Color {
	return t.bgColor
}

// GetPaddingLeft returns the text-style's current left-padding
func (t textStyle) GetPaddingLeft() int {
	return t.paddingLeft
}

// GetPaddingRight returns the text-style's current right-padding
func (t textStyle) GetPaddingRight() int {
	return t.paddingRight
}

// GetBold returns the text-style's current bold status
func (t textStyle) GetBold() bool {
	return t.bold
}
