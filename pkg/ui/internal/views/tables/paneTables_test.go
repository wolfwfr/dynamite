package tableselection

import (
	"context"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gm "go.uber.org/mock/gomock"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/tables/mocks"
	tu "github.com/wolfwfr/dynamite/test/testutils"
)

// validate keys on init
func init() {
	validateTableSelectionKeys()
}

var (
	searchKeyValid bool

	searchKey = tea.KeyPressMsg(tea.Key{Text: "/"})
)

// validateTableSelectionKeys ensures keymap-variables accurately inform tests on the validity
// of the keymap configuration in this test file.
func validateTableSelectionKeys() {
	keymap := DefaultTablePaneKeyMap()

	// ensure all are enabled
	keymap.Search.SetEnabled(true)

	// test matching
	searchKeyValid = key.Matches(searchKey, keymap.Search)
}

// fail tests on invalid keys; indicates the keymap has changed
func TestKeyMapValid(t *testing.T) {
	assert.True(t, searchKeyValid)
}
func TestTableSelectionDetails(t *testing.T) {
	var (
		tableARN = "table-arn"
		region   = "us-east-1"
	)

	// factory initialising a new system-under-test
	newSUT := func(m *mocks.MockdynamodbClient) *tableSelectionPane {
		sut := newTableSelectionPane(context.Background(), &appconfig.Config{Region: region})
		sut.dynamodbClient = m
		return sut
	}

	t.Run("table-pane should", func(t *testing.T) {
		t.Run("preview the correct table-details when paging new results", func(t *testing.T) {
			ctrl := gm.NewController(t)                  // init mock controller
			db := mocks.NewMockdynamodbClient(ctrl)      // init mocked DynamoDB client
			sut := newSUT(db)                            // init sut
			tables := []string{"table-A", "table-B"}     // page
			cmd := simpleLoadTables(sut, region, tables) // load tables
			db.EXPECT().
				DescribeTable(gm.Any(), gm.Any(), "table-A").
				Return(&apitypes.DescribeTableResponse{TableArn: &tableARN}, nil).
				Times(1) // expect a call to describe-table
			targets := tu.ExtractMessages[messages.TableDetails](cmd)     // obtain target messages
			require.Len(t, targets, 1)                                    // assert only one table-details message
			assert.EqualValues(t, tableARN, *targets[0].Details.TableArn) // assert correct table being previewed
		})
		t.Run("preview correct table after search", func(t *testing.T) {
			tu.SkipIf(t, !searchKeyValid, "skipping due to outdated search key") // skip if testing-keymap needs updating
			ctrl := gm.NewController(t)                                          // init mock controller
			db := mocks.NewMockdynamodbClient(ctrl)                              // init mocked DynamoDB client
			sut := newSUT(db)                                                    // init sut
			tables := []string{"table-A", "table-B"}                             // page
			cmd := simpleLoadTables(sut, region, tables)                         // load tables
			sut.Update(searchKey)                                                // enable search
			cmd, ok := searchTableSelection(t, sut, "tablB")                     // search for one table
			require.True(t, ok)                                                  // ensure search was successful
			db.EXPECT().
				DescribeTable(gm.Any(), gm.Any(), "table-B").
				Return(&apitypes.DescribeTableResponse{TableArn: &tableARN}, nil).
				Times(1) // expect a call to describe-table
			targets := tu.ExtractMessages[messages.TableDetails](cmd)                  // obtain target messages
			require.NotEmpty(t, targets)                                               // require presence of target messages
			assert.EqualValues(t, tableARN, *targets[len(targets)-1].Details.TableArn) // assert correct table being previewed
		})
	})
}

// convenience function to apply a search query. Returns a boolean that equals
// `true` when the search was successfully applied.
//
// Note that this function does not enable the search!
func searchTableSelection(t *testing.T, receiver *tableSelectionPane, query string) (tea.Cmd, bool) {
	updates := tu.CharsToMessages(query)
	var cmd tea.Cmd // only require the last command
	for _, msg := range updates {
		cmd = receiver.Update(msg)
	}

	// process filtering
	filtermsgs := tu.ExecuteCommand(cmd)

	// feed back filter-results to sut
	var cmds []tea.Cmd
	for _, msg := range filtermsgs {
		cmds = append(cmds, receiver.Update(msg))
	}

	valid := true

	// ensure search is properly enabled and received query
	valid = assert.Contains(t, receiver.search.View(), query) && valid // once false, stays false

	// ensure search results were processed by pane
	valid = assert.True(t, receiver.tablefiltering.enabled) && valid // once false, stays false

	return tea.Batch(cmds...), valid
}

// convenience function to send a page of tabe-names to the system-under-test.
func simpleLoadTables(sut *tableSelectionPane, region string, tables []string) tea.Cmd {
	return sut.Update(messages.TablePageReady{
		Tables:        tables,
		PaginationKey: new(string),
		Err:           nil,
		Region:        region,
	})
}
