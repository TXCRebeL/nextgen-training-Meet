package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Day12/internal/models"
	"Day12/internal/store"
)

func setupTestServer() *httptest.Server {
	catalog := store.NewProductCatalog()
	h := NewProductHandler(catalog)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /products/stats", h.GetStats)
	mux.HandleFunc("POST /products", h.AddProducts)
	mux.HandleFunc("GET /products", h.GetProducts)
	mux.HandleFunc("GET /products/{id}", h.GetProductByID)
	mux.HandleFunc("DELETE /products/{id}", h.DeleteProduct)

	return httptest.NewServer(mux)
}

func TestAPIIntegration(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}

	newProduct := models.Product{
		ID:       "TEST-1",
		Name:     "Integration Test Product",
		Category: "Testing",
		Price:    99.99,
		Rating:   4.5,
	}

	body, _ := json.Marshal(newProduct)
	req, _ := http.NewRequest("POST", server.URL+"/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /products failed: %v", err)
	}
	resp.Body.Close()

	resp, _ = client.Get(server.URL + "/products/TEST-1")
	var p models.Product
	json.NewDecoder(resp.Body).Decode(&p)
	if p.Name != "Integration Test Product" {
		t.Errorf("Expected product name 'Integration Test Product', got '%s'", p.Name)
	}
	resp.Body.Close()

	resp, _ = client.Get(server.URL + "/products?min_price=10&max_price=200")
	var products []*models.Product
	json.NewDecoder(resp.Body).Decode(&products)
	if len(products) != 1 {
		t.Errorf("Expected 1 product in range query, got %d", len(products))
	}
	resp.Body.Close()

	req, _ = http.NewRequest("DELETE", server.URL+"/products/TEST-1", nil)
	client.Do(req)

	resp, _ = client.Get(server.URL + "/products/TEST-1")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 after delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	
	// Test Invalid JSON
	req, _ = http.NewRequest("POST", server.URL+"/products", strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected bad request on invalid JSON, got %d", resp.StatusCode)
	}
}
