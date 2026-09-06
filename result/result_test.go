package result_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/23jdd/gomad/option"
	"github.com/23jdd/gomad/result"
)

func TestResultBasics(t *testing.T) {
	tests := []struct {
		name  string
		value result.Result[int, string]
		ok    bool
	}{
		{name: "ok", value: result.Ok[int, string](42), ok: true},
		{name: "err", value: result.Err[int]("failed")},
		{name: "zero value", value: result.Result[int, string]{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value.IsOk() != test.ok || test.value.IsErr() == test.ok {
				t.Fatalf("unexpected state: ok=%v err=%v", test.value.IsOk(), test.value.IsErr())
			}
		})
	}
	if result.Ok[int, string](4).Unwrap() != 4 {
		t.Fatal("Unwrap returned the wrong value")
	}
	if result.Err[int]("failed").UnwrapErr() != "failed" {
		t.Fatal("UnwrapErr returned the wrong value")
	}
}

func TestResultChaining(t *testing.T) {
	inspected := 0
	value := result.Ok[int, error](10).
		Map(func(value int) int { return value * 2 }).
		AndThen(func(value int) result.Result[int, error] { return result.Ok[int, error](value + 1) }).
		Inspect(func(value int) { inspected = value })
	if value.Unwrap() != 21 || inspected != 21 {
		t.Fatalf("chain = %v, inspected = %d", value, inspected)
	}

	mapped := result.Map(value, strconv.Itoa)
	if mapped.Unwrap() != "21" {
		t.Fatalf("cross-type Map = %q", mapped.Unwrap())
	}
	failed := result.Err[int](errors.New("failed"))
	called := false
	failed.Map(func(value int) int { called = true; return value }).Inspect(func(int) { called = true })
	failed.InspectErr(func(error) { called = true })
	if !called {
		t.Fatal("InspectErr was not called")
	}
	if result.Map(failed, strconv.Itoa).IsOk() {
		t.Fatal("Map changed Err into Ok")
	}
}

func TestResultFallbacksAndPanics(t *testing.T) {
	failed := result.Err[int]("failed")
	if failed.UnwrapOr(7) != 7 || failed.UnwrapOrElse(func(err string) int { return len(err) }) != 6 {
		t.Fatal("fallback returned the wrong value")
	}
	if got := failed.OrElse(func(string) result.Result[int, string] { return result.Ok[int, string](8) }).Unwrap(); got != 8 {
		t.Fatalf("OrElse = %d", got)
	}
	assertPanic(t, "failed", func() { failed.Unwrap() })
	assertPanic(t, "called Result.UnwrapErr on Ok", func() { result.Ok[int, string](1).UnwrapErr() })
	assertPanic(t, "context: failed", func() { failed.Expect("context") })
	assertPanic(t, "wanted error", func() { result.Ok[int, string](1).ExpectErr("wanted error") })
}

func TestFromAndConversions(t *testing.T) {
	parse := func(input string) (int, error) { return strconv.Atoi(input) }
	if result.From(parse("12")).Unwrap() != 12 || result.From(parse("x")).IsOk() {
		t.Fatal("From returned the wrong state")
	}
	missing := errors.New("missing")
	if option.None[int]().OkOr(missing).UnwrapErr() != missing {
		t.Fatal("Option.OkOr returned the wrong error")
	}
	if option.OkOr(option.Some(3), "missing").Unwrap() != 3 {
		t.Fatal("generic option.OkOr returned the wrong value")
	}
	if result.Ok[int, string](3).Ok().Unwrap() != 3 || result.Ok[int, string](3).Err().IsSome() {
		t.Fatal("Result Option conversion returned the wrong state")
	}
}

func TestMatchCollectAndPartition(t *testing.T) {
	matched := ""
	result.Match(result.Err[int]("bad"), func(int) { matched = "ok" }, func(err string) { matched = err })
	if matched != "bad" {
		t.Fatalf("Match = %q", matched)
	}
	values := []result.Result[int, string]{result.Ok[int, string](1), result.Ok[int, string](2)}
	if got := result.Collect(values).Unwrap(); len(got) != 2 || got[1] != 2 {
		t.Fatalf("Collect = %v", got)
	}
	values = append(values, result.Err[int]("bad"))
	if result.All(values) || result.Collect(values).UnwrapErr() != "bad" {
		t.Fatal("All or Collect returned the wrong error state")
	}
	okValues, errValues := result.Partition(values)
	if len(okValues) != 2 || len(errValues) != 1 || errValues[0] != "bad" {
		t.Fatalf("Partition = %v, %v", okValues, errValues)
	}
}

func TestErrorsIntegration(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := result.Err[int](sentinel).Wrap("load config")
	if !errors.Is(wrapped, sentinel) || wrapped.Error() != "load config: sentinel" {
		t.Fatalf("wrapped error = %v", wrapped)
	}
	var target *testError
	typed := result.Err[int](error(&testError{code: 7}))
	if !errors.As(typed, &target) || target.code != 7 {
		t.Fatalf("errors.As target = %#v", target)
	}
	if result.Ok[int, error](1).IntoError() != nil {
		t.Fatal("Ok.IntoError should return nil")
	}
}

func TestResultJSONAndIterator(t *testing.T) {
	okData, err := json.Marshal(result.Ok[int, string](4))
	if err != nil || string(okData) != `{"ok":4}` {
		t.Fatalf("Marshal Ok = %s, %v", okData, err)
	}
	errData, err := json.Marshal(result.Err[int](errors.New("bad")))
	if err != nil || string(errData) != `{"err":"bad"}` {
		t.Fatalf("Marshal Err = %s, %v", errData, err)
	}
	var decoded result.Result[int, error]
	if err := json.Unmarshal(errData, &decoded); err != nil || decoded.UnwrapErr().Error() != "bad" {
		t.Fatalf("Unmarshal Err = %v, %v", decoded, err)
	}
	if values := result.Ok[int, string](4).Iter().Collect(); len(values) != 1 || values[0] != 4 {
		t.Fatalf("Ok iterator = %v", values)
	}
	if result.Err[int]("bad").Iter().Len() != 0 {
		t.Fatal("Err iterator should be empty")
	}
}

func FuzzResultJSONRoundTrip(f *testing.F) {
	f.Add(1, "")
	f.Add(0, "failed")
	f.Fuzz(func(t *testing.T, value int, message string) {
		original := result.Ok[int, string](value)
		if message != "" {
			original = result.Err[int](message)
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		var decoded result.Result[int, string]
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		gotValue, gotErr, gotOK := decoded.Get()
		wantValue, wantErr, wantOK := original.Get()
		if gotValue != wantValue || gotErr != wantErr || gotOK != wantOK {
			t.Fatalf("round trip = (%v, %v, %v), want (%v, %v, %v)", gotValue, gotErr, gotOK, wantValue, wantErr, wantOK)
		}
	})
}

func BenchmarkResultMap(b *testing.B) {
	for b.Loop() {
		_ = result.Ok[int, error](10).Map(func(value int) int { return value * 2 })
	}
}

func BenchmarkFrom(b *testing.B) {
	for b.Loop() {
		_ = result.From(strconv.Atoi("42"))
	}
}

type testError struct{ code int }

func (err *testError) Error() string { return fmt.Sprintf("code %d", err.code) }

func assertPanic(t *testing.T, want any, function func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %v", got, want)
		}
	}()
	function()
}
