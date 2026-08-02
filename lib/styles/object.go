package styles

import "strings"

// ObjectStyle is an assembly of line-styles. It is a convenience type that
// offers rendering of a multi-line text object, with the specified styling.
type ObjectStyle []LineStyle

// Render renders the specified string according to the per rune index styling
// specified in the method receiver. If the string is multi-line, it will be
// split by the newline character ("\n") and each line rendered according to the
// line-styles specified in the object-style. When the number of line-styles and
// lines in the specified string are not equal, Render will apply the styles in
// order and fallback to empty styling when it runs out.
//
// Relating to per-line rendering:
//
// Note that left-padding is included in the styling, inserted lines of JSON
// will have the spaces trimmed. When the number of rune-styles specified in the
// line-style is not equal to the number of runes of the specified string, it
// will apply the styles in order and fallback to empty styling when it runs
// out.
func (o ObjectStyle) Render(in string) string {
	lines := strings.Split(in, "\n")
	res := strings.Builder{}
	for i := 0; i < len(lines); i++ {
		style := LineStyle{}
		if i < len(o) {
			style = o[i]
		}
		res.WriteString(style.Render(lines[i]))
		res.WriteString("\n")
	}
	return res.String()
}
