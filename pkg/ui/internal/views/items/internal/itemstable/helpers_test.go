package itemstable

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

func TestCompileUniqueKeys(t *testing.T) {
	t.Run("compile-unique-keys should", func(t *testing.T) {
		// testcases
		testcases := []struct {
			desc           string
			input_hasRange bool
			input_keys     [][]apitypes.KeyValue
			exp            []string
		}{
			{
				desc:           "compile a complete set when first entry has missing keys",
				input_hasRange: false,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "C"},
					},
					{
						{Key: "A"},
						{Key: "B"},
						{Key: "C"},
					},
				},
				exp: []string{"A", "B", "C"},
			}, {
				desc:           "compile a complete set when second entry has missing keys",
				input_hasRange: false,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "B"},
						{Key: "C"},
					},
					{
						{Key: "A"},
						{Key: "C"},
					},
				},
				exp: []string{"A", "B", "C"},
			}, {
				desc:           "compile a complete set when two entries each have unique keys; sort correctly in orientation 1",
				input_hasRange: false,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "B"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
					{
						{Key: "A"},
						{Key: "C"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
				},
				exp: []string{"A", "B", "C", "X", "Y", "Z"},
			}, {
				desc:           "compile a complete set when two entries each have unique keys; sort correctly in orientation 2",
				input_hasRange: false,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "C"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
					{
						{Key: "A"},
						{Key: "B"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
				},
				exp: []string{"A", "B", "C", "X", "Y", "Z"},
			}, {
				desc:           "respect range-key presence when sorting in orientation 1",
				input_hasRange: true,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "B"}, // range-key
						{Key: "D"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
					{
						{Key: "A"},
						{Key: "B"},
						{Key: "C"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
				},
				exp: []string{"A", "B", "C", "D", "X", "Y", "Z"},
			}, {
				desc:           "respect range-key presence when sorting in orientation 2",
				input_hasRange: true,
				input_keys: [][]apitypes.KeyValue{
					{
						{Key: "A"}, // first is always assumed to be shared hash-key
						{Key: "B"}, // range-key
						{Key: "C"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
					{
						{Key: "A"},
						{Key: "B"},
						{Key: "D"}, {Key: "X"}, {Key: "Y"}, {Key: "Z"},
					},
				},
				exp: []string{"A", "B", "C", "D", "X", "Y", "Z"},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				// mock raw types
				attrTypes := make([]map[string]types.AttributeValue, len(tc.input_keys))
				for i, c := range tc.input_keys {
					attrTypes[i] = make(map[string]types.AttributeValue)
					for _, r := range c {
						attrTypes[i][r.Key] = &types.AttributeValueMemberS{}
					}
				}
				// test
				res := compileUniqueKeys(tc.input_keys, attrTypes, nil, tc.input_hasRange)

				// extract
				titles := make([]string, len(res))
				for i := range res {
					titles[i] = res[i].Title
				}

				// assert
				assert.EqualValues(t, tc.exp, titles)
			})
		}
	})
}
