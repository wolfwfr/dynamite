// credits go to charm.land/bubbles/v2/list/list.go
package search

import (
	"sort"

	"github.com/sahilm/fuzzy"
)

// TODO: consider flattening to just 'string'
type Item struct {
	Content string
}

type FilteredItem struct {
	Index   int   // index in the unfiltered list
	Item    Item  // item matched
	Matches []int // rune indices of matched items
}

type FilteredItems []FilteredItem

func (f FilteredItems) items() []Item {
	agg := make([]Item, len(f))
	for i, v := range f {
		agg[i] = v.Item
	}
	return agg
}

// FilterMatchesMsg contains data about items matched during filtering. The
// message should be routed to Update for processing.
// TODO: move to messages package
type FilterMatchesMsg struct {
	ID    string
	Items []FilteredItem
}

// FilterFunc takes a term and a list of strings to search through
// (defined by Item#FilterValue).
// It should return a sorted list of ranks.
type FilterFunc func(string, []string) []Rank

// Rank defines a rank for a given item.
type Rank struct {
	// The index of the item in the original input.
	Index int
	// Indices of the actual word that were matched against the filter term.
	MatchedIndexes []int
}

// DefaultFilter uses the sahilm/fuzzy to filter through the list.
// This is set by default.
func DefaultFilter(term string, targets []string) []Rank {
	ranks := fuzzy.Find(term, targets)
	sort.Stable(ranks)
	result := make([]Rank, len(ranks))
	for i, r := range ranks {
		result[i] = Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}

// UnsortedFilter uses the sahilm/fuzzy to filter through the list. It does not
// sort the results.
func UnsortedFilter(term string, targets []string) []Rank {
	ranks := fuzzy.FindNoSort(term, targets)
	result := make([]Rank, len(ranks))
	for i, r := range ranks {
		result[i] = Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}
