package ds

import (
	"Day13/models"
	"sync"
)

// BidNode represents a single node in the bid history linked list
type BidNode struct {
	Bid  *models.Bid
	Next *BidNode
}

// BidHistoryList is a thread-safe linked list optimized for frequent appends
type BidHistoryList struct {
	mu   sync.RWMutex
	head *BidNode
	tail *BidNode
	size int
}

func NewBidHistoryList() *BidHistoryList {
	return &BidHistoryList{}
}

// Append adds a new bid to the end of the history chronologically
func (l *BidHistoryList) Append(bid *models.Bid) {
	l.mu.Lock()
	defer l.mu.Unlock()

	newNode := &BidNode{Bid: bid}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.Next = newNode
		l.tail = newNode
	}
	l.size++
}

// GetAll returns all active bids in chronological order
func (l *BidHistoryList) GetAll() []*models.Bid {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var bids []*models.Bid
	current := l.head
	for current != nil {
		bids = append(bids, current.Bid)
		current = current.Next
	}
	return bids
}

// Size returns the number of bids in the history
func (l *BidHistoryList) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.size
}
