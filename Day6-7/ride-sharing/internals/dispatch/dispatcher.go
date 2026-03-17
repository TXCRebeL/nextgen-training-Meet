package dispatch

import (
	"fmt"
	"math"
	"time"

	"github.com/meetbha/ride-sharing/internals/driver"
	"github.com/meetbha/ride-sharing/internals/models"
	"github.com/meetbha/ride-sharing/internals/queue"
	"github.com/meetbha/ride-sharing/internals/rides"
)

type Dispatcher struct {
	DriverStore driver.DriverStore
	RideQueue   queue.RideQueue
	RideManager *rides.RideManager
}

// NewDispatcher initializes the central dispatch coordinator.
func NewDispatcher(ds driver.DriverStore, rq queue.RideQueue, rm *rides.RideManager) *Dispatcher {
	return &Dispatcher{
		DriverStore: ds,
		RideQueue:   rq,
		RideManager: rm,
	}
}

// RequestRide creates a new ride request and enqueues it in the Priority Queue.
func (d *Dispatcher) RequestRide(riderID string, pickup, drop models.Location) string {
	rideID := fmt.Sprintf("ride-%d", time.Now().UnixNano())
	ride := &models.Ride{
		ID:             rideID,
		RiderID:        riderID,
		PickupLocation: pickup,
		DropLocation:   drop,
		Status:         models.RideStatusPending,
		RequestedAt:    time.Now(),
	}

	d.DriverStore.IncrementRideCount(driver.CalculateZone(pickup.Lat, pickup.Lng))
	d.RideQueue.AddRide(ride)
	return rideID
}

// AssignDriver forcefully processes the queue, finding the ride waiting the longest,
// and attempting to assign the nearest 5-unit driver to it.
// It returns the assigned Driver ID if successful, or an error if no drivers are available.
func (d *Dispatcher) AssignDriver() (string, error) {
	ride := d.RideQueue.GetNextRide()
	if ride == nil {
		return "", fmt.Errorf("no rides waiting in the queue")
	}

	// 1. Timeout constraints: If the ride has waited for longer than 10 minutes, cancel it outright.
	if time.Since(ride.RequestedAt) > 10*time.Minute {
		// Log system notification
		fmt.Printf("NOTIFICATION: Ride %s cancelled due to >10m wait time\n", ride.ID)

		// Immediately push to History slices as aborted since it was never genuinely active
		ride.Status = models.RideStatusCancelled
		d.RideManager.ArchiveRide(ride)

		return "", fmt.Errorf("ride %s aborted due to 10-minute timeout", ride.ID)
	}

	// 2. Calculate nearest available driver securely via Euclidean Zone Math
	nearestDriver, err := d.DriverStore.FindNearestDriver(ride.PickupLocation.Lat, ride.PickupLocation.Lng)
	if err != nil {
		// If no driver found, re-queue the ride! It still maintains its original old RequestedAt timestamp so it stays prioritized.
		d.RideQueue.AddRide(ride)
		return "", fmt.Errorf("could not find a nearby available driver for ride %s: %v", ride.ID, err)
	}

	// 3. Assign Driver to Ride
	ride.DriverID = nearestDriver.ID
	ride.Status = models.RideStatusAccepted
	ride.StartedAt = time.Now()

	// 4. Add Ride to the Active System (Linked List array)
	d.RideManager.AddActiveRide(ride)

	// 5. Mark the assigned driver as strictly 'Busy' in the Store framework
	err = d.DriverStore.ChangeStatus(nearestDriver.ID, models.DriverStatusBusy)
	if err != nil {
		return "", fmt.Errorf("critical failure marking assigned driver busy: %v", err)
	}

	fmt.Printf("The Ride %s has been assigned to driver %s\n", ride.ID, nearestDriver.ID)

	return nearestDriver.ID, nil
}

// CompleteRide finishes an active ride by calculating its fare based on
// geometric distance and marking the transit completed in historical archives.
func (d *Dispatcher) CompleteRide(rideID string) (float64, error) {
	// A) Fetch the active transaction from the Linked List
	ride, err := d.RideManager.GetActiveRide(rideID)
	if err != nil {
		return 0, fmt.Errorf("cannot complete ride: %v", err)
	}

	// B) Mathematical formulation of transit boundaries
	distLat := ride.PickupLocation.Lat - ride.DropLocation.Lat
	distLng := ride.PickupLocation.Lng - ride.DropLocation.Lng
	distanceKm := math.Sqrt(distLat*distLat + distLng*distLng)

	// User requested fare logic: base 50 + 12/km
	fare := 50.0 + (distanceKm * 12.0)
	ride.Fare = fare
	ride.CompletedAt = time.Now()

	// C) Remove from Active List -> Append to System History slice
	err = d.RideManager.FinishRide(rideID, models.RideStatusCompleted)
	if err != nil {
		return 0, fmt.Errorf("failed closing out active ride: %v", err)
	}

	// D) Update Driver's geographic location to the Drop-off point (this will automatically recalculate their Zone geometrically in DriverStore)
	err = d.DriverStore.UpdateLocation(ride.DriverID, ride.DropLocation.Lat, ride.DropLocation.Lng)
	if err != nil {
		return 0, fmt.Errorf("could not update driver location post-transit: %v", err)
	}

	// E) Mark Driver internally available for consecutive assignments
	err = d.DriverStore.ChangeStatus(ride.DriverID, models.DriverStatusAvailable)
	if err != nil {
		return 0, fmt.Errorf("could not mark driver %s available: %v", ride.DriverID, err)
	}

	// F) Also append the completed Ride directly onto the individual Driver struct history
	driverObj, err := d.DriverStore.GetDriver(ride.DriverID)
	if err == nil {
		driverObj.RideHistory = append(driverObj.RideHistory, *ride)
	}

	fmt.Printf("Ride %s completed. The distance was %.2f km and the fare was %.2f\n", rideID, distanceKm, fare)
	return fare, nil
}

// GetAverageWaitTime returns the rolling average wait time for all fulfilled system rides.
func (d *Dispatcher) GetAverageWaitTime() time.Duration {
	return d.RideManager.GetAverageWaitTime()
}
