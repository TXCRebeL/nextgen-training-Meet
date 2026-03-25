package bst

type BST struct {
	root *Node
}

type Node struct {
	editDistance int
	freq         int
	word         string
	left         *Node
	right        *Node
}

func NewBST() *BST {
	return &BST{
		root: nil,
	}
}

func isBetter(dist1, freq1, dist2, freq2 int) bool {
	if dist1 < dist2 {
		return true
	}
	if dist1 == dist2 && freq1 > freq2 {
		return true
	}
	return false
}

func (t *BST) Insert(word string, editDistance int, freq int) {
	t.root = t.insertNode(t.root, word, editDistance, freq)
}

func (t *BST) insertNode(node *Node, word string, editDistance int, freq int) *Node {
	if node == nil {
		return &Node{
			word:         word,
			editDistance: editDistance,
			freq:         freq,
		}
	}
	if isBetter(editDistance, freq, node.editDistance, node.freq) {
		node.left = t.insertNode(node.left, word, editDistance, freq)
	} else {
		node.right = t.insertNode(node.right, word, editDistance, freq)
	}
	return node
}

func (t *BST) GetSuggestions() []string {
	var result []string
	t.inorder(t.root, &result)
	return result
}

func (t *BST) inorder(root *Node, result *[]string) {
	if root == nil {
		return
	}
	t.inorder(root.left, result)
	*result = append(*result, root.word)
	t.inorder(root.right, result)
}
