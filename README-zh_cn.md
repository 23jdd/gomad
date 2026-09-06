# gomad

[English](README.md)

`gomad` 是一个面向 Go 1.27+ 的轻量级 Option 与 Result 库，围绕泛型方法和
零开销值类型构建。

它使用显式类型表达“值可能不存在”和“计算可能失败”，同时保持 Go 原有的
使用习惯：类型是普通结构体、零值合法，并且支持 `(value, ok)`、
`(value, error)`、JSON、`database/sql` 以及标准 `errors` 包。

## 为什么使用 gomad

指针经常同时承担“值不存在”、共享可变状态和对象标识等不同职责；`(T, error)`
则可能在保存或传递过程中被意外拆散。`Option[T]` 与 `Result[T, E]` 把状态和
数据放在同一个值中，让每个分支都清晰可见，并能通过 Go 1.27 泛型方法连续组合。

## 安装

```shell
go get github.com/23jdd/gomad@latest
```

项目需要使用 Go 1.27 或更高版本：

```go
module example.com/myapp

go 1.27
```

## Option

使用 `Some` 表示值存在，使用 `None` 表示值不存在：

```go
value := option.Some(10)

if value.IsSome() {
	fmt.Println(value.Unwrap())
}

fallback := option.None[int]().UnwrapOr(42)
```

`Option[T]` 的零值等价于 `None`。常见 Go 值可以直接转换：

```go
fromPointer := option.FromPtr(ptr)
fromLookup := option.FromValueOk(value, ok)
fromCall := option.FromValueError(value, err)
```

Go 不会把 map 索引作为二值函数参数传递，因此 map 查询需要先接收
comma-ok 结果：

```go
value, ok := values[key]
found := option.FromValueOk(value, ok)
```

## Result

使用 `Ok` 表示成功，使用 `Err` 表示失败。错误类型 `E` 不必局限于 `error`：

```go
func divide(a, b int) result.Result[int, error] {
	if b == 0 {
		return result.Err[int](errors.New("division by zero"))
	}
	return result.Ok[int, error](a / b)
}
```

`From` 可以直接接收普通 Go 函数返回的 `(T, error)`：

```go
contents := result.From(os.ReadFile("config.json"))
number := result.From(strconv.Atoi("42"))
```

`Result[T, E]` 的零值是包含 `E` 零值的 `Err`。

## Map

Go 1.27 的泛型方法允许链式调用在保持静态类型安全的同时改变值类型：

```go
length := option.Some(10).
	Map(strconv.Itoa).
	Map(func(value string) int { return len(value) })

text := result.From(strconv.Atoi("21")).
	Map(func(value int) int { return value * 2 }).
	Map(strconv.Itoa)
```

库同时提供包级 `option.Map`、`result.Map` 和 `result.MapErr`，方便函数式组合或
需要把转换器作为普通函数使用的场景。

## AndThen

当回调本身已经返回 `Option` 或 `Result` 时，使用 `AndThen` 避免产生嵌套：

```go
name := option.Some(1).
	AndThen(findUser).
	Map(func(user User) string { return user.Name }).
	Filter(func(name string) bool { return name != "" })

config := result.From(os.ReadFile("config.json")).
	Map(parseConfig).
	AndThen(validateConfig)
```

`OrElse` 用于延迟恢复失败分支。`Inspect` 与 `InspectErr` 可以观察链中的值而不
改变结果。

## 错误处理

Option 与 Result 可以直接互相转换：

```go
required := optionalValue.OkOr(errors.New("value is required"))
optional := required.Ok()
failure := required.Err()
```

`Result[T, error]` 可以通过 `IntoError` 接入标准错误链：

```go
wrapped := result.From(load()).Wrap("load configuration")

if errors.Is(wrapped.IntoError(), fs.ErrNotExist) {
	// 处理文件不存在
}

err := wrapped.IntoError() // Ok 返回 nil，Err 返回 error
```

`errors.As` 的用法相同。包级 `result.Error` 等价于 `IntoError`，`Wrap` 使用 `%w`
语义添加上下文。Result 本身不会实现 `error`，因为值解包所需的 `Unwrap() T`
与标准错误链约定的 `Unwrap() error` 无法同时存在。

## Option 与指针

当对象标识、共享修改或大对象复制成本具有实际意义时，应使用指针。当重点只是
“值是否存在”时，使用 `Option[T]` 更清晰。Option 不会把“存在的 nil 指针”和
“不存在”混为一谈，并且拥有合法零值；简单值路径不产生堆分配。

## Result 与 `(T, error)`

在传统 Go API 边界上继续使用 `(T, error)`。当多个转换需要连续组合、需要保存
完整计算状态，或者错误值不是 `error` 时，可以转为 Result。`result.From` 与
`IntoError` 让两种风格能够自然互操作。

## 集合与迭代器

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

Option 提供 `Collect`、`All` 和 `Any`；Result 提供 `Collect`、`All` 和
`Partition`。

## JSON 与 SQL

Option 使用自然的可空 JSON 表示：

```text
Some("hello") -> "hello"
None           -> null
```

Result 使用只包含一个分支的对象：

```json
{"ok": 42}
{"err": "message"}
```

`Option[T]` 实现了 `sql.Scanner` 和 `driver.Valuer`，SQL `NULL` 会映射为
`None`。它支持数据库原生标量类型，以及实现 `sql.Scanner`、`driver.Valuer`
或 `encoding.TextUnmarshaler` 的类型。

## Go 1.27

最低版本为 Go 1.27，因为 `Map`、`MapErr`、`AndThen` 和 `OrElse` 的方法声明了
新的类型参数。这使同一条链可以从 `Option[int]` 变为 `Option[string]`，或者从
`Result[T, E]` 变为 `Result[U, E]`。

如果 `PATH` 中存在多个 Go 版本，请同时确认 `gofmt` 来自 Go 1.27。旧版
`gofmt` 会拒绝泛型方法语法，即使 `go test` 已经自动选择正确的新工具链。

## 测试与基准

运行完整验证：

```shell
go test ./...
go test ./... -race
go vet ./...
go test -run '^$' -bench Benchmark -benchmem ./option ./result
```

在 windows/amd64、Intel Core i7-14650HX、Go 1.27.1 环境中，仓库内的微基准
显示 `Option.Map`、`Result.Map`、map 查询转换和 `result.From` 均为
`0 B/op`、`0 allocs/op`。时间结果会因机器而异，应以本地运行结果为准。

## 可执行示例

所有公开构造器、包级函数和方法都有带输出断言的 Example 测试：

- [Option 示例](option/example_test.go)
- [Result 示例](result/example_test.go)
- [Iterator 示例](iterator/example_test.go)
- [完整程序](examples/main.go)

这些示例会由 `go test ./...` 编译并执行，既是文档也是回归测试。

## API 参考

| 分类 | API |
| --- | --- |
| Option 构造 | `Some`、`None`、`FromPtr`、`FromZero`、`FromValueOk`、`FromValueError` |
| Option 状态 | `IsSome`、`IsNone`、`Get`、`Unwrap`、`Expect`、`UnwrapOr`、`UnwrapOrElse` |
| Option 组合 | `Map`、`AndThen`、`OrElse`、`Filter`、`Inspect`、`Flatten`、`OkOr`、`OkOrElse`、`Iter` |
| Option 集合 | `Collect`、`All`、`Any`、`Match` |
| Result 构造 | `Ok`、`Err`、`From` |
| Result 状态 | `IsOk`、`IsErr`、`Get`、`Unwrap`、`UnwrapErr`、`Expect`、`ExpectErr`、`UnwrapOr`、`UnwrapOrElse` |
| Result 组合 | `Map`、`MapErr`、`AndThen`、`OrElse`、`Inspect`、`InspectErr`、`Ok`、`Err`、`Iter` |
| Result 错误 | `Error`、`IntoError`、`Wrap`，以及标准 `errors.Is`、`errors.As` |
| Result 集合 | `Collect`、`All`、`Partition`、`Match` |
| Iterator | `Empty`、`Once`、`FromSlice`、`Map`、`Filter`、`Collect`、`Len` |
