package stack

import "errors"

type SliceStack[T any] struct {
	items []T
}

func NewSliceStack[T any]() *SliceStack[T] {
	return &SliceStack[T]{
		items: make([]T, 0),
	}
}

func (s *SliceStack[T]) Push(x T) error {
	s.items = append(s.items, x)
	return nil
}

func (s *SliceStack[T]) Pop() (T, error) {
	if s.IsEmpty() {
		var zero T
		return zero, errors.New("stack is empty")
	}
	value := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return value, nil
}

func (s *SliceStack[T]) Peek() (T, error) {
	if s.IsEmpty() {
		var zero T
		return zero, errors.New("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

func (s *SliceStack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *SliceStack[T]) Size() int {
	return len(s.items)
}
