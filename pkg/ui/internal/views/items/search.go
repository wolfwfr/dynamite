package itemselection

import (
	tea "charm.land/bubbletea/v2"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
)

func (m *ItemSelectionPane) SearchInputCallback(col string) []string {
	cols := m.table.GetColumns()
	idx := findColumnByTitle(cols, col)
	return extractColumnFromRows(m.table.GetRows(), idx)
}

func (m *ItemSelectionPane) SearchEmptyInputCallback() tea.Cmd {
	m.table.ResetSearch()
	m.table.SetSearchEnable() // keep enabled
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchResultsCallback(col string, results []search.FilteredItem) tea.Cmd {
	m.table.SetSearchResults(col, results)
	return nil
}

func (m *ItemSelectionPane) SearchResetCallback(searchHeight int) tea.Cmd {
	m.table.ResetSearch()
	m.updateSize()
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchBoxOpensCallback(searchHeight int) tea.Cmd {
	m.table.SetSearchEnable()
	m.updateKeyMaps()
	m.updateSize()
	return nil
}
