package main

import (
	"fmt"
	"testing"
)

// Helper function to generate mock packets
func generatePackets(count int) []Packet {
	packets := make([]Packet, count)
	for i := 0; i < count; i++ {
		packets[i] = Packet{
			ID:       i,
			Protocol: ProtocolTCP,
			TTL:      10,
		}
	}
	return packets
}

// benchmarkQueue is a generic runner that tests any PacketQueue implementation
func benchmarkQueue(b *testing.B, queue PacketQueue, packetCount int) {
	packets := generatePackets(packetCount)

	b.ResetTimer() // Don't count the setup time in the benchmark

	for i := 0; i < b.N; i++ {
		// Test Enqueue
		for _, p := range packets {
			queue.Enqueue(p)
		}
		// Test Dequeue
		for j := 0; j < packetCount; j++ {
			queue.Dequeue()
		}
	}
}

// 1. Benchmark the Linked List
func BenchmarkLinkedListQueue(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Packets-%d", size), func(b *testing.B) {
			queue := &LinkedListQueue{}
			benchmarkQueue(b, queue, size)
		})
	}
}

// 2. Benchmark the Slice
func BenchmarkSliceQueue(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Packets-%d", size), func(b *testing.B) {
			queue := &SliceQueue{}
			benchmarkQueue(b, queue, size)
		})
	}
}
