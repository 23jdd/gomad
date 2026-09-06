package core

import (
	"bytes"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/23jdd/gomad/iterator"
)

// Option contains either one value (Some) or no value (None).
// The fields are private so invalid states cannot be constructed.
type Option[T any] struct {
	value T
	some  bool
}

func Some[T any](value T) Option[T] { return Option[T]{value: value, some: true} }
func None[T any]() Option[T]        { return Option[T]{} }

func (opt Option[T]) IsSome() bool   { return opt.some }
func (opt Option[T]) IsNone() bool   { return !opt.some }
func (opt Option[T]) Get() (T, bool) { return opt.value, opt.some }

func (opt Option[T]) Unwrap() T {
	if !opt.some {
		panic("called Option.Unwrap on None")
	}
	return opt.value
}

// Unwarp is kept for source compatibility with gomad's initial release.
// Deprecated: use Unwrap.
func (opt Option[T]) Unwarp() T { return opt.Unwrap() }

func (opt Option[T]) Expect(message string) T {
	if !opt.some {
		panic(message)
	}
	return opt.value
}

func (opt Option[T]) UnwrapOr(fallback T) T {
	if opt.some {
		return opt.value
	}
	return fallback
}

func (opt Option[T]) UnwrapOrElse(fallback func() T) T {
	if opt.some {
		return opt.value
	}
	return fallback()
}

// Map transforms a value without changing its type. Use option.Map when the
// output type differs; Go methods cannot declare additional type parameters.
func (opt Option[T]) Map(transform func(T) T) Option[T] {
	if !opt.some {
		return None[T]()
	}
	return Some(transform(opt.value))
}

// AndThen chains an Option operation with the same value type. Use
// option.AndThen when the output type differs.
func (opt Option[T]) AndThen(transform func(T) Option[T]) Option[T] {
	if !opt.some {
		return None[T]()
	}
	return transform(opt.value)
}

func (opt Option[T]) OrElse(fallback func() Option[T]) Option[T] {
	if opt.some {
		return opt
	}
	return fallback()
}

func (opt Option[T]) Filter(predicate func(T) bool) Option[T] {
	if opt.some && !predicate(opt.value) {
		return None[T]()
	}
	return opt
}

func (opt Option[T]) Inspect(inspect func(T)) Option[T] {
	if opt.some {
		inspect(opt.value)
	}
	return opt
}

// OkOr converts Some to Ok and None to Err.
func (opt Option[T]) OkOr(err error) Result[T, error] {
	if opt.some {
		return Ok[T, error](opt.value)
	}
	return Err[T](err)
}

// OkOrElse lazily converts Some to Ok and None to Err.
func (opt Option[T]) OkOrElse(fallback func() error) Result[T, error] {
	if opt.some {
		return Ok[T, error](opt.value)
	}
	return Err[T](fallback())
}

// Flatten removes one nesting level when T itself is an Option. Prefer the
// statically constrained option.Flatten function in generic code.
func (opt Option[T]) Flatten() T {
	if opt.some {
		return opt.value
	}
	var zero T
	return zero
}

func (opt Option[T]) Iter() iterator.Iterator[T] {
	if !opt.some {
		return iterator.Empty[T]()
	}
	return iterator.Once(opt.value)
}

func (opt Option[T]) MarshalJSON() ([]byte, error) {
	if !opt.some {
		return []byte("null"), nil
	}
	return json.Marshal(opt.value)
}

func (opt *Option[T]) UnmarshalJSON(data []byte) error {
	if opt == nil {
		return errors.New("option.Option: UnmarshalJSON on nil pointer")
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*opt = None[T]()
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*opt = Some(value)
	return nil
}

// Scan implements sql.Scanner. SQL NULL becomes None.
func (opt *Option[T]) Scan(source any) error {
	if opt == nil {
		return errors.New("option.Option: Scan on nil pointer")
	}
	if source == nil {
		*opt = None[T]()
		return nil
	}

	var value T
	if scanner, ok := any(&value).(interface{ Scan(any) error }); ok {
		if err := scanner.Scan(source); err != nil {
			return err
		}
		*opt = Some(value)
		return nil
	}
	if unmarshaler, ok := any(&value).(encoding.TextUnmarshaler); ok {
		var text []byte
		switch source := source.(type) {
		case string:
			text = []byte(source)
		case []byte:
			text = source
		default:
			return fmt.Errorf("option.Option: cannot scan %T as text", source)
		}
		if err := unmarshaler.UnmarshalText(text); err != nil {
			return err
		}
		*opt = Some(value)
		return nil
	}
	if err := assignSQLValue(&value, source); err != nil {
		return err
	}
	*opt = Some(value)
	return nil
}

// Value implements driver.Valuer. None becomes SQL NULL.
func (opt Option[T]) Value() (driver.Value, error) {
	if !opt.some {
		return nil, nil
	}
	if valuer, ok := any(opt.value).(driver.Valuer); ok {
		return valuer.Value()
	}
	return driver.DefaultParameterConverter.ConvertValue(opt.value)
}

func assignSQLValue[T any](destination *T, source any) error {
	target := reflect.ValueOf(destination).Elem()
	input := reflect.ValueOf(source)
	if input.Type().AssignableTo(target.Type()) {
		if bytesValue, ok := source.([]byte); ok {
			input = reflect.ValueOf(bytes.Clone(bytesValue))
		}
		target.Set(input)
		return nil
	}
	if input.Type().ConvertibleTo(target.Type()) && target.Kind() != reflect.String {
		target.Set(input.Convert(target.Type()))
		return nil
	}

	text := ""
	switch source := source.(type) {
	case string:
		text = source
	case []byte:
		text = string(source)
	case int64, float64, bool:
		text = fmt.Sprint(source)
	default:
		return fmt.Errorf("option.Option: cannot scan %T into %T", source, *destination)
	}

	var err error
	switch target.Kind() {
	case reflect.String:
		target.SetString(text)
	case reflect.Bool:
		var parsed bool
		parsed, err = strconv.ParseBool(text)
		if err == nil {
			target.SetBool(parsed)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var parsed int64
		parsed, err = strconv.ParseInt(text, 10, target.Type().Bits())
		if err == nil {
			target.SetInt(parsed)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var parsed uint64
		parsed, err = strconv.ParseUint(text, 10, target.Type().Bits())
		if err == nil {
			target.SetUint(parsed)
		}
	case reflect.Float32, reflect.Float64:
		var parsed float64
		parsed, err = strconv.ParseFloat(text, target.Type().Bits())
		if err == nil {
			target.SetFloat(parsed)
		}
	default:
		return fmt.Errorf("option.Option: cannot scan %T into %T", source, *destination)
	}
	if err != nil {
		return fmt.Errorf("option.Option: converting %T to %T: %w", source, *destination, err)
	}
	return nil
}
