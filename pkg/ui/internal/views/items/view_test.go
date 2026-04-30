package itemselection

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

func TestCompileCompleteKeys(t *testing.T) {
	t.Run("compile-complete-keys should", func(t *testing.T) {
		// testcases
		testcases := []struct {
			desc           string
			input_hasRange bool
			input_keys     [][]types.KeyValue
			exp            []string
		}{
			{
				desc:           "compile a complete set when first entry has missing keys",
				input_hasRange: false,
				input_keys: [][]types.KeyValue{
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
				input_keys: [][]types.KeyValue{
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
				input_keys: [][]types.KeyValue{
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
				input_keys: [][]types.KeyValue{
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
				input_keys: [][]types.KeyValue{
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
				input_keys: [][]types.KeyValue{
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
				// test
				res := compileCompleteKeys(tc.input_keys, nil, tc.input_hasRange)

				// assert
				assert.EqualValues(t, tc.exp, res)
			})
		}
	})
}
