package rides

import (
	"fmt"
	"testing"
	"time"

	"github.com/meetbha/ride-sharing/internals/models"
)

func makeRide(id string, ago time.Duration) *models.Ride {
	return &models.Ride{
		ID:          id,
		Status:      models.RideStatusAccepted,
		RequestedAt: time.Now().Add(-ago),
		StartedAt:   time.Now(),
	}
}

// ── ActiveRides linked list tests ───────────────────────────────────

func TestActiveRidesAddAndTraverse(t *testing.T) {
	ar := NewActiveRides()

	ar.AddRide(makeRide("r1", 3*time.Minute))
	ar.AddRide(makeRide("r2", 2*time.Minute))
	ar.AddRide(makeRide("r3", 1*time.Minute))

	// Walk the linked list
	count := 0
	cur := ar.Head
	for cur != nil {
		count++
		cur = cur.Next
	}
	if count != 3 {
		t.Errorf("Expected 3 nodes, got %d", count)
	}

	// Verify head and tail
	if ar.Head.Ride.ID != "r1" {
		t.Errorf("Expected head=r1, got %s", ar.Head.Ride.ID)
	}
	if ar.Tail.Ride.ID != "r3" {
		t.Errorf("Expected tail=r3, got %s", ar.Tail.Ride.ID)
	}
}

func TestActiveRidesRemoveHead(t *testing.T) {
	ar := NewActiveRides()
	r1 := makeRide("r1", 0)
	r2 := makeRide("r2", 0)
	ar.AddRide(r1)
	ar.AddRide(r2)

	ar.RemoveRide(r1)

	if ar.Head.Ride.ID != "r2" {
		t.Errorf("Head should be r2 after removing r1, got %s", ar.Head.Ride.ID)
	}
	if ar.Tail.Ride.ID != "r2" {
		t.Errorf("Tail should be r2, got %s", ar.Tail.Ride.ID)
	}
}

func TestActiveRidesRemoveTail(t *testing.T) {
	ar := NewActiveRides()
	r1 := makeRide("r1", 0)
	r2 := makeRide("r2", 0)
	r3 := makeRide("r3", 0)
	ar.AddRide(r1)
	ar.AddRide(r2)
	ar.AddRide(r3)

	ar.RemoveRide(r3)

	if ar.Tail.Ride.ID != "r2" {
		t.Errorf("Tail should now be r2, got %s", ar.Tail.Ride.ID)
	}
	if ar.Tail.Next != nil {
		t.Error("Tail.Next should be nil")
	}
}

func TestActiveRidesRemoveMiddle(t *testing.T) {
	ar := NewActiveRides()
	r1 := makeRide("r1", 0)
	r2 := makeRide("r2", 0)
	r3 := makeRide("r3", 0)
	ar.AddRide(r1)
	ar.AddRide(r2)
	ar.AddRide(r3)

	ar.RemoveRide(r2)

	if ar.Head.Ride.ID != "r1" {
		t.Errorf("Head should be r1, got %s", ar.Head.Ride.ID)
	}
	if ar.Head.Next.Ride.ID != "r3" {
		t.Errorf("r1-next should be r3, got %s", ar.Head.Next.Ride.ID)
	}
}

func TestActiveRidesRemoveLast(t *testing.T) {
	ar := NewActiveRides()
	r1 := makeRide("r1", 0)
	ar.AddRide(r1)

	ar.RemoveRide(r1)

	if ar.Head != nil || ar.Tail != nil {
		t.Error("Both head and tail should be nil after removing the only node")
	}
}

// ── RideManager tests ───────────────────────────────────────────────

func TestAddAndGetActiveRide(t *testing.T) {
	rm := NewRideManager()
	ride := makeRide("r1", 2*time.Minute)

	rm.AddActiveRide(ride)

	got, err := rm.GetActiveRide("r1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got.ID != "r1" {
		t.Errorf("Expected r1, got %s", got.ID)
	}
}

func TestGetActiveRideNotFound(t *testing.T) {
	rm := NewRideManager()
	_, err := rm.GetActiveRide("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent ride")
	}
}

func TestFinishRideMovesToHistory(t *testing.T) {
	rm := NewRideManager()
	ride := makeRide("r1", 2*time.Minute)
	rm.AddActiveRide(ride)

	err := rm.FinishRide("r1", models.RideStatusCompleted)
	if err != nil {
		t.Fatalf("FinishRide failed: %v", err)
	}

	// Should no longer be in active rides
	_, err = rm.GetActiveRide("r1")
	if err == nil {
		t.Error("Ride should not be in active rides after finishing")
	}

	// Should be in history
	history := rm.GetRideHistory()
	if len(history) != 1 {
		t.Fatalf("Expected 1 ride in history, got %d", len(history))
	}
	if history[0].Status != models.RideStatusCompleted {
		t.Errorf("Expected status completed, got %s", history[0].Status)
	}
}

func TestArchiveRide(t *testing.T) {
	rm := NewRideManager()
	ride := makeRide("r1", 0)
	ride.Status = models.RideStatusCancelled

	rm.ArchiveRide(ride)

	history := rm.GetRideHistory()
	if len(history) != 1 || history[0].ID != "r1" {
		t.Error("Archived ride should appear in history")
	}
}

func TestAverageWaitTime(t *testing.T) {
	rm := NewRideManager()

	// No rides — should return 0
	if rm.GetAverageWaitTime() != 0 {
		t.Error("Expected 0 wait time with no rides")
	}

	// Add two rides with known wait times
	r1 := makeRide("r1", 4*time.Second)
	r2 := makeRide("r2", 6*time.Second)
	rm.AddActiveRide(r1)
	rm.AddActiveRide(r2)

	avg := rm.GetAverageWaitTime()
	// Average should be roughly 5 seconds (4+6)/2, allow tolerance
	if avg < 4*time.Second || avg > 7*time.Second {
		t.Errorf("Expected avg ~5s, got %v", avg)
	}
}

// ── Benchmarks ──────────────────────────────────────────────────────

func BenchmarkAddActiveRide(b *testing.B) {
	rm := NewRideManager()
	for i := 0; i < b.N; i++ {
		rm.AddActiveRide(makeRide("r", time.Duration(i)*time.Millisecond))
	}
}

func BenchmarkGetActiveRide(b *testing.B) {
	rm := NewRideManager()
	// Pre-fill with 1000 rides
	for i := 0; i < 1000; i++ {
		rm.AddActiveRide(makeRide("r"+string(rune(i)), time.Duration(i)*time.Millisecond))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.GetActiveRide("r" + string(rune(500))) // lookup in the middle
	}
}

func BenchmarkFinishRide(b *testing.B) {
	rm := NewRideManager()

	// Preload rides
	rides := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		rideID := fmt.Sprintf("ride-%d", i)
		rides[i] = rideID
		rm.AddActiveRide(makeRide(rideID, 0))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rm.FinishRide(rides[i], models.RideStatusCompleted)
	}
}
