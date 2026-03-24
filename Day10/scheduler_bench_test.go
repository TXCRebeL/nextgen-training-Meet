package main

import (
	"sync"
	"testing"
	"time"
)

// Micro-benchmarks for MinHeap
func BenchmarkMinHeap_Insert(b *testing.B) {
	h := NewMinHeap[*Task](func(a, b *Task) bool { return a.GetPriority() < b.GetPriority() })
	tasks := make([]*Task, b.N)
	for i := 0; i < b.N; i++ {
		tasks[i] = NewTask(i, "Task", i%10+1, 10*time.Millisecond, time.Now())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Insert(tasks[i])
	}
}

func BenchmarkMinHeap_ExtractMin(b *testing.B) {
	h := NewMinHeap[*Task](func(a, b *Task) bool { return a.GetPriority() < b.GetPriority() })
	for i := 0; i < b.N; i++ {
		h.Insert(NewTask(i, "Task", i%10+1, 10*time.Millisecond, time.Now()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.ExtractMin()
	}
}

// Macro-benchmarks for Schedulers
func BenchmarkPriorityScheduler_Schedule(b *testing.B) {
	s := NewPriorityScheduler()
	metrics := NewMetricsCollector()
	var wg sync.WaitGroup
	
	count := 1000
	tasks := make([]*Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = NewTask(i, "Task", i%10+1, 1*time.Microsecond, time.Now())
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < count; j++ {
			s.Add(tasks[j])
		}
		
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s.readyQueue.Size() > 0 {
				task := s.Next()
				if task != nil {
					task.SetStatus(StatusCompleted)
					metrics.RecordCompletion(task)
				}
			}
		}()
		wg.Wait()
	}
}

func BenchmarkRoundRobinScheduler_Schedule(b *testing.B) {
	s := NewRoundRobinScheduler()
	metrics := NewMetricsCollector()
	
	count := 1000
	tasks := make([]*Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = NewTask(i, "Task", i%10+1, 1*time.Microsecond, time.Now())
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < count; j++ {
			s.Add(tasks[j])
		}
		
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			processed := 0
			for processed < count {
				task := s.Next()
				if task != nil {
					task.SetStatus(StatusCompleted)
					metrics.RecordCompletion(task)
					processed++
				}
			}
		}()
		wg.Wait()
	}
}
