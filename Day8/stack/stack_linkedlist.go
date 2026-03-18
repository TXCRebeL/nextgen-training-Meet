package stack

import "errors"

type Node[T any] struct {
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}

type LinkedListStack[T any] struct {
	Head *Node[T]
	size int
}

func NewLinkedListStack[T any]() *LinkedListStack[T] {
	return &LinkedListStack[T]{
		Head: nil,
		size: 0,
	}
}

func (s *LinkedListStack[T]) Push(x T) error {
	newNode := &Node[T]{
		Value: x,
		Next:  s.Head,
		Prev:  nil,
	}
	if s.Head != nil {
		s.Head.Prev = newNode
	}
	s.Head = newNode
	s.size++
	return nil
}

func (s *LinkedListStack[T]) Pop() (T, error) {
	if s.Head == nil {
		var zero T
		return zero, errors.New("stack is empty")
	}
	value := s.Head.Value
	s.Head = s.Head.Next
	if s.Head != nil {
		s.Head.Prev = nil
	}
	s.size--
	return value, nil
}

func (s *LinkedListStack[T]) Peek() (T, error) {
	if s.Head == nil {
		var zero T
		return zero, errors.New("stack is empty")
	}
	return s.Head.Value, nil
}

func (s *LinkedListStack[T]) IsEmpty() bool {
	return s.size == 0
}

func (s *LinkedListStack[T]) Size() int {
	return s.size
}
