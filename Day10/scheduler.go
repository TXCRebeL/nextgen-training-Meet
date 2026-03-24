package main

import (
	"context"
	"sync"
)

type Scheduler interface {
	Add(task *Task)
	Next() *Task
	ScheduleTask(ctx context.Context, wg *sync.WaitGroup, metrics *MetricsCollector)
}
