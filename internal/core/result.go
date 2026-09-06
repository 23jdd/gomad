package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/23jdd/gomad/iterator"
)

// Result contains either a successful value (Ok) or an error value (Err).
// Its zero value is Err containing the zero value of E.
type Result[T, E any] struct {
	value T
	err   E
	ok    bool
}

func Ok[T, E any](value T) Result[T, E] { return Result[T, E]{value: value, ok: true} }
func Err[T, E any](err E) Result[T, E]  { return Result[T, E]{err: err} }

func (result Result[T, E]) IsOk() bool  { return result.ok }
func (result Result[T, E]) IsErr() bool { return !result.ok }

func (result Result[T, E]) Get() (T, E, bool) {
	return result.value, result.err, result.ok
}

func (result Result[T, E]) Unwrap() T {
	if !result.ok {
		panic(result.err)
	}
	return result.value
}

func (result Result[T, E]) UnwrapErr() E {
	if result.ok {
		panic("called Result.UnwrapErr on Ok")
	}
	return result.err
}

func (result Result[T, E]) Expect(message string) T {
	if !result.ok {
		panic(fmt.Sprintf("%s: %v", message, result.err))
	}
	return result.value
}

func (result Result[T, E]) ExpectErr(message string) E {
	if result.ok {
		panic(message)
	}
	return result.err
}

func (result Result[T, E]) UnwrapOr(fallback T) T {
	if result.ok {
		return result.value
	}
	return fallback
}

func (result Result[T, E]) UnwrapOrElse(fallback func(E) T) T {
	if result.ok {
		return result.value
	}
	return fallback(result.err)
}

// Map transforms an Ok value and may change its type.
func (result Result[T, E]) Map[U any](transform func(T) U) Result[U, E] {
	if !result.ok {
		return Err[U](result.err)
	}
	return Ok[U, E](transform(result.value))
}

// MapErr transforms an Err value and may change its type.
func (result Result[T, E]) MapErr[F any](transform func(E) F) Result[T, F] {
	if result.ok {
		return Ok[T, F](result.value)
	}
	return Err[T](transform(result.err))
}

// AndThen chains an operation and may change the Ok type.
func (result Result[T, E]) AndThen[U any](transform func(T) Result[U, E]) Result[U, E] {
	if !result.ok {
		return Err[U](result.err)
	}
	return transform(result.value)
}

// OrElse recovers from an Err and may change its error type.
func (result Result[T, E]) OrElse[F any](fallback func(E) Result[T, F]) Result[T, F] {
	if result.ok {
		return Ok[T, F](result.value)
	}
	return fallback(result.err)
}

func (result Result[T, E]) Inspect(inspect func(T)) Result[T, E] {
	if result.ok {
		inspect(result.value)
	}
	return result
}

func (result Result[T, E]) InspectErr(inspect func(E)) Result[T, E] {
	if !result.ok {
		inspect(result.err)
	}
	return result
}

// Ok converts the successful branch to Some and the error branch to None.
func (result Result[T, E]) Ok() Option[T] {
	if result.ok {
		return Some(result.value)
	}
	return None[T]()
}

// Err converts the error branch to Some and the successful branch to None.
func (result Result[T, E]) Err() Option[E] {
	if result.ok {
		return None[E]()
	}
	return Some(result.err)
}

func (result Result[T, E]) Iter() iterator.Iterator[T] {
	if !result.ok {
		return iterator.Empty[T]()
	}
	return iterator.Once(result.value)
}

// IntoError returns nil for Ok and an error representation for Err.
func (result Result[T, E]) IntoError() error {
	if result.ok {
		return nil
	}
	if err, ok := any(result.err).(error); ok {
		return err
	}
	return fmt.Errorf("%v", result.err)
}

// Error lets Result participate in the standard errors package. Ok returns an
// empty string; use IntoError when a nil-on-success error is needed.
func (result Result[T, E]) Error() string {
	if result.ok {
		return ""
	}
	return result.IntoError().Error()
}

// Is supports errors.Is for Err values whose E implements error.
func (result Result[T, E]) Is(target error) bool {
	return !result.ok && errors.Is(result.IntoError(), target)
}

// As supports errors.As for Err values whose E implements error.
func (result Result[T, E]) As(target any) bool {
	return !result.ok && errors.As(result.IntoError(), target)
}

// Wrap adds context to an Err value and normalizes its error type to error.
func (result Result[T, E]) Wrap(message string) Result[T, error] {
	if result.ok {
		return Ok[T, error](result.value)
	}
	return Err[T](fmt.Errorf("%s: %w", message, result.IntoError()))
}

// MarshalJSON encodes Result as exactly one of {"ok": value} or {"err": err}.
func (result Result[T, E]) MarshalJSON() ([]byte, error) {
	key := "ok"
	value := any(result.value)
	if !result.ok {
		key = "err"
		value = result.err
		if err, ok := value.(error); ok {
			value = err.Error()
		}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return fmt.Appendf(nil, "{\"%s\":%s}", key, data), nil
}

func (result *Result[T, E]) UnmarshalJSON(data []byte) error {
	if result == nil {
		return errors.New("result.Result: UnmarshalJSON on nil pointer")
	}
	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&object); err != nil {
		return err
	}
	if len(object) != 1 {
		return errors.New("result.Result: expected exactly one of ok or err")
	}
	if raw, exists := object["ok"]; exists {
		var value T
		if err := json.Unmarshal(raw, &value); err != nil {
			return err
		}
		*result = Ok[T, E](value)
		return nil
	}
	raw, exists := object["err"]
	if !exists {
		return errors.New("result.Result: expected exactly one of ok or err")
	}
	var errValue E
	if err := json.Unmarshal(raw, &errValue); err == nil {
		*result = Err[T](errValue)
		return nil
	}
	var message string
	if err := json.Unmarshal(raw, &message); err != nil {
		return err
	}
	converted, ok := any(errors.New(message)).(E)
	if !ok {
		return fmt.Errorf("result.Result: cannot decode error string into %T", errValue)
	}
	*result = Err[T](converted)
	return nil
}
