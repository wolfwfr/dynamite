// package util defines various generic go utility functions
package util

import (
	"cmp"
	"reflect"
	"strings"
)

func Ternary[T any](first, second T, cond bool) T {
	if cond {
		return first
	}
	return second
}

func ToPtr[T any](t T) *T {
	return &t
}

func IfNotNil[T any](first *T, second T) T {
	if first != nil {
		return *first
	}
	return second
}

// with appropriate condition, it escapes nil-pointers
func TernarySafe[T any](first *T, second T, cond bool) T {
	if cond {
		return *first
	}
	return second
}

func Find[S []E, E comparable](slice S, target E) int {
	for i, e := range slice {
		if e == target {
			return i
		}
	}
	return -1
}

func FindBy[S []E, E comparable](slice S, cond func(i E) bool) int {
	for i, e := range slice {
		if cond(e) {
			return i
		}
	}
	return -1
}

func RepeatString(str string, c int) string {
	b := strings.Builder{}
	for range c {
		b.WriteString(str)
	}
	return b.String()
}

// MergeMaps takes two maps and returns an identically typed map containing all
// keys present in the specified inputs and their associated values, where the
// values contained in the first argument take precedence.
func MergeMaps[K comparable, T any](m1, m2 map[K]T) (merged map[K]T) {
	merged = make(map[K]T)
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		if _, ok := merged[k]; ok {
			continue
		}
		merged[k] = v
	}
	return
}

// MergeMapsSafe takes two maps and returns two identically typed maps. The
// first containing all keys present in the specified inputs and their
// associated values, where the values contained in the first argument take
// precedence. The second map contains all key-value pairs of the map in the
// second argument that could not be safely merged without violating the
// precedence rule.
func MergeMapsSafe[K comparable, T any](m1, m2 map[K]T) (merged map[K]T, remainder map[K]T) {
	merged = make(map[K]T)
	remainder = make(map[K]T)
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		if vv, ok := merged[k]; ok && !reflect.DeepEqual(v, vv) {
			remainder[k] = v
		}
		merged[k] = v
	}
	return
}

func Clamp[T cmp.Ordered](v, low, high T) T {
	return min(max(v, low), high)
}

func ContainsBy[S ~[]E, E any](s S, f func(E) bool) bool {
	for _, e := range s {
		if f(e) {
			return true
		}
	}
	return false
}
