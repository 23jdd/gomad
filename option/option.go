// Package option provides a zero-allocation Option type.
package option

import (
	"github.com/23jdd/gomad/internal/core"
	"github.com/23jdd/gomad/iterator"
)

// Option contains either one value (Some) or no value (None).
// Its zero value is None.
type Option[T any] = core.Option[T]

// Some constructs an Option containing value.
func Some[T any](value T) Option[T] { return core.Some(value) }

// None constructs an empty Option.
func None[T any]() Option[T] { return core.None[T]() }

// Map transforms an Option and may change its value type.
func Map[T, U any](value Option[T], transform func(T) U) Option[U] {
	if value.IsNone() {
		return None[U]()
	}
	return Some(transform(value.Unwrap()))
}

// AndThen chains an Option-producing operation and may change its value type.
func AndThen[T, U any](value Option[T], transform func(T) Option[U]) Option[U] {
	if value.IsNone() {
		return None[U]()
	}
	return transform(value.Unwrap())
}

// FromPtr converts a pointer to an Option. A nil pointer becomes None.
func FromPtr[T any](value *T) Option[T] {
	if value == nil {
		return None[T]()
	}
	return Some(*value)
}

// FromZero converts the zero value of T to None and every other value to Some.
func FromZero[T comparable](value T) Option[T] {
	var zero T
	if value == zero {
		return None[T]()
	}
	return Some(value)
}

// FromValueOk converts Go's conventional (value, ok) pair to an Option.
func FromValueOk[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// FromValueError converts a (value, error) pair to an Option.
func FromValueError[T any](value T, err error) Option[T] {
	if err != nil {
		return None[T]()
	}
	return Some(value)
}

// Flatten removes one level of Option nesting.
func Flatten[T any](value Option[Option[T]]) Option[T] {
	if value.IsNone() {
		return None[T]()
	}
	return value.Unwrap()
}

// Match dispatches to some or none according to value's state.
func Match[T any](value Option[T], some func(T), none func()) {
	if value.IsSome() {
		some(value.Unwrap())
		return
	}
	none()
}

// Collect returns all contained values, or None when any input is None.
func Collect[T any](values []Option[T]) Option[[]T] {
	collected := make([]T, 0, len(values))
	for _, value := range values {
		if value.IsNone() {
			return None[[]T]()
		}
		collected = append(collected, value.Unwrap())
	}
	return Some(collected)
}

// All reports whether every Option is Some. It is true for an empty slice.
func All[T any](values []Option[T]) bool {
	for _, value := range values {
		if value.IsNone() {
			return false
		}
	}
	return true
}

// Any reports whether at least one Option is Some.
func Any[T any](values []Option[T]) bool {
	for _, value := range values {
		if value.IsSome() {
			return true
		}
	}
	return false
}

// FromIterator collects zero or one item from an iterator into an Option.
func FromIterator[T any](iter iterator.Iterator[T]) Option[T] {
	values := iter.Collect()
	if len(values) == 0 {
		return None[T]()
	}
	return Some(values[0])
}
