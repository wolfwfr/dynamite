package dialogs

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/wolfwfr/dynamite/pkg/theme"
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

func syncInputStylesWithTheme() textinput.Styles {
	s := textinput.DefaultStyles(theme.DarkTheme)

	s.Focused.Text = s.Focused.Text.Foreground(theme.InputFocusedTextFg)
	s.Blurred.Text = s.Blurred.Text.Foreground(theme.InputBlurredTextFg)
	s.Focused.Placeholder = s.Focused.Placeholder.Foreground(theme.InputFocusedPlaceholderFg)
	s.Blurred.Placeholder = s.Blurred.Placeholder.Foreground(theme.InputBlurredPlaceholderFg)

	return s
}
