package ds

import (
	"errors"
	"strings"
	"sync"
)

var ErrCategoryExists = errors.New("category already exists")
var ErrCategoryNotFound = errors.New("category not found")

// CategoryNode represents a category and its subcategories in a Tree structure
type CategoryNode struct {
	Name          string
	Subcategories map[string]*CategoryNode
}

// CategoryTree is a thread-safe graph/tree for category navigation
// (e.g., Electronics > Phones > Smartphones)
type CategoryTree struct {
	mu   sync.RWMutex
	root *CategoryNode
}

func NewCategoryTree() *CategoryTree {
	return &CategoryTree{
		root: &CategoryNode{
			Name:          "Root",
			Subcategories: make(map[string]*CategoryNode),
		},
	}
}

// AddCategory adds a path of categories (e.g., "Electronics/Phones/Smartphones")
func (t *CategoryTree) AddCategoryPath(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	parts := strings.Split(path, "/")
	current := t.root

	for _, part := range parts {
		if part == "" {
			continue
		}
		if _, exists := current.Subcategories[part]; !exists {
			current.Subcategories[part] = &CategoryNode{
				Name:          part,
				Subcategories: make(map[string]*CategoryNode),
			}
		}
		current = current.Subcategories[part]
	}
}

// GetSubcategories returns the direct subcategories of a given path
func (t *CategoryTree) GetSubcategories(path string) ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	current := t.root
	if path != "" && path != "/" && path != "Root" {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			if part == "" {
				continue
			}
			if next, exists := current.Subcategories[part]; exists {
				current = next
			} else {
				return nil, ErrCategoryNotFound
			}
		}
	}

	var subs []string
	for name := range current.Subcategories {
		subs = append(subs, name)
	}
	return subs, nil
}
