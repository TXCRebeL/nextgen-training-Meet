package main

import (
	"Day13/auction"
	"Day13/models"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func main() {
	fmt.Println("Real-Time Auction Platform (Part 1)")

	manager := auction.NewManager()

	// 1. Add Users
	for i := range 100 {
		u1 := &models.User{
			ID:      "U" + strconv.Itoa(i),
			Name:    "Alice" + strconv.Itoa(i),
			Balance: rand.ExpFloat64() * 1000,
		}
		manager.Users.AddUser(u1)
	}
	fmt.Printf("Users Added: %d\n", manager.Users.Count())

	// 2. Add Category Navigation
	manager.Categories.AddCategoryPath("Electronics/Phones/Smartphones")
	manager.Categories.AddCategoryPath("Electronics/Phones/KeypadPhone")
	manager.Categories.AddCategoryPath("Electronics/Phones/Telephone")

	subs, _ := manager.Categories.GetSubcategories("Electronics/Phones")
	fmt.Println("Subcategories of Phones:", subs)

	for i := range 100 {
		// 3. Add Item
		cat := "Electronics/Phones/Smartphones"
		switch i % 3 {
		case 0:
			cat = "Electronics/Phones/KeypadPhone"
		case 1:
			cat = "Electronics/Phones/Telephone"
		}
		item := &models.Item{
			ID:         "I" + strconv.Itoa(i),
			Name:       "iPhone " + strconv.Itoa(i),
			Category:   cat,
			StartPrice: rand.Float64() * 500,
			StartTime:  time.Now(),
			EndTime:    time.Now().Add(24 * time.Hour),
			Status:     models.StatusActive,
		}
		manager.AddItem(item)
		fmt.Println("Item Added:", item.Name, "Starting at $", item.StartPrice)
	}

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(bidIndex int) {
			defer wg.Done()
			bid := &models.Bid{
				ID:        "B" + strconv.Itoa(bidIndex),
				ItemID:    "I" + strconv.Itoa(rand.Intn(100)),
				UserID:    "U" + strconv.Itoa(rand.Intn(100)),
				Amount:    rand.Float64() * 1000,
				Timestamp: time.Now(),
			}
			fmt.Println("Bid : ", bid.ID, " placed by", bid.UserID, "for $", bid.Amount, " for product ", bid.ItemID)
			err := manager.PlaceBid(bid)
			if err != nil {
				fmt.Println("Error placing bid:", err)
			} else {
				fmt.Println("Bid placed by", bid.UserID, "for $", bid.Amount, " for product ", bid.ItemID)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			user := rand.Intn(100)
			err := manager.RetractLastBid("U" + strconv.Itoa(user))
			if err != nil {
				fmt.Println("Error retracting bid:", err)
			} else {
				fmt.Println("Bid retracted by", "U"+strconv.Itoa(user))
			}
		}(i)
	}
	wg.Wait()

	//get the winner of all items
	for i := range 100 {
		winner, err := manager.EndAuction("I" + strconv.Itoa(i))
		if err != nil {
			fmt.Println("Error ending auction:", err)
		} else if winner != nil {
			fmt.Println("Auction ended! Winner:", winner.Name, "with remaining balance $", winner.Balance)
		}
	}
}
