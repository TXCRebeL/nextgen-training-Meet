package ds

import (
	"Day13/models"
	"errors"
	"sync"
)

var ErrEmptyStack = errors.New("stack is empty")

// UndoStack is a thread-safe LIFO stack for retracting recent bids per user
type UndoStack struct {
	mu   sync.Mutex
	bids []*models.Bid
}

func NewUndoStack() *UndoStack {
	return &UndoStack{
		bids: make([]*models.Bid, 0),
	}
}

// Push adds a bid to the top of the stack
func (s *UndoStack) Push(bid *models.Bid) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bids = append(s.bids, bid)
}

// Pop removes and returns the most recent bid
func (s *UndoStack) Pop() (*models.Bid, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.bids) == 0 {
		return nil, ErrEmptyStack
	}

	lastIndex := len(s.bids) - 1
	bid := s.bids[lastIndex]
	
	// Avoid memory leak by nil-ing out the pointer
	s.bids[lastIndex] = nil 
	s.bids = s.bids[:lastIndex]
	
	return bid, nil
}

// Peek returns the most recent bid without removing it
func (s *UndoStack) Peek() (*models.Bid, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.bids) == 0 {
		return nil, ErrEmptyStack
	}

	return s.bids[len(s.bids)-1], nil
}

func (s *UndoStack) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.bids)
}
