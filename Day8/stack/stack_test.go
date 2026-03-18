package stack

import (
	"fmt"
	"testing"
)

func BenchmarkSliceStackPush(b *testing.B) {
	count := []int{100, 1000, 10000, 100000}
	for _, c := range count {
		b.Run(fmt.Sprintf("count=%d", c), func(b *testing.B) {
			stack := NewSliceStack[int]()
			for i := 0; i < b.N; i++ {
				stack.Push(i)
			}
		})
	}
}

func BenchmarkLinkedListStackPush(b *testing.B) {
	count := []int{100, 1000, 10000, 100000}
	for _, c := range count {
		b.Run(fmt.Sprintf("count=%d", c), func(b *testing.B) {
			stack := NewLinkedListStack[int]()
			for i := 0; i < b.N; i++ {
				stack.Push(i)
			}
		})
	}
}

func BenchmarkSliceStackPop(b *testing.B) {
	count := []int{100, 1000, 10000, 100000}
	for _, c := range count {
		b.Run(fmt.Sprintf("count=%d", c), func(b *testing.B) {
			stack := NewSliceStack[int]()
			for i := 0; i < b.N; i++ {
				stack.Push(i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stack.Pop()
			}
		})
	}
}
func BenchmarkLinkedListStackPop(b *testing.B) {
	count := []int{100, 1000, 10000, 100000}
	for _, c := range count {
		b.Run(fmt.Sprintf("count=%d", c), func(b *testing.B) {
			stack := NewLinkedListStack[int]()
			for i := 0; i < b.N; i++ {
				stack.Push(i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stack.Pop()
			}
		})
	}
}
