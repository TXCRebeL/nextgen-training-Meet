package main

import (
	"fmt"
	"time"
)

var parkingGarage = [5][100]ParkingSpot{}

type ParkingSpot struct {
	Id           int
	Floor        int
	Occupied     bool
	VehiclePlate string
	EntryTime    time.Time
}

func main() {

	fmt.Println("Parking Garage initialized with 5 floors and 100 spots per floor.")

	fmt.Println("------------------------------------------------------")

	for {
		var userChoice int

		fmt.Println("1 for Entry --------")
		fmt.Println("2 for Exit ---------")
		fmt.Println("3 for Display ---------")
		fmt.Println("4 for Search ---------")
		fmt.Println("0 for close the application --------")
		fmt.Println("Enter your choice : ")
		fmt.Scan(&userChoice)
		switch userChoice {
		case 1:
			var floor int
			var plate string
			fmt.Print("Enter floor number (0-4): ")
			fmt.Scan(&floor)

			if floor < 0 || floor >= len(parkingGarage) {
				fmt.Println("Invalid floor number. Please try again.")
				continue
			}

			fmt.Print("Enter vehicle plate number: ")
			fmt.Scan(&plate)

			_, _, isAlreayParked := findVehicle(plate)

			if isAlreayParked {
				fmt.Printf("Vehicle with plate %s is already parked in the garage.\n", plate)
				continue
			}

			parkingSpot, spotFound := findParkingSpot(floor)
			if spotFound {
				parkingGarage[floor][parkingSpot] = ParkingSpot{
					Id:           parkingSpot + 1,
					Floor:        floor,
					Occupied:     true,
					VehiclePlate: plate,
					EntryTime:    time.Now(),
				}
				fmt.Println("Parking spot assigned successfully.")
				fmt.Printf("Vehicle with plate %s parked at floor %d, spot %d.\n", plate, floor, parkingSpot)
			} else {
				fmt.Printf("No available parking spots on floor %d.\n", floor)
			}

		case 2:
			var plate string
			fmt.Print("Enter vehicle plate number for exit: ")
			fmt.Scan(&plate)

			floor, spot, found := findVehicle(plate)
			if found {
				timeParked := time.Since(parkingGarage[floor][spot].EntryTime)
				hours := int(timeParked.Hours())
				minutes := int(timeParked.Minutes()) % 60

				billableHours := hours

				if minutes > 0 {
					billableHours++
				}

				if billableHours == 0 {
					billableHours = 1
				}
				parkingCharge := float64(billableHours) * 20
				fmt.Printf("Parking charge for vehicle with plate %s is Rs. %.2f for parking %d hours %d minutes.\n", plate, parkingCharge, hours, minutes)
				fmt.Printf("Vehicle with plate %s has exited from floor %d, spot %d.\n", plate, floor, spot)
				parkingGarage[floor][spot] = ParkingSpot{}
			} else {
				fmt.Printf("Vehicle with plate %s not found in the parking garage.\n", plate)
			}

		case 3:
			fmt.Println("Current Parking Garage Status:")
			for floor := 0; floor < len(parkingGarage); floor++ {
				fmt.Printf("Floor %d:\n", floor)
				var availableSpots int
				for spot := 0; spot < len(parkingGarage[floor]); spot++ {
					if !parkingGarage[floor][spot].Occupied {
						availableSpots++
					}
				}
				fmt.Printf("  Available spots: %d/%d\n", availableSpots, len(parkingGarage[floor]))
			}

		case 4:
			var plate string
			fmt.Print("Enter vehicle plate number to search: ")
			fmt.Scan(&plate)

			floor, spot, found := findVehicle(plate)
			if found {
				fmt.Printf("Vehicle with plate %s is parked at floor %d, spot %d.\n", plate, floor, spot)
			} else {
				fmt.Printf("Vehicle with plate %s not found in the parking garage.\n", plate)
			}

		case 0:
			fmt.Println("Closing the application. Goodbye!")
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func findParkingSpot(floor int) (int, bool) {
	for i := 0; i < len(parkingGarage[floor]); i++ {
		if !parkingGarage[floor][i].Occupied {
			return i, true
		}
	}
	return -1, false
}

func findVehicle(plate string) (int, int, bool) {
	for floor := 0; floor < len(parkingGarage); floor++ {
		for spot := 0; spot < len(parkingGarage[floor]); spot++ {
			if parkingGarage[floor][spot].Occupied && parkingGarage[floor][spot].VehiclePlate == plate {
				return floor, spot, true
			}
		}
	}
	return -1, -1, false
}
