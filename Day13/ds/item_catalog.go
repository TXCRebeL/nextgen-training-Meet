package ds

import (
	"Day13/models"
	"sync"
)

// BSTNode represents a node in the Binary Search Tree
type BSTNode struct {
	Item  *models.Item
	Left  *BSTNode
	Right *BSTNode
}

// ItemBST is a Binary Search Tree tailored for Items, sorted by StartPrice
type ItemBST struct {
	Root *BSTNode
}

// Insert adds an item to the BST based on its StartPrice
func (t *ItemBST) Insert(item *models.Item) {
	newNode := &BSTNode{Item: item}

	if t.Root == nil {
		t.Root = newNode
		return
	}

	current := t.Root
	for {
		if item.StartPrice < current.Item.StartPrice {
			if current.Left == nil {
				current.Left = newNode
				return
			}
			current = current.Left
		} else {
			if current.Right == nil {
				current.Right = newNode
				return
			}
			current = current.Right
		}
	}
}

// GetItemsInRange performs an in-order traversal to find items within a price range
func (t *ItemBST) GetItemsInRange(minPrice, maxPrice float64) []*models.Item {
	var result []*models.Item
	var traverse func(node *BSTNode)

	traverse = func(node *BSTNode) {
		if node == nil {
			return
		}

		// If current price is strictly greater than minPrice, search left
		if node.Item.StartPrice > minPrice {
			traverse(node.Left)
		}

		// If current price is within range, add it
		if node.Item.StartPrice >= minPrice && node.Item.StartPrice <= maxPrice {
			result = append(result, node.Item)
		}

		// If current price is strictly less than maxPrice, search right
		if node.Item.StartPrice < maxPrice {
			traverse(node.Right)
		}
	}

	traverse(t.Root)
	return result
}

// ItemCatalog maintains an ItemBST per category for fast range queries
type ItemCatalog struct {
	mu         sync.RWMutex
	categories map[string]*ItemBST
}

func NewItemCatalog() *ItemCatalog {
	return &ItemCatalog{
		categories: make(map[string]*ItemBST),
	}
}

// AddItem adds an item to its corresponding category's BST
func (c *ItemCatalog) AddItem(item *models.Item) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.categories[item.Category]; !exists {
		c.categories[item.Category] = &ItemBST{}
	}
	c.categories[item.Category].Insert(item)
}

// SearchRange queries items in a category within a price range
func (c *ItemCatalog) SearchRange(category string, minPrice, maxPrice float64) []*models.Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if bst, exists := c.categories[category]; exists {
		return bst.GetItemsInRange(minPrice, maxPrice)
	}
	return nil
}
