package store

import (
	"Day12/internal/btree"
	"Day12/internal/models"
)

type CatalogStats struct {
	TotalProducts     int            `json:"total_products"`
	CategoryCounts    map[string]int `json:"category_counts"`
	PriceDistribution map[string]int `json:"price_distribution"`
	AverageRating     float64        `json:"average_rating"`
}

// ProductCatalog maintains products by ID and a B-Tree index by Price.
type ProductCatalog struct {
	products map[string]*models.Product
	priceIdx *btree.BTree[float64, []string]
}

func NewProductCatalog() *ProductCatalog {
	return &ProductCatalog{
		products: make(map[string]*models.Product),
		priceIdx: btree.NewBTree[float64, []string](50), // Scale up minimum degree t
	}
}

// AddProduct adds or updates a product
func (c *ProductCatalog) AddProduct(p *models.Product) {
	if existing, exists := c.products[p.ID]; exists {
		if existing.Price != p.Price {
			c.removePriceIndex(existing.ID, existing.Price)
			c.addPriceIndex(p.ID, p.Price)
		}
	} else {
		c.addPriceIndex(p.ID, p.Price)
	}

	c.products[p.ID] = p
}

func (c *ProductCatalog) addPriceIndex(id string, price float64) {
	if val, ok := c.priceIdx.Search(price); ok {
		found := false
		for _, v := range val {
			if v == id {
				found = true
				break
			}
		}
		if !found {
			val = append(val, id)
			c.priceIdx.Insert(price, val)
		}
	} else {
		c.priceIdx.Insert(price, []string{id})
	}
}

func (c *ProductCatalog) removePriceIndex(id string, oldPrice float64) {
	if val, ok := c.priceIdx.Search(oldPrice); ok {
		newVal := make([]string, 0, len(val))
		for _, v := range val {
			if v != id {
				newVal = append(newVal, v)
			}
		}
		if len(newVal) == 0 {
			c.priceIdx.Delete(oldPrice)
		} else {
			c.priceIdx.Insert(oldPrice, newVal)
		}
	}
}

func (c *ProductCatalog) GetProductByID(id string) (*models.Product, bool) {
	p, ok := c.products[id]
	return p, ok
}

func (c *ProductCatalog) GetProductsByPriceRange(min, max float64) []*models.Product {
	results := make([]*models.Product, 0, 500)
	c.priceIdx.WalkRange(min, max, func(price float64, ids []string) {
		for _, id := range ids {
			if p, ok := c.products[id]; ok {
				results = append(results, p)
			}
		}
	})
	return results
}

func (c *ProductCatalog) RemoveProduct(id string) {
	if p, exists := c.products[id]; exists {
		c.removePriceIndex(id, p.Price)
		delete(c.products, id)
	}
}

func (c *ProductCatalog) GetAllProducts() []*models.Product {
	var results []*models.Product
	for _, p := range c.products {
		results = append(results, p)
	}
	return results
}

func (c *ProductCatalog) GetAllProductsSortedByPrice() []*models.Product {
	items := c.priceIdx.InOrder()
	results := make([]*models.Product, 0, len(c.products))
	for _, item := range items {
		for _, id := range item.Value {
			if p, ok := c.products[id]; ok {
				results = append(results, p)
			}
		}
	}
	return results
}

func (c *ProductCatalog) GetStats() CatalogStats {
	stats := CatalogStats{
		TotalProducts:     len(c.products),
		CategoryCounts:    make(map[string]int),
		PriceDistribution: map[string]int{
			"0-50":    0,
			"51-100":  0,
			"101-500": 0,
			"501+":    0,
		},
	}

	var totalRating float64
	for _, p := range c.products {
		stats.CategoryCounts[p.Category]++
		totalRating += p.Rating

		if p.Price <= 50 {
			stats.PriceDistribution["0-50"]++
		} else if p.Price <= 100 {
			stats.PriceDistribution["51-100"]++
		} else if p.Price <= 500 {
			stats.PriceDistribution["101-500"]++
		} else {
			stats.PriceDistribution["501+"]++
		}
	}

	if stats.TotalProducts > 0 {
		stats.AverageRating = totalRating / float64(stats.TotalProducts)
	}

	return stats
}
