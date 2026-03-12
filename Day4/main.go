package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Helper array and function from your original code to generate random traffic
var protocols = []Protocol{
	ProtocolTCP,
	ProtocolUDP,
	ProtocolICMP,
}

func RandomProtocol() Protocol {
	return protocols[rand.Intn(len(protocols))]
}

func main() {
	fmt.Println("=== Starting Network Router Simulation ===")

	// 1. Initialize the Classifier
	classifier := &DefaultClassifier{}

	// 2. Define our Queue Factory
	// This is the beauty of interfaces! Want to test the SliceQueue instead?
	// Just change `&LinkedListQueue{}` to `&SliceQueue{}` below.
	queueFactory := func() PacketQueue {
		return &LinkedListQueue{}
	}

	// 3. Create the Router
	route := NewRoute(classifier, queueFactory)

	// 4. Simulate 10,000 packets
	fmt.Println("\n[1] Generating and routing 10,000 packets...")
	for i := 0; i < 10000; i++ {
		packet := Packet{
			ID:        i,
			SourceIP:  fmt.Sprintf("192.168.1.%d", i%255),
			DestIP:    fmt.Sprintf("10.0.0.%d", i%255),
			Protocol:  RandomProtocol(),
			Priority:  0, // Will be set by the classifier
			Payload:   []byte("Hello, World!"),
			Timestamp: int(time.Now().Unix()),
			TTL:       rand.Intn(5), // Generates 0-4. ~20% chance of TTL being 0
		}
		route.RoutePacket(packet)
	}

	fmt.Println("\n--- Initial Queue Status ---")
	route.DisplayQueueStatus()

	// 5. Drop expired packets (TTL = 0)
	fmt.Println("\n[2] Dropping expired packets (TTL=0)...")
	route.DropExpiredPackets()

	fmt.Println("\n--- Queue Status After Drops ---")
	route.DisplayQueueStatus()

	// 6. Process (Dequeue) all remaining packets and benchmark the time
	fmt.Println("\n[3] Processing remaining packets...")
	var processedCount int

	// Start the stopwatch!
	startTime := time.Now()

	for {
		_, err := route.DequeuePacket()
		if err != nil {
			// An error here means all queues are completely empty
			break
		}
		processedCount++
	}

	// Stop the stopwatch!
	elapsedTime := time.Since(startTime)

	// 7. Print the final results
	fmt.Printf("\n=== Simulation Complete ===\n")
	fmt.Printf("Total packets successfully processed: %d\n", processedCount)
	fmt.Printf("Total processing time: %s\n", elapsedTime)
}
