package mytypes

// Add shared types here

// Tuple is a generic type that holds two values of potentially different types.
type Tuple[T any, U any] struct {
	First  T
	Second U
}

// Set implements a generic set data structure using a map.
// This is a thin abstraction over idiomatic use of map[]struct{} in Go as a set.
type Set[E comparable] map[E]struct{}

// NewSet creates a new empty Set.
// The type parameter E must be comparable to allow its use as a map key.
func NewSet[E comparable]() Set[E] {
	return Set[E]{}
}

// Contains checks if an element exists in the set.
func (s Set[E]) Contains(e E) bool {
	_, exists := s[e]
	return exists
}

// Add adds an element to the set and returns the modified set for chaining.
func (s Set[E]) Add(e E) Set[E] {
	s[e] = struct{}{}
	return s
}

// AddAll adds multiple elements to the set and returns the modified set for chaining.
func (s Set[E]) AddAll(elements ...E) Set[E] {
	for _, e := range elements {
		s[e] = struct{}{}
	}
	return s
}
