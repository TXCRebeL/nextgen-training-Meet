package queue

import (
	"errors"
	"sync"
)

type CircularQueue[T any] struct {
	items    []T
	front    int
	rear     int
	size     int
	capacity int
	mu       sync.Mutex
}

func NewCircularQueue[T any](capacity int) *CircularQueue[T] {
	return &CircularQueue[T]{
		items:    make([]T, capacity),
		front:    0,
		rear:     0,
		size:     0,
		capacity: capacity,
	}
}

func (q *CircularQueue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == q.capacity {
		return errors.New("queue is full")
	}
	q.items[q.rear] = item
	q.rear = (q.rear + 1) % q.capacity
	q.size++
	return nil
}

func (q *CircularQueue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		var zero T
		return zero, errors.New("queue is empty")
	}
	item := q.items[q.front]
	q.front = (q.front + 1) % q.capacity
	q.size--
	return item, nil
}

func (q *CircularQueue[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size == 0
}

func (q *CircularQueue[T]) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size == q.capacity
}

func (q *CircularQueue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

func (q *CircularQueue[T]) Capacity() int {
	return q.capacity
}

func (q *CircularQueue[T]) Peek() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.size == 0 {
		var zero T
		return zero, errors.New("queue is empty")
	}
	return q.items[q.front], nil
}
