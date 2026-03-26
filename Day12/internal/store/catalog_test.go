package store

import (
	"testing"

	"Day12/internal/models"
)

func TestProductCatalog(t *testing.T) {
	catalog := NewProductCatalog()

	p1 := &models.Product{ID: "p1", Name: "Laptop", Price: 1500.0}
	p2 := &models.Product{ID: "p2", Name: "Mouse", Price: 50.0}
	p3 := &models.Product{ID: "p3", Name: "Keyboard", Price: 150.0}
	p4 := &models.Product{ID: "p4", Name: "Monitor", Price: 300.0}
	p5 := &models.Product{ID: "p5", Name: "Gaming Mouse", Price: 50.0}

	catalog.AddProduct(p1)
	catalog.AddProduct(p2)
	catalog.AddProduct(p3)
	catalog.AddProduct(p4)
	catalog.AddProduct(p5)

	if p, ok := catalog.GetProductByID("p3"); !ok || p.Name != "Keyboard" {
		t.Errorf("Failed to retrieve product by ID")
	}

	res := catalog.GetProductsByPriceRange(40.0, 200.0)
	if len(res) != 3 {
		t.Errorf("Expected 3 products in range [40, 200], got %d", len(res))
	}

	newP2 := *p2
	newP2.Price = 250.0
	catalog.AddProduct(&newP2)

	resAfter := catalog.GetProductsByPriceRange(40.0, 200.0)
	if len(resAfter) != 2 {
		t.Errorf("Expected 2 products in range after price update, got %d", len(resAfter))
	}

	catalog.RemoveProduct("p1")
	if _, ok := catalog.GetProductByID("p1"); ok {
		t.Errorf("Product p1 should be removed")
	}

	emptyRes := catalog.GetProductsByPriceRange(1400.0, 1600.0)
	if len(emptyRes) != 0 {
		t.Errorf("Secondary index not cleared upon Delete, found %d items", len(emptyRes))
	}
}
