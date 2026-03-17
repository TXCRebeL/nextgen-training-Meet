package driver

import (
	"testing"

	"github.com/meetbha/ride-sharing/internals/models"
)

func TestCalculateZone2D(t *testing.T) {
	// Points that should snap to the same 5-unit 2D grid
	lat1, lng1 := 12.0000, 77.0100 // mapped to 10_75
	lat2, lng2 := 14.9900, 79.9900 // mapped to 10_75

	zone1 := CalculateZone(lat1, lng1)
	zone2 := CalculateZone(lat2, lng2)

	if zone1 != zone2 {
		t.Errorf("Expected zone1 and zone2 to be the same, but got %s and %s", zone1, zone2)
	}

	// Point outside the 5.0 unit band
	lat3, lng3 := 15.0000, 80.0000 // mapped to 15_80
	zone3 := CalculateZone(lat3, lng3)
	if zone1 == zone3 {
		t.Errorf("Expected zone3 to be different from zone1, but both got %s", zone1)
	}
}

func TestFindNearestAvailableDrivers(t *testing.T) {
	zm := NewZoneManager()

	// Driver 1: Very close, within the same grid (Distance ~1.41)
	zm.AddDriver(&models.Driver{
		ID: "d1", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 11.0, Lng: 11.0},
	})

	// Driver 2: Neighbor grid, but distance is exactly 5.0
	zm.AddDriver(&models.Driver{
		ID: "d2", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 15.0, Lng: 10.0},
	})

	// Driver 3: Neighbor grid, but distance is too far (e.g. 6.0)
	zm.AddDriver(&models.Driver{
		ID: "d3", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 16.0, Lng: 10.0},
	})

	// Driver 4: Close, but busy
	zm.AddDriver(&models.Driver{
		ID: "d4", Status: models.DriverStatusBusy, Location: models.Location{Lat: 10.5, Lng: 10.5},
	})

	riderLat, riderLng := 10.0, 10.0 // Rider is exactly on a grid line (maps to 10_10 zone)

	drivers := zm.FindNearestDrivers(riderLat, riderLng)

	if len(drivers) != 2 {
		t.Fatalf("Expected 2 drivers, found %d", len(drivers))
	}

	// Let's test FindNearestDriver (should return d1 at dist 1.41)
	nearest, err := zm.FindNearestDriver(riderLat, riderLng)
	if err != nil {
		t.Fatalf("Unexpected error retrieving nearest driver: %v", err)
	}

	if nearest.ID != "d1" {
		t.Errorf("Expected nearest driver to be d1, got %s", nearest.ID)
	}
}

func TestFindNearestDriversNoAvailability(t *testing.T) {
	zm := NewZoneManager()

	// All drivers are too far
	zm.AddDriver(&models.Driver{
		ID: "d1", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 20.0, Lng: 20.0},
	})
	zm.AddDriver(&models.Driver{
		ID: "d2", Status: models.DriverStatusAvailable, Location: models.Location{Lat: 10.0, Lng: 16.0}, // dist = 6
	})

	riderLat, riderLng := 10.0, 10.0
	_, err := zm.FindNearestDriver(riderLat, riderLng)

	if err == nil {
		t.Fatalf("Expected an error since no driver is within 5 units, but got nil")
	}
}
