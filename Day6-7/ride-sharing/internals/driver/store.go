package driver

import (
	"fmt"
	"time"

	"github.com/meetbha/ride-sharing/internals/models"
)

// DriverStore defines the interface for driver lifecycle and state management operations.
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

// DriverManager implements the DriverStore interface and manages the lifecycle of drivers: registration, updates, status changes, and removal.
type DriverManager struct {
	Drivers     map[string]*models.Driver
	ZoneManager *ZoneManager
}

func NewDriverManager(zm *ZoneManager) *DriverManager {
	return &DriverManager{
		Drivers:     make(map[string]*models.Driver),
		ZoneManager: zm,
	}
}

// RegisterDriver adds a new driver and places them into the correct zone based on their location.
func (d *DriverManager) RegisterDriver(driver *models.Driver) {
	d.Drivers[driver.ID] = driver

	// Calculate zone from driver's lat/lng and add to zone manager
	d.ZoneManager.AddDriver(driver)
}

// GetDriver retrieves a driver by ID.
func (d *DriverManager) GetDriver(id string) (*models.Driver, error) {
	driver, exists := d.Drivers[id]
	if !exists {
		return nil, fmt.Errorf("driver %s not found", id)
	}
	return driver, nil
}

// EditDriver updates the name and rating of an existing driver.
func (d *DriverManager) EditDriver(id string, name string, rating float64) error {
	driver, exists := d.Drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}
	driver.Name = name
	driver.Rating = rating
	return nil
}

// UpdateLocation changes the driver's location and moves them to the correct zone if it changed.
func (d *DriverManager) UpdateLocation(id string, lat, lng float64) error {
	driver, exists := d.Drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}

	oldZone := CalculateZone(driver.Location.Lat, driver.Location.Lng)
	driver.Location = models.Location{Lat: lat, Lng: lng}

	d.ZoneManager.MoveDriver(driver, oldZone)

	return nil
}

// ChangeStatus updates the driver's availability status.
func (d *DriverManager) ChangeStatus(id string, status models.DriverStatus) error {
	driver, exists := d.Drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}
	driver.Status = status
	return nil
}

// RemoveDriver removes a driver from the registry and from their zone.
func (d *DriverManager) RemoveDriver(id string) error {
	_, exists := d.Drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}

	zone := CalculateZone(d.Drivers[id].Location.Lat, d.Drivers[id].Location.Lng)
	d.ZoneManager.RemoveDriver(id, zone)

	delete(d.Drivers, id)

	return nil
}

// GetDriverEarnings calculates the total earnings for a driver based on their ride history.
// timeframe can be "today" (last 24 hours) or "week" (last 7 days).
func (d *DriverManager) GetDriverEarnings(id string, timeframe string) (float64, error) {
	driver, err := d.GetDriver(id)
	if err != nil {
		return 0, err
	}

	totalEarnings := 0.0
	now := time.Now()

	for _, ride := range driver.RideHistory {
		if ride.Status != models.RideStatusCompleted {
			continue // Only count completed rides
		}

		timeSinceCompletion := now.Sub(ride.CompletedAt)

		switch timeframe {
		case "today":
			if timeSinceCompletion <= 24*time.Hour {
				totalEarnings += ride.Fare
			}
		case "week":
			if timeSinceCompletion <= 7*24*time.Hour {
				totalEarnings += ride.Fare
			}
		case "all":
			totalEarnings += ride.Fare
		default:
			return 0, fmt.Errorf("invalid timeframe: %s", timeframe)
		}
	}

	return totalEarnings, nil
}

// FindNearestDriver passes through to the underlying ZoneManager to fulfill the DriverStore interface.
func (d *DriverManager) FindNearestDriver(riderLat, riderLng float64) (*models.Driver, error) {
	return d.ZoneManager.FindNearestDriver(riderLat, riderLng)
}

func (d *DriverManager) IncrementRideCount(zone Zone) {
	d.ZoneManager.IncrementRideCount(zone)
}
