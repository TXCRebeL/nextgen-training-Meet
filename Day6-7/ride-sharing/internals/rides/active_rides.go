package rides

import "github.com/meetbha/ride-sharing/internals/models"

// RideNode represents a node in the custom Linked List holding an active ride.
type RideNode struct {
	Ride *models.Ride
	Next *RideNode
}

// ActiveRides manages in-transit rides via a singly-linked list for pure sequential operations.
type ActiveRides struct {
	Head *RideNode
	Tail *RideNode
}

// NewActiveRides initializes an empty active rides queue.
func NewActiveRides() *ActiveRides {
	return &ActiveRides{
		Head: nil,
		Tail: nil,
	}
}

// AddRide inserts a new ride node at the Tail of the linked list.
func (ar *ActiveRides) AddRide(ride *models.Ride) {
	newNode := &RideNode{
		Ride: ride,
		Next: nil,
	}

	if ar.Head == nil {
		ar.Head = newNode
		ar.Tail = newNode
		return
	}

	ar.Tail.Next = newNode
	ar.Tail = newNode
}

// RemoveRide explicitly searches for and deletes a ride node from the linked list.
func (ar *ActiveRides) RemoveRide(ride *models.Ride) {
	if ar.Head == nil {
		return
	}

	// Case 1: The ride to remove is the Head
	if ar.Head.Ride.ID == ride.ID {
		ar.Head = ar.Head.Next
		if ar.Head == nil {
			ar.Tail = nil // If list is now empty, reset Tail too
		}
		return
	}

	// Case 2: The ride to remove is somewhere in the body or at the Tail
	current := ar.Head
	for current.Next != nil {
		if current.Next.Ride.ID == ride.ID {
			// If we are removing the current Tail node, point the Tail backward
			if current.Next == ar.Tail {
				ar.Tail = current
			}

			// Bypass the deleted node
			current.Next = current.Next.Next
			return
		}
		current = current.Next
	}
}
