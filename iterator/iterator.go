// Package iterator supplies the small value iterator used by Option and Result.
package iterator

// Iterator is an immutable, reusable iterator over a finite sequence.
type Iterator[T any] struct {
	values []T
}

// Empty returns an iterator with no items.
func Empty[T any]() Iterator[T] { return Iterator[T]{} }

// Once returns an iterator containing one item.
func Once[T any](value T) Iterator[T] { return Iterator[T]{values: []T{value}} }

// FromSlice returns an iterator over a defensive copy of values.
func FromSlice[T any](values []T) Iterator[T] {
	return Iterator[T]{values: append([]T(nil), values...)}
}

// Map transforms every item and may change its type.
func (iter Iterator[T]) Map[U any](transform func(T) U) Iterator[U] {
	values := make([]U, len(iter.values))
	for index, value := range iter.values {
		values[index] = transform(value)
	}
	return Iterator[U]{values: values}
}

// Map transforms every item and may change its type.
func Map[T, U any](iter Iterator[T], transform func(T) U) Iterator[U] {
	values := make([]U, len(iter.values))
	for index, value := range iter.values {
		values[index] = transform(value)
	}
	return Iterator[U]{values: values}
}

// Filter keeps items accepted by predicate.
func (iter Iterator[T]) Filter(predicate func(T) bool) Iterator[T] {
	values := make([]T, 0, len(iter.values))
	for _, value := range iter.values {
		if predicate(value) {
			values = append(values, value)
		}
	}
	return Iterator[T]{values: values}
}

// Collect returns a defensive copy of the sequence.
func (iter Iterator[T]) Collect() []T {
	return append([]T(nil), iter.values...)
}

// Len returns the number of items in the iterator.
func (iter Iterator[T]) Len() int { return len(iter.values) }
