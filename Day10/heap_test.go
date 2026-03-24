package main

import (
	"testing"
)

func TestMinHeap_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			"Ascending",
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"Descending",
			[]int{5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"Random",
			[]int{3, 1, 4, 1, 5, 9, 2, 6, 5},
			[]int{1, 1, 2, 3, 4, 5, 5, 6, 9},
		},
		{
			"Duplicates",
			[]int{2, 2, 2, 1, 1, 1},
			[]int{1, 1, 1, 2, 2, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			less := func(a, b int) bool { return a < b }
			h := NewMinHeap(less)
			for _, v := range tt.input {
				h.Insert(v)
				if err := h.Verify(); err != nil {
					t.Fatalf("Verify failed after insert %d: %v", v, err)
				}
			}

			if h.Size() != len(tt.input) {
				t.Errorf("Expected size %d, got %d", len(tt.input), h.Size())
			}

			for _, exp := range tt.expected {
				val, err := h.ExtractMin()
				if err != nil {
					t.Fatalf("ExtractMin failed: %v", err)
				}
				if val != exp {
					t.Errorf("Expected %d, got %d", exp, val)
				}
				if err := h.Verify(); err != nil {
					t.Fatalf("Verify failed after extract: %v", err)
				}
			}
		})
	}
}

type testItem struct {
	id   int
	prio int
}

func TestMinHeap_UpdateAndRemove(t *testing.T) {
	less := func(a, b *testItem) bool { return a.prio < b.prio }
	h := NewMinHeap(less)

	i1 := &testItem{1, 10}
	i2 := &testItem{2, 20}
	i3 := &testItem{3, 30}
	i4 := &testItem{4, 5}

	h.Insert(i1)
	h.Insert(i2)
	h.Insert(i3)
	h.Insert(i4)

	// Update i3's priority to 1
	i3.prio = 1
	err := h.Update(i3)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	if err := h.Verify(); err != nil {
		t.Errorf("Verify failed after update: %v", err)
	}
	
	min, _ := h.Peek()
	if min.id != 3 {
		t.Errorf("Expected min ID 3, got %d", min.id)
	}

	// Remove i1 (priority 10)
	_, err = h.Remove(i1)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if err := h.Verify(); err != nil {
		t.Errorf("Verify failed after remove: %v", err)
	}
	
	if h.Size() != 3 {
		t.Errorf("Expected size 3, got %d", h.Size())
	}
}
