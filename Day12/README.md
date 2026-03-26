# High Performance Go Product Catalog

A scalable REST API designed to simulate an e-commerce product catalog with a custom built **Generic B-Tree index**. The architecture provides $O(\log N)$ price range filtering and sort analytics completely independent of standard linear array limitations.

## Folder Structure
```text
catalog/
├── cmd/server/main.go        # Entrypoint mapping router and booting processes
├── internal/
│ ├── btree/                  # Custom B-Tree implementation (Generics)
│ ├── models/                 # Product JSON definitions and shared structs
│ ├── store/                  # Catalog holding HashMaps and the B-Tree
│ ├── handlers/               # REST HTTP processing controllers
│ └── middleware/             # Intercepts latency, recovery, and logging
├── testdata/                 # Bootstrapping environment tests
└── README.md
```

## Profiling & Performance Results
The project contains load benchmark verification tests contrasting our `B-Tree` algorithm to standard `Linear Scanning` under **100,000 Concurrent Products**.

### CPU Execution Speeds (`go tool pprof`)
*Using `-benchtime=100x` enforcing 100 benchmark loops across exactly 100K stored objects:*
- **Linear Scan:** `~300 µs` per query
- **B-Tree Range:** `~69 µs` per query (**4.3x Faster**)

### Memory Management and Allocations
- **Zero-Allocation Traversals:** We modified our custom B-Tree's `RangeQuery` engine into an allocation-free callback tracker `WalkRange()`. This completely bypasses iterative slice appending along tree ascensions.
- During standard concurrent benchmark loads, memory explicitly allocated on the Heap by the B-Tree algorithms dropped from `240 KB` down to just `55 KB`, representing purely the pointers allocated natively inside the final API JSON boundary wrapper.

### JSON Streaming Startup Optimization
The server boots up parsing a massive 100,000 item `products.json` file. By explicitly refactoring the loader to use `json.Decoder` tokens and streaming individual product structs into memory sequentially (`decoder.More()`), we achieved the following metric improvements on the Heap (`go tool pprof -memprofile`):
- **Max Physical Memory (MaxRSS)**: Reduced from `~120 MB` down strictly to **`~45 MB`**
- **Total Startup Allocations (`TotalAlloc`)**: Reduced from `~165 MB` down to **`~39 MB`** (saving effectively 126 MB of bloated slice garbage blocks).
- **Final Heap Usage (`HeapAlloc`)**: Stabilized comfortably at **`~35.5 MB`** resting natively in the B-Tree.

The B-Tree structural index ensures the actual API handlers provide near-instant endpoint evaluation regardless of the scale of the mapped catalog map.
