package main

import (
	"fmt"
	"math/rand"
	"time"
)

type PricePoint struct {
	Price     float64
	Timestamp time.Time
	Volume    int
}

type Stock struct {
	Symbol      string
	CompanyName string
	PricePoints []PricePoint
	WindowSize  int
	StartIndex  int
	Size        int
}

func newStock(symbol, companyName string, windowSize int) *Stock {

	if windowSize <= 0 {
		windowSize = 10
	}

	return &Stock{
		Symbol:      symbol,
		CompanyName: companyName,
		PricePoints: make([]PricePoint, 0, 1),
		WindowSize:  windowSize,
		StartIndex:  0,
	}
}

func (s *Stock) AddPricePoint(pricePoint PricePoint) {

	oldCapacity := cap(s.PricePoints)

	if s.Size < s.WindowSize {
		s.PricePoints = append(s.PricePoints, pricePoint)
		s.Size++
	} else {
		s.PricePoints[s.StartIndex] = pricePoint
		s.StartIndex = (s.StartIndex + 1) % s.WindowSize
	}

	if cap(s.PricePoints) != oldCapacity {
		fmt.Printf("Resized PricePoints slice for %s from capacity %d to %d\n", s.Symbol, oldCapacity, cap(s.PricePoints))
	}
}

func (s *Stock) SMA() float64 {
	if s.Size == 0 {
		return 0
	}
	sum := 0.0
	for i := 0; i < s.Size; i++ {

		index := (s.StartIndex + i) % s.WindowSize
		sum += s.PricePoints[index].Price
	}
	return sum / float64(s.Size)
}

func (s *Stock) minMax() (float64, float64) {
	if s.Size == 0 {
		return 0.0, 0.0
	}
	firstIndex := s.StartIndex
	min := s.PricePoints[firstIndex].Price
	max := min

	for i := 0; i < s.Size; i++ {

		index := (s.StartIndex + i) % s.WindowSize
		price := s.PricePoints[index].Price

		if price < min {
			min = price
		}

		if price > max {
			max = price
		}
	}
	return max, min
}

func (s Stock) String() string {

	if s.Size == 0 {
		return fmt.Sprintf("%s (%s) - No data", s.CompanyName, s.Symbol)
	}
	max, min := s.minMax()
	return fmt.Sprintf("%s (%s) - Current Price: %.2f, SMA: %.2f, Min: %.2f, Max: %.2f",
		s.CompanyName, s.Symbol, s.PricePoints[(s.StartIndex-1+s.Size)%s.WindowSize].Price, s.SMA(), min, max)
}

var stocksList = [][]string{
	{"RELIANCE", "Reliance Industries"},
	{"TCS", "Tata Consultancy Services"},
	{"INFY", "Infosys"},
	{"HDFCBANK", "HDFC Bank"},
	{"ITC", "ITC"},
}

type stockMarket struct {
	stocks []*Stock
}

func newStockMarket() *stockMarket {
	stocks := make([]*Stock, 0, len(stocksList))
	for _, company := range stocksList {
		stocks = append(stocks, newStock(company[0], company[1], 10))
	}
	return &stockMarket{stocks: stocks}
}

func (sm *stockMarket) simulatePriceUpdates() {
	ticker := time.NewTicker(1000 * time.Millisecond)

	for range ticker.C {
		for i := range sm.stocks {

			stock := sm.stocks[i]

			var price float64

			if stock.Size == 0 {
				price = rand.Float64()*1000 + 500
			} else {
				price = stock.PricePoints[(stock.StartIndex-1+stock.Size)%stock.WindowSize].Price + rand.Float64()*4 - 2
			}
			pricePoint := PricePoint{
				Price:     price,
				Timestamp: time.Now(),
				Volume:    rand.Intn(1000),
			}
			stock.AddPricePoint(pricePoint)

		}

		fmt.Print("\033[H\033[2J")

		fmt.Println("--------------------------------------------------")
		fmt.Println("📈 LIVE STOCK DASHBOARD")

		for _, stock := range sm.stocks {
			fmt.Println(stock)
		}
		fmt.Println("--------------------------------------------------")

	}
}

func main() {
	stockMarket := newStockMarket()
	go stockMarket.simulatePriceUpdates()

	select {}
}
