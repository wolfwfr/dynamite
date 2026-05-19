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
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

// validate keys on init
func init() {
	validateItemSelectionKeys()
}

// validateItemSelectionKeys ensures keymap-variables accurately inform tests on the validity
// of the keymap configuration in this test file.
func validateItemSelectionKeys() {
	keymap := DefaultItemPaneKeyMap()
	searchKeyValid = key.Matches(searchKey, keymap.Search)
	keysValidated = true
}

// fail tests on invalid keys; indicates the keymap has changed
func TestKeyMapValid(t *testing.T) {
	assert.True(t, searchKeyValid)
}

func genJSONItems(b, n int) apitypes.Items {
	res := apitypes.Items{}

	ln := n - b
	res.JSON = make([]string, ln)
	res.JSONStyled = make([]styles.ObjectStyle, ln)
	res.Raw = make([]map[string]types.AttributeValue, ln)
	res.TableKeys = make([][]apitypes.KeyValue, ln)

	for i := range ln {
		id := fmt.Sprintf("id-%d", b+i)

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

func extractColumns(in []apitypes.KeyValue) []string {
	res := make([]string, len(in))
	for i, kv := range in {
		res[i] = kv.Key
	}
	return res
}

var (
	keysValidated  bool
	searchKeyValid bool

	searchKey = tea.KeyPressMsg(tea.Key{Text: "/"})
)

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
			var (
				sut = newSUT()
			)

			// prepare message
			inputMessage := messages.PageReady{
				Table:    apitypes.DescribeTableResponse{TableArn: &tableARN},
				Response: &messages.Page{Items: genJSONItems(0, 1)},
			}

			// call
			cmd := sut.Update(inputMessage)

			// obtain target messages
			targets := extractMessages[messages.PreviewItem](cmd)

			// assert only one preview-item message
			require.Len(t, targets, 1)
			// assert correct item being previewed
			assert.EqualValues(t, inputMessage.Response.Items.JSON[0], targets[0].RawItem)
		})
		t.Run("preview correct item after loading new page that is sorted to table top", func(t *testing.T) {
			var (
				sut   = newSUT()
				page1 = genJSONItems(0, 3)
				page2 = genJSONItems(3, 6)
			)

			// prepare first page
			sut.Update(messages.PageReady{
				Table:    apitypes.DescribeTableResponse{TableArn: &tableARN},
				Response: &messages.Page{Items: page1},
			})

			// enable sorting
			sut.Update(messages.ColumnSortingUpdate{
				TableARN:   tableARN,
				AllColumns: extractColumns(page1.TableKeys[0]),
				SortingOn:  page1.TableKeys[0][0].Key,
				Ascending:  false,
			})

			// load next page
			cmd := sut.Update(messages.PageReady{
				Table:    apitypes.DescribeTableResponse{TableArn: &tableARN},
				Response: &messages.Page{Items: page2},
			})

			// obtain target messages
			targets := extractMessages[messages.PreviewItem](cmd)

			// assert only one preview-item message
			require.Len(t, targets, 1)
			// assert correct item being previewed
			assert.EqualValues(t, page2.JSON[2], targets[0].RawItem)
		})
		t.Run("preview correct item after search", func(t *testing.T) {
			if !searchKeyValid {
				t.Skip("skipping due to outdated search key")
			}

			var (
				sut   = newSUT()
				items = genJSONItems(0, 3)
			)

			// load items
			sut.Update(messages.PageReady{
				Table:    apitypes.DescribeTableResponse{TableArn: &tableARN},
				Response: &messages.Page{Items: items},
			})

			// enable search
			sut.Update(searchKey)

			// search for first item
			cmd, ok := searchItemSelection(t, sut, "id=id-1")
			require.True(t, ok)

			// obtain target messages
			targets := extractMessages[messages.PreviewItem](cmd)

			// assert only one preview-item message
			require.NotEmpty(t, targets)
			// assert correct item being previewed
			assert.EqualValues(t, items.JSON[1], targets[len(targets)-1].RawItem)
		})
	})
}

// convenience function to apply a search query. Returns a boolean that equals
// `true` when the search was successfully applied.
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
	valid = assert.True(t, receiver.itemfiltering.enabled) && valid          // once false, stays false
	valid = assert.NotEmpty(t, receiver.itemfiltering.matchedItems) && valid // once false, stays false

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
