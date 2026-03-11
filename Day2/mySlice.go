package main

import "fmt"

const (
	CAP_THRESHOLD = 10
)

type mySlice[T any] struct {
	header *[]T
	len    int
	cap    int
}

func newMySlice[T any](capacity int) *mySlice[T] {
	slice := make([]T, capacity)
	return &mySlice[T]{
		header: &slice,
		len:    0,
		cap:    capacity,
	}
}

func (s *mySlice[T]) append(value T) {
	if s.len < s.cap {
		(*s.header)[s.len] = value
		s.len++
	} else {
		fmt.Printf("the cap changed from %d", s.cap)
		if s.cap < CAP_THRESHOLD {
			s.cap *= 2
		} else {
			s.cap = int(float64(s.cap) * 1.25)
		}
		fmt.Printf("to %d\n", s.cap)
		newSlice := make([]T, s.cap)
		copy(newSlice, *s.header)
		s.header = &newSlice
		s.append(value)
	}
}

func (s *mySlice[T]) get(index int) (T, bool) {
	if index < 0 || index >= s.len {
		var zeroValue T
		return zeroValue, false
	}
	return (*s.header)[index], true
}

func (s *mySlice[T]) length() int {
	return s.len
}

func (s *mySlice[T]) capacity() int {
	return s.cap
}
