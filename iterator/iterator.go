// Package iterator 提供 Option 与 Result 使用的小型值迭代器。
package iterator

// Iterator 表示一个不可变、可重复使用的有限序列迭代器。
type Iterator[T any] struct {
	values []T
}

// Empty 创建一个空迭代器。
func Empty[T any]() Iterator[T] { return Iterator[T]{} }

// Once 创建一个只包含 value 的迭代器。
func Once[T any](value T) Iterator[T] { return Iterator[T]{values: []T{value}} }

// FromSlice 基于 values 的副本创建迭代器。
func FromSlice[T any](values []T) Iterator[T] {
	return Iterator[T]{values: append([]T(nil), values...)}
}

// Map 转换每个元素，并允许改变元素类型。
func (iter Iterator[T]) Map[U any](transform func(T) U) Iterator[U] {
	values := make([]U, len(iter.values))
	for index, value := range iter.values {
		values[index] = transform(value)
	}
	return Iterator[U]{values: values}
}

// Map 转换每个元素，并允许改变元素类型。
func Map[T, U any](iter Iterator[T], transform func(T) U) Iterator[U] {
	values := make([]U, len(iter.values))
	for index, value := range iter.values {
		values[index] = transform(value)
	}
	return Iterator[U]{values: values}
}

// Filter 仅保留 predicate 接受的元素。
func (iter Iterator[T]) Filter(predicate func(T) bool) Iterator[T] {
	values := make([]T, 0, len(iter.values))
	for _, value := range iter.values {
		if predicate(value) {
			values = append(values, value)
		}
	}
	return Iterator[T]{values: values}
}

// Collect 返回序列的副本。
func (iter Iterator[T]) Collect() []T {
	return append([]T(nil), iter.values...)
}

// Len 返回迭代器中的元素数量。
func (iter Iterator[T]) Len() int { return len(iter.values) }
