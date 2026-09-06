package option

type Option[T any] struct {
	value T
	some  bool
}

func Some[T any](value T) *Option[T] {
	return &Option[T]{
		value: value,
		some:  true,
	}
}
func None[T any]() *Option[T] {
	return &Option[T]{
		some: false,
	}
}

func (opt *Option[T]) IsSome() bool {
	return opt.some
}
func (opt *Option[T]) IsNone() bool {
	return !opt.some
}
func (opt *Option[T]) Get() (T, bool) {
	return opt.value, opt.some
}
func (opt *Option[T]) Unwarp() T {
	if !opt.some {
		panic("called Unwrap on None")
	}
	return opt.value
}
func (opt *Option[T]) Expect(msg string) T {
	if !opt.some {
		panic(msg)
	}
	return opt.value
}
func (opt *Option[T]) UnwrapOr(value T) T {
	if opt.some {
		return opt.value
	}
	return value
}
func (opt *Option[T]) UnwrapOrElse(f func() T) T {
	if opt.some {
		return opt.value
	}
	return f()
}
