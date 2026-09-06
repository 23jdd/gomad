package option_test

import (
	"errors"
	"fmt"

	"github.com/23jdd/gomad/iterator"
	"github.com/23jdd/gomad/option"
)

func ExampleSome() {
	value := option.Some(10)
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleNone() {
	value := option.None[int]()
	fmt.Println(value.IsNone())
	// Output: true
}

func ExampleMap() {
	value := option.Map(option.Some(10), func(value int) string {
		return fmt.Sprintf("value=%d", value)
	})
	fmt.Println(value.Unwrap())
	// Output: value=10
}

func ExampleAndThen() {
	value := option.AndThen(option.Some(10), func(value int) option.Option[string] {
		return option.Some(fmt.Sprintf("%d", value))
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOkOr() {
	value := option.OkOr(option.None[int](), "missing")
	fmt.Println(value.UnwrapErr())
	// Output: missing
}

func ExampleOkOrElse() {
	value := option.OkOrElse(option.None[int](), func() string { return "missing" })
	fmt.Println(value.UnwrapErr())
	// Output: missing
}

func ExampleFromPtr() {
	value := 10
	fmt.Println(option.FromPtr(&value).Unwrap())
	fmt.Println(option.FromPtr[int](nil).IsNone())
	// Output:
	// 10
	// true
}

func ExampleFromZero() {
	fmt.Println(option.FromZero(0).IsNone())
	fmt.Println(option.FromZero(10).Unwrap())
	// Output:
	// true
	// 10
}

func ExampleFromValueOk() {
	values := map[string]int{"answer": 42}
	value, ok := values["answer"]
	fmt.Println(option.FromValueOk(value, ok).Unwrap())
	// Output: 42
}

func ExampleFromValueError() {
	fmt.Println(option.FromValueError(10, nil).Unwrap())
	fmt.Println(option.FromValueError(10, errors.New("failed")).IsNone())
	// Output:
	// 10
	// true
}

func ExampleFlatten() {
	value := option.Flatten(option.Some(option.Some(10)))
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleMatch() {
	option.Match(
		option.Some(10),
		func(value int) { fmt.Println("some", value) },
		func() { fmt.Println("none") },
	)
	// Output: some 10
}

func ExampleCollect() {
	value := option.Collect([]option.Option[int]{option.Some(1), option.Some(2)})
	fmt.Println(value.Unwrap())
	// Output: [1 2]
}

func ExampleAll() {
	values := []option.Option[int]{option.Some(1), option.Some(2)}
	fmt.Println(option.All(values))
	// Output: true
}

func ExampleAny() {
	values := []option.Option[int]{option.None[int](), option.Some(2)}
	fmt.Println(option.Any(values))
	// Output: true
}

func ExampleFromIterator() {
	value := option.FromIterator(iterator.Once(10))
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOption_IsSome() {
	fmt.Println(option.Some(10).IsSome())
	// Output: true
}

func ExampleOption_IsNone() {
	fmt.Println(option.None[int]().IsNone())
	// Output: true
}

func ExampleOption_Get() {
	value, ok := option.Some(10).Get()
	fmt.Println(value, ok)
	// Output: 10 true
}

func ExampleOption_Unwrap() {
	fmt.Println(option.Some(10).Unwrap())
	// Output: 10
}

func ExampleOption_Unwarp() {
	fmt.Println(option.Some(10).Unwarp())
	// Output: 10
}

func ExampleOption_Expect() {
	fmt.Println(option.Some(10).Expect("value is required"))
	// Output: 10
}

func ExampleOption_UnwrapOr() {
	fmt.Println(option.None[int]().UnwrapOr(10))
	// Output: 10
}

func ExampleOption_UnwrapOrElse() {
	value := option.None[int]().UnwrapOrElse(func() int { return 10 })
	fmt.Println(value)
	// Output: 10
}

func ExampleOption_Map() {
	value := option.Some(10).Map(func(value int) string {
		return fmt.Sprintf("value=%d", value)
	})
	fmt.Println(value.Unwrap())
	// Output: value=10
}

func ExampleOption_AndThen() {
	value := option.Some(10).AndThen(func(value int) option.Option[string] {
		return option.Some(fmt.Sprintf("%d", value))
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOption_OrElse() {
	value := option.None[int]().OrElse(func() option.Option[int] {
		return option.Some(10)
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOption_Filter() {
	fmt.Println(option.Some(10).Filter(func(value int) bool { return value > 5 }).IsSome())
	// Output: true
}

func ExampleOption_Inspect() {
	value := option.Some(10).Inspect(func(value int) { fmt.Println("inspect", value) })
	fmt.Println(value.Unwrap())
	// Output:
	// inspect 10
	// 10
}

func ExampleOption_OkOr() {
	value := option.None[int]().OkOr(errors.New("missing"))
	fmt.Println(value.UnwrapErr())
	// Output: missing
}

func ExampleOption_OkOrElse() {
	value := option.None[int]().OkOrElse(func() error { return errors.New("missing") })
	fmt.Println(value.UnwrapErr())
	// Output: missing
}

func ExampleOption_Flatten() {
	value := option.Some(option.Some(10)).Flatten()
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOption_Iter() {
	values := option.Some(10).Iter().Collect()
	fmt.Println(values)
	// Output: [10]
}

func ExampleOption_MarshalJSON() {
	data, _ := option.Some("hello").MarshalJSON()
	fmt.Println(string(data))
	// Output: "hello"
}

func ExampleOption_UnmarshalJSON() {
	var value option.Option[string]
	_ = value.UnmarshalJSON([]byte(`"hello"`))
	fmt.Println(value.Unwrap())
	// Output: hello
}

func ExampleOption_Scan() {
	var value option.Option[int]
	_ = value.Scan(int64(10))
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOption_Value() {
	value, _ := option.Some("hello").Value()
	fmt.Println(value)
	// Output: hello
}
