package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskRunsOnStart(t *testing.T) {
	var count atomic.Int32

	task := Task{
		Name:     "test-on-start",
		Interval: 10 * time.Second, // long interval so it won't fire again
		Run: func(ctx context.Context) error {
			count.Add(1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := New(nil, task)
	go s.Start(ctx)

	// Give the task time to run once on start
	time.Sleep(100 * time.Millisecond)
	cancel()

	if n := count.Load(); n < 1 {
		t.Errorf("task should have run at least once on start, got %d", n)
	}
}

func TestTaskRepeats(t *testing.T) {
	var count atomic.Int32

	task := Task{
		Name:     "test-repeat",
		Interval: 50 * time.Millisecond,
		Run: func(ctx context.Context) error {
			count.Add(1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := New(nil, task)
	go s.Start(ctx)

	// Wait for a few ticks
	time.Sleep(180 * time.Millisecond)
	cancel()

	n := count.Load()
	if n < 3 {
		t.Errorf("task should have run at least 3 times (start + 2 ticks), got %d", n)
	}
}

func TestTaskStopsOnCancel(t *testing.T) {
	var count atomic.Int32

	task := Task{
		Name:     "test-cancel",
		Interval: 20 * time.Millisecond,
		Run: func(ctx context.Context) error {
			count.Add(1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := New(nil, task)
	go s.Start(ctx)

	time.Sleep(80 * time.Millisecond)
	before := count.Load()
	cancel()
	time.Sleep(80 * time.Millisecond)
	after := count.Load()

	// After cancellation, count should stabilise (no more increments)
	if after != before {
		t.Errorf("task kept running after cancel: before=%d after=%d", before, after)
	}
}

func TestMultipleTasks(t *testing.T) {
	var a, b atomic.Int32

	tasks := []Task{
		{
			Name:     "task-a",
			Interval: 30 * time.Millisecond,
			Run: func(ctx context.Context) error {
				a.Add(1)
				return nil
			},
		},
		{
			Name:     "task-b",
			Interval: 30 * time.Millisecond,
			Run: func(ctx context.Context) error {
				b.Add(1)
				return nil
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := New(nil, tasks...)
	go s.Start(ctx)

	time.Sleep(120 * time.Millisecond)
	cancel()

	if n := a.Load(); n < 3 {
		t.Errorf("task-a should have run at least 3 times, got %d", n)
	}
	if n := b.Load(); n < 3 {
		t.Errorf("task-b should have run at least 3 times, got %d", n)
	}
}

func TestTaskErrorDoesNotCrash(t *testing.T) {
	var count atomic.Int32

	task := Task{
		Name:     "test-error",
		Interval: 30 * time.Millisecond,
		Run: func(ctx context.Context) error {
			count.Add(1)
			// always returns an error — scheduler should keep going
			return context.DeadlineExceeded
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := New(nil, task)
	go s.Start(ctx)

	time.Sleep(100 * time.Millisecond)
	cancel()

	if n := count.Load(); n < 2 {
		t.Errorf("task should have kept running despite errors, got %d", n)
	}
}

func TestNilLoggerDoesNotPanic(t *testing.T) {
	task := Task{
		Name:     "test-nil-logger",
		Interval: time.Hour,
		Run: func(ctx context.Context) error {
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	s := New(nil, task)
	// Should not panic
	s.Start(ctx)
}
