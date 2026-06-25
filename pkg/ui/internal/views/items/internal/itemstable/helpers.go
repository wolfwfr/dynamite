package itemstable

import (
	"slices"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
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
func compileUniqueKeys(table [][]apitypes.KeyValue, existing []string, hasRangeKey bool) []string {
	res := make([]string, 0)
	seen := map[string]struct{}{}
	if len(existing) > 0 {
		res = existing
	}
	for _, e := range existing {
		seen[e] = struct{}{}
	}
	for _, row := range table {
		for _, col := range row {
			key := col.Key
			if _, ok := seen[key]; !ok {
				res = append(res, key)
				seen[key] = struct{}{}
			}
		}
	}

	sortLenOffset := u.Ternary(2, 1, hasRangeKey)
	toSort := make([]string, len(res)-sortLenOffset)
	copy(toSort, res[sortLenOffset:])
	slices.Sort(toSort)
	copy(res[sortLenOffset:], toSort)

	return res
}
