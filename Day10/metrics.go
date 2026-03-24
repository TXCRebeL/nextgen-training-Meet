package main

import (
	"fmt"
	"sync"
	"time"
)

type MetricsCollector struct {
	mu              sync.Mutex
	startTime       time.Time
	totalTasks      int
	completedTasks  int
	waitTimes       map[int][]time.Duration // priority -> wait times
	contextSwitches int
	starvationCount int
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
		waitTimes: make(map[int][]time.Duration),
	}
}

func (m *MetricsCollector) RecordCompletion(t *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.completedTasks++
	m.waitTimes[t.Priority] = append(m.waitTimes[t.Priority], t.WaitTime)
	if t.WaitTime > 5*time.Second {
		m.starvationCount++
	}
}

func (m *MetricsCollector) RecordContextSwitch() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contextSwitches++
}

func (m *MetricsCollector) RecordNewTask() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalTasks++
}

func (m *MetricsCollector) GetRelativeTime() string {
	elapsed := time.Since(m.startTime).Seconds()
	return fmt.Sprintf("[T=%.2fs]", elapsed)
}

func (m *MetricsCollector) PrintSummary() {
	fmt.Println("\n" + "================ SIMULATION SUMMARY ================")
	elapsed := time.Since(m.startTime).Seconds()
	fmt.Printf("Total Simulation Time: %.2fs\n", elapsed)
	fmt.Printf("Total Tasks Produced:  %d\n", m.totalTasks)
	fmt.Printf("Total Tasks Completed: %d\n", m.completedTasks)
	fmt.Printf("Throughput:            %.2f tasks/sec\n", float64(m.completedTasks)/elapsed)
	fmt.Printf("Context Switches:      %d\n", m.contextSwitches)
	fmt.Printf("Starvation Count (>5s): %d\n", m.starvationCount)

	fmt.Println("\nAverage Wait Time per Priority Level:")
	for p := 1; p <= 10; p++ {
		times, ok := m.waitTimes[p]
		if !ok || len(times) == 0 {
			continue
		}
		var total time.Duration
		for _, dur := range times {
			total += dur
		}
		avg := total / time.Duration(len(times))
		fmt.Printf("  Priority %2d: %v (count: %d)\n", p, avg, len(times))
	}
	fmt.Println("====================================================")
}
