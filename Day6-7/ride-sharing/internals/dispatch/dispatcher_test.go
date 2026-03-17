package dispatch

import (
	"testing"

	"github.com/meetbha/ride-sharing/internals/driver"
	"github.com/meetbha/ride-sharing/internals/models"
	"github.com/meetbha/ride-sharing/internals/queue"
	"github.com/meetbha/ride-sharing/internals/rides"
)

func TestDispatcherEndToEnd(t *testing.T) {
	// 1. Initialize core system modules
	zm := driver.NewZoneManager()
	ds := driver.NewDriverManager(zm)
	rq := queue.NewPriorityQueue()
	rm := rides.NewRideManager()

	d := NewDispatcher(ds, rq, rm)

	// 2. Setup Drivers
	d1 := &models.Driver{
		ID: "d1", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 10.0, Lng: 10.0},
	}
	d2 := &models.Driver{ // Out of range and busy
		ID: "d2", Status: models.DriverStatusBusy, Location: models.Location{Lat: 50.0, Lng: 50.0},
	}
	ds.RegisterDriver(d1)
	ds.RegisterDriver(d2)

	// 3. User Requests a Ride
	pickup := models.Location{Lat: 11.0, Lng: 11.0} // Distance to d1 is ~1.41
	drop := models.Location{Lat: 20.0, Lng: 20.0}

	rideID := d.RequestRide("rider_X", pickup, drop)

	if rq.GetRideCount() != 1 {
		t.Fatalf("Ride was not cleanly appended to priority queue")
	}

	// 4. Dispatch matches ride to the nearest driver (d1)
	assignedDriverID, err := d.AssignDriver()
	if err != nil {
		t.Fatalf("Failed associating ride: %v", err)
	}

	if assignedDriverID != "d1" {
		t.Errorf("Expected driver d1 to be assigned, instead got %s", assignedDriverID)
	}

	// Ensure ride is off the queue
	if rq.GetRideCount() != 0 {
		t.Errorf("Queue wasn't forcefully popped on assignment")
	}

	// Ensure driver is busy and ride is Active
	d1Obj, _ := ds.GetDriver("d1")
	if d1Obj.Status != models.DriverStatusBusy {
		t.Errorf("Driver state wasn't transitioned to busy")
	}

	activeRide, err := rm.GetActiveRide(rideID)
	if err != nil || activeRide.Status != models.RideStatusAccepted {
		t.Errorf("Ride wasn't pushed effectively to Active array")
	}

	// 5. Complete Transit logic execution
	fare, err := d.CompleteRide(rideID)
	if err != nil {
		t.Fatalf("CompleteRide failed cleanly routing history transition: %v", err)
	}

	// Base 50 + distance(sqrt((9)^2 + (9)^2) = ~12.72) * 12 = ~202.73
	if fare <= 50.0 {
		t.Errorf("Fare mathematically wasn't evaluated linearly")
	}

	// Ensure Driver became geographically re-located to the dropoff (20.0, 20.0) and Available!
	d1Obj, _ = ds.GetDriver("d1")
	if d1Obj.Location.Lat != 20.0 || d1Obj.Status != models.DriverStatusAvailable {
		t.Errorf("Driver geo-zone constraints were not rewritten post transit")
	}

	if len(d1Obj.RideHistory) != 1 {
		t.Errorf("Ride wasn't structurally archived on individual driver history slice")
	}

	sysHistory := rm.GetRideHistory()
	if len(sysHistory) != 1 || sysHistory[0].Status != models.RideStatusCompleted {
		t.Errorf("Chronological Append-Only History state machine failure")
	}
}
