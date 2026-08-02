package tableselection

import (
	"log/slog"
	"regexp"

	"charm.land/lipgloss/v2"
)

var regexMatchs []*regexp.Regexp

func (m *tableSelectionPane) initialiseRegex(exprs []string) {
	for i, expr := range exprs {
		c, err := regexp.Compile(expr)
		if err != nil {
			m.logger.Error("failed to parse regular expression",
				slog.Int("index", i),
				slog.String("expression", expr),
				slog.Any("error", err),
			)
			continue
		}
		regexMatchs = append(regexMatchs, c)
	}
}

func compileMatchedStyles(name string, highlights []lipgloss.Style, defaultStyle lipgloss.Style) ([]string, []lipgloss.Style) {
	if name == "" || len(regexMatchs) == 0 {
		return []string{name}, []lipgloss.Style{defaultStyle}
	}

	var captures []string
	for _, re := range regexMatchs {
		captures = re.FindStringSubmatch(name)
		if len(captures) > 0 {
			break
		}
	}
	if len(captures) == 0 {
		return []string{name}, []lipgloss.Style{defaultStyle}
	}

	// index 0 always contains the full match
	captures = captures[1:]
	styles := make([]lipgloss.Style, len(captures))

	j := -1
	for i := range captures {
		j++
		if j >= len(highlights) {
			j = 0
		}
		styles[i] = highlights[j]
	}

	return captures, styles
}
