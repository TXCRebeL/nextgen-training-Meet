package driver

import (
	"fmt"
	"math"

	"github.com/meetbha/ride-sharing/internals/models"
)

type Zone string

const ZoneSizeKm = 5.0

// ZoneManager manages drivers grouped by geographic zones and tracks ride activity per zone.
type ZoneManager struct {
	DriverLocation map[Zone][]*models.Driver
	RideCount      map[Zone]int
}

func NewZoneManager() *ZoneManager {
	return &ZoneManager{
		DriverLocation: make(map[Zone][]*models.Driver),
		RideCount:      make(map[Zone]int),
	}
}

// CalculateZone computes a zone string from lat/lng in a 2D plane space.
// Each zone represents a 5x5 grid cell.
func CalculateZone(lat, lng float64) Zone {
	// Snap lat and lng to the nearest multiple of ZoneSizeKm (5.0)
	zoneLat := math.Floor(lat/ZoneSizeKm) * ZoneSizeKm
	zoneLng := math.Floor(lng/ZoneSizeKm) * ZoneSizeKm

	return Zone(fmt.Sprintf("%.0f_%.0f", zoneLat, zoneLng))
}

// AddDriver adds a driver to the given zone.
func (z *ZoneManager) AddDriver(driver *models.Driver) {
	zone := CalculateZone(driver.Location.Lat, driver.Location.Lng)
	z.DriverLocation[zone] = append(z.DriverLocation[zone], driver)
}

// GetDrivers returns all drivers currently in the given zone.
func (z *ZoneManager) GetDrivers(zone Zone) []*models.Driver {
	return z.DriverLocation[zone]
}

// RemoveDriver removes a specific driver from the given zone.
func (z *ZoneManager) RemoveDriver(driverID string, zone Zone) {
	drivers := z.DriverLocation[zone]
	for i, d := range drivers {
		if d.ID == driverID {
			z.DriverLocation[zone] = append(drivers[:i], drivers[i+1:]...)
			return
		}
	}
}

// MoveDriver moves a driver from an old zone to a new zone.
// This is called when a driver's location is updated and the zone changes.
func (z *ZoneManager) MoveDriver(driver *models.Driver, oldZone Zone) {
	z.RemoveDriver(driver.ID, oldZone)
	z.AddDriver(driver)
}

// IncrementRideCount increments the ride count for a zone when a ride request is made.
func (z *ZoneManager) IncrementRideCount(zone Zone) {
	z.RideCount[zone]++
}

// GetRideCount returns the number of rides that have been taken from the given zone.
func (z *ZoneManager) GetRideCount(zone Zone) int {
	return z.RideCount[zone]
}

// GetBusyZone returns the busiest zone.
func (z *ZoneManager) GetBusyZone() Zone {
	var busiestZone Zone
	count := math.MinInt
	for zone := range z.RideCount {
		if z.RideCount[zone] > count {
			count = z.RideCount[zone]
			busiestZone = zone
		}
	}
	return busiestZone
}

// FindNearestDrivers finds all available drivers within a 5-unit radius (assuming 2D coordinates)
// by checking the rider's zone and all 8 surrounding zones.
func (z *ZoneManager) FindNearestDrivers(riderLat, riderLng float64) []*models.Driver {
	zoneLat := math.Floor(riderLat/ZoneSizeKm) * ZoneSizeKm
	zoneLng := math.Floor(riderLng/ZoneSizeKm) * ZoneSizeKm

	var nearbyDrivers []*models.Driver
	offsets := []float64{-ZoneSizeKm, 0, ZoneSizeKm}

	for _, offsetLat := range offsets {
		for _, offsetLng := range offsets {
			checkZone := Zone(fmt.Sprintf("%.0f_%.0f", zoneLat+offsetLat, zoneLng+offsetLng))
			driversInZone := z.GetDrivers(checkZone)

			for _, driver := range driversInZone {
				if driver.Status != models.DriverStatusAvailable {
					continue
				}

				// Basic 2D distance calculation (Pythagorean theorem)
				dLat := driver.Location.Lat - riderLat
				dLng := driver.Location.Lng - riderLng
				dist := dLat*dLat + dLng*dLng

				if dist <= ZoneSizeKm*ZoneSizeKm {
					nearbyDrivers = append(nearbyDrivers, driver)
				}
			}
		}
	}

	return nearbyDrivers
}

// FindNearestDriver finds the single closest available driver within a 5-unit radius.
func (z *ZoneManager) FindNearestDriver(riderLat, riderLng float64) (*models.Driver, error) {
	drivers := z.FindNearestDrivers(riderLat, riderLng)
	if len(drivers) == 0 {
		return nil, fmt.Errorf("no available driver found within %v units", ZoneSizeKm)
	}

	var nearestDriver *models.Driver
	shortestDistance := math.MaxFloat64

	for _, driver := range drivers {
		dLat := driver.Location.Lat - riderLat
		dLng := driver.Location.Lng - riderLng
		dist := dLat*dLat + dLng*dLng

		if dist < shortestDistance {
			shortestDistance = dist
			nearestDriver = driver
		}
	}

	return nearestDriver, nil
}
