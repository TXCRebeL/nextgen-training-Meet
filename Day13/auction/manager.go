package auction

import (
	"container/heap"
	"Day13/ds"
	"Day13/models"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrItemNotFound        = errors.New("item not found")
	ErrBidTooLow           = errors.New("bid amount must be higher than current bid and start price")
	ErrAuctionEnded        = errors.New("auction for this item has ended")
	ErrSellerCannotBid     = errors.New("seller cannot bid on their own item")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrUserNotFound        = errors.New("user not found")
	ErrNoWinner            = errors.New("no winner found")
)

// Manager orchestrates all core operations of the real-time auction platform
type Manager struct {
	Users      *ds.UserRegistry
	Categories *ds.CategoryTree
	Catalog    *ds.ItemCatalog

	mu           sync.RWMutex
	Items        map[string]*models.Item
	ActiveBids   map[string]*ds.BidMaxHeap     // ItemID -> MaxHeap of Bids
	BidHistories map[string]*ds.BidHistoryList // ItemID -> Linked List
	UndoStacks   map[string]*ds.UndoStack      // UserID -> LIFO Stack

	LiveUpdates *ds.LiveUpdateQueue
}

func NewManager() *Manager {
	return &Manager{
		Users:        ds.NewUserRegistry(),
		Categories:   ds.NewCategoryTree(),
		Catalog:      ds.NewItemCatalog(),
		Items:        make(map[string]*models.Item),
		ActiveBids:   make(map[string]*ds.BidMaxHeap),
		BidHistories: make(map[string]*ds.BidHistoryList),
		UndoStacks:   make(map[string]*ds.UndoStack),
		LiveUpdates:  ds.NewLiveUpdateQueue(100), // Buffer of 100
	}
}

// AddItem registers a new item
func (m *Manager) AddItem(item *models.Item) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Items[item.ID] = item
	h := &ds.BidMaxHeap{}
	heap.Init(h)
	m.ActiveBids[item.ID] = h
	m.BidHistories[item.ID] = ds.NewBidHistoryList()
	m.Catalog.AddItem(item)
}

// PlaceBid attempts to place a new bid on an item
func (m *Manager) PlaceBid(bid *models.Bid) error {
	m.mu.RLock()
	item, exists := m.Items[bid.ItemID]
	h := m.ActiveBids[bid.ItemID]
	history := m.BidHistories[bid.ItemID]
	m.mu.RUnlock()

	if !exists {
		return ErrItemNotFound
	}

	user, err := m.Users.GetUser(bid.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// 1. Lock User for balance check and deduction
	user.Mu.Lock()
	if user.Balance < bid.Amount {
		user.Mu.Unlock()
		return ErrInsufficientBalance
	}

	// 2. Lock Item for bid validation
	item.Mu.Lock()

	if time.Now().After(item.EndTime) || item.Status != models.StatusActive {
		item.Mu.Unlock()
		user.Mu.Unlock()
		return ErrAuctionEnded
	}
	if bid.Amount <= item.CurrentBid || bid.Amount < item.StartPrice {
		item.Mu.Unlock()
		user.Mu.Unlock()
		fmt.Printf("user %s amount %f and start price %f current bid %f error %v\n", bid.UserID, bid.Amount, item.StartPrice, item.CurrentBid, ErrBidTooLow)
		return ErrBidTooLow
	}
	if bid.UserID == item.SellerID {
		item.Mu.Unlock()
		user.Mu.Unlock()
		return ErrSellerCannotBid
	}

	// Deduct balance and update bid
	user.Balance -= bid.Amount
	user.Mu.Unlock()

	item.CurrentBid = bid.Amount
	heap.Push(h, bid)
	item.Mu.Unlock()

	// Append to Bid History (Linked List is thread-safe internally)
	history.Append(bid)

	// Add to User's Undo Stack
	m.mu.Lock()
	if _, exists := m.UndoStacks[bid.UserID]; !exists {
		m.UndoStacks[bid.UserID] = ds.NewUndoStack()
	}
	stack := m.UndoStacks[bid.UserID]
	m.mu.Unlock()

	stack.Push(bid)

	// Publish to Live Update Queue
	m.LiveUpdates.Publish(bid)

	return nil
}

// RetractLastBid retracts the user's most recent bid and restores balance + previous bid
func (m *Manager) RetractLastBid(userID string) error {
	m.mu.RLock()
	stack, exists := m.UndoStacks[userID]
	m.mu.RUnlock()

	if !exists {
		return ds.ErrEmptyStack
	}

	bid, err := stack.Pop()
	if err != nil {
		return err
	}

	m.mu.RLock()
	item, itemExists := m.Items[bid.ItemID]
	h := m.ActiveBids[bid.ItemID]
	m.mu.RUnlock()

	if !itemExists {
		return ErrItemNotFound
	}

	user, err := m.Users.GetUser(bid.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	user.Mu.Lock()
	item.Mu.Lock()

	// Remove the specific bid from the MaxHeap in-place using O(log N) heap.Remove
	for i, b := range *h {
		if b.ID == bid.ID {
			heap.Remove(h, i)
			break
		}
	}

	// Restore Item's Current Bid
	if h.Len() > 0 {
		highest := (*h)[0]
		item.CurrentBid = highest.Amount
	} else {
		item.CurrentBid = item.StartPrice
	}

	// Refund User Balance
	user.Balance += bid.Amount

	item.Mu.Unlock()
	user.Mu.Unlock()

	return nil
}

// EndAuction extracts the max bid from the heap and returns the winner
func (m *Manager) EndAuction(itemID string) (*models.User, error) {
	m.mu.RLock()
	item, exists := m.Items[itemID]
	h := m.ActiveBids[itemID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrItemNotFound
	}

	item.Mu.Lock()
	defer item.Mu.Unlock()

	if item.Status == models.StatusEnded {
		return nil, ErrAuctionEnded
	}

	item.Status = models.StatusEnded

	if h.Len() == 0 {
		return nil, ErrNoWinner // No winner
	}

	highest := (*h)[0]
	winner, err := m.Users.GetUser(highest.UserID)
	if err != nil {
		return nil, err
	}

	return winner, nil
}

// GetHighestBid returns the current highest bid from the MaxHeap in O(1)
func (m *Manager) GetHighestBid(itemID string) (*models.Bid, error) {
	m.mu.RLock()
	item, exists := m.Items[itemID]
	h := m.ActiveBids[itemID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrItemNotFound
	}

	item.Mu.Lock()
	defer item.Mu.Unlock()

	if h.Len() == 0 {
		return nil, ErrItemNotFound
	}
	return (*h)[0], nil
}
