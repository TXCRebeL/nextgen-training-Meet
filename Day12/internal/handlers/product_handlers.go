package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"Day12/internal/models"
	"Day12/internal/store"
)

type ProductHandler struct {
	Catalog *store.ProductCatalog
}

func NewProductHandler(catalog *store.ProductCatalog) *ProductHandler {
	return &ProductHandler{Catalog: catalog}
}

func (h *ProductHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.Catalog.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *ProductHandler) AddProducts(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	h.Catalog.AddProduct(&p)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	minStr := r.URL.Query().Get("min_price")
	maxStr := r.URL.Query().Get("max_price")
	category := r.URL.Query().Get("category")
	sortParam := r.URL.Query().Get("sort")
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	var products []*models.Product
	var err error

	if minStr != "" || maxStr != "" {
		products, err = h.getProductsByPriceRange(minStr, maxStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if sortParam == "price" {
		products = h.Catalog.GetAllProductsSortedByPrice()
	} else {
		products = h.Catalog.GetAllProducts()
	}

	if category != "" {
		var filtered []*models.Product
		for _, p := range products {
			if strings.EqualFold(p.Category, category) {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}

	if pageStr != "" || sizeStr != "" {
		page := 1
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		size := 20
		if sizeStr != "" {
			if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
				size = s
			}
		}

		start := (page - 1) * size
		end := start + size

		if start > len(products) {
			products = []*models.Product{}
		} else {
			if end > len(products) {
				end = len(products)
			}
			products = products[start:end]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if p, ok := h.Catalog.GetProductByID(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	} else {
		http.Error(w, "Product not found", http.StatusNotFound)
	}
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := h.Catalog.GetProductByID(id); !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	h.Catalog.RemoveProduct(id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) getProductsByPriceRange(minStr, maxStr string) ([]*models.Product, error) {
	if minStr == "" {
		minStr = "0"
	}
	if maxStr == "" {
		maxStr = strconv.FormatInt(math.MaxInt64, 10)
	}

	min, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		return nil, err
	}

	max, err := strconv.ParseFloat(maxStr, 64)
	if err != nil {
		return nil, err
	}

	products := h.Catalog.GetProductsByPriceRange(min, max)
	if products == nil {
		products = []*models.Product{}
	}
	return products, nil
}
