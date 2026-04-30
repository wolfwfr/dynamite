// package util defines various utility functions
package util

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

func Clamp(v, low, high int) int {
	return min(max(v, low), high)
}
