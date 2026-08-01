package itemselection

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
)

func (m *ItemSelectionPane) SearchInputCallback(col string) []string {
	m.logger.Log(m.ctx, logging.LevelTrace, "executing search-input-callback", slog.String("column", col))
	cols := m.table.GetColumns()
	idx := findColumnByTitle(cols, col)
	return extractColumnFromRows(m.table.GetRows(), idx)
}

func (m *ItemSelectionPane) SearchEmptyInputCallback() tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace, "executing empty-search-input-callback")
	m.table.ResetSearch()
	m.table.SetSearchEnable() // keep enabled
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchResultsCallback(col string, results []search.FilteredItem) tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace, "executing search-results-callback",
		slog.String("column", col),
		slog.String("results", fmt.Sprintf("%+v", results)),
	)
	m.table.SetSearchResults(col, results)
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchResetCallback(searchHeight int) tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace, "executing search-reset-callback", slog.Int("search_height", searchHeight))
	m.table.ResetSearch()
	m.updateSize()
	m.updateKeyMaps()
	return m.MaybePreviewItem(true)
}

func (m *ItemSelectionPane) SearchBoxOpensCallback(searchHeight int) tea.Cmd {
	m.logger.Log(m.ctx, logging.LevelTrace, "executing search-box-open-callback", slog.Int("search_height", searchHeight))
	m.table.SetSearchEnable()
	m.updateKeyMaps()
	m.updateSize()
	return nil
}
