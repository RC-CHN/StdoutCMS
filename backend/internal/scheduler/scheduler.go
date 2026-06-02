package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Task describes a recurring background job.
type Task struct {
	Name     string
	Interval time.Duration // how often to run after the first execution
	Run      func(ctx context.Context) error
}

// Scheduler runs a set of Tasks concurrently, each in its own goroutine.
// Every task fires once immediately on Start, then repeats at its Interval.
type Scheduler struct {
	tasks  []Task
	logger *slog.Logger
}

// New creates a Scheduler with the given tasks.
func New(logger *slog.Logger, tasks ...Task) *Scheduler {
	return &Scheduler{
		tasks:  tasks,
		logger: logger,
	}
}

// Start launches every task in a background goroutine.
// It blocks until ctx is cancelled, then lets all goroutines exit.
func (s *Scheduler) Start(ctx context.Context) {
	for _, t := range s.tasks {
		go s.runLoop(ctx, t)
	}
	s.log("scheduler started", "tasks", len(s.tasks))
	<-ctx.Done()
	s.log("scheduler shutting down")
}

func (s *Scheduler) log(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

func (s *Scheduler) logError(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Error(msg, args...)
	}
}

func (s *Scheduler) runLoop(ctx context.Context, t Task) {
	// run immediately on start
	s.runOne(ctx, t)

	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOne(ctx, t)
		}
	}
}

func (s *Scheduler) runOne(ctx context.Context, t Task) {
	s.log("scheduler: running task", "name", t.Name)
	start := time.Now()
	if err := t.Run(ctx); err != nil {
		s.logError("scheduler: task failed", "name", t.Name, "error", err, "elapsed", time.Since(start))
	} else {
		s.log("scheduler: task completed", "name", t.Name, "elapsed", time.Since(start))
	}
}
