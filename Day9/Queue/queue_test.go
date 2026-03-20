package queue

import (
	"sync"
	"testing"
)

func TestCircularQueue_Operations_TableDriven(t *testing.T) {
	type testCase struct {
		name       string
		capacity   int
		operations []struct {
			op       string
			val      int
			wantErr  bool
			expected int
		}
	}

	testCases := []testCase{
		{
			name:     "Basic Enqueue and Dequeue",
			capacity: 3,
			operations: []struct {
				op       string
				val      int
				wantErr  bool
				expected int
			}{
				{op: "enqueue", val: 10, wantErr: false},
				{op: "enqueue", val: 20, wantErr: false},
				{op: "enqueue", val: 30, wantErr: false},
				{op: "size", expected: 3},
				{op: "dequeue", wantErr: false, expected: 10},
				{op: "dequeue", wantErr: false, expected: 20},
				{op: "dequeue", wantErr: false, expected: 30},
				{op: "size", expected: 0},
				{op: "isempty", expected: 1}, // 1 for true
			},
		},
		{
			name:     "Queue Full and Empty Errors",
			capacity: 2,
			operations: []struct {
				op       string
				val      int
				wantErr  bool
				expected int
			}{
				{op: "dequeue", wantErr: true},
				{op: "enqueue", val: 1, wantErr: false},
				{op: "enqueue", val: 2, wantErr: false},
				{op: "enqueue", val: 3, wantErr: true},
				{op: "isfull", expected: 1},
				{op: "peek", wantErr: false, expected: 1},
			},
		},
		{
			name:     "Wrap Around Behavior",
			capacity: 3,
			operations: []struct {
				op       string
				val      int
				wantErr  bool
				expected int
			}{
				{op: "enqueue", val: 1, wantErr: false},
				{op: "enqueue", val: 2, wantErr: false},
				{op: "dequeue", wantErr: false, expected: 1},
				{op: "enqueue", val: 3, wantErr: false},
				{op: "enqueue", val: 4, wantErr: false},
				{op: "isfull", expected: 1},
				{op: "dequeue", wantErr: false, expected: 2},
				{op: "dequeue", wantErr: false, expected: 3},
				{op: "dequeue", wantErr: false, expected: 4},
				{op: "isempty", expected: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewCircularQueue[int](tc.capacity)
			for i, op := range tc.operations {
				var err error
				var val int
				switch op.op {
				case "enqueue":
					err = q.Enqueue(op.val)
				case "dequeue":
					val, err = q.Dequeue()
				case "peek":
					val, err = q.Peek()
				case "size":
					val = q.Size()
				case "isempty":
					if q.IsEmpty() {
						val = 1
					} else {
						val = 0
					}
				case "isfull":
					if q.IsFull() {
						val = 1
					} else {
						val = 0
					}
				default:
					t.Fatalf("Operation %d: unknown operation %s", i, op.op)
				}

				if (err != nil) != op.wantErr {
					t.Errorf("Operation %d (%s): wantErr = %v, got err = %v", i, op.op, op.wantErr, err)
				}

				if op.op == "dequeue" || op.op == "peek" || op.op == "size" || op.op == "isempty" || op.op == "isfull" {
					if !op.wantErr && val != op.expected {
						t.Errorf("Operation %d (%s): expected %d, got %d", i, op.op, op.expected, val)
					}
				}
			}
		})
	}
}

func TestCircularQueue_Race(t *testing.T) {
	q := NewCircularQueue[int](10)
	var wg sync.WaitGroup
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = q.Enqueue(i)
		}
	}()
	
	for i := 0; i < 100; i++ {
		_, _ = q.Dequeue()
	}
	
	wg.Wait()
}
