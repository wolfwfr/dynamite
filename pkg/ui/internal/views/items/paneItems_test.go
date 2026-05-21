package itemselection

import (
	"context"
	"fmt"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

// validate keys on init
func init() {
	validateItemSelectionKeys()
}

var (
	searchKeyValid bool
	queryKeyValid  bool
	scanKeyValid   bool

	searchKey = tea.KeyPressMsg(tea.Key{Text: "/"})
	queryKey  = tea.KeyPressMsg(tea.Key{Text: "Q", Mod: tea.ModShift, Code: 'q', ShiftedCode: 'Q'})
	scanKey   = tea.KeyPressMsg(tea.Key{Text: "S", Mod: tea.ModShift, Code: 's', ShiftedCode: 'S'})
)

// validateItemSelectionKeys ensures keymap-variables accurately inform tests on the validity
// of the keymap configuration in this test file.
func validateItemSelectionKeys() {
	keymap := DefaultItemPaneKeyMap()

	// ensure all are enabled
	keymap.Search.SetEnabled(true)
	keymap.Query.SetEnabled(true)
	keymap.Scan.SetEnabled(true)

	// test matching
	searchKeyValid = key.Matches(searchKey, keymap.Search)
	queryKeyValid = key.Matches(queryKey, keymap.Query)
	scanKeyValid = key.Matches(scanKey, keymap.Scan)
}

// fail tests on invalid keys; indicates the keymap has changed
func TestKeyMapValid(t *testing.T) {
	assert.True(t, searchKeyValid)
	assert.True(t, queryKeyValid)
	assert.True(t, scanKeyValid)
}

type genOpts struct {
	begin int
	idFmt string
}

// generate some simple items with JSON contents (skipping YAML) for testing
// purporses. When used without options, the first argument both counts to the
// number of items being created and their granted ID enumeration (starting at
// 0). When providing an option with 'begin', the function will return 'n -
// begin' number of items, with ID enumeration starting at 'begin'.
func genJSONItems(n int, opts ...genOpts) apitypes.Items {
	res := apitypes.Items{}

	var (
		b     = 0
		idFmt = "id-%d"
	)

	if len(opts) > 0 {
		b = opts[0].begin
		idFmt = opts[0].idFmt
	}

	ln := n - b
	res.JSON = make([]string, ln)
	res.JSONStyled = make([]styles.ObjectStyle, ln)
	res.Raw = make([]map[string]types.AttributeValue, ln)
	res.TableKeys = make([][]apitypes.KeyValue, ln)

	for i := range ln {
		id := fmt.Sprintf(idFmt, b+i)

		res.JSON[i] = `{
  "id": "` + id + `",
  "configured": true
}`
		res.JSONStyled[i] = styles.ObjectStyle{}

		res.Raw[i] = map[string]types.AttributeValue{
			"id":         &types.AttributeValueMemberS{Value: id},
			"configured": &types.AttributeValueMemberBOOL{Value: true},
		}

		res.TableKeys[i] = []apitypes.KeyValue{
			{Key: "id", Value: fmt.Sprintf("\"%s\"", id)},
			{Key: "configured", Value: "true"},
		}
	}

	return res
}

func TestItemSelectionPreviews(t *testing.T) {
	var (
		tableARN = "table"
	)

	// factory initialising a new system-under-test
	newSUT := func() *ItemSelectionPane {
		sut := newItemSelectionPane(context.Background(), &appconfig.Config{})
		sut.previewFormat = JSONformat
		sut.selectedTable.TableArn = &tableARN
		return sut
	}

	t.Run("items-pane should", func(t *testing.T) {
		t.Run("preview the correct item when paging new results", func(t *testing.T) {
			sut := newSUT()                                          // init
			items := genJSONItems(1)                                 // page
			cmd := simpleLoadItems(sut, &tableARN, items)            // load items
			targets := extractMessages[messages.PreviewItem](cmd)    // obtain target messages
			require.Len(t, targets, 1)                               // assert only one preview-item message
			assert.EqualValues(t, items.JSON[0], targets[0].RawItem) // assert correct item being previewed
		})
		t.Run("preview correct item after loading new page that is sorted to table top", func(t *testing.T) {
			sut := newSUT()                                                  // init
			page1 := genJSONItems(3, genOpts{begin: 0})                      // page 1
			page2 := genJSONItems(6, genOpts{begin: 3})                      // page 2
			simpleLoadItems(sut, &tableARN, page1)                           // prepare first page
			simpleSortItems(sut, tableARN, page1.TableKeys[0][0].Key, false) // enable sorting
			cmd := simpleLoadItems(sut, &tableARN, page2)                    // load next page
			targets := extractMessages[messages.PreviewItem](cmd)            // obtain target messages
			require.Len(t, targets, 1)                                       // assert only one preview-item message
			assert.EqualValues(t, page2.JSON[2], targets[0].RawItem)         // assert correct item being previewed
		})
		t.Run("preview correct item after search", func(t *testing.T) {
			skipIf(t, !searchKeyValid, "skipping due to outdated search key")     // skip if testing-keymap needs updating
			sut := newSUT()                                                       // init
			items := genJSONItems(3)                                              // page
			simpleLoadItems(sut, &tableARN, items)                                // load items
			sut.Update(searchKey)                                                 // enable search
			cmd, ok := searchItemSelection(t, sut, "id=id-1")                     // search for first item
			require.True(t, ok)                                                   // ensure search was successful
			targets := extractMessages[messages.PreviewItem](cmd)                 // obtain target messages
			require.NotEmpty(t, targets)                                          // assert only one preview-item message
			assert.EqualValues(t, items.JSON[1], targets[len(targets)-1].RawItem) // assert correct item being previewed
		})
	})
}

func TestItemSelectionCacheInvalidation(t *testing.T) {
	var (
		tableARN = "table"
	)

	cacheKey := func(r, c, cw int) string {
		return fmt.Sprintf("%d-%d-%d", r, c, cw)
	}

	// factory initialising a new system-under-test
	newSUT := func() *ItemSelectionPane {
		sut := newItemSelectionPane(context.Background(), &appconfig.Config{})
		sut.selectedTable.TableArn = &tableARN
		sut.applySize(100, 200) // required for underlying table to properly render items

		// simple delegate that does not consider any kind of styling, only caching
		sut.content.SetFieldDelegate(func(row table.Row, col table.Column, colIdx, rowIdx, colW, padL, padR int, selected bool) string {
			key := cacheKey(rowIdx, colIdx, colW)
			if f, ok := sut.renderCache[key]; ok { // return from cache if found
				return f
			}
			f := row.Fields[colIdx].Value() // no styling for this test
			sut.renderCache[key] = f        // store in cache
			return f                        // return
		})

		return sut
	}

	// mustPassInitialCacheCheck fails the test immediately with a 'failed test
	// initialisation' message. It is intended for test setup, not test result
	// verification.
	mustPassInitialCacheCheck := func(t *testing.T, cache map[string]string, table *table.Model, expChecks int) {
		var n int
		cols := table.Columns()
		for ri, r := range table.VisualRows() {
			for ci, c := range r.Fields {
				n++
				cw := cols[ci].Width
				k := cacheKey(ri, ci, cw)
				if v, ok := cache[k]; !ok {
					t.Fatalf("failed test initialisation: render-cache did not contain entryfor key '%s'", k)
				} else if v != c.Value() {
					t.Fatalf("failed test initialisation: render-cache did not contain expected field for key '%s', expected '%s', got '%s'", k, c.Value(), v)
				}
			}
		}
		if n != expChecks {
			t.Fatalf("failed test initialisation: expected '%d' cached field checks but got '%d'", expChecks, n)
		}
	}

	// assertPassCacheCheck asserts cache contents exactly matches the contents
	// of the table.
	assertPassCacheCheck := func(t *testing.T, cache map[string]string, table *table.Model, expChecks int) {
		var n int
		cols := table.Columns()
		for ri, r := range table.VisualRows() {
			for ci, c := range r.Fields {
				n++
				cw := cols[ci].Width
				k := cacheKey(ri, ci, cw)
				v, ok := cache[k]
				require.True(t, ok, "did not find entry for cache-key '%s'", k)
				require.EqualValues(t, c.Value(), v, "did not find expected value for cache-key '%s'", k)
			}
		}
		assert.EqualValues(t, expChecks, n, "expected '%d' cached field checks, got '%d'", expChecks, n)
	}

	t.Run("item-selection-pane should", func(t *testing.T) {
		t.Run("refresh cache when", func(t *testing.T) {
			t.Run("updating search results", func(t *testing.T) {
				skipIf(t, !searchKeyValid, "skipping test due to outdated search key") // skip if testing-keymap needs updating
				sut := newSUT()                                                        // init
				itemsNotMatching := genJSONItems(3)                                    // first half page
				itemsMatching := genJSONItems(3, genOpts{idFmt: "op%d"})               // second half page
				items := mergeItems(itemsNotMatching, itemsMatching)                   // full page
				simpleLoadItems(sut, &tableARN, items)                                 // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 6*2)        // ensure cache is initialised
				sut.Update(searchKey)                                                  // enable search
				_, ok := searchItemSelection(t, sut, "id=op")                          // matching second 3 only
				require.True(t, ok, "failed to apply search")                          // ensure search is successful
				assertPassCacheCheck(t, sut.renderCache, sut.content, 3*2)             // assert test passed
			})
			t.Run("resetting search", func(t *testing.T) {
				skipIf(t, !searchKeyValid, "skipping test due to outdated search key") // skip if testing-keymap needs updating
				sut := newSUT()                                                        // init
				itemsNotMatching := genJSONItems(3)                                    // first half page
				itemsMatching := genJSONItems(3, genOpts{idFmt: "op%d"})               // second half page
				items := mergeItems(itemsNotMatching, itemsMatching)                   // full page
				simpleLoadItems(sut, &tableARN, items)                                 // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 6*2)        // ensure cache is initialised
				sut.Update(searchKey)                                                  // enable search
				_, ok := searchItemSelection(t, sut, "id=o")                           // matching second 3 only
				require.True(t, ok, "failed to apply search")                          // ensure search is successful
				sut.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))                 // escape once to blur search
				sut.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))                 // escape twice to reset search
				assertPassCacheCheck(t, sut.renderCache, sut.content, 6*2)             // assert test passed
			})
			t.Run("obtaining empty search input", func(t *testing.T) {
				skipIf(t, !searchKeyValid, "skipping test due to outdated search key") // skip if testing-keymap needs updating
				sut := newSUT()                                                        // init
				itemsNotMatching := genJSONItems(3)                                    // first half page
				itemsMatching := genJSONItems(3, genOpts{idFmt: "op%d"})               // second half page
				items := mergeItems(itemsNotMatching, itemsMatching)                   // full page
				simpleLoadItems(sut, &tableARN, items)                                 // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 6*2)        // ensure cache is initialised
				sut.Update(searchKey)                                                  // enable search
				_, ok := searchItemSelection(t, sut, "id=o")                           // matching second 3 only
				require.True(t, ok, "failed to apply search")                          // ensure search is successful
				sut.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))           // backspace once to trigger empty-input
				assertPassCacheCheck(t, sut.renderCache, sut.content, 6*2)             // assert test passed
			})
			t.Run("updating sort parameters", func(t *testing.T) {
				sut := newSUT()
				items := genJSONItems(3)                                         // page
				simpleLoadItems(sut, &tableARN, items)                           // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)  // ensure cache is initialised
				simpleSortItems(sut, tableARN, items.TableKeys[0][0].Key, false) // sort items
				assertPassCacheCheck(t, sut.renderCache, sut.content, 3*2)       // assert test passed
				simpleSortItems(sut, tableARN, "id", true)                       // sort items the other way for good measure
				assertPassCacheCheck(t, sut.renderCache, sut.content, 3*2)       // assert test passed
			})
			t.Run("resetting sort", func(t *testing.T) {
				sut := newSUT()
				items := genJSONItems(3)                                         // page
				simpleLoadItems(sut, &tableARN, items)                           // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)  // ensure cache is initialised
				simpleSortItems(sut, tableARN, items.TableKeys[0][0].Key, false) // sort items
				sut.Update(messages.ColumnSortingReset{tableARN})                // reset sorting
				assertPassCacheCheck(t, sut.renderCache, sut.content, 3*2)       // assert test passed
			})
			t.Run("processing a new page (could change the sorting of existing records)", func(t *testing.T) {
				sut := newSUT()
				items1 := genJSONItems(6, genOpts{begin: 3})                                   // page 1
				items2 := genJSONItems(3, genOpts{begin: 0})                                   // page 2
				simpleLoadItems(sut, &tableARN, items1)                                        // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)                // ensure cache is initialised
				simpleSortItems(sut, tableARN, items1.TableKeys[0][0].Key, true)               // sort items
				top := sut.content.VisualRows()[0].Fields[0].Value()                           // obtain first item for later verification
				simpleLoadItems(sut, &tableARN, items2)                                        // load second page, that puts new items at the top
				newtop := sut.content.VisualRows()[0].Fields[0].Value()                        // obtain first item for later verification
				fatalIf(t, top == newtop, "test initialisation failed: expected new top item") // ensure sorting still in effect
				assertPassCacheCheck(t, sut.renderCache, sut.content, 6*2)                     // assert test passed
			})
		})
		t.Run("clear cache when", func(t *testing.T) {
			t.Run("switching from scan to query", func(t *testing.T) {
				skipIf(t, !queryKeyValid, "skipping test because query-keymap is outdated") // skip when testing-keymap needs updating
				sut := newSUT()                                                             // defaults to scan
				items := genJSONItems(3)                                                    // page
				simpleLoadItems(sut, &tableARN, items)                                      // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)             // ensure cache is initialised
				sut.Update(queryKey)                                                        // switch to query mode
				assert.Empty(t, sut.renderCache)                                            // assert cache has been cleared
			})
			t.Run("switching from query to scan", func(t *testing.T) {
				skipIf(t, !queryKeyValid, "skipping test because query-keymap is outdated") // skip when testing-keymap needs updating
				skipIf(t, !scanKeyValid, "skipping test because scan-keymap is outdated")   // skip when testing-keymap needs updating
				sut := newSUT()                                                             // defaults to scan
				sut.Update(queryKey)                                                        // first enable query mode before switching back
				items := genJSONItems(3)                                                    // page
				simpleLoadItems(sut, &tableARN, items)                                      // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)             // ensure cache is initialised
				sut.Update(scanKey)                                                         // switch to scan mode
				assert.Empty(t, sut.renderCache)                                            // assert cache has been cleared
			})
			t.Run("changing scan parameters", func(t *testing.T) {
				sut := newSUT()                                                 // defaults to scan
				items := genJSONItems(3)                                        // page
				simpleLoadItems(sut, &tableARN, items)                          // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2) // ensure cache is initialised
				simpleChangeScanIndex(sut, tableARN, "new")                     // change index
				assert.Empty(t, sut.renderCache)                                // assert cache has been cleared
			})
			t.Run("changing query parameters", func(t *testing.T) {
				skipIf(t, !queryKeyValid, "skipping test because query-keymap is outdated") // skip when testing-keymap needs updating
				sut := newSUT()                                                             // defaults to scan
				sut.Update(queryKey)                                                        // first enable query mode to accept query settings
				items := genJSONItems(3)                                                    // page
				simpleLoadItems(sut, &tableARN, items)                                      // load items
				mustPassInitialCacheCheck(t, sut.renderCache, sut.content, 3*2)             // ensure cache is initialised
				simpleChangeQParams(sut, tableARN, "new")                                   // change query index
				assert.Empty(t, sut.renderCache)                                            // assert cache has been cleared
			})
		})
	})
}

func TestItemSelectionURLResolution(t *testing.T) {
	var (
		tableARN  = "table"
		tableName = "testing-table"
	)

	// factory initialising a new system-under-test
	newSUT := func() *ItemSelectionPane {
		sut := newItemSelectionPane(context.Background(), &appconfig.Config{})
		sut.selectedTable.TableArn = &tableARN
		sut.selectedTable.TableName = &tableName
		sut.config = &appconfig.Config{}
		sut.config.Region = "us-east-1"
		return sut
	}

	t.Run("item-selection-pane should", func(t *testing.T) {
		t.Run("resolve URLs with correct path-escaping", func(t *testing.T) {
			sut := newSUT()
			items := genJSONItems(1, genOpts{idFmt: "id %d"})                                                                                     // include space in ids
			simpleLoadItems(sut, &tableARN, items)                                                                                                // load items
			url := sut.resolveBrowserURL()                                                                                                        // obtain resolved url
			exp := "https://us-east-1.console.aws.amazon.com/dynamodbv2/home?region=us-east-1#edit-item?itemMode=2&pk=id%200&table=testing-table" // define expectation
			assert.EqualValues(t, exp, url)                                                                                                       // assert expectation
		})
		t.Run("resolve URLs with sort-key", func(t *testing.T) {
			sut := newSUT()
			items := genJSONItems(1, genOpts{idFmt: "id %d"})        // include space
			sut.selectedTable.KeySchema = []types.KeySchemaElement{{ // define primary keys
				AttributeName: &items.TableKeys[0][0].Key,
				KeyType:       types.KeyTypeHash}, {
				AttributeName: &items.TableKeys[0][1].Key,
				KeyType:       types.KeyTypeRange},
			}
			simpleLoadItems(sut, &tableARN, items)                                                                                                        // load items
			url := sut.resolveBrowserURL()                                                                                                                // obtain resolved url
			exp := "https://us-east-1.console.aws.amazon.com/dynamodbv2/home?region=us-east-1#edit-item?itemMode=2&pk=id%200&table=testing-table&sk=true" // define expectation
			assert.EqualValues(t, exp, url)                                                                                                               // assert expectation
		})
	})
}

// convenience function to apply a search query. Returns a boolean that equals
// `true` when the search was successfully applied.
//
// Note that this function does not enable the search!
func searchItemSelection(t *testing.T, receiver *ItemSelectionPane, query string) (tea.Cmd, bool) {
	updates := charsToMessages(query)
	var cmd tea.Cmd // only require the last command
	for _, msg := range updates {
		cmd = receiver.Update(msg)
	}

	// process filtering
	filtermsgs := executeCommand(cmd)

	// feed back filter-results to sut
	var cmds []tea.Cmd
	for _, msg := range filtermsgs {
		cmds = append(cmds, receiver.Update(msg))
	}

	valid := true

	// ensure search is properly enabled and received query
	valid = assert.Contains(t, receiver.search.View(), query) && valid // once false, stays false

	// ensure search results were processed by pane
	valid = assert.True(t, receiver.itemfiltering.enabled) && valid // once false, stays false

	return tea.Batch(cmds...), valid
}

// basic implementation, supports only lowercase basic characters
func charsToMessages(in string) []tea.Msg {
	msgs := make([]tea.Msg, len(in))
	for i, c := range in {
		msgs[i] = tea.KeyPressMsg(tea.Key{Text: string(c)})
	}
	return msgs
}

// executeCommand executes the given commands in linear fashion, for simplicity
// & greater reproducability
// TODO: execute linearly in DFS style
func executeCommand(cmd tea.Cmd) []tea.Msg {
	var (
		msgs []tea.Msg
		i    = -1
		cmds = []tea.Cmd{cmd}
	)

	for {
		i++
		if i >= len(cmds) {
			break
		}

		cmd := cmds[i]
		if cmd == nil {
			continue
		}
		msg := cmd()

		if batch, ok := msg.(tea.BatchMsg); ok {
			cmds = append(cmds, batch...)
			continue
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func extractMessages[T any](cmd tea.Cmd) []T {
	var (
		targets []T
		i       = -1
		cmds    = []tea.Cmd{cmd}
	)

	for {
		i++
		if i >= len(cmds) {
			break
		}

		cmd := cmds[i]
		if cmd == nil {
			continue
		}
		msg := cmd()

		if pr, ok := msg.(T); ok {
			targets = append(targets, pr)
		}

		if batch, ok := msg.(tea.BatchMsg); ok {
			cmds = append(cmds, batch...)
		}
	}

	return targets
}

// convenience function to send a page of items for the table-index to the
// system-under-test.
func simpleLoadItems(sut *ItemSelectionPane, tableARN *string, items apitypes.Items) tea.Cmd {
	return sut.Update(messages.PageReady{
		Table:    apitypes.DescribeTableResponse{TableArn: tableARN},
		Response: &messages.Page{Items: items},
	})
}

// convenience function to send a 'ColumnSortingUpdate' message to the
// system-under-test
func simpleSortItems(sut *ItemSelectionPane, tableARN string, sortOn string, asc bool) tea.Cmd {
	cols := sut.content.Columns()
	colsS := make([]string, 0, len(cols))
	for _, c := range cols {
		colsS = append(colsS, c.Title)
	}
	return sut.Update(messages.ColumnSortingUpdate{
		TableARN:   tableARN,
		AllColumns: colsS,
		SortingOn:  sortOn,
		Ascending:  asc,
	})
}

// convenience function to send a 'ScanIndexChanged' message to the
// system-under-test
func simpleChangeScanIndex(sut *ItemSelectionPane, tableARN, index string) tea.Cmd {
	return sut.Update(messages.ScanIndexChanged{
		TableARN:  tableARN,
		IndexName: index,
	})
}

// convenience function to send a 'QueryParametersChanged' message to the
// system-under-test
func simpleChangeQParams(sut *ItemSelectionPane, tableARN, index string) tea.Cmd {
	return sut.Update(messages.QueryParametersChanged{
		TableARN:  tableARN,
		IndexName: index,
	})
}

// convenience function to merge multiple items together. Slices will be appended.
func mergeItems(items ...apitypes.Items) apitypes.Items {
	res := apitypes.Items{}

	for _, itm := range items {
		res.JSON = append(res.JSON, itm.JSON...)
		res.JSONStyled = append(res.JSONStyled, itm.JSONStyled...)
		res.YAML = append(res.YAML, itm.YAML...)
		res.YAMLStyled = append(res.YAMLStyled, itm.YAMLStyled...)
		res.Raw = append(res.Raw, itm.Raw...)
		res.TableKeys = append(res.TableKeys, itm.TableKeys...)
	}

	return res
}

// convenience function for more concise test expressions
func fatalIf(t *testing.T, cond bool, msg ...any) {
	if cond {
		t.Fatal(msg...)
	}
}

// convenience function for more concise test expressions
func skipIf(t *testing.T, cond bool, msg ...any) {
	if cond {
		t.Skip(msg...)
	}
}
