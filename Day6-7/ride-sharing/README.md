# 🚗 Ride-Sharing Dispatch System

A Go-based ride-sharing dispatch system built with custom data structures — no external dependencies.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                       CLI (cmd/main.go)                  │
│              Interactive command-line interface           │
└──────────────────┬───────────────────────────────────────┘
                   │
┌──────────────────▼───────────────────────────────────────┐
│                  Dispatcher (dispatch/)                   │
│   Orchestrates ride lifecycle: request → assign → complete│
│   - 10-min timeout cancellation                          │
│   - Re-queue if no driver found                          │
│   - Fare calculation: ₹50 base + ₹12/km                 │
└───────┬──────────────┬───────────────┬──────────────────┘
        │              │               │
┌───────▼──────┐ ┌─────▼──────┐ ┌──────▼──────────┐
│ DriverStore  │ │ RideQueue  │ │  RideManager    │
│  (driver/)   │ │  (queue/)  │ │   (rides/)      │
│              │ │            │ │                  │
│ HashMap by ID│ │  Min-Heap  │ │  Linked List    │
│ + Zone Index │ │ (wait time)│ │  (active rides) │
│  (grid-based)│ │            │ │ + History Slice  │
└──────────────┘ └────────────┘ └─────────────────┘
```

## Project Structure

```
ride-sharing/
├── cmd/
│   └── main.go              # CLI entry point
├── internals/
│   ├── models/
│   │   └── model.go          # Driver, Rider, Ride, Location
│   ├── driver/
│   │   ├── store.go          # DriverStore interface + DriverManager (HashMap)
│   │   ├── zone.go           # ZoneManager (grid-based spatial index)
│   │   ├── store_test.go     # Earnings tests
│   │   ├── zone_test.go      # Zone calculation + nearest driver tests
│   │   ├── driver_registry_test.go  # Registry sync tests
│   │   └── benchmark_test.go # Benchmarks
│   ├── queue/
│   │   ├── priority_queue.go      # RideQueue interface + min-heap
│   │   └── priority_queue_test.go # Ordering, removal, benchmarks
│   ├── rides/
│   │   ├── active_rides.go        # Singly-linked list for active rides
│   │   ├── ride_manager.go        # RideManager (active + history)
│   │   └── ride_manager_test.go   # Linked list ops, manager, benchmarks
│   └── dispatch/
│       ├── dispatcher.go          # Central dispatch coordinator
│       └── dispatcher_test.go     # End-to-end test
├── go.mod
└── README.md
```

## Data Structure Trade-offs

| Data Structure | Used In | Why | Trade-off |
|---|---|---|---|
| **HashMap** (`map[string]*Driver`) | Driver Store | O(1) lookup/insert/delete by ID | No ordering; uses extra memory for hash table |
| **HashMap + Zone Grid** (`map[Zone][]*Driver`) | Zone Manager | O(1) zone lookup; spatial queries search only 9 zones | Drivers near zone boundaries may be missed by single-zone queries (solved by checking 8 neighbors) |
| **Min-Heap** (array-based) | Priority Queue | O(log n) insert/extract-min for wait-time priority | O(n) for arbitrary removal; no random access |
| **Singly Linked List** | Active Rides | O(1) append; O(n) search — appropriate for in-transit rides (small N) | No random access; O(n) lookup by ID |
| **Append-Only Slice** | Ride History | O(1) append; sequential iteration for reporting | Immutable — never delete from history |

## Complexity Analysis

| Operation | Time | Space | Implementation |
|---|---|---|---|
| Register Driver | O(1) | O(1) | HashMap insert + zone list append |
| Get Driver by ID | O(1) | - | HashMap lookup |
| Update Location | O(k) | O(1) | Remove from old zone O(k), add to new zone O(1), k = drivers in zone |
| Change Status | O(1) | - | HashMap lookup + field mutation |
| Find Nearest Driver | O(9k) | O(k) | Check 9 adjacent zones, filter by distance ≤ 5km |
| Request Ride | O(log n) | O(1) | Heap insert |
| Get Next Ride | O(log n) | O(1) | Heap extract-min |
| Remove Ride from Queue | O(n) | O(1) | Linear search + heap rebalance |
| Assign Driver | O(log n + 9k) | O(1) | Queue pop + nearest driver search |
| Complete Ride | O(m) | O(1) | Active list search (m = active rides) |
| Get Average Wait Time | O(1) | - | Pre-computed running total |
| Get Driver Earnings | O(h) | O(1) | Scan driver's ride history |
| Get Busiest Zone | O(z) | O(1) | Scan all zones with ride counts |

Where: n = queued rides, k = drivers per zone, m = active rides, h = driver's ride count, z = total zones

## Interfaces

```go
// DriverStore — driver lifecycle management
type DriverStore interface {
    RegisterDriver(driver *models.Driver)
    GetDriver(id string) (*models.Driver, error)
    EditDriver(id string, name string, rating float64) error
    UpdateLocation(id string, lat, lng float64) error
    ChangeStatus(id string, status models.DriverStatus) error
    RemoveDriver(id string) error
    GetDriverEarnings(id string, timeframe string) (float64, error)
    FindNearestDriver(riderLat, riderLng float64) (*models.Driver, error)
    IncrementRideCount(zone Zone)
}

// RideQueue — priority queue for ride requests
type RideQueue interface {
    AddRide(ride *models.Ride)
    RemoveRide(ride *models.Ride)
    GetNextRide() *models.Ride
    GetRideCount() int
}
```

## How to Run

```bash
# Run the CLI
cd cmd && go run .

# Run all tests with coverage
go test ./... -cover -v

# Run benchmarks
go test ./... -bench=. -benchmem -run=^$
```

## CLI Commands

| Command | Description |
|---|---|
| `register-driver <name> <lat> <lng>` | Register a driver (ID & rating auto-generated) |
| `get-driver <id>` | Show driver details |
| `edit-driver <id> <name> <rating>` | Update driver name and rating |
| `update-location <id> <lat> <lng>` | Move driver to new coordinates |
| `change-status <id> available\|busy\|offline` | Change driver availability |
| `remove-driver <id>` | Remove driver from system |
| `driver-earnings <id> today\|week\|all` | Query driver earnings |
| `find-nearest <lat> <lng>` | Find nearest available driver |
| `list-drivers` | List all registered drivers |
| `zone-drivers <lat> <lng>` | List drivers in a zone |
| `busy-zone` | Show busiest zone |
| `request-ride <pLat> <pLng> <dLat> <dLng>` | Request a ride (riderID auto-generated) |
| `assign-driver` | Process queue — assign oldest request to nearest driver |
| `complete-ride <rideID>` | Complete an active ride, calculate fare |
| `queue-count` | Number of rides waiting |
| `active-rides` | List in-transit rides |
| `ride-history` | List completed/cancelled rides |
| `avg-wait-time` | Average rider wait time |
