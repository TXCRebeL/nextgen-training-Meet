package main

import "fmt"

type Node struct {
	Packet Packet
	Next   *Node
	Prev   *Node
}

type LinkedListQueue struct {
	Head *Node
	Tail *Node
	size int // Added for O(1) length lookups
}

func (q *LinkedListQueue) Enqueue(packet Packet) {
	newNode := &Node{Packet: packet}
	if q.Tail == nil {
		q.Head = newNode
		q.Tail = newNode
	} else {
		q.Tail.Next = newNode
		newNode.Prev = q.Tail
		q.Tail = newNode
	}
	q.size++
}

func (q *LinkedListQueue) Dequeue() (Packet, error) {
	if q.Head == nil {
		return Packet{}, fmt.Errorf("queue is empty")
	}
	packet := q.Head.Packet
	q.Head = q.Head.Next
	if q.Head != nil {
		q.Head.Prev = nil
	} else {
		q.Tail = nil
	}
	q.size--
	return packet, nil
}

func (q *LinkedListQueue) Peek() (Packet, error) {
	if q.Head == nil {
		return Packet{}, fmt.Errorf("queue is empty")
	}
	return q.Head.Packet, nil
}

func (q *LinkedListQueue) Len() int {
	return q.size // Now O(1) time!
}

func (q *LinkedListQueue) Drop(packetID int) error {
	current := q.Head
	for current != nil {
		if current.Packet.ID == packetID {
			if current.Prev != nil {
				current.Prev.Next = current.Next
			} else {
				q.Head = current.Next
			}
			if current.Next != nil {
				current.Next.Prev = current.Prev
			} else {
				q.Tail = current.Prev
			}
			q.size--
			return nil
		}
		current = current.Next
	}
	return fmt.Errorf("packet with ID %d not found", packetID)
}

func (q *LinkedListQueue) DropExpired() int {
	droppedCount := 0
	current := q.Head

	for current != nil {
		next := current.Next
		if current.Packet.TTL <= 0 {
			//fmt.Printf("⚠️ Dropped Packet ID %d: TTL Expired\n", current.Packet.ID)

			if current.Prev != nil {
				current.Prev.Next = current.Next
			} else {
				q.Head = current.Next
			}
			if current.Next != nil {
				current.Next.Prev = current.Prev
			} else {
				q.Tail = current.Prev
			}

			q.size--
			droppedCount++
		}
		current = next
	}
	return droppedCount
}
