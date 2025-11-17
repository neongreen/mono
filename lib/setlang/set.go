package setlang

// Set represents a set of items of type T.
// It uses a map for O(1) membership testing and efficient set operations.
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet creates a new empty set.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{items: make(map[T]struct{})}
}

// NewSetFrom creates a new set containing the given items.
func NewSetFrom[T comparable](items ...T) *Set[T] {
	s := NewSet[T]()
	for _, item := range items {
		s.Add(item)
	}
	return s
}

// Add adds an item to the set.
func (s *Set[T]) Add(item T) {
	if s.items == nil {
		s.items = make(map[T]struct{})
	}
	s.items[item] = struct{}{}
}

// Has returns true if the item is in the set.
func (s *Set[T]) Has(item T) bool {
	if s.items == nil {
		return false
	}
	_, ok := s.items[item]
	return ok
}

// Remove removes an item from the set.
func (s *Set[T]) Remove(item T) {
	if s.items == nil {
		return
	}
	delete(s.items, item)
}

// Size returns the number of items in the set.
func (s *Set[T]) Size() int {
	if s.items == nil {
		return 0
	}
	return len(s.items)
}

// IsEmpty returns true if the set is empty.
func (s *Set[T]) IsEmpty() bool {
	return s.Size() == 0
}

// Items returns a slice of all items in the set.
// The order is not guaranteed.
func (s *Set[T]) Items() []T {
	if s.items == nil {
		return nil
	}
	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// Union returns a new set containing all items from both sets.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()

	if s.items != nil {
		for item := range s.items {
			result.Add(item)
		}
	}

	if other != nil && other.items != nil {
		for item := range other.items {
			result.Add(item)
		}
	}

	return result
}

// Intersect returns a new set containing only items present in both sets.
func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	result := NewSet[T]()

	if s.items == nil || other == nil || other.items == nil {
		return result
	}

	// Iterate over the smaller set for efficiency
	var smaller, larger *Set[T]
	if s.Size() <= other.Size() {
		smaller, larger = s, other
	} else {
		smaller, larger = other, s
	}

	for item := range smaller.items {
		if larger.Has(item) {
			result.Add(item)
		}
	}

	return result
}

// Diff returns a new set containing items in s but not in other.
func (s *Set[T]) Diff(other *Set[T]) *Set[T] {
	result := NewSet[T]()

	if s.items == nil {
		return result
	}

	for item := range s.items {
		if other == nil || !other.Has(item) {
			result.Add(item)
		}
	}

	return result
}

// Clone returns a shallow copy of the set.
func (s *Set[T]) Clone() *Set[T] {
	result := NewSet[T]()
	if s.items != nil {
		for item := range s.items {
			result.Add(item)
		}
	}
	return result
}
