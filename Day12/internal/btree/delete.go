package btree

func (bt *BTree[K, V]) Delete(k K) {
	if bt.root == nil {
		return
	}

	bt.deleteNode(bt.root, k)

	if len(bt.root.keys) == 0 {
		if bt.root.leaf {
			bt.root = nil
		} else {
			bt.root = bt.root.children[0]
		}
	}
}

func (bt *BTree[K, V]) findKey(x *Node[K, V], k K) int {
	idx := 0
	for idx < len(x.keys) && x.keys[idx] < k {
		idx++
	}
	return idx
}

func (bt *BTree[K, V]) deleteNode(x *Node[K, V], k K) {
	idx := bt.findKey(x, k)

	if idx < len(x.keys) && x.keys[idx] == k {
		if x.leaf {
			bt.removeFromLeaf(x, idx)
		} else {
			bt.removeFromNonLeaf(x, idx)
		}
	} else {
		if x.leaf {
			return
		}

		flag := false
		if idx == len(x.keys) {
			flag = true
		}

		if len(x.children[idx].keys) < bt.t {
			bt.fill(x, idx)
		}

		if flag && idx > len(x.keys) {
			bt.deleteNode(x.children[idx-1], k)
		} else {
			bt.deleteNode(x.children[idx], k)
		}
	}
}

func (bt *BTree[K, V]) removeFromLeaf(x *Node[K, V], idx int) {
	copy(x.keys[idx:], x.keys[idx+1:])
	x.keys = x.keys[:len(x.keys)-1]

	copy(x.values[idx:], x.values[idx+1:])
	x.values = x.values[:len(x.values)-1]
}

func (bt *BTree[K, V]) removeFromNonLeaf(x *Node[K, V], idx int) {
	k := x.keys[idx]

	if len(x.children[idx].keys) >= bt.t {
		predK, predV := bt.getPred(x, idx)
		x.keys[idx] = predK
		x.values[idx] = predV
		bt.deleteNode(x.children[idx], predK)
	} else if len(x.children[idx+1].keys) >= bt.t {
		succK, succV := bt.getSucc(x, idx)
		x.keys[idx] = succK
		x.values[idx] = succV
		bt.deleteNode(x.children[idx+1], succK)
	} else {
		bt.merge(x, idx)
		bt.deleteNode(x.children[idx], k)
	}
}

func (bt *BTree[K, V]) getPred(x *Node[K, V], idx int) (K, V) {
	cur := x.children[idx]
	for !cur.leaf {
		cur = cur.children[len(cur.keys)]
	}
	return cur.keys[len(cur.keys)-1], cur.values[len(cur.values)-1]
}

func (bt *BTree[K, V]) getSucc(x *Node[K, V], idx int) (K, V) {
	cur := x.children[idx+1]
	for !cur.leaf {
		cur = cur.children[0]
	}
	return cur.keys[0], cur.values[0]
}

func (bt *BTree[K, V]) fill(x *Node[K, V], idx int) {
	if idx != 0 && len(x.children[idx-1].keys) >= bt.t {
		bt.borrowFromPrev(x, idx)
	} else if idx != len(x.keys) && len(x.children[idx+1].keys) >= bt.t {
		bt.borrowFromNext(x, idx)
	} else {
		if idx != len(x.keys) {
			bt.merge(x, idx)
		} else {
			bt.merge(x, idx-1)
		}
	}
}

func (bt *BTree[K, V]) borrowFromPrev(x *Node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx-1]

	var zeroK K
	var zeroV V

	child.keys = append([]K{zeroK}, child.keys...)
	child.keys[0] = x.keys[idx-1]

	child.values = append([]V{zeroV}, child.values...)
	child.values[0] = x.values[idx-1]

	if !child.leaf {
		child.children = append([]*Node[K, V]{nil}, child.children...)
		child.children[0] = sibling.children[len(sibling.keys)]
	}

	x.keys[idx-1] = sibling.keys[len(sibling.keys)-1]
	x.values[idx-1] = sibling.values[len(sibling.values)-1]

	sibling.keys = sibling.keys[:len(sibling.keys)-1]
	sibling.values = sibling.values[:len(sibling.values)-1]
	if !sibling.leaf {
		sibling.children = sibling.children[:len(sibling.children)-1]
	}
}

func (bt *BTree[K, V]) borrowFromNext(x *Node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx+1]

	child.keys = append(child.keys, x.keys[idx])
	child.values = append(child.values, x.values[idx])

	if !child.leaf {
		child.children = append(child.children, sibling.children[0])
	}

	x.keys[idx] = sibling.keys[0]
	x.values[idx] = sibling.values[0]

	copy(sibling.keys, sibling.keys[1:])
	sibling.keys = sibling.keys[:len(sibling.keys)-1]

	copy(sibling.values, sibling.values[1:])
	sibling.values = sibling.values[:len(sibling.values)-1]

	if !sibling.leaf {
		copy(sibling.children, sibling.children[1:])
		sibling.children = sibling.children[:len(sibling.children)-1]
	}
}

func (bt *BTree[K, V]) merge(x *Node[K, V], idx int) {
	child := x.children[idx]
	sibling := x.children[idx+1]

	child.keys = append(child.keys, x.keys[idx])
	child.values = append(child.values, x.values[idx])

	child.keys = append(child.keys, sibling.keys...)
	child.values = append(child.values, sibling.values...)

	if !child.leaf {
		child.children = append(child.children, sibling.children...)
	}

	copy(x.keys[idx:], x.keys[idx+1:])
	x.keys = x.keys[:len(x.keys)-1]

	copy(x.values[idx:], x.values[idx+1:])
	x.values = x.values[:len(x.values)-1]

	copy(x.children[idx+1:], x.children[idx+2:])
	x.children = x.children[:len(x.children)-1]
}
