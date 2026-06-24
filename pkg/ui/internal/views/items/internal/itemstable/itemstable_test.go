package itemstable

import (
	"fmt"
	"maps"
	"testing"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

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
	res.Raw = make([]map[string]dynamodbtypes.AttributeValue, ln)
	res.TableKeys = make([][]apitypes.KeyValue, ln)

	for i := range ln {
		id := fmt.Sprintf(idFmt, b+i)

		res.JSON[i] = `{
  "id": "` + id + `",
  "configured": true
}`
		res.JSONStyled[i] = styles.ObjectStyle{}

		res.Raw[i] = map[string]dynamodbtypes.AttributeValue{
			"id":         &dynamodbtypes.AttributeValueMemberS{Value: id},
			"configured": &dynamodbtypes.AttributeValueMemberBOOL{Value: true},
		}

		res.TableKeys[i] = []apitypes.KeyValue{
			{Key: "id", Value: fmt.Sprintf("\"%s\"", id)},
			{Key: "configured", Value: "true"},
		}
	}

	return res
}

type searchGenOpts struct {
	begin int
}

// generate some simple search results for testing
// purporses. When used without options, the first argument both counts to the
// number of items being created and their recorded indices (starting at
// 0). When providing an option with 'begin', the function will return 'n -
// begin' number of items, with ID enumeration starting at 'begin'.
func genSearchResults(n int, opts ...searchGenOpts) []search.FilteredItem {
	b := 0

	if len(opts) > 0 {
		b = opts[0].begin
	}

	res := make([]search.FilteredItem, n-b)

	for nn := range len(res) {
		res[nn] = search.FilteredItem{
			Index:   b + nn,
			Item:    search.Item{Content: "some item"},
			Matches: []int{0},
		}
	}

	return res
}

func TestCacheInvalidation(t *testing.T) {
	var (
	// cacheKey = func(r, c, cw int) string {
	// 	return fmt.Sprintf("%d-%d-%d", r, c, cw)
	// }
	)

	// factory initialising a new system-under-test
	newSUT := func() *ItemsTable {
		sut := NewItemsTable()
		sut.UpdateSize(100, 200) // required for underlying table to properly render items

		// simple delegate that does not consider any kind of styling, only caching
		// sut.table.SetFieldDelegate(func(row table.Row, col table.Column, colIdx, rowIdx, colW, padL, padR int, selected, inview bool) string {
		// 	key := cacheKey(rowIdx, colIdx, colW)
		// 	if f, ok := sut.renderCache[key]; ok { // return from cache if found
		// 		return f
		// 	}
		// 	f := row.Fields[colIdx].Value() // no styling for this test
		// 	sut.renderCache[key] = f        // store in cache
		// 	return f                        // return
		// })

		return sut
	}

	// assertContainsExclusively asserts that cache contains only values
	// contained in the specified table. If cache is not required to contain all
	// table values, set `strict` to false.
	assertContainsExclusively := func(t *testing.T, cache map[string]string, table *table.Model, strict bool) {
		var n int
		cols := table.Columns()
		for ri, r := range table.VisualRows() {
			for ci, c := range r.Fields {
				cw := cols[ci].Width
				k := cacheKey(ri, ci, cw)
				v, ok := cache[k]
				if strict {
					require.True(t, ok, "did not find entry for cache-key '%s'", k)
				}
				if !ok {
					continue
				}
				n++
				require.EqualValues(t, c.Value(), v[:len(c.Value())], "did not find expected value for cache-key '%s'", k)
			}
		}
		assert.EqualValues(t, n, len(cache), "cache contained more values than allowed by table: expected '%d', got '%d'", n, len(cache))
	}

	t.Run("item-selection-pane should", func(t *testing.T) {
		t.Run("refresh cache when", func(t *testing.T) {
			t.Run("setting search results", func(t *testing.T) {
				sut := newSUT()                                                      // init
				n_items := 6                                                         // define initial table length
				items := genJSONItems(n_items)                                       // first half page
				sut.AddItems(items, false)                                           // initialise table with items
				require.NotEmpty(t, sut.renderCache)                                 // ensure cache is initialised
				oldState := make(map[string]string)                                  // record existing cache-state for comparison
				maps.Copy(oldState, sut.renderCache)                                 // copy between vars
				n_searchResults := n_items - 3                                       // reduce table len after search
				s := genSearchResults(n_searchResults)                               // generate search results
				sut.SetSearchResults("id", s)                                        // set search results on table
				require.EqualValues(t, n_searchResults, len(sut.table.VisualRows())) // ensure table shows only search results
				assert.NotEmpty(t, sut.renderCache)                                  // assert new cache is not empty
				assert.Less(t, len(sut.renderCache), len(oldState))                  // assert less items in cache
				assertContainsExclusively(t, sut.renderCache, sut.table, false)      // assert cached values all match current table items
			})
			t.Run("resetting search", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				n_searchResults := n_items - 3                                  // reduce table len after search
				s := genSearchResults(n_searchResults)                          // generate search results
				sut.SetSearchResults("id", s)                                   // set search results on table
				oldState := make(map[string]string)                             // record existing cache-state for comparison
				maps.Copy(oldState, sut.renderCache)                            // copy between vars
				sut.ResetSearch()                                               // reset search
				require.EqualValues(t, n_items, len(sut.table.VisualRows()))    // ensure table shows original results
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assert.Greater(t, len(sut.renderCache), len(oldState))          // assert more items in cache
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("refresh cache is called", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.refreshCache()                                              // call refresh cache
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after update to table contents - columns", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.updateTable(sut.table.Columns(), nil, nil)                  // call update-table
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after update to table contents - rows", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.updateTable(nil, sut.table.Rows(), nil)                     // call update-table
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after update to table contents - virtual-rows", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.updateTable(nil, nil, sut.table.Rows())                     // call update-table
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after update to table contents - rows & colums", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.updateTable(sut.table.Columns(), sut.table.Rows(), nil)     // call update-table
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after updating table size", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{}                           // override cache for later refresh validation
				sut.UpdateSize(200, 400)                                        // call update-table
				assert.NotEmpty(t, sut.renderCache)                             // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
			t.Run("after calling full reset", func(t *testing.T) {
				sut := newSUT()                               // init
				n_items := 6                                  // define initial table length
				items := genJSONItems(n_items)                // first half page
				sut.AddItems(items, false)                    // initialise table with items
				sut.renderCache = map[string]string{"a": "b"} // override cache for later refresh validation
				sut.Reset()                                   // call full reset
				assert.Empty(t, sut.renderCache)              // assert new cache is empty (as table after full reset)
			})
		})
		t.Run("clear cache when", func(t *testing.T) {
			t.Run("clear cache is called", func(t *testing.T) {
				sut := newSUT()                                                 // init
				n_items := 6                                                    // define initial table length
				items := genJSONItems(n_items)                                  // first half page
				sut.AddItems(items, false)                                      // initialise table with items
				sut.renderCache = map[string]string{"a": "b"}                   // override cache for later refresh validation
				sut.clearCache()                                                // call refresh cache
				assert.Empty(t, sut.renderCache)                                // assert new cache is not empty
				assertContainsExclusively(t, sut.renderCache, sut.table, false) // assert cached values all match current table items
			})
		})
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
