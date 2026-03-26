package btree

import (
	"cmp"
	"fmt"
	"strings"
)

type Item[K cmp.Ordered, V any] struct {
	Key   K
	Value V
}

// Node represents a generic node in the B-Tree
type Node[K cmp.Ordered, V any] struct {
	keys     []K
	values   []V
	children []*Node[K, V]
	leaf     bool
}

// BTree represents a generic B-Tree
type BTree[K cmp.Ordered, V any] struct {
	root *Node[K, V]
	t    int // Minimum degree
}

// NewBTree creates a new B-Tree with minimum degree t
func NewBTree[K cmp.Ordered, V any](t int) *BTree[K, V] {
	if t < 2 {
		panic("Minimum degree t must be >= 2")
	}
	return &BTree[K, V]{
		root: nil,
		t:    t,
	}
}

// Search checks if a key exists in the B-Tree
func (bt *BTree[K, V]) Search(k K) (V, bool) {
	if bt.root == nil {
		var zero V
		return zero, false
	}
	return searchNode(bt.root, k)
}

func searchNode[K cmp.Ordered, V any](node *Node[K, V], k K) (V, bool) {
	i := 0
	for i < len(node.keys) && k > node.keys[i] {
		i++
	}

	if i < len(node.keys) && k == node.keys[i] {
		return node.values[i], true
	}

	if node.leaf {
		var zero V
		return zero, false
	}

	return searchNode(node.children[i], k)
}

// InOrder returns a sorted list of all key-value items in the tree
func (bt *BTree[K, V]) InOrder() []Item[K, V] {
	if bt.root == nil {
		return nil
	}
	var res []Item[K, V]
	inOrderNode(bt.root, &res)
	return res
}

func inOrderNode[K cmp.Ordered, V any](node *Node[K, V], res *[]Item[K, V]) {
	var i int
	for i = 0; i < len(node.keys); i++ {
		if !node.leaf {
			inOrderNode(node.children[i], res)
		}
		*res = append(*res, Item[K, V]{Key: node.keys[i], Value: node.values[i]})
	}
	if !node.leaf {
		inOrderNode(node.children[i], res)
	}
}

// RangeQuery returns all items in the tree with keys between min and max (inclusive)
func (bt *BTree[K, V]) RangeQuery(min, max K) []Item[K, V] {
	if bt.root == nil {
		return nil
	}
	var res []Item[K, V]
	rangeQueryNode(bt.root, min, max, &res)
	return res
}

func rangeQueryNode[K cmp.Ordered, V any](node *Node[K, V], min, max K, res *[]Item[K, V]) {
	var i int
	for i = 0; i < len(node.keys); i++ {
		if !node.leaf && min <= node.keys[i] {
			rangeQueryNode(node.children[i], min, max, res)
		}

		if node.keys[i] >= min && node.keys[i] <= max {
			*res = append(*res, Item[K, V]{Key: node.keys[i], Value: node.values[i]})
		}

		if node.keys[i] > max {
			return
		}
	}

	if !node.leaf {
		rangeQueryNode(node.children[i], min, max, res)
	}
}

// WalkRange is an allocation-free DFS that calls the provided function for every matching item within bounds
func (bt *BTree[K, V]) WalkRange(min, max K, fn func(K, V)) {
	if bt.root == nil {
		return
	}
	walkRangeNode(bt.root, min, max, fn)
}

func walkRangeNode[K cmp.Ordered, V any](node *Node[K, V], min, max K, fn func(K, V)) {
	var i int
	for i = 0; i < len(node.keys); i++ {
		if !node.leaf && min <= node.keys[i] {
			walkRangeNode(node.children[i], min, max, fn)
		}

		if node.keys[i] >= min && node.keys[i] <= max {
			fn(node.keys[i], node.values[i])
		}

		if node.keys[i] > max {
			return
		}
	}

	if !node.leaf {
		walkRangeNode(node.children[i], min, max, fn)
	}
}

// Visualize prints a textual representation of the tree level by level
func (bt *BTree[K, V]) Visualize() {
	if bt.root == nil {
		fmt.Println("Empty B-Tree")
		return
	}
	visualizeNode(bt.root, 0)
}

func visualizeNode[K cmp.Ordered, V any](node *Node[K, V], level int) {
	prefix := strings.Repeat("  ", level)
	fmt.Printf("%sLevel %d: %v\n", prefix, level, node.keys)
	if !node.leaf {
		for _, child := range node.children {
			visualizeNode(child, level+1)
		}
	}
}
