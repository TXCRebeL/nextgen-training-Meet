package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DNSRecord represents a cached DNS entry.
type DNSRecord struct {
	Domain    string        `json:"domain"`
	IP        string        `json:"ip"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
	HitCount  int           `json:"hit_count"`
}

// Ensure it's not expired based on TTL
func (r *DNSRecord) isExpired() bool {
	return time.Since(r.CreatedAt) > r.TTL
}

// DNSCache implements a TTL-based cache with wildcard matching.
// It leverages the custom MyMap[string] implementation from main.go
type DNSCache struct {
	mu          sync.RWMutex
	records     *MyMap[string]
	hits        int
	misses      int
	evictedHits int // track hits for records we had to evict
}

func NewDNSCache() *DNSCache {
	c := &DNSCache{
		records: NewMyMap[string](),
	}
	
	// Background cleanup every 30 seconds removes expired entries
	go c.startBackgroundCleanup()
	
	return c
}

// startBackgroundCleanup runs a time.Ticker to periodically evict expired records
func (c *DNSCache) startBackgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	// Routine runs indefinitely
	for range ticker.C {
		// Log statement can be uncommented for debugging
		// fmt.Println("[Background] TTL cleanup running...")
		c.EvictExpired()
	}
}

// AddRecord inserts a new domain entry into the cache with a specified TTL.
func (c *DNSCache) AddRecord(domain, ip string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.records.Insert(domain, &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       ttl,
		CreatedAt: time.Now(),
		HitCount:  0,
	})
}

// Resolve looks up a domain in the cache.
// It supports exact matches first, then falls back to wildcard (*.example.com) matches.
func (c *DNSCache) Resolve(domain string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Try Exact Match
	if val, exists := c.records.Get(domain); exists {
		record := val.(*DNSRecord)
		if record.isExpired() {
			c.evictedHits += record.HitCount
			c.records.Delete(domain)
			// Expired counts as a miss because we must fetch fresh data
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

		// We need to iterate over all entries in the MyMap to find a matching subdomain.
		// Since MyMap is an array of Linked Lists, we must traverse it.
		for i := range c.records.Buckets {
			curr := &c.records.Buckets[i]
			for curr != nil {
				for j := 0; j < curr.Count; j++ {
					record := curr.Slots[j].Value.(*DNSRecord)

					// Check if this record's domain ends with the suffix (e.g., ".example.com")
					// and it's not strictly equal to just ".example.com"
					if strings.HasSuffix(record.Domain, suffix) && len(record.Domain) > len(suffix) {
						if record.isExpired() {
							c.evictedHits += record.HitCount
							c.records.Delete(record.Domain)
						} else {
							record.HitCount++
							c.hits++
							return record.IP
						}
					}
				}
				curr = curr.Next
			}
		}
	}

	// 3. Cache Miss - Simulate upstream lookup
	c.misses++
	return c.simulateUpstreamLookup(domain)
}

// EvictExpired manually cleans up all expired entries from the custom map.
func (c *DNSCache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Since MyMap doesn't currently easily expose an iterator (only print),
	// here is a brute-force approach to walk the custom map for eviction.
	// We extract all keys that need deletion to a separate slice
	// to avoid modifying the map while walking it.
	var domainsToDelete []string

	for i := range c.records.Buckets {
		curr := &c.records.Buckets[i]
		for curr != nil {
			for j := 0; j < curr.Count; j++ {
				record := curr.Slots[j].Value.(*DNSRecord)
				if record.isExpired() {
					c.evictedHits += record.HitCount
					domainsToDelete = append(domainsToDelete, curr.Slots[j].Key)
				}
			}
			curr = curr.Next
		}
	}

	// Delete them after walking the tree
	for _, k := range domainsToDelete {
		c.records.Delete(k)
	}
}

// simulateUpstreamLookup simulates fetching IP from a real DNS server.
func (c *DNSCache) simulateUpstreamLookup(domain string) string {
	fmt.Printf("[Upstream Lookup] Resolving %s...\n", domain)
	time.Sleep(50 * time.Millisecond) // Simulating network latency

	// Dummy IP generation for simulation
	ip := fmt.Sprintf("192.168.1.%d", len(domain)%255)

	// Caching it for 5 seconds by default as simulating the upstream's TTL returned
	c.records.Insert(domain, &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       5 * time.Second,
		CreatedAt: time.Now(),
		HitCount:  0,
	})
	return ip
}

// PrintStats outputs cache metrics: hit rate, miss rate, total entries, etc.
func (c *DNSCache) PrintStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalLookups := c.hits + c.misses
	hitRate := 0.0
	missRate := 0.0

	if totalLookups > 0 {
		hitRate = float64(c.hits) / float64(totalLookups) * 100
		missRate = float64(c.misses) / float64(totalLookups) * 100
	}

	// Calculate map size securely
	mapEntries := c.records.Size

	entrySizeEstimate := 16 + 8 + int(16+16+8+24+8) // roughly domain + ip string header, ttl, time, hitcount
	totalMemEstimate := mapEntries * entrySizeEstimate

	fmt.Println("--- DNS Cache Statistics ---")
	fmt.Printf("Total Entries : %d\n", mapEntries)
	fmt.Printf("Total Lookups : %d\n", totalLookups)
	fmt.Printf("Hits          : %d (%.2f%%)\n", c.hits, hitRate)
	fmt.Printf("Misses        : %d (%.2f%%)\n", c.misses, missRate)
	fmt.Printf("Memory Est.   : ~%d bytes\n", totalMemEstimate)
	fmt.Println("----------------------------")
}
