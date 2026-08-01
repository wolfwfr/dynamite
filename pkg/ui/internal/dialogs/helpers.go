package dialogs

import (
	"regexp"

	"charm.land/lipgloss/v2"
)

func getPadWidth(s lipgloss.Style) int {
	return s.GetPaddingLeft() + s.GetPaddingRight()
}

func getPadHeight(s lipgloss.Style) int {
	return s.GetPaddingTop() + s.GetPaddingBottom()
}

func getBorderWidth(s lipgloss.Style) int {
	return s.GetBorderLeftSize() + s.GetBorderRightSize()
}

func getBorderHeight(s lipgloss.Style) int {
	return s.GetBorderTopSize() + s.GetBorderBottomSize()
}

var alphanum *regexp.Regexp
var singleChar *regexp.Regexp

func init() {
	var err error
	alphanum, err = regexp.Compile("[a-zA-Z0-9]")
	if err != nil {
		panic(err)
	}
	singleChar, err = regexp.Compile("^.{1,1}$")
	if err != nil {
		panic(err)
	}
}
