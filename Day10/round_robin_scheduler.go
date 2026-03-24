package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RoundRobinScheduler struct {
	queues [3]chan *Task
}

func NewRoundRobinScheduler() *RoundRobinScheduler {
	return &RoundRobinScheduler{
		queues: [3]chan *Task{
			make(chan *Task, 100), // High (0)
			make(chan *Task, 100), // Medium (1)
			make(chan *Task, 100), // Low (2)
		},
	}
}

func (s *RoundRobinScheduler) Add(t *Task) {
	p := t.GetPriority()
	switch {
	case p <= 3:
		s.queues[0] <- t
	case p <= 7:
		s.queues[1] <- t
	default:
		s.queues[2] <- t
	}
}

func (s *RoundRobinScheduler) Next() *Task {
	// Strictly check high, then medium, then low
	for i := 0; i < 3; i++ {
		select {
		case t := <-s.queues[i]:
			return t
		default:
			continue
		}
	}
	return nil
}

func (s *RoundRobinScheduler) ScheduleTask(ctx context.Context, wg *sync.WaitGroup, metrics *MetricsCollector) {
	defer wg.Done()
	for {
		task := s.Next()
		if task != nil {
			s.executeTask(task, metrics)
		} else {
			select {
			case <-ctx.Done():
				fmt.Printf("%s [Round-Robin Scheduler] Shutting down...\n", metrics.GetRelativeTime())
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (s *RoundRobinScheduler) executeTask(t *Task, metrics *MetricsCollector) {
	prio := t.GetPriority()
	quantum := s.GetQuantum(prio)
	startTime := time.Now()
	
	metrics.RecordContextSwitch()
	fmt.Printf("%s [Round-Robin] PID=%d (P=%d) START (Remaining %v)\n", 
		metrics.GetRelativeTime(), t.PID, prio, t.GetRemaining())
	
	isHigh := prio <= 3
	rem := t.GetRemaining()
	runTime := quantum
	if rem < quantum {
		runTime = rem
	}
	
	runTimer := time.NewTimer(runTime)
	defer runTimer.Stop()

	for {
		select {
		case <-runTimer.C:
			elapsed := time.Since(startTime)
			currentRem := t.GetRemaining()
			if elapsed >= currentRem {
				t.SetRemaining(0)
				t.SetStatus(StatusCompleted)
				t.SetWaitTime(time.Since(t.ArrivalTime) - t.CPUBurst)
				if t.GetWaitTime() < 0 { t.SetWaitTime(0) }
				metrics.RecordCompletion(t)
				fmt.Printf("%s [Round-Robin] PID=%d DONE\n", metrics.GetRelativeTime(), t.PID)
			} else {
				t.SetRemaining(currentRem - elapsed)
				t.SetStatus(StatusReady)
				fmt.Printf("%s [Round-Robin] PID=%d QUANTUM EXPIRED (Remaining %v). Re-queueing...\n", 
					metrics.GetRelativeTime(), t.PID, t.GetRemaining())
				s.Add(t)
			}
			return
		default:
			if !isHigh {
				select {
				case highTask := <-s.queues[0]:
					elapsed := time.Since(startTime)
					currentRem := t.GetRemaining()
					if elapsed >= currentRem {
						t.SetRemaining(0)
					} else {
						t.SetRemaining(currentRem - elapsed)
					}
					t.SetStatus(StatusReady)
					fmt.Printf("%s [Round-Robin] PID=%d PREEMPTED by High-Priority PID=%d (Remaining %v)\n", 
						metrics.GetRelativeTime(), t.PID, highTask.PID, t.GetRemaining())
					
					s.Add(t)
					s.Add(highTask)
					return 
				default:
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (s *RoundRobinScheduler) GetQuantum(priority int) time.Duration {
	if priority <= 3 {
		return 50 * time.Millisecond
	}
	if priority <= 7 {
		return 100 * time.Millisecond
	}
	return 200 * time.Millisecond
}
