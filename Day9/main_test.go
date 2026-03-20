package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDetectAnomaly_TableDriven(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name         string
		sensor       Sensor
		avg          float64
		wantAnomaly  bool
		wantSeverity string
	}{
		{
			name: "Normal Temperature",
			sensor: Sensor{
				SensorID:  "S1",
				Type:      Temperature,
				Value:     25.0,
				Timestamp: now,
			},
			avg:          24.5,
			wantAnomaly:  false,
			wantSeverity: "High",
		},
		{
			name: "High Temperature Anomaly",
			sensor: Sensor{
				SensorID:  "S1",
				Type:      Temperature,
				Value:     55.0,
				Timestamp: now,
			},
			avg:          25.0,
			wantAnomaly:  true,
			wantSeverity: "High",
		},
		{
			name: "Normal Humidity",
			sensor: Sensor{
				SensorID:  "S2",
				Type:      Humidity,
				Value:     50.0,
				Timestamp: now,
			},
			avg:          48.0,
			wantAnomaly:  false,
			wantSeverity: "High",
		},
		{
			name: "High Humidity Anomaly",
			sensor: Sensor{
				SensorID:  "S2",
				Type:      Humidity,
				Value:     96.0,
				Timestamp: now,
			},
			avg:          50.0,
			wantAnomaly:  true,
			wantSeverity: "High",
		},
		{
			name: "Motion Normal Hours",
			sensor: Sensor{
				SensorID:  "S3",
				Type:      Motion,
				Value:     1.0,
				Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), // 10 AM
			},
			avg:          0.0,
			wantAnomaly:  false,
			wantSeverity: "High",
		},
		{
			name: "Motion Restricted Hours (Evening)",
			sensor: Sensor{
				SensorID:  "S3",
				Type:      Motion,
				Value:     1.0,
				Timestamp: time.Date(2023, 1, 1, 20, 0, 0, 0, time.UTC), // 8 PM
			},
			avg:          0.0,
			wantAnomaly:  true,
			wantSeverity: "High",
		},
		{
			name: "Motion Restricted Hours (Morning)",
			sensor: Sensor{
				SensorID:  "S3",
				Type:      Motion,
				Value:     1.0,
				Timestamp: time.Date(2023, 1, 1, 4, 0, 0, 0, time.UTC), // 4 AM
			},
			avg:          0.0,
			wantAnomaly:  true,
			wantSeverity: "High",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAnomaly, _, gotSeverity := DetectAnomaly(tc.sensor, tc.avg)
			if gotAnomaly != tc.wantAnomaly {
				t.Errorf("DetectAnomaly() gotAnomaly = %v, want %v", gotAnomaly, tc.wantAnomaly)
			}
			if gotAnomaly && gotSeverity != tc.wantSeverity {
				t.Errorf("DetectAnomaly() gotSeverity = %v, want %v", gotSeverity, tc.wantSeverity)
			}
		})
	}
}

func BenchmarkQueueThroughput(b *testing.B) {
	queueSizes := []int{10, 100, 1000, 10000}
	for _, size := range queueSizes {
		b.Run(fmt.Sprintf("QueueSize-%d", size), func(b *testing.B) {
			sensorChan := make(chan Sensor, size)
			// alertChan := make(chan Alert, size)
			alertChan := make(chan Alert, size)

			// Minimal processing for benchmark
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// // Sink for alerts
			// go func() {
			// 	for range alertChan {
			// 	}
			// }()

			// Sink for alerts
			go func() {
				for range alertChan {
				}
			}()

			// Start a simplified processor
			go func() {
				for {
					select {
					case <-sensorChan:
						// Simulate some work
					case <-ctx.Done():
						return
					}
				}
			}()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sensorChan <- Sensor{
					SensorID: "Bench",
					Type:     Temperature,
					Value:    25.0,
				}
			}
			b.StopTimer()
		})
	}
}
