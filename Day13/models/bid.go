package models

import "time"

type Bid struct {
	ID        string
	ItemID    string
	UserID    string
	Amount    float64
	Timestamp time.Time
}
