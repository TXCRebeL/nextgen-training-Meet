package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"Day12/internal/handlers"
	"Day12/internal/middleware"
	"Day12/internal/models"
	"Day12/internal/store"
)

func main() {
	catalog := store.NewProductCatalog()

	// Optionally switch to testdata/products.json
	file, err := os.Open("../../testdata/products.json")
	if err != nil {
		file, err = os.Open("testdata/products.json")
	}

	if err == nil {
		decoder := json.NewDecoder(file)

		_, _ = decoder.Token()

		for decoder.More() {
			var p models.Product
			if err := decoder.Decode(&p); err != nil {
				continue
			}
			catalog.AddProduct(&p)
		}

		_, _ = decoder.Token()

		file.Close()
	} else {
		fmt.Println("No products.json found, starting with empty catalog.", err)
	}

	h := handlers.NewProductHandler(catalog)
	mux := http.NewServeMux()

	mux.Handle("GET /products/stats", middleware.Logging(http.HandlerFunc(h.GetStats)))
	mux.Handle("POST /products", middleware.Logging(http.HandlerFunc(h.AddProducts)))
	mux.Handle("GET /products", middleware.Logging(http.HandlerFunc(h.GetProducts)))
	mux.Handle("GET /products/{id}", middleware.Logging(http.HandlerFunc(h.GetProductByID)))
	mux.Handle("DELETE /products/{id}", middleware.Logging(http.HandlerFunc(h.DeleteProduct)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	pprofServer := &http.Server{
		Addr: ":8090",
	}

	fmt.Println("Server running on :8080")
	fmt.Println("pprof Server is running on :8090")

	go func() {
		if err := pprofServer.ListenAndServe(); err != nil {
			fmt.Println("Stopped Pprof:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server stopped:", err)
	}
}
