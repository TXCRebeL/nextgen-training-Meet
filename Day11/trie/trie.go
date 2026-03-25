package trie

type TrieNode struct {
	children map[rune]*TrieNode
	IsWord   bool
	Freq     int
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

func (t *Trie) Insert(word string) {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[ch]
	}
	node.IsWord = true
	node.Freq++
}

// get all words with freq
func (t *Trie) GetAllWords() []string {
	var results []string
	var dfs func(*TrieNode, string)
	dfs = func(node *TrieNode, prefix string) {
		if node.IsWord {
			results = append(results, prefix)
		}
		for ch, child := range node.children {
			dfs(child, prefix+string(ch))
		}
	}
	dfs(t.root, "")
	return results
}

func (t *Trie) Search(word string) bool {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			return false
		}
		node = node.children[ch]
	}
	return node.IsWord
}

func (t *Trie) StartsWith(prefix string) bool {
	node := t.root
	for _, ch := range prefix {
		if _, ok := node.children[ch]; !ok {
			return false
		}
		node = node.children[ch]
	}
	return true
}

func (t *Trie) Delete(word string) bool {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			return false
		}
		node = node.children[ch]
	}
	if !node.IsWord {
		return false
	}
	node.IsWord = false
	node.Freq--
	return true
}

func (t *Trie) AutoComplete(prefix string, limit int) []string {
	node := t.root
	for _, ch := range prefix {
		if _, ok := node.children[ch]; !ok {
			return nil
		}
		node = node.children[ch]
	}
	var results []string
	var dfs func(*TrieNode, string)
	dfs = func(node *TrieNode, prefix string) {
		if len(results) >= limit {
			return
		}
		if node.IsWord {
			results = append(results, prefix)
		}
		for ch, child := range node.children {
			dfs(child, prefix+string(ch))
		}
	}
	dfs(node, prefix)
	return results
}
func (t *Trie) GetFreq(word string) int {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			return 0
		}
		node = node.children[ch]
	}
	if node.IsWord {
		return node.Freq
	}
	return 0
}
func (t *Trie) DidYouMean(word string, maxDist int) []string {
	var candidates []string
	current := make([]int, len(word)+1)
	for i := range current {
		current[i] = i
	}

	for ch, child := range t.root.children {
		t.dymSearch(child, string(ch), word, current, maxDist, &candidates)
	}
	return candidates
}

func (t *Trie) dymSearch(node *TrieNode, currentWord string, target string, prevRow []int, maxDist int, results *[]string) {
	rowSize := len(target) + 1
	currentRow := make([]int, rowSize)
	currentRow[0] = prevRow[0] + 1

	minDist := currentRow[0]
	for i := 1; i < rowSize; i++ {
		cost := 1
		if rune(target[i-1]) == rune(currentWord[len(currentWord)-1]) {
			cost = 0
		}
		currentRow[i] = minVal(
			currentRow[i-1]+1,   // Insertion
			prevRow[i]+1,        // Deletion
			prevRow[i-1]+cost,   // Substitution
		)
		if currentRow[i] < minDist {
			minDist = currentRow[i]
		}
	}

	if currentRow[rowSize-1] <= maxDist && node.IsWord {
		*results = append(*results, currentWord)
	}

	if minDist <= maxDist {
		for ch, child := range node.children {
			t.dymSearch(child, currentWord+string(ch), target, currentRow, maxDist, results)
		}
	}
}

func minVal(vals ...int) int {
	res := vals[0]
	for _, v := range vals[1:] {
		if v < res {
			res = v
		}
	}
	return res
}
