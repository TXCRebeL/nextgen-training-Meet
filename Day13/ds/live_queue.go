package ds

import (
	"Day13/models"
)

// LiveUpdateQueue handles broadcasting bid updates to active watchers
type LiveUpdateQueue struct {
	updates chan *models.Bid
}

func NewLiveUpdateQueue(bufferSize int) *LiveUpdateQueue {
	return &LiveUpdateQueue{
		updates: make(chan *models.Bid, bufferSize),
	}
}

// Publish sends a new bid update into the queue
func (q *LiveUpdateQueue) Publish(bid *models.Bid) {
	q.updates <- bid
}

// GetUpdatesChannel returns the read-only channel for watchers
func (q *LiveUpdateQueue) GetUpdatesChannel() <-chan *models.Bid {
	return q.updates
}

// Close gracefully closes the broadcast channel
func (q *LiveUpdateQueue) Close() {
	close(q.updates)
}
