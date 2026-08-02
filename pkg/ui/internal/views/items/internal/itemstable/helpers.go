package itemstable

import (
	"slices"
	"sort"

	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/common"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

func mergeSlices[S ~[]E, E any](s1, s2 S) S {
	n := make([]E, len(s1)+len(s2))
	copy(n[:len(s1)], s1)
	copy(n[len(s1):], s2)
	return n
}

// compileUniqueKeys takes a table of key-value pairs, observes all keys and
// compiles a complete, in-order list of all unique key observed.
// This ensures that when individual table rows have keys missing, the final
// result still contains these keys when they are present in other rows in the
// specified table.
func compileUniqueKeys(items apitypes.Items, existing []ColumnAttributes, hasRangeKey bool) []ColumnAttributes {
	res := make([]ColumnAttributes, 0)
	seen := map[string]struct{}{}
	if len(existing) > 0 {
		res = existing
	}
	for _, e := range existing {
		seen[e.Title] = struct{}{}
	}
	for _, item := range items {
		for _, col := range item.TableKeys {
			key := col.Key
			typ := common.ParseDynamoAttributeType(item.Raw[key])
			if _, ok := seen[key]; !ok {
				res = append(res, ColumnAttributes{key, typ})
				seen[key] = struct{}{}
			}
		}
	}

	sortLenOffset := u.Ternary(2, 1, hasRangeKey)
	toSort := make([]ColumnAttributes, len(res)-sortLenOffset)
	copy(toSort, res[sortLenOffset:])
	slices.SortFunc(toSort, func(a, b ColumnAttributes) int {
		if sort.StringsAreSorted([]string{a.Title, b.Title}) {
			return -1
		}
		return 1 // equal is not possible in this context
	})
	copy(res[sortLenOffset:], toSort)

	return res
}
