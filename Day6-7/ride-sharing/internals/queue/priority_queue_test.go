package queue

import (
	"testing"
	"time"

	"github.com/meetbha/ride-sharing/internals/models"
)

func makeRide(id string, ago time.Duration) *models.Ride {
	return &models.Ride{
		ID:          id,
		RequestedAt: time.Now().Add(-ago),
	}
}

func TestEmptyQueue(t *testing.T) {
	pq := NewPriorityQueue()

	if pq.GetRideCount() != 0 {
		t.Errorf("Expected 0, got %d", pq.GetRideCount())
	}
	if pq.GetNextRide() != nil {
		t.Error("Expected nil from empty queue")
	}
}

func TestAddAndGetNextRide(t *testing.T) {
	pq := NewPriorityQueue()

	// Add rides: r2 (oldest) → r1 → r3 (newest)
	r1 := makeRide("r1", 5*time.Minute)
	r2 := makeRide("r2", 10*time.Minute)
	r3 := makeRide("r3", 1*time.Minute)

	pq.AddRide(r1)
	pq.AddRide(r2)
	pq.AddRide(r3)

	if pq.GetRideCount() != 3 {
		t.Fatalf("Expected 3 rides in queue, got %d", pq.GetRideCount())
	}

	// Oldest ride (r2) should come out first
	first := pq.GetNextRide()
	if first.ID != "r2" {
		t.Errorf("Expected r2 (oldest), got %s", first.ID)
	}

	second := pq.GetNextRide()
	if second.ID != "r1" {
		t.Errorf("Expected r1 (second oldest), got %s", second.ID)
	}

	third := pq.GetNextRide()
	if third.ID != "r3" {
		t.Errorf("Expected r3 (newest), got %s", third.ID)
	}

	if pq.GetRideCount() != 0 {
		t.Errorf("Queue should be empty, got %d", pq.GetRideCount())
	}
}

func TestRemoveRide(t *testing.T) {
	pq := NewPriorityQueue()

	r1 := makeRide("r1", 10*time.Minute)
	r2 := makeRide("r2", 5*time.Minute)
	r3 := makeRide("r3", 1*time.Minute)

	pq.AddRide(r1)
	pq.AddRide(r2)
	pq.AddRide(r3)

	// Remove the middle-age ride
	pq.RemoveRide(r2)

	if pq.GetRideCount() != 2 {
		t.Fatalf("Expected 2 after removal, got %d", pq.GetRideCount())
	}

	// r1 (oldest) should still come out first
	first := pq.GetNextRide()
	if first.ID != "r1" {
		t.Errorf("Expected r1, got %s", first.ID)
	}

	second := pq.GetNextRide()
	if second.ID != "r3" {
		t.Errorf("Expected r3, got %s", second.ID)
	}
}

func TestRemoveHead(t *testing.T) {
	pq := NewPriorityQueue()

	r1 := makeRide("r1", 10*time.Minute)
	r2 := makeRide("r2", 5*time.Minute)

	pq.AddRide(r1)
	pq.AddRide(r2)

	// Remove the oldest (heap root)
	pq.RemoveRide(r1)

	if pq.GetRideCount() != 1 {
		t.Fatalf("Expected 1, got %d", pq.GetRideCount())
	}

	next := pq.GetNextRide()
	if next.ID != "r2" {
		t.Errorf("Expected r2, got %s", next.ID)
	}
}

func TestManyRidesOrdering(t *testing.T) {
	pq := NewPriorityQueue()

	// Add 10 rides with decreasing age (r9 oldest → r0 newest)
	rides := make([]*models.Ride, 10)
	for i := 0; i < 10; i++ {
		rides[i] = makeRide("r"+string(rune('0'+i)), time.Duration(10-i)*time.Minute)
		pq.AddRide(rides[i])
	}

	// They should come out oldest first (r0 was created 10min ago)
	for i := 0; i < 10; i++ {
		r := pq.GetNextRide()
		if r == nil {
			t.Fatalf("Got nil at position %d", i)
		}
	}

	if pq.GetRideCount() != 0 {
		t.Errorf("Queue should be empty")
	}
}

// ── Benchmarks ──────────────────────────────────────────────────────

func BenchmarkAddRide(b *testing.B) {
	pq := NewPriorityQueue()
	for i := 0; i < b.N; i++ {
		pq.AddRide(makeRide("r", time.Duration(i)*time.Millisecond))
	}
}

func BenchmarkGetNextRide(b *testing.B) {
	pq := NewPriorityQueue()
	for i := 0; i < b.N; i++ {
		pq.AddRide(makeRide("r", time.Duration(i)*time.Millisecond))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.GetNextRide()
	}
}
