package styles

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestJSONStyles(t *testing.T) {
	first := "{"
	second := "\"key\": \"value\""
	third := "}"

	lines := make([]JSONLineStyling, 3)
	lines[0] = JSONLineStyling{
		next: 1,
		styles: map[int]JSONStyle{
			0: JSONStyle{
				fgColor: lipgloss.Color("160"),
			},
		},
	}
	lines[2] = lines[0]

	lines[1] = JSONLineStyling{
		next: 0,
		styles: map[int]JSONStyle{
			0: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			1: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			2: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			3: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			4: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			5: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			6: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			7: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			8: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			9: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			10: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			11: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			12: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
			13: JSONStyle{
				fgColor: lipgloss.Color("170"),
			},
		},
	}

	renderedLines := make([]string, 3)
	renderedLines[0] = lines[0].Render(first)
	renderedLines[1] = lines[1].Render(second)
	renderedLines[2] = lines[2].Render(third)
	res := strings.Join(renderedLines, "\n")
	panic(fmt.Sprintf("%q", res))
	// without combining styling on consequtive runes
	exp := "\x1b[38;5;160m{\x1b[m\n\x1b[38;5;170m\"\x1b[m\x1b[38;5;170mk\x1b[m\x1b[38;5;170me\x1b[m\x1b[38;5;170my\x1b[m\x1b[38;5;170m\"\x1b[m\x1b[38;5;170m:\x1b[m\x1b[38;5;170m \x1b[m\x1b[38;5;170m\"\x1b[m\x1b[38;5;170mv\x1b[m\x1b[38;5;170ma\x1b[m\x1b[38;5;170ml\x1b[m\x1b[38;5;170mu\x1b[m\x1b[38;5;170me\x1b[m\x1b[38;5;170m\"\x1b[m\n\x1b[38;5;160m}\x1b[m"
	// with combining styling on consiqutive runes
	exp = "\x1b[38;5;160m{\x1b[m\n\x1b[38;5;170m\"key\": \"value\"\x1b[m\n\x1b[38;5;160m}\x1b[m"
	assert.Equal(t, exp, res)
}
