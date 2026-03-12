package main

import "fmt"

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolICMP Protocol = "ICMP"
)

type Packet struct {
	ID        int
	SourceIP  string
	DestIP    string
	Protocol  Protocol
	Priority  int
	Payload   []byte
	Timestamp int
	TTL       int
}

func (p Packet) String() string {
	return fmt.Sprintf("Packet(ID: %d, SourceIP: %s, DestIP: %s, Protocol: %s, Priority: %d, TTL: %d)",
		p.ID, p.SourceIP, p.DestIP, p.Protocol, p.Priority, p.TTL)
}

// PacketQueue Interface - Notice how clean this makes our dependencies!
type PacketQueue interface {
	Enqueue(packet Packet)
	Dequeue() (Packet, error)
	Peek() (Packet, error)
	Len() int
	Drop(packetID int) error
	DropExpired() int
}

type PacketClassifier interface {
	Classify(packet Packet) int
}
