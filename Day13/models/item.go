package models

import (
	"sync"
	"time"
)

type ItemStatus string

const (
	StatusActive    ItemStatus = "active"
	StatusEnded     ItemStatus = "ended"
	StatusCancelled ItemStatus = "cancelled"
)

type Item struct {
	ID          string
	Name        string
	Category    string
	Description string
	SellerID    string
	StartPrice  float64
	CurrentBid  float64
	BidHistory  []string // Array of Bid IDs or can be managed through the separate DS layer
	StartTime   time.Time
	EndTime     time.Time
	Status      ItemStatus
	Mu          sync.Mutex // Fine-grained locking per item
}

