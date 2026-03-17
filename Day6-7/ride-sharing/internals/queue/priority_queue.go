package queue

import (
	"github.com/meetbha/ride-sharing/internals/models"
)

// RideQueue defines the interactions for prioritizing rides.
type RideQueue interface {
	AddRide(ride *models.Ride)
	RemoveRide(ride *models.Ride)
	GetNextRide() *models.Ride
	GetRideCount() int
}

// priorityQueue implements RideQueue using a custom min-heap array.
type priorityQueue struct {
	rides []*models.Ride
}

// NewPriorityQueue instantiates a new min-heap ride queue from scratch.
func NewPriorityQueue() RideQueue {
	return &priorityQueue{
		rides: make([]*models.Ride, 0),
	}
}

// bubbleUp moves the element at index i up to restore the heap property.
func (pq *priorityQueue) bubbleUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		// If current is older than parent, swap them
		if pq.rides[i].RequestedAt.Before(pq.rides[parent].RequestedAt) {
			pq.rides[i], pq.rides[parent] = pq.rides[parent], pq.rides[i]
			i = parent
		} else {
			break
		}
	}
}

// sinkDown moves the element at index i down to restore the heap property.
func (pq *priorityQueue) sinkDown(i int) {
	n := len(pq.rides)
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i

		if left < n && pq.rides[left].RequestedAt.Before(pq.rides[smallest].RequestedAt) {
			smallest = left
		}
		if right < n && pq.rides[right].RequestedAt.Before(pq.rides[smallest].RequestedAt) {
			smallest = right
		}
		if smallest != i {
			pq.rides[i], pq.rides[smallest] = pq.rides[smallest], pq.rides[i]
			i = smallest
		} else {
			break
		}
	}
}

// AddRide inserts a new ride and heapifies.
func (pq *priorityQueue) AddRide(ride *models.Ride) {
	pq.rides = append(pq.rides, ride)
	pq.bubbleUp(len(pq.rides) - 1)
}

// RemoveRide iterates over the heap to find the matching ride and removes it, heapifying from there.
func (pq *priorityQueue) RemoveRide(ride *models.Ride) {
	for i, r := range pq.rides {
		if r.ID == ride.ID {
			n := len(pq.rides) - 1
			// Swap the requested element with the last element
			pq.rides[i], pq.rides[n] = pq.rides[n], pq.rides[i]
			// Cut off the last element (which is the one we want deleted)
			pq.rides = pq.rides[:n]

			// Only sink/bubble if we didn't just delete the very last slot outright
			if i < n {
				// Re-balance the heap from index `i` (which is now the item swapped from the end)
				parent := (i - 1) / 2
				if i > 0 && pq.rides[i].RequestedAt.Before(pq.rides[parent].RequestedAt) {
					pq.bubbleUp(i)
				} else {
					pq.sinkDown(i)
				}
			}
			break
		}
	}
}

// GetNextRide pops the ride that has been waiting the longest.
func (pq *priorityQueue) GetNextRide() *models.Ride {
	if len(pq.rides) == 0 {
		return nil
	}

	n := len(pq.rides) - 1
	pq.rides[0], pq.rides[n] = pq.rides[n], pq.rides[0]
	oldest := pq.rides[n]
	pq.rides = pq.rides[:n]

	if len(pq.rides) > 0 {
		pq.sinkDown(0)
	}

	return oldest
}

// GetRideCount returns the number of rides waiting in the queue.
func (pq *priorityQueue) GetRideCount() int {
	return len(pq.rides)
}
