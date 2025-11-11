package gamerules

// Set is a collection of unique elements of type T
type Set[T comparable] map[T]struct{}

// NewSet creates and returns an empty set
func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func MakeSet[T comparable](items []T) Set[T] {
	set := NewSet[T]()
	for _, item := range items {
		set.Add(item)
	}
	return set
}

// Add adds an element to the set
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Remove removes an element from the set
func (s Set[T]) Remove(item T) {
	delete(s, item)
}

// Contains checks if the set contains an element
func (s Set[T]) Contains(item T) bool {
	_, exists := s[item]
	return exists
}

// Size returns the number of elements in the set
func (s Set[T]) Size() int {
	return len(s)
}

// ForEach executes a function for each element in the set
func (s Set[T]) ForEach(fn func(T)) {
	for item := range s {
		fn(item)
	}
}

// ToSlice returns a slice of all elements in the set
func (s Set[T]) ToSlice() []T {
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

// Union returns a new set containing all elements from both sets
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		result.Add(item)
	}
	for item := range other {
		result.Add(item)
	}
	return result
}

// Intersection returns a new set containing elements present in both sets
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference returns a new set containing elements present in the first set but not in the second
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// IsSubset returns true if all elements in the set are in the other set
func (s Set[T]) IsSubset(other Set[T]) bool {
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

func (s Set[T]) IsEqual(other Set[T]) bool {
	return s.IsSubset(other) && other.IsSubset(s)
}

// Clear removes all elements from the set
func (s Set[T]) Clear() {
	for item := range s {
		delete(s, item)
	}
}

// Copy returns a new set with the same elements
func (s Set[T]) Copy() Set[T] {
	result := NewSet[T]()
	for item := range s {
		result.Add(item)
	}
	return result
}
