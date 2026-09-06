package option

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"testing"
)

func TestOptionBasics(t *testing.T) {
	tests := []struct {
		name  string
		value Option[int]
		some  bool
		want  int
	}{
		{name: "some", value: Some(42), some: true, want: 42},
		{name: "none", value: None[int]()},
		{name: "zero value", value: Option[int]{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value.IsSome() != test.some || test.value.IsNone() == test.some {
				t.Fatalf("unexpected state: some=%v none=%v", test.value.IsSome(), test.value.IsNone())
			}
			got, ok := test.value.Get()
			if ok != test.some || got != test.want {
				t.Fatalf("Get() = (%v, %v), want (%v, %v)", got, ok, test.want, test.some)
			}
		})
	}
}

func TestOptionChaining(t *testing.T) {
	inspected := 0
	got := Some(10).
		Map(func(value int) string { return strconv.Itoa(value) }).
		AndThen(func(value string) Option[int] { return Some(len(value)) }).
		Filter(func(value int) bool { return value > 1 }).
		Inspect(func(value int) { inspected = value })
	if got.Unwrap() != 2 || inspected != 2 {
		t.Fatalf("chain = %v, inspected = %d", got, inspected)
	}

	called := false
	none := None[int]().
		Map(func(value int) string { called = true; return strconv.Itoa(value) }).
		AndThen(func(value string) Option[int] { called = true; return Some(len(value)) }).
		Inspect(func(int) { called = true })
	if none.IsSome() || called {
		t.Fatal("None chain invoked a callback")
	}
}

func TestOptionFallbacksAndPanics(t *testing.T) {
	if got := None[int]().UnwrapOr(7); got != 7 {
		t.Fatalf("UnwrapOr = %d", got)
	}
	if got := None[int]().UnwrapOrElse(func() int { return 8 }); got != 8 {
		t.Fatalf("UnwrapOrElse = %d", got)
	}
	if got := None[int]().OrElse(func() Option[int] { return Some(9) }).Unwrap(); got != 9 {
		t.Fatalf("OrElse = %d", got)
	}
	assertPanic(t, "called Option.Unwrap on None", func() { None[int]().Unwrap() })
	assertPanic(t, "missing", func() { None[int]().Expect("missing") })
}

func TestOptionConstructorsAndCollections(t *testing.T) {
	value := 3
	if FromPtr(&value).Unwrap() != 3 || FromPtr[int](nil).IsSome() {
		t.Fatal("FromPtr returned the wrong state")
	}
	if FromZero(0).IsSome() || FromZero(3).Unwrap() != 3 {
		t.Fatal("FromZero returned the wrong state")
	}
	if FromValueOk("x", true).Unwrap() != "x" || FromValueOk("x", false).IsSome() {
		t.Fatal("FromValueOk returned the wrong state")
	}
	values := []Option[int]{Some(1), Some(2)}
	if got := Collect(values).Unwrap(); len(got) != 2 || got[1] != 2 {
		t.Fatalf("Collect = %v", got)
	}
	if !All(values) || Any([]Option[int]{None[int](), None[int]()}) {
		t.Fatal("All or Any returned the wrong result")
	}
	if Collect([]Option[int]{Some(1), None[int]()}).IsSome() {
		t.Fatal("Collect should stop at None")
	}
	if Flatten(Some(Some(5))).Unwrap() != 5 || Flatten(Some(None[int]())).IsSome() {
		t.Fatal("Flatten returned the wrong state")
	}
}

func TestOptionMatchAndIterator(t *testing.T) {
	matched := 0
	Match(Some(4), func(value int) { matched = value }, func() { matched = -1 })
	if matched != 4 {
		t.Fatalf("Match selected the wrong branch: %d", matched)
	}
	values := Some(4).Iter().Map(func(value int) int { return value * 2 }).Filter(func(value int) bool { return value > 4 }).Collect()
	if len(values) != 1 || values[0] != 8 || None[int]().Iter().Len() != 0 {
		t.Fatalf("unexpected iterator output: %v", values)
	}
}

func TestOptionJSON(t *testing.T) {
	data, err := json.Marshal(Some("hello"))
	if err != nil || string(data) != `"hello"` {
		t.Fatalf("Marshal = %s, %v", data, err)
	}
	data, err = json.Marshal(None[string]())
	if err != nil || string(data) != "null" {
		t.Fatalf("Marshal None = %s, %v", data, err)
	}
	var some Option[string]
	if err := json.Unmarshal([]byte(`"hello"`), &some); err != nil || some.Unwrap() != "hello" {
		t.Fatalf("Unmarshal Some = %v, %v", some, err)
	}
	if err := json.Unmarshal([]byte("null"), &some); err != nil || some.IsSome() {
		t.Fatalf("Unmarshal None = %v, %v", some, err)
	}
}

func TestOptionSQL(t *testing.T) {
	var value Option[int]
	if err := value.Scan(int64(12)); err != nil || value.Unwrap() != 12 {
		t.Fatalf("Scan = %v, %v", value, err)
	}
	if err := value.Scan(nil); err != nil || value.IsSome() {
		t.Fatalf("Scan NULL = %v, %v", value, err)
	}
	databaseValue, err := Some("hello").Value()
	if err != nil || databaseValue != driver.Value("hello") {
		t.Fatalf("Value = %v, %v", databaseValue, err)
	}
	databaseValue, err = None[string]().Value()
	if err != nil || databaseValue != nil {
		t.Fatalf("None Value = %v, %v", databaseValue, err)
	}
}

func FuzzOptionJSONRoundTrip(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Fuzz(func(t *testing.T, input string) {
		data, err := json.Marshal(Some(input))
		if err != nil {
			t.Fatal(err)
		}
		var output Option[string]
		if err := json.Unmarshal(data, &output); err != nil {
			t.Fatal(err)
		}
		if output.Unwrap() != input {
			t.Fatalf("round trip = %q, want %q", output.Unwrap(), input)
		}
	})
}

func BenchmarkOptionMap(b *testing.B) {
	for b.Loop() {
		_ = Some(10).Map(func(value int) int { return value * 2 })
	}
}

func BenchmarkValueOk(b *testing.B) {
	values := map[string]int{"answer": 42}
	b.Run("plain", func(b *testing.B) {
		for b.Loop() {
			_, _ = values["answer"]
		}
	})
	b.Run("option", func(b *testing.B) {
		for b.Loop() {
			value, ok := values["answer"]
			_ = FromValueOk(value, ok)
		}
	})
}

func assertPanic(t *testing.T, want any, function func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %v", got, want)
		}
	}()
	function()
}
