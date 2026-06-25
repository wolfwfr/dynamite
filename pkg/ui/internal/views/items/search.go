package itemselection

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
)

func (m *ItemSelectionPane) SearchInputCallback(col string) []string {
	cols := m.content.GetColumns()
	idx := findColumnByTitle(cols, col)
	return extractColumnFromRows(m.content.GetRows(), idx)
}

func (m *ItemSelectionPane) SearchEmptyInputCallback() tea.Cmd {
	m.content.ResetSearch()
	m.content.SetSearchEnable() // keep enabled
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchResultsCallback(col string, results []search.FilteredItem) tea.Cmd {
	m.content.SetSearchResults(col, results)
	return nil
}

func (m *ItemSelectionPane) SearchResetCallback(searchHeight int) tea.Cmd {
	m.content.ResetSearch()
	m.updateSize()
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchBoxOpensCallback(searchHeight int) tea.Cmd {
	m.content.SetSearchEnable()
	m.updateKeyMaps()
	m.updateSize()
	return nil
}
