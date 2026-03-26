package store

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"Day12/internal/models"
)

var (
	benchCatalog  *ProductCatalog
	benchFlatList []*models.Product
	setupOnce     sync.Once
)

func setupBenchmarkData(numProducts int) (*ProductCatalog, []*models.Product) {
	rng := rand.New(rand.NewSource(42))
	catalog := NewProductCatalog()
	flatList := make([]*models.Product, 0, numProducts)

	for i := 0; i < numProducts; i++ {
		p := &models.Product{
			ID:    fmt.Sprintf("PROD-%d", i),
			Price: float64(rng.Intn(5000)) + rng.Float64(),
		}
		catalog.AddProduct(p)
		flatList = append(flatList, p)
	}
	return catalog, flatList
}

func BenchmarkBTreeRangeQuery(b *testing.B) {
	counter := []int{100, 1000, 10000, 100000}
	for _, count := range counter {
		catalog, _ := setupBenchmarkData(count)
		b.Run(fmt.Sprintf("size_%d", count), func(b *testing.B) {
			minPrice := 500.0
			maxPrice := 600.0
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = catalog.GetProductsByPriceRange(minPrice, maxPrice)
			}
		})
	}
}

func BenchmarkLinearScan(b *testing.B) {
	counter := []int{100, 1000, 10000, 100000}
	for _, count := range counter {
		_, flatList := setupBenchmarkData(count)
		b.Run(fmt.Sprintf("size_%d", count), func(b *testing.B) {
			minPrice := 500.0
			maxPrice := 600.0
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var results []*models.Product
				for _, p := range flatList {
					if p.Price >= minPrice && p.Price <= maxPrice {
						results = append(results, p)
					}
				}
				_ = results
			}
		})
	}
}
