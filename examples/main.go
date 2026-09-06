package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/23jdd/gomad/option"
	"github.com/23jdd/gomad/result"
)

type user struct {
	name string
}

func findUser(id int) option.Option[user] {
	if id != 1 {
		return option.None[user]()
	}
	return option.Some(user{name: "Gopher"})
}

func parsePositive(input string) result.Result[int, error] {
	return result.From(strconv.Atoi(input)).
		AndThen(func(value int) result.Result[int, error] {
			if value <= 0 {
				return result.Err[int](errors.New("value must be positive"))
			}
			return result.Ok[int, error](value)
		})
}

func main() {
	name := option.Some(1).
		AndThen(findUser).
		Map(func(value user) string { return value.name }).
		Filter(func(value string) bool { return value != "" })

	option.Match(
		name,
		func(value string) { fmt.Println("user:", value) },
		func() { fmt.Println("user not found") },
	)

	text := parsePositive("21").
		Map(func(value int) int { return value * 2 }).
		Map(strconv.Itoa).
		Inspect(func(value string) { fmt.Println("answer:", value) }).
		InspectErr(func(err error) { fmt.Println("error:", err) })

	result.Match(
		text,
		func(value string) { fmt.Println("result:", value) },
		func(err error) { fmt.Println("failed:", err) },
	)
}
