package rides

import (
	"fmt"
	"time"

	"github.com/meetbha/ride-sharing/internals/models"
)

// RideManager coordinates active rides traversing a Custom Linked List
// and finished/aborted rides archiving to an Append-Only slice.
type RideManager struct {
	ActiveRides         *ActiveRides
	ActiveIndex         map[string]*models.Ride // O(1) Lookup
	RideHistory         []*models.Ride
	TotalWaitTime       time.Duration
	TotalRidesFulfilled int
}

// NewRideManager initializes the manager bridging active state and history.
func NewRideManager() *RideManager {
	return &RideManager{
		ActiveRides: NewActiveRides(),
		ActiveIndex: make(map[string]*models.Ride),
		RideHistory: make([]*models.Ride, 0),
	}
}

// AddActiveRide assigns a driver and registers the ride into the Linked List,
// immediately resolving and tallying its wait-time footprint.
func (rm *RideManager) AddActiveRide(ride *models.Ride) {
	// 1. Calculate the exact starvation wait time and tally it in O(1) state memory.
	waitDuration := time.Since(ride.RequestedAt)
	rm.TotalWaitTime += waitDuration
	rm.TotalRidesFulfilled++

	// 2. Append visually to Active linked list memory
	rm.ActiveRides.AddRide(ride)
	// 3. Update the quick lookup index
	rm.ActiveIndex[ride.ID] = ride
}

// GetActiveRide recursively simulates an O(N) lookup over the Linked List because
// it lacks true indices, fetching an Active Ride to permit status mutations.
func (rm *RideManager) GetActiveRide(rideID string) (*models.Ride, error) {
	ride, ok := rm.ActiveIndex[rideID]
	if !ok {
		return nil, fmt.Errorf("active ride %s not found", rideID)
	}
	return ride, nil
}

// FinishRide removes the ride node from the Active Linked List
// and Appends it chronologically to the immutable RideHistory.
func (rm *RideManager) FinishRide(rideID string, finalStatus models.RideStatus) error {
	ride, err := rm.GetActiveRide(rideID)
	if err != nil {
		return err // Cannot finish an untracked active ride
	}
	// Update terminal status
	ride.Status = finalStatus

	// O(N) deletion operation bridging head/tail states
	rm.ActiveRides.RemoveRide(ride)

	// Remove from quick lookup index
	delete(rm.ActiveIndex, rideID)

	// Highly efficient O(1) mathematical memory trailing append
	rm.RideHistory = append(rm.RideHistory, ride)

	return nil
}

// GetRideHistory returns the chronological (Append-Only) slice list of all archived rides.
func (rm *RideManager) GetRideHistory() []*models.Ride {
	return rm.RideHistory
}

// ArchiveRide directly appends a ride to the history slice without it ever being active.
// Useful for rides that are cancelled directly from the queue.
func (rm *RideManager) ArchiveRide(ride *models.Ride) {
	rm.RideHistory = append(rm.RideHistory, ride)
}

// GetAverageWaitTime returns the O(1) mathematical calculation of average queue starvation.
func (rm *RideManager) GetAverageWaitTime() time.Duration {
	if rm.TotalRidesFulfilled == 0 {
		return 0
	}
	return rm.TotalWaitTime / time.Duration(rm.TotalRidesFulfilled)
}
