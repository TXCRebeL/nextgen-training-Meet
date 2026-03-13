package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("------------------------------------------------------")
	fmt.Println("Starting Custom DNS Cache implementation testing...")
	fmt.Println("------------------------------------------------------")

	// Initialize Custom Cache which wraps our custom HashMap `MyMap`
	cache := NewDNSCache()

	// 1. Add specific records
	fmt.Println("Adding initial DNS records to Cache...")
	cache.AddRecord("google.com", "142.250.190.46", 3*time.Second)
	cache.AddRecord("amazon.com", "205.251.242.103", 5*time.Second)

	// Add subdomains to demonstrate the reverse wildcard querying
	cache.AddRecord("sub.example.com", "192.168.1.50", 10*time.Second)
	cache.AddRecord("api.example.com", "192.168.1.51", 10*time.Second)

	// -------------------------------------------------------------------------------- //

	// 2. Exact Matches Tests
	fmt.Println("\n[TEST: Exact Matches]")
	fmt.Println("Resolving google.com (Expect Exact HIT) -> IP:", cache.Resolve("google.com"))

	// Expect it to log upstream simulation
	fmt.Println("Resolving netflix.com (Expect MISS - Fetch from Simulate Upstream) -> IP:", cache.Resolve("netflix.com"))

	// Fast follow to see it cached from the upstream call above
	fmt.Println("Resolving netflix.com again (Expect HIT from earlier upstream fetch) -> IP:", cache.Resolve("netflix.com"))

	// -------------------------------------------------------------------------------- //

	// 3. User Specific Wildcard Matching Requirement Test
	// Requirement: If I Resolve(*.example.com) then it should return an IP for sub.example.com or meta.example.com if they exist.
	fmt.Println("\n[TEST: Resolving Wildcard Query directly -> `*.example.com`]")
	fmt.Println("Querying `*.example.com`...")

	// Should hit "sub.example.com" or "api.example.com" inside the custom cache map list!
	ip := cache.Resolve("*.example.com")
	fmt.Printf("Resolved `*.example.com` to -> IP: %s (Hit an existing subdomain record!)\n", ip)

	// -------------------------------------------------------------------------------- //

	fmt.Println("\n[CURRENT CACHE STATS]")
	cache.PrintStats()

	// -------------------------------------------------------------------------------- //

	// 4. Test Eviction/Expiration
	fmt.Println("\n[TEST: Expiration / Eviction Lifecycle]")
	fmt.Println("Waiting 4 seconds to let 'google.com' TTL (3s) expire...")
	time.Sleep(4 * time.Second)

	// First we should see our manual eviction clean it up (or it dropping next resolve call)
	fmt.Println("Running manual EvictExpired()...")
	cache.EvictExpired() // Explicit eviction

	fmt.Println("Resolving google.com post expiration (Expect MISS - Upstream due to it being evicted) -> IP:", cache.Resolve("google.com"))

	fmt.Println("\n------------------------------------------------------")
	fmt.Println("[FINAL CACHE STATS]")
	cache.PrintStats()

}
