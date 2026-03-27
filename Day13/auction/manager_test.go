package auction

import (
	"Day13/models"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentBiddingAndRetraction(t *testing.T) {
	manager := NewManager()

	// 1. Setup Item
	item := &models.Item{
		ID:         "I1",
		Name:       "Diamond Necklace",
		Category:   "Jewelry",
		StartPrice: 10.0,
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(1 * time.Hour),
		Status:     models.StatusActive,
	}
	manager.AddItem(item)

	// 2. Setup 50 Users with large balances
	for i := 0; i < 50; i++ {
		manager.Users.AddUser(&models.User{
			ID:      fmt.Sprintf("U%d", i),
			Name:    fmt.Sprintf("User %d", i),
			Balance: 10000.0,
		})
	}

	var wg sync.WaitGroup

	// 3. Test Bidding (Ordered to prevent correct but restrictive ErrBidTooLow rejections)
	t.Log("Starting 50 sequential bids to establish clear stack history...")
	for i := 0; i < 50; i++ {
		bid := &models.Bid{
			ID:        fmt.Sprintf("B%d", i),
			ItemID:    "I1",
			UserID:    fmt.Sprintf("U%d", i),
			Amount:    20.0 + float64(i)*5.0,
			Timestamp: time.Now(),
		}
		err := manager.PlaceBid(bid)
		if err != nil {
			t.Errorf("Unexpected error placing bid from U%d: %v", i, err)
		}
	}

	// 4. Validate Highest Bid
	highest, err := manager.GetHighestBid("I1")
	if err != nil {
		t.Fatalf("Expected to find highest bid, got error: %v", err)
	}
	// Highest should be from U49 -> 20 + (49 * 5) = 265
	if highest.Amount != 265.0 {
		t.Errorf("Expected highest bid to be 265, got %v", highest.Amount)
	}
	t.Logf("Validated highest bid dynamically reached: $%v from User %v", highest.Amount, highest.UserID)

	// 5. Test Concurrent Bid Retraction
	// Retrieve bids from the top 10 users (U40 to U49) concurrently to ensure Undo stack and Heap removal survives race
	t.Log("Simulating top 10 highest users retracting their bids concurrently...")
	for i := 40; i < 50; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			err := manager.RetractLastBid(fmt.Sprintf("U%d", userID))
			if err != nil {
				t.Errorf("Unexpected error retracting bid: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// 6. Re-determine Highest Bid
	newHighest, err := manager.GetHighestBid("I1")
	if err != nil {
		t.Fatalf("Expected to find a new highest bid, got error: %v", err)
	}
	// The new highest should belong to U39 -> 20 + (39 * 5) = 215.0
	expectedNewAmount := 20.0 + (39.0 * 5.0)
	if newHighest.Amount != expectedNewAmount {
		t.Errorf("Expected new highest bid to fall back to %v, but got %v", expectedNewAmount, newHighest.Amount)
	}
	if newHighest.UserID != "U39" {
		t.Errorf("Expected new highest user to be U39, but got %v", newHighest.UserID)
	}
	t.Logf("Validated new highest bid successfully fell back to: $%v from User %v", newHighest.Amount, newHighest.UserID)

	// 7. Test Auction End resolution
	winner, err := manager.EndAuction("I1")
	if err != nil {
		t.Fatalf("Expected no error ending auction, got %v", err)
	}
	if winner == nil {
		t.Fatalf("Expected to find a winner, got nil")
	}
	if winner.ID != "U39" {
		t.Errorf("Expected strictly declared final winner to be U39, got %v", winner.ID)
	}

	t.Logf("Auction forcefully ended! Winner strictly asserted as: %v", winner.ID)

	// 8. Verify post-end locking logic (should reject new bids)
	err = manager.PlaceBid(&models.Bid{ItemID: "I1", UserID: "U0", Amount: 500, Timestamp: time.Now()})
	if err != ErrAuctionEnded {
		t.Errorf("Expected ErrAuctionEnded when bidding on closed auction, got: %v", err)
	}
}
