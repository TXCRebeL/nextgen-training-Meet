package models

import "sync"

type User struct {
	ID         string
	Name       string
	Balance    float64
	ActiveBids []string // List of BidIDs
	Mu         sync.Mutex // Fine-grained locking per user
}
