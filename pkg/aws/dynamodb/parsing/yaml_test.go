package parsing

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func genericTestItemYAML() string {
	tabsize = 2
	return `string-key: "string-value"
bool-false-key: false
bool-true-key: true
byte-key: byte-value
byte-set-key: 
  - byte-set-value-1
  - byte-set-value-2
empty-map-key: 
empty-set-key: 
list-key-1: 
  - "list-1-value-1"
  - "list-1-value-2"
list-key-2: 
  - list-2-value-1-field-1: "map-1-value-1"
    list-2-value-1-field-2: "map-1-value-2"
  - list-2-value-2-field-1: "map-2-value-1"
    list-2-value-2-field-2: "map-2-value-2"
map-key: 
  map-value-field-1: "map-value-field-1-value"
  map-value-field-2: "map-value-field-2-value"
null-false-key: NOT NULL
null-true-key: NULL
number-key: 100.5
number-set-key: 
  - 52.5
  - 35
string-set-key: 
  - "string-set-value-1"
  - "string-set-value-2"
`
}

func TestYAMLParsing(t *testing.T) {
	tabsize = 2
	item := genericTestItem()
	res := ParseItemToYAML(item, "string-key", nil)
	exp := genericTestItemYAML()
	assert.EqualValues(t, exp, res)
	fmt.Print(res)
}

func toPtr[T any](t T) *T {
	return &t
}

func TestKeySorting(t *testing.T) {
	tr := func(in []string) map[string]types.AttributeValue {
		res := make(map[string]types.AttributeValue)
		for _, k := range in {
			res[k] = &types.AttributeValueMemberS{Value: ""}
		}
		return res
	}

	t.Run("get-sorted-keys should", func(t *testing.T) {
		// convenience resources

		// testcases
		testcases := []struct {
			desc           string
			input_keys     []string
			input_hashkey  string
			input_rangekey *string
			rootlevel      bool
			exp            []string
		}{
			{
				desc:           "simply return alphabetically sorted keys when root-level == false",
				input_keys:     []string{"Z", "X", "B", "U", "L"},
				input_hashkey:  "U",        // should have no effect
				input_rangekey: toPtr("L"), // should have no effect
				rootlevel:      false,
				exp:            []string{"B", "L", "U", "X", "Z"},
			}, {
				desc:           "return alphabetically sorted keys with hashkey at idx 0 when root-level == true",
				input_keys:     []string{"Z", "X", "B", "U", "L"},
				input_hashkey:  "U", // should have no effect
				input_rangekey: nil, // should have no effect
				rootlevel:      true,
				exp:            []string{"U", "B", "L", "X", "Z"},
			}, {
				desc:           "return alphabetically sorted keys with hashkey at idx 0, & rangekey at idx 1 when root-level == true",
				input_keys:     []string{"Z", "X", "B", "U", "L"},
				input_hashkey:  "U",        // should have no effect
				input_rangekey: toPtr("Z"), // should have no effect
				rootlevel:      true,
				exp:            []string{"U", "Z", "B", "L", "X"},
			}, {
				desc:           "sort hash-key correctly when already at correct position",
				input_keys:     []string{"U", "X", "B", "Z", "L"},
				input_hashkey:  "U", // should have no effect
				input_rangekey: nil, // should have no effect
				rootlevel:      true,
				exp:            []string{"U", "B", "L", "X", "Z"},
			}, {
				desc:           "sort hash- & range-keys correctly when already at correct position",
				input_keys:     []string{"U", "X", "B", "Z", "L"},
				input_hashkey:  "U",        // should have no effect
				input_rangekey: toPtr("X"), // should have no effect
				rootlevel:      true,
				exp:            []string{"U", "X", "B", "L", "Z"},
			}, {
				desc:           "sort hash- & range-keys correctly when already at each other's position",
				input_keys:     []string{"X", "U", "B", "Z", "L"},
				input_hashkey:  "U",        // should have no effect
				input_rangekey: toPtr("X"), // should have no effect
				rootlevel:      true,
				exp:            []string{"U", "X", "B", "L", "Z"},
			}, {
				desc:           "sort hash- & range-keys correctly when only range is at correct position",
				input_keys:     []string{"X", "U", "B", "Z", "L"},
				input_hashkey:  "L",        // should have no effect
				input_rangekey: toPtr("U"), // should have no effect
				rootlevel:      true,
				exp:            []string{"L", "U", "B", "X", "Z"},
			}, {
				desc:           "sort hash- & range-keys correctly when only hash is at correct position",
				input_keys:     []string{"X", "U", "B", "Z", "L"},
				input_hashkey:  "X",        // should have no effect
				input_rangekey: toPtr("Z"), // should have no effect
				rootlevel:      true,
				exp:            []string{"X", "Z", "B", "L", "U"},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				// test
				res := getSortedKeys(tc.input_hashkey, tc.input_rangekey, tr(tc.input_keys), tc.rootlevel)

				// assert
				assert.EqualValues(t, tc.exp, res)
			})
		}
	})
}
