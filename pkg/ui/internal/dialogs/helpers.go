package dialogs

import "charm.land/lipgloss/v2"

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
