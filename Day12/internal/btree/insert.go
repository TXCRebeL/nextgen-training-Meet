package btree

import "cmp"

// Insert inserts a new key into the B-Tree or updates it if it exists
func (bt *BTree[K, V]) Insert(k K, v V) {
	if bt.root == nil {
		bt.root = &Node[K, V]{
			keys:     []K{k},
			values:   []V{v},
			children: make([]*Node[K, V], 0),
			leaf:     true,
		}
		return
	}

	// If the key already exists, simply update its value
	if updateExisting(bt.root, k, v) {
		return
	}

	r := bt.root
	// If root is full, then tree grows in height
	if len(r.keys) == 2*bt.t-1 {
		s := &Node[K, V]{
			keys:     make([]K, 0),
			values:   make([]V, 0),
			children: []*Node[K, V]{r},
			leaf:     false,
		}
		bt.root = s
		bt.splitChild(s, 0, r)
		bt.insertNonFull(s, k, v)
	} else {
		bt.insertNonFull(r, k, v)
	}
}

func updateExisting[K cmp.Ordered, V any](node *Node[K, V], k K, v V) bool {
	i := 0
	for i < len(node.keys) && k > node.keys[i] {
		i++
	}
	if i < len(node.keys) && k == node.keys[i] {
		node.values[i] = v
		return true
	}
	if node.leaf {
		return false
	}
	return updateExisting(node.children[i], k, v)
}

// splitChild splits a full child y of node x at index i
func (bt *BTree[K, V]) splitChild(x *Node[K, V], i int, y *Node[K, V]) {
	t := bt.t
	z := &Node[K, V]{
		keys:     make([]K, t-1),
		values:   make([]V, t-1),
		children: make([]*Node[K, V], 0),
		leaf:     y.leaf,
	}

	// Copy the last t-1 keys and values of y to z
	copy(z.keys, y.keys[t:])
	copy(z.values, y.values[t:])

	// Copy the last t children of y to z if y is not a leaf
	if !y.leaf {
		z.children = make([]*Node[K, V], t)
		copy(z.children, y.children[t:])
		y.children = y.children[:t] // Shrink y's children
	}

	// The median key of y moves up to x
	medianKey := y.keys[t-1]
	medianVal := y.values[t-1]
	
	y.keys = y.keys[:t-1] // Shrink y's keys
	y.values = y.values[:t-1] // Shrink y's values

	// Insert z into x's children
	x.children = append(x.children, nil)     // Expand by one
	copy(x.children[i+2:], x.children[i+1:]) // Shift right
	x.children[i+1] = z

	// Insert medianKey into x's keys/values
	var zeroK K
	x.keys = append(x.keys, zeroK)         // Expand by one
	copy(x.keys[i+1:], x.keys[i:])         // Shift right
	x.keys[i] = medianKey

	var zeroV V
	x.values = append(x.values, zeroV)
	copy(x.values[i+1:], x.values[i:])
	x.values[i] = medianVal
}

func (bt *BTree[K, V]) insertNonFull(x *Node[K, V], k K, v V) {
	i := len(x.keys) - 1

	if x.leaf {
		var zeroK K
		var zeroV V
		x.keys = append(x.keys, zeroK)
		x.values = append(x.values, zeroV)

		for i >= 0 && k < x.keys[i] {
			x.keys[i+1] = x.keys[i]
			x.values[i+1] = x.values[i]
			i--
		}
		x.keys[i+1] = k
		x.values[i+1] = v
	} else {
		for i >= 0 && k < x.keys[i] {
			i--
		}
		i++

		if len(x.children[i].keys) == 2*bt.t-1 {
			bt.splitChild(x, i, x.children[i])
			if k > x.keys[i] {
				i++
			}
		}
		bt.insertNonFull(x.children[i], k, v)
	}
}
