package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AgingScheduler struct {
	readyQueue *MinHeap[*Task]
}

func NewAgingScheduler(ctx context.Context) *AgingScheduler {
	s := &AgingScheduler{
		readyQueue: NewMinHeap[*Task](func(a, b *Task) bool {
			ap := a.GetPriority()
			bp := b.GetPriority()
			if ap != bp {
				return ap < bp
			}
			return a.ArrivalTime.Before(b.ArrivalTime)
		}),
	}

	// Start aging goroutine
	go s.startAging(ctx)
	return s
}

func (s *AgingScheduler) Add(t *Task) {
	s.readyQueue.Insert(t)
}

func (s *AgingScheduler) Next() *Task {
	task, err := s.readyQueue.ExtractMin()
	if err != nil {
		return nil
	}
	return task
}

func (s *AgingScheduler) ScheduleTask(ctx context.Context, wg *sync.WaitGroup, metrics *MetricsCollector) {
	defer wg.Done()
	for {
		task := s.Next()
		if task != nil {
			metrics.RecordContextSwitch()
			fmt.Printf("%s [Aging] PID=%d (P=%d) START\n", metrics.GetRelativeTime(), task.PID, task.GetPriority())
			
			task.SetStatus(StatusRunning)
			task.SetWaitTime(time.Since(task.ArrivalTime))
			
			time.Sleep(task.CPUBurst)
			
			task.SetStatus(StatusCompleted)
			metrics.RecordCompletion(task)
			fmt.Printf("%s [Aging] PID=%d DONE (Wait: %v)\n", metrics.GetRelativeTime(), task.PID, task.GetWaitTime())
		} else {
			select {
			case <-ctx.Done():
				if s.readyQueue.Size() == 0 {
					fmt.Printf("%s [Aging Scheduler] Queue empty. Exiting...\n", metrics.GetRelativeTime())
					return
				}
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (s *AgingScheduler) startAging(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.readyQueue.mu.Lock()
			count := len(s.readyQueue.data)
			for i := 0; i < count; i++ {
				t := s.readyQueue.data[i]
				p := t.GetPriority()
				if p > 1 {
					t.SetPriority(p - 1)
				}
			}
			// Re-heapify
			for i := (count / 2) - 1; i >= 0; i-- {
				s.readyQueue.bubbleDown(i)
			}
			s.readyQueue.mu.Unlock()
		}
	}
}
