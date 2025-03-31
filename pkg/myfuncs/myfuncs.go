package myfuncs

// Ternary is a generic function that returns the first value if the condition is true,
// otherwise it returns the second value.
func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// Mergeable is a constraint for types that support addition
type Mergeable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// MergeMaps merges the source map into the destination map by adding values for matching keys
func MergeMaps[K comparable, V Mergeable](dest map[K]V, src map[K]V) {
	for k, v := range src {
		dest[k] += v
	}
}

// Map applies a function to each element of a slice and returns a new slice with the results.
// It does not modify the original slice.
func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

// Filter returns a new slice containing only the elements of the input slice that satisfy the predicate function.
// It does not modify the original slice.
func Filter[T any](s []T, f func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce applies a function to each element in the slice to accumulate a single result,
// starting with the initial value provided.
func Reduce[T any, U any](s []T, init U, f func(U, T) U) U {
	result := init
	for _, v := range s {
		result = f(result, v)
	}
	return result
}
