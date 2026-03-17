package driver

import (
	"fmt"
	"testing"

	"github.com/meetbha/ride-sharing/internals/models"
)

func BenchmarkRegisterDriver(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			zm := NewZoneManager()
			dm := NewDriverManager(zm)
			for i := 0; i < b.N; i++ {
				dm.RegisterDriver(&models.Driver{
					ID:       fmt.Sprintf("d%d", i),
					Name:     "Bench",
					Status:   models.DriverStatusAvailable,
					Location: models.Location{Lat: float64(i % 100), Lng: float64(i % 100)},
				})
			}
		})
	}
}

func BenchmarkFindNearestDriver(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			zm := NewZoneManager()
			dm := NewDriverManager(zm)
			// Pre-fill with drivers spread across zones
			for i := 0; i < size; i++ {
				dm.RegisterDriver(&models.Driver{
					ID:       fmt.Sprintf("d%d", i),
					Name:     "Bench",
					Status:   models.DriverStatusAvailable,
					Location: models.Location{Lat: float64(i % 50), Lng: float64(i % 50)},
				})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dm.FindNearestDriver(25.0, 25.0)
			}
		})
	}
}

func BenchmarkUpdateLocation(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			zm := NewZoneManager()
			dm := NewDriverManager(zm)
			// Fill with drivers
			for i := 0; i < size; i++ {
				dm.RegisterDriver(&models.Driver{
					ID:       fmt.Sprintf("d%d", i),
					Name:     "Bench",
					Status:   models.DriverStatusAvailable,
					Location: models.Location{Lat: float64(i % 100), Lng: float64(i % 100)},
				})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dm.UpdateLocation("d1", float64(i%100), float64(i%100))
			}
		})
	}
}
