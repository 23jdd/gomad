# gomad

[![Go Reference](https://pkg.go.dev/badge/github.com/23jdd/gomad.svg)](https://pkg.go.dev/github.com/23jdd/gomad)
[![Go 1.27+](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![MIT License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Runnable Examples](https://img.shields.io/badge/examples-runnable-2563eb)](option/example_test.go)

> Rust-style Option and Result ergonomics, designed for modern Go.

[Quick start](#30-second-tour) · [Why gomad](#why) · [API reference](#api-reference) ·
[Examples](examples/main.go) · [简体中文](README-zh_cn.md)

`gomad` is a lightweight Option and Result library for Go 1.27+, built around
generic methods and zero-cost value types.

It brings explicit optional values and typed failures to Go while keeping the
API familiar: values are plain structs, zero values are valid, and adapters are
provided for `(value, ok)`, `(value, error)`, JSON, `database/sql`, and the
standard `errors` package.

If you are looking for type-safe null handling, composable Go error handling,
or Rust-inspired functional primitives without abandoning Go conventions,
`gomad` is built for that exact middle ground.

## Highlights

| | What you get |
| --- | --- |
| **Go 1.27 native** | Type-changing generic method chains such as `Option[int].Map(...) -> Option[string]` |
| **Go-friendly boundaries** | Adapters for pointers, map lookups, `(T, error)`, `errors.Is/As`, JSON, and SQL |
| **Predictable values** | Valid zero values, no hidden global state, and no mandatory heap allocation |
| **Small dependency surface** | Standard library only; no runtime framework or code generation |
| **Executable documentation** | Every public Option, Result, and Iterator API has a tested example |

## 30-second tour

```go
name := option.Some(1).
	AndThen(findUser).
	Map(func(user User) string { return user.Name }).
	Filter(func(name string) bool { return name != "" })

config := result.From(os.ReadFile("config.json")).
	Map(parseConfig).
	AndThen(validateConfig).
	Inspect(func(Config) { log.Println("config loaded") }).
	InspectErr(func(err error) { log.Println("config error:", err) })
```

Both chains remain statically typed from end to end. None and Err branches skip
success callbacks automatically, so the happy path stays readable without
hiding failure handling.

## Why

Pointers overload “absent” with allocation and mutability, while `(T, error)`
pairs are easy to accidentally separate. `Option[T]` and `Result[T, E]` keep
the state and payload together, make every branch explicit, and compose with
Go 1.27 generic method chains.

gomad is a good fit for parsers, configuration loaders, API clients, database
boundaries, validation pipelines, and domain models where “missing” and
“failed” should be impossible to confuse with ordinary values.

## Install

```shell
go get github.com/23jdd/gomad@latest
```

Your module must use Go 1.27 or newer:

```go
module example.com/myapp

go 1.27
```

## Option

```go
value := option.Some(10)

if value.IsSome() {
	fmt.Println(value.Unwrap())
}

fallback := option.None[int]().UnwrapOr(42)
```

The zero value of `Option[T]` is `None`. Constructors cover common Go forms:

```go
fromPointer := option.FromPtr(ptr)
fromLookup := option.FromValueOk(value, ok)
fromCall := option.FromValueError(value, err)
```

A map lookup must first bind its comma-ok result because Go does not pass a map
index as a two-value function argument:

```go
value, ok := values[key]
found := option.FromValueOk(value, ok)
```

## Result

```go
func divide(a, b int) result.Result[int, error] {
	if b == 0 {
		return result.Err[int](errors.New("division by zero"))
	}
	return result.Ok[int, error](a / b)
}
```

Use `From` with ordinary Go functions that return `(T, error)`:

```go
contents := result.From(os.ReadFile("config.json"))
number := result.From(strconv.Atoi("42"))
```

The zero value of `Result[T, E]` is `Err` containing the zero value of `E`.

## Map

Go 1.27 generic methods let a chain change its value type without losing static
type safety:

```go
length := option.Some(10).
	Map(strconv.Itoa).
	Map(func(value string) int { return len(value) })

text := result.From(strconv.Atoi("21")).
	Map(func(value int) int { return value * 2 }).
	Map(strconv.Itoa)
```

Package-level `option.Map`, `result.Map`, and `result.MapErr` equivalents are
also available when function composition is more convenient.

## AndThen

Use `AndThen` when the callback already returns an `Option` or `Result`:

```go
name := option.Some(1).
	AndThen(findUser).
	Map(func(user User) string { return user.Name }).
	Filter(func(name string) bool { return name != "" })

config := result.From(os.ReadFile("config.json")).
	Map(parseConfig).
	AndThen(validateConfig)
```

`OrElse` provides lazy recovery. `Inspect` and `InspectErr` observe a chain
without changing its value.

## Error handling

Convert between the two types without unpacking them:

```go
required := optionalValue.OkOr(errors.New("value is required"))
optional := required.Ok()
failure := required.Err()
```

For `Result[T, error]`, standard error traversal works directly:

```go
wrapped := result.From(load()).Wrap("load configuration")

if errors.Is(wrapped.IntoError(), fs.ErrNotExist) {
	// handle a missing file
}

err := wrapped.IntoError() // nil for Ok, an error for Err
```

`errors.As` is supported in the same way. The package-level `result.Error`
helper is equivalent to `IntoError`, and `Wrap` adds context with `%w`
semantics. Result itself deliberately does not implement `error`: its required
`Unwrap() T` value method would conflict with the standard `Unwrap() error`
error-chain convention.

## Option vs pointer

Use a pointer when identity, shared mutation, or a large object makes pointer
semantics meaningful. Use `Option[T]` when the important fact is simply whether
a value exists. `Option` cannot be confused with a present nil pointer, has a
valid zero value, and simple values stay allocation-free.

## Result vs `(T, error)`

Use ordinary `(T, error)` at conventional Go API boundaries. Convert to
`Result` when several transformations need to be composed, a typed non-error
failure is useful, or the combined value needs to be stored or passed around.
`result.From` and `IntoError` make both styles interoperable.

## Collections and iterators

```go
values := []result.Result[int, error]{
	result.Ok[int, error](1),
	result.Ok[int, error](2),
}
collected := result.Collect(values) // Result[[]int, error]

doubled := option.Some(10).
	Iter().
	Map(func(value int) int { return value * 2 }).
	Filter(func(value int) bool { return value > 10 }).
	Collect()
```

Option provides `Collect`, `All`, and `Any`. Result provides `Collect`, `All`,
and `Partition`.

## JSON and SQL

Option uses the natural nullable JSON representation:

```text
Some("hello") -> "hello"
None           -> null
```

Result uses an object with exactly one branch:

```json
{"ok": 42}
{"err": "message"}
```

`Option[T]` implements `sql.Scanner` and `driver.Valuer`; SQL `NULL` maps to
`None`. Database-native scalar types and types implementing `sql.Scanner`,
`driver.Valuer`, or `encoding.TextUnmarshaler` are supported.

## Go 1.27

The minimum version is Go 1.27 because `Map`, `MapErr`, `AndThen`, and `OrElse`
declare method type parameters. This enables a single chain to move from
`Option[int]` to `Option[string]`, or from `Result[T, E]` to `Result[U, E]`.

If multiple Go installations are on `PATH`, ensure `gofmt` also comes from Go
1.27; older formatters reject generic method syntax even when `go test` selects
the correct toolchain automatically.

## Benchmarks

Run the complete validation suite with:

```shell
go test ./...
go test ./... -race
go test -run '^$' -bench Benchmark -benchmem ./option ./result
```

On windows/amd64 with an Intel Core i7-14650HX and Go 1.27.1, the included
microbenchmarks report zero allocations for `Option.Map`, `Result.Map`, map
lookup conversion, and `result.From`. Benchmark timings vary by machine; run
them locally before drawing performance conclusions.

## API reference

| Area | API |
| --- | --- |
| Option constructors | `Some`, `None`, `FromPtr`, `FromZero`, `FromValueOk`, `FromValueError` |
| Option state | `IsSome`, `IsNone`, `Get`, `Unwrap`, `Expect`, `UnwrapOr`, `UnwrapOrElse` |
| Option composition | `Map`, `AndThen`, `OrElse`, `Filter`, `Inspect`, `Flatten`, `OkOr`, `OkOrElse`, `Iter` |
| Option collections | `Collect`, `All`, `Any`, `Match` |
| Result constructors | `Ok`, `Err`, `From` |
| Result state | `IsOk`, `IsErr`, `Get`, `Unwrap`, `UnwrapErr`, `Expect`, `ExpectErr`, `UnwrapOr`, `UnwrapOrElse` |
| Result composition | `Map`, `MapErr`, `AndThen`, `OrElse`, `Inspect`, `InspectErr`, `Ok`, `Err`, `Iter` |
| Result errors | `Error`, `IntoError`, `Wrap` (with standard `errors.Is` and `errors.As`) |
| Result collections | `Collect`, `All`, `Partition`, `Match` |
| Iterator | `Empty`, `Once`, `FromSlice`, `Map`, `Filter`, `Collect`, `Len` |

See [`examples/main.go`](examples/main.go) for a runnable end-to-end example.

## Support and contribute

If gomad makes your Go code clearer, consider
[starring the repository](https://github.com/23jdd/gomad). Stars help other Go
developers discover the project.

- Run the [end-to-end example](examples/main.go) or browse the executable
  [Option](option/example_test.go), [Result](result/example_test.go), and
  [Iterator](iterator/example_test.go) examples.
- [Open an issue](https://github.com/23jdd/gomad/issues) for bugs, API ideas, or
  real-world integration gaps.
- Pull requests with focused tests and examples are welcome.
- Share gomad with teams exploring explicit optional values or composable error
  handling in Go.
