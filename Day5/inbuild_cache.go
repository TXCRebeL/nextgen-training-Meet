package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// InbuiltDNSCache implements a TTL-based cache with wildcard matching,
// using Go's built-in map[string]*DNSRecord structure.
type InbuiltDNSCache struct {
	mu          sync.RWMutex
	records     map[string]*DNSRecord
	hits        int
	misses      int
	evictedHits int
}

func NewInbuiltDNSCache() *InbuiltDNSCache {
	c := &InbuiltDNSCache{
		records: make(map[string]*DNSRecord),
	}
	
	// Background cleanup every 30 seconds removes expired entries
	go c.startBackgroundCleanup()
	
	return c
}

// startBackgroundCleanup runs a time.Ticker to periodically evict expired records
func (c *InbuiltDNSCache) startBackgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	// Routine runs indefinitely
	for range ticker.C {
		c.EvictExpired()
	}
}

func (c *InbuiltDNSCache) AddRecord(domain, ip string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.records[domain] = &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       ttl,
		CreatedAt: time.Now(),
		HitCount:  0,
	}
}

func (c *InbuiltDNSCache) Resolve(domain string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Try Exact Match
	if record, exists := c.records[domain]; exists {
		if record.isExpired() {
			c.evictedHits += record.HitCount
			delete(c.records, domain)
		} else {
			record.HitCount++
			c.hits++
			return record.IP
		}
	}

	// 2. Try Wildcard Match if the query itself is a wildcard
	// e.g., querying "*.example.com" should find an IP for "sub.example.com" or "meta.example.com" if they exist in the cache
	if strings.HasPrefix(domain, "*.") {
		suffix := domain[1:] // Extract ".example.com"

		// Need to iterate over the built-in map
		for k, record := range c.records {
			if strings.HasSuffix(k, suffix) && len(k) > len(suffix) {
				if record.isExpired() {
					c.evictedHits += record.HitCount
					delete(c.records, k)
				} else {
					record.HitCount++
					c.hits++
					return record.IP
				}
			}
		}
	}

	// 3. Cache Miss - Simulate upstream lookup
	c.misses++
	return c.simulateUpstreamLookup(domain)
}

func (c *InbuiltDNSCache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var domainsToDelete []string
	for k, record := range c.records {
		if record.isExpired() {
			c.evictedHits += record.HitCount
			domainsToDelete = append(domainsToDelete, k)
		}
	}

	for _, k := range domainsToDelete {
		delete(c.records, k)
	}
}

func (c *InbuiltDNSCache) simulateUpstreamLookup(domain string) string {
	// Reduce latency in simulation to avoid muddying benchmark metrics if hit heavily
	// For actual benchmark of just resolve matching, we shouldn't hit this often if pre-populated.
	// time.Sleep(50 * time.Millisecond)

	ip := fmt.Sprintf("192.168.1.%d", len(domain)%255)
	c.records[domain] = &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       5 * time.Second,
		CreatedAt: time.Now(),
		HitCount:  0,
	}
	return ip
}

func (c *InbuiltDNSCache) PrintStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalLookups := c.hits + c.misses
	hitRate := 0.0
	missRate := 0.0

	if totalLookups > 0 {
		hitRate = float64(c.hits) / float64(totalLookups) * 100
		missRate = float64(c.misses) / float64(totalLookups) * 100
	}

	mapEntries := len(c.records)
	entrySizeEstimate := 16 + 8 + int(16+16+8+24+8) // domain, ip strings, ttl, struct size
	totalMemEstimate := mapEntries * entrySizeEstimate

	fmt.Println("--- Inbuilt DNS Cache Statistics ---")
	fmt.Printf("Total Entries : %d\n", mapEntries)
	fmt.Printf("Total Lookups : %d\n", totalLookups)
	fmt.Printf("Hits          : %d (%.2f%%)\n", c.hits, hitRate)
	fmt.Printf("Misses        : %d (%.2f%%)\n", c.misses, missRate)
	fmt.Printf("Memory Est.   : ~%d bytes\n", totalMemEstimate)
	fmt.Println("------------------------------------")
}
