package gomad

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

// Option 表示一个可能存在的值：Some 包含值，None 不包含值。
// Option 的零值等价于 None。
type Option[T any] struct {
	value T
	some  bool
}

// Some 创建一个包含 value 的 Option。
func Some[T any](value T) Option[T] { return Option[T]{value: value, some: true} }

// None 创建一个不包含值的 Option。
func None[T any]() Option[T] { return Option[T]{} }

// IsSome 判断 Option 是否包含值。
func (opt Option[T]) IsSome() bool { return opt.some }

// IsNone 判断 Option 是否不包含值。
func (opt Option[T]) IsNone() bool { return !opt.some }

// Get 返回内部值以及表示值是否存在的布尔值。
func (opt Option[T]) Get() (T, bool) { return opt.value, opt.some }

// Unwrap 返回 Some 中的值；对 None 调用时会触发 panic。
func (opt Option[T]) Unwrap() T {
	if !opt.some {
		panic("called Option.Unwrap on None")
	}
	return opt.value
}

// Unwarp 为早期版本的拼写错误保留源码兼容性。
// Deprecated: 请使用 Unwrap。
func (opt Option[T]) Unwarp() T { return opt.Unwrap() }

// Expect 返回 Some 中的值；对 None 调用时使用 message 触发 panic。
func (opt Option[T]) Expect(message string) T {
	if !opt.some {
		panic(message)
	}
	return opt.value
}

// UnwrapOr 返回 Some 中的值，None 则返回 fallback。
func (opt Option[T]) UnwrapOr(fallback T) T {
	if opt.some {
		return opt.value
	}
	return fallback
}

// UnwrapOrElse 返回 Some 中的值，None 则调用 fallback 生成默认值。
func (opt Option[T]) UnwrapOrElse(fallback func() T) T {
	if opt.some {
		return opt.value
	}
	return fallback()
}

// Map 转换 Some 中的值，并允许改变值类型；None 会直接传播。
func (opt Option[T]) Map[U any](transform func(T) U) Option[U] {
	if !opt.some {
		return None[U]()
	}
	return Some(transform(opt.value))
}

// AndThen 串联另一个返回 Option 的操作，并允许改变值类型。
func (opt Option[T]) AndThen[U any](transform func(T) Option[U]) Option[U] {
	if !opt.some {
		return None[U]()
	}
	return transform(opt.value)
}

// OrElse 为 None 延迟生成另一个 Option，Some 保持不变。
func (opt Option[T]) OrElse(fallback func() Option[T]) Option[T] {
	if opt.some {
		return opt
	}
	return fallback()
}

// Filter 仅在谓词接受 Some 中的值时保留该值。
func (opt Option[T]) Filter(predicate func(T) bool) Option[T] {
	if opt.some && !predicate(opt.value) {
		return None[T]()
	}
	return opt
}

// Inspect 在 Some 分支执行只读回调，并返回原 Option。
func (opt Option[T]) Inspect(inspect func(T)) Option[T] {
	if opt.some {
		inspect(opt.value)
	}
	return opt
}

// OkOr 将 Some 转为 Ok，将 None 转为包含 err 的 Err。
func (opt Option[T]) OkOr[E any](err E) Result[T, E] {
	if opt.some {
		return Ok[T, E](opt.value)
	}
	return Err[T](err)
}

// OkOrElse 将 Some 转为 Ok，None 则调用 fallback 生成 Err。
func (opt Option[T]) OkOrElse[E any](fallback func() E) Result[T, E] {
	if opt.some {
		return Ok[T, E](opt.value)
	}
	return Err[T](fallback())
}

// Flatten 去掉一层 Option 嵌套。
// 泛型代码中建议优先使用 option.Flatten，以获得更严格的类型约束。
func (opt Option[T]) Flatten() T {
	if opt.some {
		return opt.value
	}
	var zero T
	return zero
}

// Iter 将 Some 转为单元素迭代器，将 None 转为空迭代器。
func (opt Option[T]) Iter() iterator.Iterator[T] {
	if !opt.some {
		return iterator.Empty[T]()
	}
	return iterator.Once(opt.value)
}

// MarshalJSON 将 Some 编码为内部值，将 None 编码为 null。
func (opt Option[T]) MarshalJSON() ([]byte, error) {
	if !opt.some {
		return []byte("null"), nil
	}
	return json.Marshal(opt.value)
}

// UnmarshalJSON 将 null 解码为 None，其他 JSON 值解码为 Some。
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

// Scan 实现 sql.Scanner；SQL NULL 会转换为 None。
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

// Value 实现 driver.Valuer；None 会转换为 SQL NULL。
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
