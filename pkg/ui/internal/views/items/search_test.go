package itemselection

import (
	"context"
	"testing"

	gm "go.uber.org/mock/gomock"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/mocks"
)

func TestSearchCallbacks(t *testing.T) {
	var (
		tableARN = "table"
	)

	// factory initialising a new system-under-test
	newSUT := func(t *testing.T) (*ItemSelectionPane, *mocks.MockitemsTable) {
		sut := newItemSelectionPane(context.Background(), &appconfig.Config{})
		sut.selectedTable.TableArn = &tableARN
		sut.applySize(100, 200) // required for underlying table to properly render items

		ctrl := gm.NewController(t)
		m := mocks.NewMockitemsTable(ctrl)
		sut.table = m

		return sut, m
	}

	viewAll := viewoptions.Check{
		SearchAllowed:           true,
		ColumnSortingAllowed:    true,
		ColumnVisibilityAllowed: true,
	}

	t.Run("item-selection-pane should", func(t *testing.T) {
		t.Run("reset search-state when", func(t *testing.T) {
			t.Run("resetting search", func(t *testing.T) {
				sut, tbl := newSUT(t)                                       // init
				tbl.EXPECT().ResetSearch().Times(1)                         // expect call to reset-search
				tbl.EXPECT().GetVisualRows().Return([]table.Row{}).Times(1) // expect call to visual rows for table-render in update-size
				tbl.EXPECT().UpdateSize(gm.Any(), gm.Any()).Times(1)        // expect call to update-size for disappearing search-box
				tbl.EXPECT().GetAllowedOptions().Return(viewAll).Times(1)   // expect call to get-view-options for keymap update
				sut.SearchResetCallback(0)                                  // execute callback
			})
			t.Run("obtaining empty search input", func(t *testing.T) {
				sut, tbl := newSUT(t)                                     // init
				tbl.EXPECT().ResetSearch().Times(1)                       // expect call to reset-search
				tbl.EXPECT().SetSearchEnable().Return(true).Times(1)      // expect call to reset-search
				tbl.EXPECT().GetAllowedOptions().Return(viewAll).Times(1) // expect call to get-view-options for keymap update
				sut.SearchEmptyInputCallback()                            // execute callback
			})
		})
	})
}
