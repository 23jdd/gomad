// Package result 提供显式表示成功与失败的 Result 类型及辅助函数。
package result

import "github.com/23jdd/gomad"

// Result 表示一次成功或失败的计算。
type Result[T, E any] = gomad.Result[T, E]

// Ok 创建一个成功的 Result。
func Ok[T, E any](value T) Result[T, E] { return gomad.Ok[T, E](value) }

// Err 创建一个失败的 Result；可以显式指定 T 并让编译器推断 E。
func Err[T, E any](err E) Result[T, E] { return gomad.Err[T](err) }

// From 将 Go 常见的 (value, error) 返回值转换为 Result。
func From[T any](value T, err error) Result[T, error] {
	if err != nil {
		return Err[T](err)
	}
	return Ok[T, error](value)
}

// Map 转换 Ok 中的值，并允许改变成功值类型。
func Map[T, U, E any](value Result[T, E], transform func(T) U) Result[U, E] {
	if value.IsErr() {
		return Err[U](value.UnwrapErr())
	}
	return Ok[U, E](transform(value.Unwrap()))
}

// MapErr 转换 Err 中的值，并允许改变错误值类型。
func MapErr[T, E, F any](value Result[T, E], transform func(E) F) Result[T, F] {
	if value.IsOk() {
		return Ok[T, F](value.Unwrap())
	}
	return Err[T](transform(value.UnwrapErr()))
}

// AndThen 串联另一个返回 Result 的操作，并允许改变成功值类型。
func AndThen[T, U, E any](value Result[T, E], transform func(T) Result[U, E]) Result[U, E] {
	if value.IsErr() {
		return Err[U](value.UnwrapErr())
	}
	return transform(value.Unwrap())
}

// OrElse 恢复 Err 分支，并允许改变错误值类型。
func OrElse[T, E, F any](value Result[T, E], fallback func(E) Result[T, F]) Result[T, F] {
	if value.IsOk() {
		return Ok[T, F](value.Unwrap())
	}
	return fallback(value.UnwrapErr())
}

// Match 根据状态调用 ok 或 err 分支。
func Match[T, E any](value Result[T, E], ok func(T), err func(E)) {
	if value.IsOk() {
		ok(value.Unwrap())
		return
	}
	err(value.UnwrapErr())
}

// Collect 收集全部 Ok 值；遇到第一个 Err 时立即返回该错误。
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

// All 判断所有 Result 是否都是 Ok；空切片返回 true。
func All[T, E any](values []Result[T, E]) bool {
	for _, value := range values {
		if value.IsErr() {
			return false
		}
	}
	return true
}

// Partition 按原顺序把 Ok 与 Err 值分到两个切片中。
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

// Wrap 为错误 Result 添加上下文。
func Wrap[T any](value Result[T, error], message string) Result[T, error] {
	return value.Wrap(message)
}

// Error 将 Err 转为 error，将 Ok 转为 nil；返回值可用于 errors.Is 与 errors.As。
func Error[T, E any](value Result[T, E]) error {
	return value.IntoError()
}
