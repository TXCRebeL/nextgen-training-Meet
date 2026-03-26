package btree

import (
	"reflect"
	"testing"
)

func TestInsertAndSearch(t *testing.T) {
	bt := NewBTree[int, string](3)
	keys := []int{10, 20, 5, 6, 12, 30, 7, 17}

	for _, k := range keys {
		bt.Insert(k, "val")
	}

	for _, k := range keys {
		if _, ok := bt.Search(k); !ok {
			t.Errorf("Expected to find key %d in tree", k)
		}
	}

	if _, ok := bt.Search(15); ok {
		t.Errorf("Expected not to find key 15")
	}
}

func TestInOrder(t *testing.T) {
	bt := NewBTree[int, string](3)
	keys := []int{10, 20, 5, 6, 12, 30, 7, 17}

	for _, k := range keys {
		bt.Insert(k, "val")
	}

	expected := []int{5, 6, 7, 10, 12, 17, 20, 30}
	items := bt.InOrder()
	var result []int
	for _, item := range items {
		result = append(result, item.Key)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("InOrder failed. Expected %v, got %v", expected, result)
	}
}

func TestRangeQuery(t *testing.T) {
	bt := NewBTree[int, string](3)
	keys := []int{10, 20, 5, 6, 12, 30, 7, 17, 25, 2, 8, 15}

	for _, k := range keys {
		bt.Insert(k, "val")
	}

	expected := []int{7, 8, 10, 12, 15}
	items := bt.RangeQuery(7, 15)
	var result []int
	for _, item := range items {
		result = append(result, item.Key)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RangeQuery failed. Expected %v, got %v", expected, result)
	}
}

func TestDelete(t *testing.T) {
	bt := NewBTree[int, string](3)
	keys := []int{1, 3, 7, 10, 11, 13, 14, 15, 18, 16, 19, 24, 25, 26, 21, 4, 5, 20, 22, 2, 17, 12, 6}

	for _, k := range keys {
		bt.Insert(k, "val")
	}

	if len(bt.InOrder()) != len(keys) {
		t.Fatalf("Insert failed, expected %d keys, got %d", len(keys), len(bt.InOrder()))
	}

	bt.Delete(6)
	if _, ok := bt.Search(6); ok {
		t.Errorf("Failed to delete 6")
	}

	bt.Delete(13)
	if _, ok := bt.Search(13); ok {
		t.Errorf("Failed to delete 13")
	}

	bt.Delete(14)
	if _, ok := bt.Search(14); ok {
		t.Errorf("Failed to delete 14")
	}

	for _, k := range keys {
		bt.Delete(k) // Safe if the key is already deleted
	}

	if len(bt.InOrder()) != 0 {
		t.Errorf("Tree should be empty after deleting all keys, got %v", bt.InOrder())
	}
}
