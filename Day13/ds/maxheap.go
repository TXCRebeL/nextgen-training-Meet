package ds

import (
	"Day13/models"
)

// BidMaxHeap implements container/heap.Interface for *models.Bid.
// We want the maximum amount to be at the root, so Less returns true if amount i > amount j.
type BidMaxHeap []*models.Bid

func (h BidMaxHeap) Len() int { return len(h) }

// Less is inverted for Max-Heap
func (h BidMaxHeap) Less(i, j int) bool {
	return h[i].Amount > h[j].Amount
}

func (h BidMaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *BidMaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*models.Bid))
}

func (h *BidMaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*h = old[0 : n-1]
	return item
}
