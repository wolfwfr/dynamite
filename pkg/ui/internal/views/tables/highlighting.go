package tableselection

import (
	"regexp"

	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

var regexMatchs []*regexp.Regexp

func (m *tableSelectionPane) initialiseRegex(exprs []string) {
	for _, expr := range exprs {
		c, err := regexp.Compile(expr)
		if err != nil {
			// TODO: log
			continue
		}
		regexMatchs = append(regexMatchs, c)
	}
}

// TODO: configurability
var highlights []lipgloss.Style = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault1),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault2),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault3),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault4),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault5),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault6),
	lipgloss.NewStyle().Foreground(styles.TableHighlightDefault7),
}

var defaultStyle = lipgloss.NewStyle().Foreground(styles.TableHighlightDefault1)

func compileMatchedStyles(name string) ([]string, []lipgloss.Style) {
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
