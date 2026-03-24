package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

// generateTasks creates tasks and adds them to all schedulers.
func generateTasks(ctx context.Context, schedulers []Scheduler, wg *sync.WaitGroup, metrics *MetricsCollector, count int, loadMode bool) {
	defer wg.Done()
	pid := 1
	for i := 0; i < count || count == -1; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("%s [Producer] Stopping task generation...\n", metrics.GetRelativeTime())
			return
		default:
			if !loadMode {
				// Random interval between 200ms and 600ms in normal mode
				time.Sleep(time.Duration(rand.Intn(400)+200) * time.Millisecond)
				// Check context again after sleep
				select {
				case <-ctx.Done():
					return
				default:
				}
			}

			p := rand.Intn(10) + 1
			burst := time.Duration(rand.Intn(100)+50) * time.Millisecond
			if loadMode {
				// Shorter bursts for load testing to process many tasks quickly
				burst = time.Duration(rand.Intn(10)+1) * time.Millisecond
			}

			t := NewTask(pid, fmt.Sprintf("Task-%d", pid), p, burst, time.Now().Add(10*time.Second))
			metrics.RecordNewTask()

			if !loadMode || pid%100 == 0 {
				fmt.Printf("%s [Producer] New Task: PID=%d, Priority=%d, Burst=%v\n", metrics.GetRelativeTime(), pid, p, burst)
			}

			for _, s := range schedulers {
				// Each scheduler gets its own copy to avoid state sharing during simulation
				copyTask := *t
				s.Add(&copyTask)
			}
			pid++

			if loadMode && pid%500 == 0 {
				// Yield to avoid starving schedulers in tight loop
				runtime.Gosched()
			}
		}
	}
	fmt.Printf("%s [Producer] Generated all %d tasks.\n", metrics.GetRelativeTime(), pid-1)
}

func main() {
	loadCount := flag.Int("load", -1, "Number of tasks to generate for load testing (-1 for infinite)")
	pprofAddr := flag.String("pprof", ":6060", "Address for pprof HTTP server")
	flag.Parse()

	// Enable Mutex and Block profiling
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	// Start pprof server
	go func() {
		fmt.Printf("Starting pprof server on %s\n", *pprofAddr)
		if err := http.ListenAndServe(*pprofAddr, nil); err != nil {
			fmt.Printf("pprof server failed: %v\n", err)
		}
	}()

	// Root context for signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Initialize Metrics
	metrics := NewMetricsCollector()

	// Initialize Schedulers
	pSched := NewPriorityScheduler()
	aSched := NewAgingScheduler(ctx)
	rrSched := NewRoundRobinScheduler()

	schedulers := []Scheduler{pSched, aSched, rrSched}

	var wg sync.WaitGroup

	// Run task generator
	wg.Add(1)
	loadMode := *loadCount > 0
	go generateTasks(ctx, schedulers, &wg, metrics, *loadCount, loadMode)

	// Start Simulation worker loops for each scheduler
	for _, s := range schedulers {
		wg.Add(1)
		go s.ScheduleTask(ctx, &wg, metrics)
	}

	fmt.Println("CPU Scheduler Simulation Started...")
	fmt.Printf("Profiling enabled. Visit http://localhost%s/debug/pprof/\n", *pprofAddr)

	if loadMode {
		fmt.Printf("LOAD MODE: Generating %d tasks...\n", *loadCount)
	} else {
		fmt.Println("NORMAL MODE: Press Ctrl+C to stop.")
	}

	// Wait for ALL goroutines to finish (Producer + Schedulers)
	wg.Wait()

	fmt.Printf("%s Graceful shutdown complete. All tasks processed.\n", metrics.GetRelativeTime())

	// Print Final Metrics
	metrics.PrintSummary()

	// Small delay to allow pprof to capture final state if needed
	time.Sleep(1 * time.Second)
}
