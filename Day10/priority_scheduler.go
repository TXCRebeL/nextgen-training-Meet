package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type PriorityScheduler struct {
	readyQueue *MinHeap[*Task]
}

func NewPriorityScheduler() *PriorityScheduler {
	return &PriorityScheduler{
		readyQueue: NewMinHeap[*Task](func(a, b *Task) bool { 
			return a.GetPriority() < b.GetPriority() 
		}),
	}
}

func (s *PriorityScheduler) Add(t *Task) {
	s.readyQueue.Insert(t)
}

func (s *PriorityScheduler) Next() *Task {
	task, err := s.readyQueue.ExtractMin()
	if err != nil {
		return nil
	}
	return task
}

func (s *PriorityScheduler) ScheduleTask(ctx context.Context, wg *sync.WaitGroup, metrics *MetricsCollector) {
	defer wg.Done()
	for {
		task := s.Next()
		if task != nil {
			metrics.RecordContextSwitch()
			fmt.Printf("%s [Priority] PID=%d (P=%d) START\n", metrics.GetRelativeTime(), task.PID, task.GetPriority())

			task.SetStatus(StatusRunning)
			task.SetWaitTime(time.Since(task.ArrivalTime))

			// Simulate CPU execution
			time.Sleep(task.CPUBurst)

			task.SetStatus(StatusCompleted)
			metrics.RecordCompletion(task)
			fmt.Printf("%s [Priority] PID=%d DONE (Wait: %v)\n", metrics.GetRelativeTime(), task.PID, task.GetWaitTime())
		} else {
			select {
			case <-ctx.Done():
				// Exit only when queue is empty after context is done
				if s.readyQueue.Size() == 0 {
					fmt.Printf("%s [Priority Scheduler] Queue empty. Exiting...\n", metrics.GetRelativeTime())
					return
				}
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
