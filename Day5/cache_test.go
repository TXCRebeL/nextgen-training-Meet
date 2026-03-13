package main

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkResolve_CustomCache(b *testing.B) {
	cache := NewDNSCache()
	
	// Pre-populate cash to avoid miss latencies in benchmark
	for i := 0; i < 1000; i++ {
		domain := fmt.Sprintf("sub%d.example.com", i)
		cache.AddRecord(domain, "192.168.1.1", 1*time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test wildcard iterating resolving over 1000 items
		// Using the fast-path exact match for half, and wildcard for half
		if i%2 == 0 {
			domain := fmt.Sprintf("sub%d.example.com", i%1000)
			cache.Resolve(domain)
		} else {
			cache.Resolve("*.example.com")
		}
	}
}

func BenchmarkResolve_InbuiltCache(b *testing.B) {
	cache := NewInbuiltDNSCache()
	
	// Pre-populate the cache
	for i := 0; i < 1000; i++ {
		domain := fmt.Sprintf("sub%d.example.com", i)
		cache.AddRecord(domain, "192.168.1.1", 1*time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			domain := fmt.Sprintf("sub%d.example.com", i%1000)
			cache.Resolve(domain)
		} else {
			cache.Resolve("*.example.com")
		}
	}
}
