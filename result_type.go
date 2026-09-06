package gomad

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/23jdd/gomad/iterator"
)

// Result 表示一次可能成功或失败的计算：Ok 包含成功值，Err 包含错误值。
// Result 的零值是包含 E 零值的 Err。
type Result[T, E any] struct {
	value T
	err   E
	ok    bool
}

// Ok 创建一个包含成功值的 Result。
func Ok[T, E any](value T) Result[T, E] { return Result[T, E]{value: value, ok: true} }

// Err 创建一个包含错误值的 Result。
func Err[T, E any](err E) Result[T, E] { return Result[T, E]{err: err} }

// IsOk 判断 Result 是否为成功分支。
func (result Result[T, E]) IsOk() bool { return result.ok }

// IsErr 判断 Result 是否为错误分支。
func (result Result[T, E]) IsErr() bool { return !result.ok }

// Get 返回成功值、错误值以及是否成功。
func (result Result[T, E]) Get() (T, E, bool) {
	return result.value, result.err, result.ok
}

// Unwrap 返回 Ok 中的值；对 Err 调用时会触发 panic。
func (result Result[T, E]) Unwrap() T {
	if !result.ok {
		panic(result.err)
	}
	return result.value
}

// UnwrapErr 返回 Err 中的错误值；对 Ok 调用时会触发 panic。
func (result Result[T, E]) UnwrapErr() E {
	if result.ok {
		panic("called Result.UnwrapErr on Ok")
	}
	return result.err
}

// Expect 返回 Ok 中的值；对 Err 调用时使用 message 触发 panic。
func (result Result[T, E]) Expect(message string) T {
	if !result.ok {
		panic(fmt.Sprintf("%s: %v", message, result.err))
	}
	return result.value
}

// ExpectErr 返回 Err 中的错误值；对 Ok 调用时使用 message 触发 panic。
func (result Result[T, E]) ExpectErr(message string) E {
	if result.ok {
		panic(message)
	}
	return result.err
}

// UnwrapOr 返回 Ok 中的值，Err 则返回 fallback。
func (result Result[T, E]) UnwrapOr(fallback T) T {
	if result.ok {
		return result.value
	}
	return fallback
}

// UnwrapOrElse 返回 Ok 中的值，Err 则调用 fallback 生成默认值。
func (result Result[T, E]) UnwrapOrElse(fallback func(E) T) T {
	if result.ok {
		return result.value
	}
	return fallback(result.err)
}

// Map 转换 Ok 中的值，并允许改变成功值类型；Err 会直接传播。
func (result Result[T, E]) Map[U any](transform func(T) U) Result[U, E] {
	if !result.ok {
		return Err[U](result.err)
	}
	return Ok[U, E](transform(result.value))
}

// MapErr 转换 Err 中的值，并允许改变错误值类型；Ok 会直接传播。
func (result Result[T, E]) MapErr[F any](transform func(E) F) Result[T, F] {
	if result.ok {
		return Ok[T, F](result.value)
	}
	return Err[T](transform(result.err))
}

// AndThen 串联另一个返回 Result 的操作，并允许改变成功值类型。
func (result Result[T, E]) AndThen[U any](transform func(T) Result[U, E]) Result[U, E] {
	if !result.ok {
		return Err[U](result.err)
	}
	return transform(result.value)
}

// OrElse 恢复 Err 分支，并允许改变错误值类型。
func (result Result[T, E]) OrElse[F any](fallback func(E) Result[T, F]) Result[T, F] {
	if result.ok {
		return Ok[T, F](result.value)
	}
	return fallback(result.err)
}

// Inspect 在 Ok 分支执行只读回调，并返回原 Result。
func (result Result[T, E]) Inspect(inspect func(T)) Result[T, E] {
	if result.ok {
		inspect(result.value)
	}
	return result
}

// InspectErr 在 Err 分支执行只读回调，并返回原 Result。
func (result Result[T, E]) InspectErr(inspect func(E)) Result[T, E] {
	if !result.ok {
		inspect(result.err)
	}
	return result
}

// Ok 将成功分支转为 Some，将错误分支转为 None。
func (result Result[T, E]) Ok() Option[T] {
	if result.ok {
		return Some(result.value)
	}
	return None[T]()
}

// Err 将错误分支转为 Some，将成功分支转为 None。
func (result Result[T, E]) Err() Option[E] {
	if result.ok {
		return None[E]()
	}
	return Some(result.err)
}

// Iter 将 Ok 转为单元素迭代器，将 Err 转为空迭代器。
func (result Result[T, E]) Iter() iterator.Iterator[T] {
	if !result.ok {
		return iterator.Empty[T]()
	}
	return iterator.Once(result.value)
}

// IntoError 将 Ok 转为 nil，将 Err 转为标准 error。
func (result Result[T, E]) IntoError() error {
	if result.ok {
		return nil
	}
	if err, ok := any(result.err).(error); ok {
		return err
	}
	return fmt.Errorf("%v", result.err)
}

// Wrap 为 Err 添加上下文，并将错误类型统一为 error。
func (result Result[T, E]) Wrap(message string) Result[T, error] {
	if result.ok {
		return Ok[T, error](result.value)
	}
	return Err[T](fmt.Errorf("%s: %w", message, result.IntoError()))
}

// MarshalJSON 将 Result 编码为 {"ok": value} 或 {"err": err}。
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

// UnmarshalJSON 从只包含 ok 或 err 的 JSON 对象解码 Result。
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
