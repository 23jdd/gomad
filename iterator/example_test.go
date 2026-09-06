package iterator_test

import (
	"fmt"
	"strconv"

	"github.com/23jdd/gomad/iterator"
)

func ExampleEmpty() {
	fmt.Println(iterator.Empty[int]().Len())
	// Output: 0
}

func ExampleOnce() {
	fmt.Println(iterator.Once(10).Collect())
	// Output: [10]
}

func ExampleFromSlice() {
	values := []int{1, 2}
	iter := iterator.FromSlice(values)
	values[0] = 9
	fmt.Println(iter.Collect())
	// Output: [1 2]
}

func ExampleMap() {
	iter := iterator.Map(iterator.FromSlice([]int{1, 2}), strconv.Itoa)
	fmt.Println(iter.Collect())
	// Output: [1 2]
}

func ExampleIterator_Map() {
	iter := iterator.FromSlice([]int{1, 2}).Map(strconv.Itoa)
	fmt.Println(iter.Collect())
	// Output: [1 2]
}

func ExampleIterator_Filter() {
	iter := iterator.FromSlice([]int{1, 2, 3}).Filter(func(value int) bool {
		return value%2 == 1
	})
	fmt.Println(iter.Collect())
	// Output: [1 3]
}

func ExampleIterator_Collect() {
	fmt.Println(iterator.Once(10).Collect())
	// Output: [10]
}

func ExampleIterator_Len() {
	fmt.Println(iterator.FromSlice([]int{1, 2}).Len())
	// Output: 2
}
