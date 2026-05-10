package styles

import (
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// Per rune index styling for each line of a text object.
// This exhaustive registration of styling allows per rune styling adjustments, such
// as when adding search highlighting (e.g. painting background of individual runes).
type LineStyle struct {
	styles []textStyle
}

// Len returns the number of styles currently recorded for the given line
func (l LineStyle) Len() int {
	return len(l.styles)
}

// copy prevents making changes to the original's reference types (slice)
func (l LineStyle) copy() LineStyle {
	cp := make([]textStyle, len(l.styles))
	copy(cp, l.styles)
	l.styles = cp
	return l
}

// has returns whether the given line contains a style for a rune at the
// specified index
func (l LineStyle) has(n int) bool {
	return n < len(l.styles)
}

// GetAt returns the text-style for the specified rune-index, as well as a
// boolean status signyfying whether a style for the specified index was found.
// An empty text-style is returned when no item was present at the given index.
func (l LineStyle) GetAt(n int) (textStyle, bool) {
	if l.has(n) {
		return l.styles[n], true
	}
	return textStyle{}, false
}

// Override overrides the given style when it existed at the given index. It
// returns a copy of the JSONLineStyling, so with GetAt one can check whether
// the override was successful.
func (l LineStyle) Override(n int, st textStyle) LineStyle {
	if l.has(n) {
		l = l.copy()
		l.styles[n] = st
	}
	return l
}

// TrimEnd removes the specified number of styles from the line's end. If the
// line contains less than the `n` styles, it removes all styles and returns the
// new line-style object.
func (l LineStyle) TrimEnd(n int) LineStyle {
	l = l.copy()
	if len(l.styles) < n {
		n = len(l.styles)
	}
	l.styles = slices.Clip(l.styles[:len(l.styles)-n])
	return l
}

// UnsetPaddingAll removes any left- or right-padding from all the line's styles
// and returns the new line-style object.
func (l LineStyle) UnsetPaddingAll() LineStyle {
	l = l.copy()
	for k, st := range l.styles {
		st.paddingLeft = 0
		st.paddingRight = 0
		l.styles[k] = st
	}
	return l
}

// SetLeftPaddingAll adds left-padding to all the line's styles and returns the
// new line-style object.
func (l LineStyle) SetLeftPaddingAll(n int) LineStyle {
	l = l.copy()
	for k, st := range l.styles {
		st.paddingLeft = n
		l.styles[k] = st
	}
	return l
}

// SetRightPaddingAll adds right-padding to all the line's styles and returns the
// new line-style object.
func (l LineStyle) SetRightPaddingAll(n int) LineStyle {
	l = l.copy()
	for k, st := range l.styles {
		st.paddingRight = n
		l.styles[k] = st
	}
	return l
}

// SetBackgroundAll sets the specified background colour to all the line's
// styles and returns the new line-style object.
func (l LineStyle) SetBackgroundAll(c color.Color) LineStyle {
	l = l.copy()
	for k, st := range l.styles {
		st.bgColor = c
		l.styles[k] = st
	}
	return l
}

// SetLeftPaddingFirst adds left-padding only to the first rune in the line,
// essentially prefixing the line with the padding, instead of adding padding
// between each rune. It returns the new line-style object.
func (l LineStyle) SetLeftPaddingFirst(n int) LineStyle {
	if l.has(0) {
		l = l.copy()
		l.styles[0].paddingLeft = n
	}
	return l
}

// SetRightPaddingLast adds right-padding only to the last rune in the line,
// essentially suffixing the line with the padding, instead of adding padding
// between each rune. It returns the new line-style object.
func (l LineStyle) SetRightPaddingLast(n int) LineStyle {
	if len(l.styles) == 0 {
		return l
	}
	l = l.copy()
	l.styles[len(l.styles)-1].paddingRight = n
	return l
}

// AppendLines takes a slice of LineStyle types and appends all the items
// in order to the method receiver. It then returns the method receiver. Only
// the returned value will be updated, the original values (method receiver and
// argument) will remain unaffected.
func (l LineStyle) AppendLines(sts []LineStyle) LineStyle {
	l = l.copy()
	for _, st := range sts {
		l = l.appendLineNoCopy(st)
	}
	return l
}

// AppendLine takes a LineStyle and appends all its ordered items to the
// method receiver. It then returns the method receiver. Only the returned value
// will be updated, the original values (method receiver and argument) will
// remain unaffected.
func (l LineStyle) AppendLine(st LineStyle) LineStyle {
	l = l.copy()
	return l.appendLineNoCopy(st)
}

// appendLineNoCopy is a private function that improves peformance by
// skipping the copy step. When called without a LineStyle.copy executed prior,
// this function updates the styles map of the original item too. Caution is advised.
func (l LineStyle) appendLineNoCopy(st LineStyle) LineStyle {
	l.styles = append(l.styles, st.styles...)
	return l
}

// AppendRune appends the given style to the line and returns the new line-style
// object.
func (l LineStyle) AppendRune(st textStyle) LineStyle {
	l = l.copy()
	l.styles = append(l.styles, st)
	return l
}

// AppendRuneLG accepts a lipgloss style, takes the relevant text-related
// styling and appends the text-style to the line. It then returns the new
// line-style object.
func (l LineStyle) AppendRuneLG(st lipgloss.Style) LineStyle {
	l = l.copy()
	return l.AppendRune(textStyle{}.FromLipgloss(st))
}

// stringStylingOptions collects styling options to be applied when appending
// styles to the line for a string of a given length.
type stringStylingOptions struct {
	// stringInitialPadding adds the specified padding to the left of the
	// string's first rune
	stringInitialPadding int
}

type StringStyleOption func(o *stringStylingOptions)

func WithStringInitialPadding(n int) StringStyleOption {
	return func(o *stringStylingOptions) {
		o.stringInitialPadding = n
	}
}

// AppendString takes a string and appends the specified style for each of the
// specified string's runes. The specified string is only required to determine
// the rune-count, the particular content is irrelevant. AppendString returns
// the new line-style object.
func (l LineStyle) AppendString(in string, style textStyle, opts ...StringStyleOption) LineStyle {
	options := &stringStylingOptions{}
	for _, o := range opts {
		o(options)
	}
	for i := range []rune(in) {
		st := style
		if i == 0 && options.stringInitialPadding > 0 {
			st.paddingLeft += options.stringInitialPadding
		}
		l = l.AppendRune(st)
	}
	return l
}

// AppendStringLG takes a string and appends the specified lipgloss style for
// each of the specified string's runes, by taiking the relevant text-related
// styling. The specified string is only required to determine
// the rune-count, the particular content is irrelevant. AppendString returns
// the new line-style object.
func (l LineStyle) AppendStringLG(in string, style lipgloss.Style, opts ...StringStyleOption) LineStyle {
	return l.AppendString(in, textStyle{}.FromLipgloss(style), opts...)
}

// Render renders the specified string according to the per rune index styling
// specified in the method receiver. Note that left-padding is included in the
// styling, inserted lines of JSON will have the spaces trimmed.
// When the number of rune-styles specified in the line-style is not equal to
// the number of runes of the specified string, it will apply the styles in
// order and fallback to empty styling when it runs out.
func (l LineStyle) Render(in string) string {
	in = strings.TrimSpace(in)
	b := strings.Builder{}

	var stylingThis string
	runes := []rune(in)

	for i, rn := range runes {
		style := textStyle{}
		if l.has(i) {
			style = l.styles[i]
		}

		bb := strings.Builder{}
		bb.WriteString(stylingThis)
		bb.WriteRune(rn)
		stylingThis = bb.String()

		if i < len(runes)-1 && l.has(i+1) && style.equals(l.styles[i+1]) {
			continue
		}
		b.WriteString(style.toLipgloss().Render(stylingThis))
		stylingThis = ""
	}
	return b.String()
}
