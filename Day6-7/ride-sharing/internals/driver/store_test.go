package driver

import (
	"testing"
	"time"

	"github.com/meetbha/ride-sharing/internals/models"
)

func TestGetDriverEarnings(t *testing.T) {
	zm := NewZoneManager()
	dm := NewDriverManager(zm)

	now := time.Now()

	driver := &models.Driver{
		ID:       "d1",
		Name:     "Meet",
		Status:   models.DriverStatusBusy,
		Location: models.Location{Lat: 12.9716, Lng: 77.5946},
		RideHistory: []models.Ride{
			{ID: "r1", Status: models.RideStatusCompleted, Fare: 100.0, CompletedAt: now.Add(-2 * time.Hour)},   // Today
			{ID: "r2", Status: models.RideStatusCompleted, Fare: 250.0, CompletedAt: now.Add(-50 * time.Hour)},  // This Week
			{ID: "r3", Status: models.RideStatusCompleted, Fare: 150.0, CompletedAt: now.Add(-300 * time.Hour)}, // Last Month
			{ID: "r4", Status: models.RideStatusCancelled, Fare: 50.0, CompletedAt: now.Add(-1 * time.Hour)},    // Cancelled (ignore)
		},
	}

	dm.RegisterDriver(driver)

	// Test Today
	todayEarnings, err := dm.GetDriverEarnings("d1", "today")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if todayEarnings != 100.0 {
		t.Errorf("Expected today's earnings to be 100.0, got %f", todayEarnings)
	}

	// Test Week
	weekEarnings, err := dm.GetDriverEarnings("d1", "week")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if weekEarnings != 350.0 { // 100 (today) + 250 (this week)
		t.Errorf("Expected week's earnings to be 350.0, got %f", weekEarnings)
	}

	// Test All Time
	allEarnings, err := dm.GetDriverEarnings("d1", "all")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if allEarnings != 500.0 { // 100 + 250 + 150
		t.Errorf("Expected all earnings to be 500.0, got %f", allEarnings)
	}
}
