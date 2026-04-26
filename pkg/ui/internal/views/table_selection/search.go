package tableselection

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/internal/fuzzy"
)

const (
	searchHeight int = 2
)

var (
	searchBox = lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4F4F4F")).
		PaddingLeft(2).
		Height(searchHeight)
)

type search struct {
	f       fuzzy.FilterFunc
	enabled bool // enabled determines whether searchbox is visible
	active  bool // active determines whether searchbox is actively receiving input
	input   textinput.Model
}
