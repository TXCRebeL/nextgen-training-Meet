package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/meetbha/ride-sharing/internals/dispatch"
	"github.com/meetbha/ride-sharing/internals/driver"
	"github.com/meetbha/ride-sharing/internals/models"
	"github.com/meetbha/ride-sharing/internals/queue"
	"github.com/meetbha/ride-sharing/internals/rides"
)

func genID(prefix string) string {
	return fmt.Sprintf("%s-%04d", prefix, rand.Intn(9000)+1000)
}

func randomRating() float64 {
	// Random rating between 3.5 and 5.0
	return 3.5 + rand.Float64()*1.5
}

func main() {
	zm := driver.NewZoneManager()
	dm := driver.NewDriverManager(zm)
	rq := queue.NewPriorityQueue()
	rm := rides.NewRideManager()
	d := dispatch.NewDispatcher(dm, rq, rm)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ride-Sharing CLI — type 'help' for commands")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		cmd, args := parts[0], parts[1:]

		switch cmd {

		case "help":
			fmt.Println(`Commands:
  register-driver  <name> <lat> <lng>          (id & rating auto-generated)
  get-driver       <id>
  edit-driver      <id> <name> <rating>
  update-location  <id> <lat> <lng>
  change-status    <id> available|busy|offline
  remove-driver    <id>
  driver-earnings  <id> today|week|all
  find-nearest     <lat> <lng>
  list-drivers

  zone-drivers     <lat> <lng>
  busy-zone

  request-ride     <pLat> <pLng> <dLat> <dLng> (riderID auto-generated)
  assign-driver
  complete-ride    <rideID>

  queue-count
  active-rides
  ride-history
  avg-wait-time

  exit`)

		// ── Drivers ──────────────────────────────────────────────────────

		case "register-driver":
			if len(args) != 3 {
				fmt.Println("Usage: register-driver <name> <lat> <lng>")
				continue
			}
			lat, _ := strconv.ParseFloat(args[1], 64)
			lng, _ := strconv.ParseFloat(args[2], 64)
			id := genID("drv")
			rating := randomRating()
			dm.RegisterDriver(&models.Driver{
				ID:       id,
				Name:     args[0],
				Location: models.Location{Lat: lat, Lng: lng},
				Status:   models.DriverStatusAvailable,
				Rating:   rating,
			})
			fmt.Printf("Driver registered: id=%s name=%s rating=%.1f zone=%s\n", id, args[0], rating, driver.CalculateZone(lat, lng))

		case "get-driver":
			if len(args) != 1 {
				fmt.Println("Usage: get-driver <id>")
				continue
			}
			drv, err := dm.GetDriver(args[0])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("ID=%s Name=%s Status=%s Loc=(%.2f,%.2f) Rating=%.1f Rides=%d\n",
				drv.ID, drv.Name, drv.Status, drv.Location.Lat, drv.Location.Lng, drv.Rating, len(drv.RideHistory))

		case "edit-driver":
			if len(args) != 3 {
				fmt.Println("Usage: edit-driver <id> <name> <rating>")
				continue
			}
			rating, _ := strconv.ParseFloat(args[2], 64)
			if err := dm.EditDriver(args[0], args[1], rating); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Driver %s updated\n", args[0])

		case "update-location":
			if len(args) != 3 {
				fmt.Println("Usage: update-location <id> <lat> <lng>")
				continue
			}
			lat, _ := strconv.ParseFloat(args[1], 64)
			lng, _ := strconv.ParseFloat(args[2], 64)
			if err := dm.UpdateLocation(args[0], lat, lng); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Driver %s moved to (%.2f,%.2f) zone=%s\n", args[0], lat, lng, driver.CalculateZone(lat, lng))

		case "change-status":
			if len(args) != 2 {
				fmt.Println("Usage: change-status <id> available|busy|offline")
				continue
			}
			statusMap := map[string]models.DriverStatus{
				"available": models.DriverStatusAvailable,
				"busy":      models.DriverStatusBusy,
				"offline":   models.DriverStatusOffline,
			}
			status, ok := statusMap[args[1]]
			if !ok {
				fmt.Println("Unknown status. Use: available | busy | offline")
				continue
			}
			if err := dm.ChangeStatus(args[0], status); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Driver %s → %s\n", args[0], args[1])

		case "remove-driver":
			if len(args) != 1 {
				fmt.Println("Usage: remove-driver <id>")
				continue
			}
			if err := dm.RemoveDriver(args[0]); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Driver %s removed\n", args[0])

		case "driver-earnings":
			if len(args) != 2 {
				fmt.Println("Usage: driver-earnings <id> today|week|all")
				continue
			}
			e, err := dm.GetDriverEarnings(args[0], args[1])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Earnings [%s] for %s: %.2f\n", args[1], args[0], e)

		case "find-nearest":
			if len(args) != 2 {
				fmt.Println("Usage: find-nearest <lat> <lng>")
				continue
			}
			lat, _ := strconv.ParseFloat(args[0], 64)
			lng, _ := strconv.ParseFloat(args[1], 64)
			drv, err := dm.FindNearestDriver(lat, lng)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Nearest: ID=%s Name=%s Loc=(%.2f,%.2f)\n",
				drv.ID, drv.Name, drv.Location.Lat, drv.Location.Lng)

		case "list-drivers":
			if len(dm.Drivers) == 0 {
				fmt.Println("No drivers registered")
				continue
			}
			for _, drv := range dm.Drivers {
				fmt.Printf("  %s | %s | %s | (%.2f,%.2f) | rating=%.1f\n",
					drv.ID, drv.Name, drv.Status, drv.Location.Lat, drv.Location.Lng, drv.Rating)
			}

		// ── Zones ────────────────────────────────────────────────────────

		case "zone-drivers":
			if len(args) != 2 {
				fmt.Println("Usage: zone-drivers <lat> <lng>")
				continue
			}
			lat, _ := strconv.ParseFloat(args[0], 64)
			lng, _ := strconv.ParseFloat(args[1], 64)
			zone := driver.CalculateZone(lat, lng)
			drvs := zm.GetDrivers(zone)
			fmt.Printf("Zone %s: %d driver(s)\n", zone, len(drvs))
			for _, drv := range drvs {
				fmt.Printf("  %s (%s)\n", drv.ID, drv.Status)
			}

		case "busy-zone":
			z := zm.GetBusyZone()
			if z == "" {
				fmt.Println("No zone activity yet")
				continue
			}
			fmt.Printf("Busiest zone: %s (%d rides)\n", z, zm.GetRideCount(z))

		// ── Dispatch ─────────────────────────────────────────────────────

		case "request-ride":
			if len(args) != 4 {
				fmt.Println("Usage: request-ride <pLat> <pLng> <dLat> <dLng>")
				continue
			}
			pLat, _ := strconv.ParseFloat(args[0], 64)
			pLng, _ := strconv.ParseFloat(args[1], 64)
			dLat, _ := strconv.ParseFloat(args[2], 64)
			dLng, _ := strconv.ParseFloat(args[3], 64)
			riderID := genID("rider")
			rideID := d.RequestRide(riderID,
				models.Location{Lat: pLat, Lng: pLng},
				models.Location{Lat: dLat, Lng: dLng},
			)
			fmt.Printf("Ride requested: rideID=%s riderID=%s (queue=%d)\n", rideID, riderID, rq.GetRideCount())

		case "assign-driver":
			drvID, err := d.AssignDriver()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Assigned driver: %s\n", drvID)

		case "complete-ride":
			if len(args) != 1 {
				fmt.Println("Usage: complete-ride <rideID>")
				continue
			}
			fare, err := d.CompleteRide(args[0])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			fmt.Printf("Ride %s completed. Fare: %.2f\n", args[0], fare)

		// ── Ride Info ────────────────────────────────────────────────────

		case "queue-count":
			fmt.Printf("Queue: %d ride(s) waiting\n", rq.GetRideCount())

		case "active-rides":
			cur := rm.ActiveRides.Head
			if cur == nil {
				fmt.Println("No active rides")
				continue
			}
			for cur != nil {
				r := cur.Ride
				fmt.Printf("  %s | rider=%s driver=%s status=%s\n", r.ID, r.RiderID, r.DriverID, r.Status)
				cur = cur.Next
			}

		case "ride-history":
			history := rm.GetRideHistory()
			if len(history) == 0 {
				fmt.Println("No ride history")
				continue
			}
			for _, r := range history {
				fmt.Printf("  %s | rider=%s driver=%s status=%s fare=%.2f\n",
					r.ID, r.RiderID, r.DriverID, r.Status, r.Fare)
			}

		case "avg-wait-time":
			w := d.GetAverageWaitTime()
			if w == 0 {
				fmt.Println("No fulfilled rides yet")
			} else {
				fmt.Printf("Avg wait time: %v\n", w)
			}

		case "exit", "quit":
			fmt.Println("Bye!")
			os.Exit(0)

		default:
			fmt.Printf("Unknown command %q — type 'help'\n", cmd)
		}
	}
}
