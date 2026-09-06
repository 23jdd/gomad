// Package result provides a Result type for explicit success and failure.
package result

import "github.com/23jdd/gomad/internal/core"

// Result contains either a successful value (Ok) or an error value (Err).
type Result[T, E any] = core.Result[T, E]

// Ok constructs a successful Result.
func Ok[T, E any](value T) Result[T, E] { return core.Ok[T, E](value) }

// Err constructs a failed Result. T can be specified while E is inferred.
func Err[T, E any](err E) Result[T, E] { return core.Err[T](err) }

// From converts Go's conventional (value, error) return into a Result.
func From[T any](value T, err error) Result[T, error] {
	if err != nil {
		return Err[T](err)
	}
	return Ok[T, error](value)
}

// Map transforms an Ok value and may change its type.
func Map[T, U, E any](value Result[T, E], transform func(T) U) Result[U, E] {
	if value.IsErr() {
		return Err[U](value.UnwrapErr())
	}
	return Ok[U, E](transform(value.Unwrap()))
}

// MapErr transforms an Err value and may change its type.
func MapErr[T, E, F any](value Result[T, E], transform func(E) F) Result[T, F] {
	if value.IsOk() {
		return Ok[T, F](value.Unwrap())
	}
	return Err[T](transform(value.UnwrapErr()))
}

// AndThen chains a Result-producing operation and may change its Ok type.
func AndThen[T, U, E any](value Result[T, E], transform func(T) Result[U, E]) Result[U, E] {
	if value.IsErr() {
		return Err[U](value.UnwrapErr())
	}
	return transform(value.Unwrap())
}

// OrElse recovers from an Err and may change its error type.
func OrElse[T, E, F any](value Result[T, E], fallback func(E) Result[T, F]) Result[T, F] {
	if value.IsOk() {
		return Ok[T, F](value.Unwrap())
	}
	return fallback(value.UnwrapErr())
}

// Match dispatches to ok or err according to value's state.
func Match[T, E any](value Result[T, E], ok func(T), err func(E)) {
	if value.IsOk() {
		ok(value.Unwrap())
		return
	}
	err(value.UnwrapErr())
}

// Collect returns all Ok values, or the first Err.
func Collect[T, E any](values []Result[T, E]) Result[[]T, E] {
	collected := make([]T, 0, len(values))
	for _, value := range values {
		if value.IsErr() {
			return Err[[]T](value.UnwrapErr())
		}
		collected = append(collected, value.Unwrap())
	}
	return Ok[[]T, E](collected)
}

// All reports whether every Result is Ok. It is true for an empty slice.
func All[T, E any](values []Result[T, E]) bool {
	for _, value := range values {
		if value.IsErr() {
			return false
		}
	}
	return true
}

// Partition separates Ok and Err values while preserving their order.
func Partition[T, E any](values []Result[T, E]) ([]T, []E) {
	okValues := make([]T, 0, len(values))
	errValues := make([]E, 0)
	for _, value := range values {
		if value.IsOk() {
			okValues = append(okValues, value.Unwrap())
		} else {
			errValues = append(errValues, value.UnwrapErr())
		}
	}
	return okValues, errValues
}

// Wrap adds context to an error Result.
func Wrap[T any](value Result[T, error], message string) Result[T, error] {
	return value.Wrap(message)
}

// Error converts an Err Result to error and an Ok Result to nil. The returned
// error can be used with errors.Is and errors.As.
func Error[T, E any](value Result[T, E]) error {
	return value.IntoError()
}
