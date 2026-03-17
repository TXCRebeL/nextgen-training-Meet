package driver

import (
	"testing"

	"github.com/meetbha/ride-sharing/internals/models"
)

func TestDriverRegistrySync(t *testing.T) {
	zm := NewZoneManager()
	dm := NewDriverManager(zm)

	driver := &models.Driver{
		ID:     "d1",
		Name:   "Meet",
		Status: models.DriverStatusBusy,
		Location: models.Location{
			Lat: 12.9716,
			Lng: 77.5946,
		},
	}

	dm.RegisterDriver(driver)

	zone := CalculateZone(driver.Location.Lat, driver.Location.Lng)
	zoneDrivers := zm.GetDrivers(zone)

	if len(zoneDrivers) != 1 {
		t.Errorf("Expected 1 driver in zone, got %d", len(zoneDrivers))
	}

	if zoneDrivers[0].Status != models.DriverStatusBusy {
		t.Errorf("Expected driver status to be %s, got %s", models.DriverStatusBusy, zoneDrivers[0].Status)
	}

	// Change status via driver manager
	err := dm.ChangeStatus("d1", models.DriverStatusAvailable)
	if err != nil {
		t.Errorf("Unexpected error changing status: %v", err)
	}

	// Verify that the status in the zone manager is also updated
	zoneDriversAfter := zm.GetDrivers(zone)
	if zoneDriversAfter[0].Status != models.DriverStatusAvailable {
		t.Errorf("Expected zone manager driver status to be synced to %s, got %s", models.DriverStatusAvailable, zoneDriversAfter[0].Status)
	}
}
