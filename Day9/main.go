package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type SensorType string

const (
	Temperature SensorType = "Temperature"
	Humidity    SensorType = "Humidity"
	Pressure    SensorType = "Pressure"
	Motion      SensorType = "Motion"
	Light       SensorType = "Light"
)

var sensorTypes = []SensorType{
	Temperature,
	Humidity,
	Pressure,
	Motion,
	Light,
}

func RandomSensorType() SensorType {
	return sensorTypes[rand.Intn(len(sensorTypes))]
}

type Alert struct {
	SensorID string
	Type     SensorType
	Message  string
	Severity string
	Time     time.Time
}

type Sensor struct {
	SensorID  string
	Type      SensorType
	Value     float64
	Timestamp time.Time
	Location  string
}

type Statistics struct {
	Sum           float64
	Count         int
	LastAlertTime time.Time
	AlertCount    int
}

var (
	SensorData     map[string]Statistics
	mu             sync.Mutex
	producerWG     sync.WaitGroup
	processorWG    sync.WaitGroup
	handlerWG      sync.WaitGroup
	totalProcessed uint64
	totalLatencyNs int64
)

func sensorProducer(ctx context.Context, ch chan<- Sensor) {
	defer producerWG.Done()
	ticker := time.NewTicker(time.Duration(rand.Intn(400)+100) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Producer shutting down...")
			return
		case <-ticker.C:
			if float64(len(ch)) > float64(cap(ch))*0.85 {
				fmt.Println("Queue 85% full — applying backpressure")
			}
			sensor := Sensor{
				SensorID:  fmt.Sprintf("Sensor-%d", rand.Intn(10)),
				Type:      RandomSensorType(),
				Value:     rand.Float64() * 100,
				Timestamp: time.Now(),
				Location:  fmt.Sprintf("Location-%d", rand.Intn(10)),
			}
			select {
			case ch <- sensor:
				fmt.Printf("Sensor %s sent to queue\n", sensor.SensorID)
			default:
				fmt.Printf("Queue full, dropping sensor %s\n", sensor.SensorID)
			}
		}
	}
}

// validate data, compute rolling average, detect anomalies
// Anomaly: temperature > 50°C, humidity > 95%, motion during restricted hours
func sensorProcessor(wp int, sensorChan <-chan Sensor, alertChan chan<- Alert) {
	for i := 0; i < wp; i++ {
		processorWG.Add(1)
		go func(id int) {
			defer processorWG.Done()
			for {
				select {
				case sensor, ok := <-sensorChan:
					if !ok {
						fmt.Printf("Processor %d: channel closed, exiting...\n", id)
						return
					}
					startTime := time.Now()
					time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
					// 1. Validate data
					if sensor.Value < 0 || sensor.Value > 100 {
						alertChan <- Alert{
							SensorID: sensor.SensorID,
							Type:     sensor.Type,
							Message:  fmt.Sprintf("Invalid sensor value: %.2f", sensor.Value),
							Severity: "Low",
							Time:     sensor.Timestamp,
						}
						atomic.AddUint64(&totalProcessed, 1)
						atomic.AddInt64(&totalLatencyNs, time.Since(startTime).Nanoseconds())
						continue
					}

					// 2. Compute rolling average
					key := fmt.Sprintf("%s-%s", sensor.SensorID, sensor.Type)
					mu.Lock()
					val := SensorData[key]
					val.Sum += sensor.Value
					val.Count++
					SensorData[key] = val
					mu.Unlock()

					avg := val.Sum / float64(val.Count)

					// 3. Detect anomalies
					isAnomaly, message, severity := DetectAnomaly(sensor, avg)

					if isAnomaly {
						alertChan <- Alert{
							SensorID: sensor.SensorID,
							Type:     sensor.Type,
							Message:  message,
							Severity: severity,
							Time:     sensor.Timestamp,
						}
					}
					atomic.AddUint64(&totalProcessed, 1)
					atomic.AddInt64(&totalLatencyNs, time.Since(startTime).Nanoseconds())
				case <-time.After(5 * time.Second):
					fmt.Printf("Processor %d: no data for 5s — warning\n", id)
				}
			}
		}(i)
	}
}

// Alert Printer and deduplicates handler
// Alert Printer and deduplicates handler
func alertHandler(alertChan <-chan Alert) {
	handlerWG.Add(1)
	go func() {
		defer handlerWG.Done()
		for alert := range alertChan {
			key := fmt.Sprintf("%s-%s", alert.SensorID, alert.Type)
			mu.Lock()
			val := SensorData[key]

			if time.Since(val.LastAlertTime) > 60*time.Second {
				fmt.Printf("[%s] ALERT from %s (%s): %s [%s]\n",
					alert.Time.Format("15:04:05"), alert.SensorID, alert.Type, alert.Message, alert.Severity)
				val.LastAlertTime = time.Now()
				val.AlertCount++
				SensorData[key] = val
			}

			mu.Unlock()
		}
		fmt.Println("Alert Handler: channel closed, exiting...")
	}()
}

func metricsReporter(ctx context.Context, sensorChan chan Sensor) {
	producerWG.Add(1)
	go func() {
		defer producerWG.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		var lastProcessed uint64
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Metrics Reporter shutting down...")
				return
			case <-ticker.C:
				currentProcessed := atomic.LoadUint64(&totalProcessed)
				processedInInterval := currentProcessed - lastProcessed
				lastProcessed = currentProcessed

				throughput := float64(processedInInterval)
				var avgLatency time.Duration
				if currentProcessed > 0 {
					avgLatency = time.Duration(atomic.LoadInt64(&totalLatencyNs) / int64(currentProcessed))
				}

				utilization := float64(len(sensorChan)) / float64(cap(sensorChan)) * 100

				fmt.Println("\n--- SYSTEM METRICS ---")
				fmt.Printf("Readings processed/sec: %.2f\n", throughput)
				fmt.Printf("Average processing latency: %v\n", avgLatency)
				fmt.Printf("Queue utilization: %.2f%%\n", utilization)
				fmt.Println("Alerts per sensor:")
				mu.Lock()
				for id, stats := range SensorData {
					if stats.AlertCount > 0 {
						fmt.Printf("  %s: %d alerts\n", id, stats.AlertCount)
					}
				}
				mu.Unlock()
				fmt.Println("----------------------")
			}
		}
	}()
}

func DetectAnomaly(sensor Sensor, avg float64) (bool, string, string) {
	isAnomaly := false
	message := ""
	severity := "High"

	switch sensor.Type {
	case Temperature:
		if sensor.Value > 50 {
			isAnomaly = true
			message = fmt.Sprintf("High Temperature Detected: %.2f°C (Avg: %.2f°C)", sensor.Value, avg)
		}
	case Humidity:
		if sensor.Value > 95 {
			isAnomaly = true
			message = fmt.Sprintf("High Humidity Detected: %.2f%% (Avg: %.2f%%)", sensor.Value, avg)
		}
	case Motion:
		hour := sensor.Timestamp.Hour()
		if hour >= 16 || hour < 6 {
			isAnomaly = true
			message = fmt.Sprintf("Motion Detected during restricted hours: %02d:%02d", hour, sensor.Timestamp.Minute())
		}
	}
	return isAnomaly, message, severity
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sensorChan := make(chan Sensor, 100)
	alertChan := make(chan Alert, 100)
	SensorData = make(map[string]Statistics)

	wp := 3

	//10 simulated sensors
	for range 10 {
		producerWG.Add(1)
		go sensorProducer(ctx, sensorChan)
	}

	sensorProcessor(wp, sensorChan, alertChan)
	alertHandler(alertChan)
	metricsReporter(ctx, sensorChan)

	// Wait for interrupt
	<-stop
	fmt.Println("\nShutdown signal received. Cleaning up...")
	cancel()

	// Graceful shutdown sequence:
	// 1. Wait for producers to stop after signal
	producerWG.Wait()
	close(sensorChan)

	// 2. Wait for processors to finish current work (drain sensorChan) and exit
	processorWG.Wait()
	close(alertChan)

	// 3. Wait for alert handler to drain alerts and exit
	handlerWG.Wait()

	fmt.Println("All goroutines finished. Goodbye!")
}
