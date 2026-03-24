package main

import (
	"errors"
	"fmt"
	"sync"
)

type MinHeap[T comparable] struct {
	data []T
	less func(a, b T) bool
	mu   sync.RWMutex
}

func NewMinHeap[T comparable](less func(a, b T) bool) *MinHeap[T] {
	return &MinHeap[T]{
		data: make([]T, 0),
		less: less,
	}
}

func (h *MinHeap[T]) Insert(val T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data = append(h.data, val)
	h.bubbleUp(len(h.data) - 1)
}

func (h *MinHeap[T]) ExtractMin() (T, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	var zero T
	if len(h.data) == 0 {
		return zero, errors.New("heap is empty")
	}

	min := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]

	if len(h.data) > 0 {
		h.bubbleDown(0)
	}

	return min, nil
}

func (h *MinHeap[T]) Peek() (T, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var zero T
	if len(h.data) == 0 {
		return zero, errors.New("heap is empty")
	}
	return h.data[0], nil
}

func (h *MinHeap[T]) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.data)
}

func (h *MinHeap[T]) Update(val T) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	idx := -1
	for i, v := range h.data {
		if v == val {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("value not found in heap")
	}

	// Re-heapify at the modified value's position
	// Since we don't know if priority increased or decreased, check both
	h.bubbleUp(idx)
	h.bubbleDown(idx)
	return nil
}

func (h *MinHeap[T]) Remove(val T) (T, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	var zero T
	
	idx := -1
	for i, v := range h.data {
		if v == val {
			idx = i
			break
		}
	}
	if idx == -1 {
		return zero, errors.New("value not found in heap")
	}

	res := h.data[idx]
	last := len(h.data) - 1
	h.data[idx] = h.data[last]
	h.data = h.data[:last]

	if idx < len(h.data) {
		h.bubbleDown(idx)
		h.bubbleUp(idx)
	}

	return res, nil
}

func (h *MinHeap[T]) bubbleUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.less(h.data[i], h.data[parent]) {
			h.data[i], h.data[parent] = h.data[parent], h.data[i]
			i = parent
		} else {
			break
		}
	}
}

func (h *MinHeap[T]) bubbleDown(i int) {
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i

		if left < len(h.data) && h.less(h.data[left], h.data[smallest]) {
			smallest = left
		}
		if right < len(h.data) && h.less(h.data[right], h.data[smallest]) {
			smallest = right
		}

		if smallest != i {
			h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
			i = smallest
		} else {
			break
		}
	}
}

// Verify checks the heap property (useful for tests)
func (h *MinHeap[T]) Verify() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for i := 0; i < len(h.data); i++ {
		left := 2*i + 1
		right := 2*i + 2
		if left < len(h.data) && h.less(h.data[left], h.data[i]) {
			return fmt.Errorf("heap property violated at index %d (parent) and %d (left child)", i, left)
		}
		if right < len(h.data) && h.less(h.data[right], h.data[i]) {
			return fmt.Errorf("heap property violated at index %d (parent) and %d (right child)", i, right)
		}
	}
	return nil
}
