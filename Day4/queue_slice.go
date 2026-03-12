package main

import "fmt"

type SliceQueue struct {
	packets []Packet
}

func (q *SliceQueue) Enqueue(packet Packet) {
	q.packets = append(q.packets, packet)
}

func (q *SliceQueue) Dequeue() (Packet, error) {
	if len(q.packets) == 0 {
		return Packet{}, fmt.Errorf("queue is empty")
	}
	packet := q.packets[0]
	q.packets = q.packets[1:] // Shift slice forward
	return packet, nil
}

func (q *SliceQueue) Peek() (Packet, error) {
	if len(q.packets) == 0 {
		return Packet{}, fmt.Errorf("queue is empty")
	}
	return q.packets[0], nil
}

func (q *SliceQueue) Len() int {
	return len(q.packets)
}

func (q *SliceQueue) Drop(packetID int) error {
	for i, p := range q.packets {
		if p.ID == packetID {
			// Remove element at index i and shift the rest
			q.packets = append(q.packets[:i], q.packets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("packet with ID %d not found", packetID)
}

func (q *SliceQueue) DropExpired() int {
	droppedCount := 0
	kept := 0 // Tracks the index of where to place valid packets

	for _, p := range q.packets {
		if p.TTL <= 0 {
			// fmt.Printf("⚠️ Dropped Packet ID %d: TTL Expired\n", p.ID)
			droppedCount++
		} else {
			q.packets[kept] = p
			kept++
		}
	}

	q.packets = q.packets[:kept]
	return droppedCount
}
