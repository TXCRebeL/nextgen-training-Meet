package main

import (
	"fmt"
)

const (
	THRESHOLDLOADFACTOR = 0.75
	INITIAL_BUCKETS     = 8
	BUCKET_CAPACITY     = 8 // Fixed size of slots within a bucket
)

type Slot[T comparable] struct {
	Key   T
	Value interface{}
}

// Bucket uses a fixed array of size 8 for Slots
type Bucket[T comparable] struct {
	Slots      [BUCKET_CAPACITY]Slot[T]
	Count      int
	isOverflow bool
	Next       *Bucket[T]
}

type MyMap[T comparable] struct {
	Buckets []Bucket[T]
	Size    int
}

// NewMyMap creates a new map with a constant initial number of buckets (8).
// User input is no longer required for initial sizing.
func NewMyMap[T comparable]() *MyMap[T] {
	return &MyMap[T]{
		Buckets: make([]Bucket[T], INITIAL_BUCKETS),
		Size:    0,
	}
}

// hash implements the FNV-1a hash algorithm.
// FNV-1a is chosen because it's fast, simple, and provides great dispersion,
// thereby minimizing collisions.
func (m *MyMap[T]) hash(key T) uint32 {
	str := fmt.Sprintf("%v", key)
	var hash uint32 = 2166136261
	for i := 0; i < len(str); i++ {
		hash ^= uint32(str[i])
		hash *= 16777619
	}
	return hash
}

// Insert adds a key-value or updates an existing one.
// Checks threshold and resizes, handles overflow within buckets.
func (m *MyMap[T]) Insert(key T, value interface{}) {
	// 1. Check Load Factor and resize if needed (Size / len(Buckets) >= THRESHOLD)
	if float64(m.Size)/float64(len(m.Buckets)) >= THRESHOLDLOADFACTOR {
		m.resize()
	}

	h := m.hash(key)
	idx := h % uint32(len(m.Buckets))

	// 2. Iterate through bucket and any linked overflow buckets to see if Key exists
	curr := &m.Buckets[idx]
	for curr != nil {
		for i := 0; i < curr.Count; i++ {
			if curr.Slots[i].Key == key {
				curr.Slots[i].Value = value
				return
			}
		}
		curr = curr.Next
	}

	// 3. Key does not exist, insert into the first available slot
	curr = &m.Buckets[idx]
	for {
		if curr.Count < BUCKET_CAPACITY {
			// Found space in the current bucket array
			curr.Slots[curr.Count] = Slot[T]{Key: key, Value: value}
			curr.Count++
			m.Size++
			return
		}
		// Current bucket array is full, go to next overflow bucket or create it
		if curr.Next == nil {
			curr.Next = &Bucket[T]{
				isOverflow: true,
			}
		}
		curr = curr.Next
	}
}

// resize doubles the bucket slice capacity and redistributes elements
func (m *MyMap[T]) resize() {
	oldBuckets := m.Buckets
	m.Buckets = make([]Bucket[T], len(oldBuckets)*2)
	m.Size = 0 // Reset size; Insert will recount it

	for i := range oldBuckets {
		curr := &oldBuckets[i]
		for curr != nil {
			for j := 0; j < curr.Count; j++ {
				// Re-insert redistributes the element over the newly doubled capacity
				m.Insert(curr.Slots[j].Key, curr.Slots[j].Value)
			}
			curr = curr.Next
		}
	}
}

func (m *MyMap[T]) Get(key T) (interface{}, bool) {
	if len(m.Buckets) == 0 {
		return nil, false
	}
	h := m.hash(key)
	idx := h % uint32(len(m.Buckets))

	curr := &m.Buckets[idx]
	for curr != nil {
		for i := 0; i < curr.Count; i++ {
			if curr.Slots[i].Key == key {
				return curr.Slots[i].Value, true
			}
		}
		curr = curr.Next
	}
	return nil, false
}

func (m *MyMap[T]) Delete(key T) bool {
	if len(m.Buckets) == 0 {
		return false
	}
	h := m.hash(key)
	idx := h % uint32(len(m.Buckets))

	curr := &m.Buckets[idx]
	for curr != nil {
		for i := 0; i < curr.Count; i++ {
			if curr.Slots[i].Key == key {
				// Swap with the last element in THIS bucket to delete in O(1)
				lastIdx := curr.Count - 1
				curr.Slots[i] = curr.Slots[lastIdx]
				var zero T
				curr.Slots[lastIdx] = Slot[T]{Key: zero, Value: nil}
				curr.Count--
				m.Size--
				return true
			}
		}
		curr = curr.Next
	}
	return false
}

// PrintMap is a helper function to visualize map structure
func (m *MyMap[T]) PrintMap() {
	fmt.Printf("Map Size: %d, Number of Buckets: %d (Load Factor Limit: %.2f)\n",
		m.Size, len(m.Buckets), THRESHOLDLOADFACTOR)

	for i := range m.Buckets {
		b := &m.Buckets[i]
		if b.Count == 0 && b.Next == nil {
			continue // Hide completely empty buckets
		}
		fmt.Printf("Bucket %d:", i)
		curr := b
		for curr != nil {
			fmt.Printf(" [ ")
			for j := 0; j < curr.Count; j++ {
				fmt.Printf("{%v: %v} ", curr.Slots[j].Key, curr.Slots[j].Value)
			}
			fmt.Printf("] ")
			if curr.Next != nil {
				fmt.Printf("-> (Overflow) -> ")
			}
			curr = curr.Next
		}
		fmt.Println()
	}
	fmt.Println()
}
