// Package option 提供零分配的 Option 类型及其常用辅助函数。
package option

import (
	"github.com/23jdd/gomad"
	"github.com/23jdd/gomad/iterator"
)

// Option 表示一个可能存在的值；其零值为 None。
type Option[T any] = gomad.Option[T]

// Some 创建一个包含 value 的 Option。
func Some[T any](value T) Option[T] { return gomad.Some(value) }

// None 创建一个不包含值的 Option。
func None[T any]() Option[T] { return gomad.None[T]() }

// Map 转换 Option 中的值，并允许改变值类型。
func Map[T, U any](value Option[T], transform func(T) U) Option[U] {
	if value.IsNone() {
		return None[U]()
	}
	return Some(transform(value.Unwrap()))
}

// AndThen 串联另一个返回 Option 的操作，并允许改变值类型。
func AndThen[T, U any](value Option[T], transform func(T) Option[U]) Option[U] {
	if value.IsNone() {
		return None[U]()
	}
	return transform(value.Unwrap())
}

// OkOr 将 Some 转为 Ok，将 None 转为包含 err 的 Err。
func OkOr[T, E any](value Option[T], err E) gomad.Result[T, E] {
	if value.IsSome() {
		return gomad.Ok[T, E](value.Unwrap())
	}
	return gomad.Err[T](err)
}

// OkOrElse 将 Some 转为 Ok，None 则调用 fallback 生成 Err。
func OkOrElse[T, E any](value Option[T], fallback func() E) gomad.Result[T, E] {
	if value.IsSome() {
		return gomad.Ok[T, E](value.Unwrap())
	}
	return gomad.Err[T](fallback())
}

// FromPtr 将指针转换为 Option；nil 指针会变为 None。
func FromPtr[T any](value *T) Option[T] {
	if value == nil {
		return None[T]()
	}
	return Some(*value)
}

// FromZero 将 T 的零值转换为 None，其他值转换为 Some。
func FromZero[T comparable](value T) Option[T] {
	var zero T
	if value == zero {
		return None[T]()
	}
	return Some(value)
}

// FromValueOk 将 Go 常见的 (value, ok) 组合转换为 Option。
func FromValueOk[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// FromValueError 将 (value, error) 组合转换为 Option。
func FromValueError[T any](value T, err error) Option[T] {
	if err != nil {
		return None[T]()
	}
	return Some(value)
}

// Flatten 去掉一层 Option 嵌套。
func Flatten[T any](value Option[Option[T]]) Option[T] {
	if value.IsNone() {
		return None[T]()
	}
	return value.Unwrap()
}

// Match 根据状态调用 some 或 none 分支。
func Match[T any](value Option[T], some func(T), none func()) {
	if value.IsSome() {
		some(value.Unwrap())
		return
	}
	none()
}

// Collect 收集全部 Some 值；任何一个 None 都会使结果为 None。
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

// All 判断所有 Option 是否都是 Some；空切片返回 true。
func All[T any](values []Option[T]) bool {
	for _, value := range values {
		if value.IsNone() {
			return false
		}
	}
	return true
}

// Any 判断是否至少存在一个 Some。
func Any[T any](values []Option[T]) bool {
	for _, value := range values {
		if value.IsSome() {
			return true
		}
	}
	return false
}

// FromIterator 将迭代器中的第一个元素转换为 Some，空迭代器转换为 None。
func FromIterator[T any](iter iterator.Iterator[T]) Option[T] {
	values := iter.Collect()
	if len(values) == 0 {
		return None[T]()
	}
	return Some(values[0])
}
