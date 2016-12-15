package container

// Iterator is an interface for iterators of any container type.
type Iterator interface {
	// IsValid returns true if the iterator is valid for use, false otherwise.
	// We must not call Next, Key, or Value if IsValid returns false.
	IsValid() bool
	// Next advances the iterator to the next element of the map
	Next()
	// Value returns the value of the underlying element
	Value() interface{}
}

// MapIterator is a common interface for iterators of maps.
type MapIterator interface {
	Iterator
	// Key returns the key of the underlying element
	Key() interface{}
}
