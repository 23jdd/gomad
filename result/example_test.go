package result_test

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/23jdd/gomad/result"
)

func ExampleOk() {
	value := result.Ok[int, error](10)
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleErr() {
	value := result.Err[int](errors.New("failed"))
	fmt.Println(value.UnwrapErr())
	// Output: failed
}

func ExampleFrom() {
	value := result.From(strconv.Atoi("10"))
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleMap() {
	value := result.Map(result.Ok[int, error](10), strconv.Itoa)
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleMapErr() {
	value := result.MapErr(result.Err[int]("failed"), errors.New)
	fmt.Println(value.UnwrapErr())
	// Output: failed
}

func ExampleAndThen() {
	value := result.AndThen(result.Ok[int, error](10), func(value int) result.Result[string, error] {
		return result.Ok[string, error](strconv.Itoa(value))
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleOrElse() {
	value := result.OrElse(result.Err[int]("failed"), func(string) result.Result[int, error] {
		return result.Ok[int, error](10)
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleMatch() {
	result.Match(
		result.Ok[int, error](10),
		func(value int) { fmt.Println("ok", value) },
		func(err error) { fmt.Println("err", err) },
	)
	// Output: ok 10
}

func ExampleCollect() {
	values := []result.Result[int, error]{
		result.Ok[int, error](1),
		result.Ok[int, error](2),
	}
	fmt.Println(result.Collect(values).Unwrap())
	// Output: [1 2]
}

func ExampleAll() {
	values := []result.Result[int, error]{
		result.Ok[int, error](1),
		result.Ok[int, error](2),
	}
	fmt.Println(result.All(values))
	// Output: true
}

func ExamplePartition() {
	values := []result.Result[int, string]{
		result.Ok[int, string](1),
		result.Err[int]("failed"),
		result.Ok[int, string](2),
	}
	okValues, errValues := result.Partition(values)
	fmt.Println(okValues)
	fmt.Println(errValues)
	// Output:
	// [1 2]
	// [failed]
}

func ExampleWrap() {
	value := result.Wrap(result.Err[int](error(errors.New("failed"))), "load")
	fmt.Println(value.IntoError())
	// Output: load: failed
}

func ExampleError() {
	err := result.Error(result.Err[int]("failed"))
	fmt.Println(err)
	// Output: failed
}

func ExampleResult_IsOk() {
	fmt.Println(result.Ok[int, error](10).IsOk())
	// Output: true
}

func ExampleResult_IsErr() {
	fmt.Println(result.Err[int]("failed").IsErr())
	// Output: true
}

func ExampleResult_Get() {
	value, err, ok := result.Err[int]("failed").Get()
	fmt.Printf("%d %q %t\n", value, err, ok)
	// Output: 0 "failed" false
}

func ExampleResult_Unwrap() {
	fmt.Println(result.Ok[int, error](10).Unwrap())
	// Output: 10
}

func ExampleResult_UnwrapErr() {
	fmt.Println(result.Err[int]("failed").UnwrapErr())
	// Output: failed
}

func ExampleResult_Expect() {
	fmt.Println(result.Ok[int, error](10).Expect("value is required"))
	// Output: 10
}

func ExampleResult_ExpectErr() {
	fmt.Println(result.Err[int]("failed").ExpectErr("error is required"))
	// Output: failed
}

func ExampleResult_UnwrapOr() {
	fmt.Println(result.Err[int]("failed").UnwrapOr(10))
	// Output: 10
}

func ExampleResult_UnwrapOrElse() {
	value := result.Err[int]("failed").UnwrapOrElse(func(err string) int { return len(err) })
	fmt.Println(value)
	// Output: 6
}

func ExampleResult_Map() {
	value := result.Ok[int, error](10).Map(strconv.Itoa)
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleResult_MapErr() {
	value := result.Err[int]("failed").MapErr(errors.New)
	fmt.Println(value.UnwrapErr())
	// Output: failed
}

func ExampleResult_AndThen() {
	value := result.Ok[int, error](10).AndThen(func(value int) result.Result[string, error] {
		return result.Ok[string, error](strconv.Itoa(value))
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleResult_OrElse() {
	value := result.Err[int]("failed").OrElse(func(string) result.Result[int, error] {
		return result.Ok[int, error](10)
	})
	fmt.Println(value.Unwrap())
	// Output: 10
}

func ExampleResult_Inspect() {
	value := result.Ok[int, error](10).Inspect(func(value int) { fmt.Println("inspect", value) })
	fmt.Println(value.Unwrap())
	// Output:
	// inspect 10
	// 10
}

func ExampleResult_InspectErr() {
	value := result.Err[int]("failed").InspectErr(func(err string) { fmt.Println("inspect", err) })
	fmt.Println(value.UnwrapErr())
	// Output:
	// inspect failed
	// failed
}

func ExampleResult_Ok() {
	fmt.Println(result.Ok[int, error](10).Ok().Unwrap())
	// Output: 10
}

func ExampleResult_Err() {
	fmt.Println(result.Err[int]("failed").Err().Unwrap())
	// Output: failed
}

func ExampleResult_Iter() {
	fmt.Println(result.Ok[int, error](10).Iter().Collect())
	// Output: [10]
}

func ExampleResult_IntoError() {
	fmt.Println(result.Err[int]("failed").IntoError())
	// Output: failed
}

func ExampleResult_Wrap() {
	value := result.Err[int](errors.New("failed")).Wrap("load")
	fmt.Println(value.IntoError())
	// Output: load: failed
}

func ExampleResult_MarshalJSON() {
	data, _ := result.Ok[int, string](10).MarshalJSON()
	fmt.Println(string(data))
	// Output: {"ok":10}
}

func ExampleResult_UnmarshalJSON() {
	var value result.Result[int, string]
	_ = value.UnmarshalJSON([]byte(`{"err":"failed"}`))
	fmt.Println(value.UnwrapErr())
	// Output: failed
}
