package iterator

import "testing"

func TestIterator(t *testing.T) {
	got := FromSlice([]int{1, 2, 3}).
		Map(func(value int) string { return string(rune('0' + value)) }).
		Filter(func(value string) bool { return value != "2" }).
		Collect()
	if len(got) != 2 || got[0] != "1" || got[1] != "3" {
		t.Fatalf("Collect = %v", got)
	}
}
