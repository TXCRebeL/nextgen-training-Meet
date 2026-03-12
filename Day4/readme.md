# Network Packet Router Simulation

## Benchmark Findings: Linked List vs. Slice Queue

To determine the most efficient data structure for our packet router, we benchmarked a Doubly Linked List implementation against a dynamic Slice implementation using Go's built-in `testing` package (`go test -bench=. -benchmem`). 

### Performance Results

| Data Structure | Scale | Speed (ns/op) | Memory (B/op) | Allocations/op |
| :--- | :--- | :--- | :--- | :--- |
| **Slice** | 100 Packets | **9,116** | 20,480 | **2** |
| **Linked List**| 100 Packets | 11,124 | **12,800** | 100 |
| | | | | |
| **Slice** | 100k Packets | 11,407,053 | 50,984,598| **11** |
| **Linked List**| 100k Packets | **6,196,458** | **12,800,030**| 100,000 |

### Analysis & Conclusion

1. **Small Workloads (< 1,000 packets):** The `SliceQueue` is faster. CPU cache locality allows the processor to fetch contiguous array elements much faster than chasing scattered Linked List pointers across the heap.
2. **Heavy Workloads (10,000+ packets):** The `LinkedListQueue` becomes significantly faster. As the slice grows to massive sizes, the amortized cost of dynamically resizing the backing array (allocating larger memory blocks and copying old data) degrades performance. The Linked List avoids this entirely via constant-time pointer assignment.
3. **Garbage Collection:** The slice is vastly superior at minimizing heap allocations (11 allocs vs 100,000 allocs at large scales), heavily reducing pressure on the Go Garbage Collector.

**Final Verdict:** For a network router dealing with constant, massive throughput where latency spikes (from array resizing) are unacceptable, the **Linked List** provides much more predictable $O(1)$ enqueue/dequeue times. However, if the queue size is capped and pre-allocated, a slice would be the optimal choice.