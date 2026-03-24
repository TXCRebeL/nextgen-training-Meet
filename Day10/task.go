package main

import (
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusReady     TaskStatus = "ready"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusStarved   TaskStatus = "starved"
)

type Task struct {
	mu           sync.RWMutex
	PID          int
	Name         string
	Priority     int // 1-10, 1=highest
	CPUBurst     time.Duration
	Remaining    time.Duration
	Deadline     time.Time
	ArrivalTime  time.Time
	WaitTime     time.Duration
	LastServedAt time.Time
	Status       TaskStatus
}

func NewTask(pid int, name string, priority int, burst time.Duration, deadline time.Time) *Task {
	now := time.Now()
	return &Task{
		PID:         pid,
		Name:        name,
		Priority:    priority,
		CPUBurst:    burst,
		Remaining:   burst,
		Deadline:    deadline,
		ArrivalTime: now,
		Status:      StatusReady,
	}
}

// GetStatus returns the current status thread-safely
func (t *Task) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}

// SetStatus sets the status thread-safely
func (t *Task) SetStatus(s TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = s
}

// GetPriority returns the priority thread-safely
func (t *Task) GetPriority() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Priority
}

// SetPriority sets the priority thread-safely
func (t *Task) SetPriority(p int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Priority = p
}

// GetRemaining returns the remaining burst time thread-safely
func (t *Task) GetRemaining() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Remaining
}

// SetRemaining sets the remaining burst time thread-safely
func (t *Task) SetRemaining(r time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Remaining = r
}

// GetWaitTime returns the wait time thread-safely
func (t *Task) GetWaitTime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.WaitTime
}

// SetWaitTime sets the wait time thread-safely
func (t *Task) SetWaitTime(w time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.WaitTime = w
}
