package models

import "time"

type Driver struct {
	ID          string
	Name        string
	Location    Location
	Status      DriverStatus
	Rating      float64
	RideHistory []Ride
}

type Location struct {
	Lat float64
	Lng float64
}

type DriverStatus string

const (
	DriverStatusAvailable DriverStatus = "available"
	DriverStatusBusy      DriverStatus = "busy"
	DriverStatusOffline   DriverStatus = "offline"
)

type Rider struct {
	ID          string
	Name        string
	Location    Location
	PaymentType PaymentType
}

type PaymentType string

const (
	PaymentTypeCash PaymentType = "cash"
	PaymentTypeUPI  PaymentType = "upi"
	PaymentTypeCard PaymentType = "card"
)

type RideStatus string

const (
	RideStatusPending   RideStatus = "pending"
	RideStatusAccepted  RideStatus = "accepted"
	RideStatusCompleted RideStatus = "completed"
	RideStatusCancelled RideStatus = "cancelled"
)

type Ride struct {
	ID             string
	RiderID        string
	DriverID       string
	PickupLocation Location
	DropLocation   Location
	Status         RideStatus
	Fare           float64
	RequestedAt    time.Time
	StartedAt      time.Time
	CompletedAt    time.Time
}
