package main

import "fmt"

type DefaultClassifier struct{}

func (c *DefaultClassifier) Classify(packet Packet) int {
	switch packet.Protocol {
	case ProtocolICMP:
		return 1
	case ProtocolTCP:
		return 2
	case ProtocolUDP:
		return 3
	default:
		return 5
	}
}

type Route struct {
	Queue        map[int]PacketQueue
	Classifier   PacketClassifier
	QueueFactory func() PacketQueue // Injects the type of queue we want to use
}

func NewRoute(classifier PacketClassifier, factory func() PacketQueue) *Route {
	return &Route{
		Queue:        make(map[int]PacketQueue),
		Classifier:   classifier,
		QueueFactory: factory,
	}
}

func (r *Route) RoutePacket(packet Packet) {
	priority := r.Classifier.Classify(packet)
	packet.Priority = priority // Ensure priority is set on the packet
	if _, exists := r.Queue[priority]; !exists {
		r.Queue[priority] = r.QueueFactory() // Use the injected factory
	}
	r.Queue[priority].Enqueue(packet)
}

func (r *Route) DequeuePacket() (Packet, error) {
	// Fixed: Now checks priorities 1 through 5
	for i := 1; i <= 5; i++ {
		if queue, exists := r.Queue[i]; exists && queue.Len() > 0 {
			return queue.Dequeue()
		}
	}
	return Packet{}, fmt.Errorf("no packets to dequeue")
}

func (r *Route) ReorderPackets(packet Packet, newPriority int) error {
	// Fixed: Search 1-5, drop from old queue, update priority, add to new queue
	for i := 1; i <= 5; i++ {
		if queue, exists := r.Queue[i]; exists {
			err := queue.Drop(packet.ID)
			if err == nil {
				packet.Priority = newPriority
				if _, ok := r.Queue[newPriority]; !ok {
					r.Queue[newPriority] = r.QueueFactory()
				}
				r.Queue[newPriority].Enqueue(packet)
				return nil
			}
		}
	}
	return fmt.Errorf("packet with ID %d not found", packet.ID)
}

func (r *Route) DropExpiredPackets() {
	droppedCount := 0
	for i := 1; i <= 5; i++ {
		if queue, exists := r.Queue[i]; exists {
			queueDroppedCount := queue.DropExpired()
			droppedCount += queueDroppedCount
			fmt.Printf("Priority %d: Dropped %d expired packets.\n", i, queueDroppedCount)
		}
	}
	fmt.Printf("Dropped %d expired packets.\n", droppedCount)
}

func (r *Route) DisplayQueueStatus() {
	for i := 1; i <= 5; i++ {
		if queue, exists := r.Queue[i]; exists {
			fmt.Printf("Priority %d: %d packets\n", i, queue.Len())
		}
	}
}
