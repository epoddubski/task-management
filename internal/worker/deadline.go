package worker

import (
	"context"
	"log/slog"
	"time"
)

type TaskRepository interface {
	MarkOverdue(ctx context.Context, now time.Time) (int64, error)
}

type DeadlineWorker struct {
	tasks    TaskRepository
	interval time.Duration
}

func NewDeadlineWorker(tasks TaskRepository, interval time.Duration) *DeadlineWorker {
	return &DeadlineWorker{tasks: tasks, interval: interval}
}

func (w *DeadlineWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("deadline worker started", "interval", w.interval.String())

	for {
		select {
		case <-ctx.Done():
			slog.Info("deadline worker stopping")
			return
		case t := <-ticker.C:
			w.process(ctx, t)
		}
	}
}

func (w *DeadlineWorker) process(ctx context.Context, now time.Time) {
	count, err := w.tasks.MarkOverdue(ctx, now)
	if err != nil {
		slog.Error("failed to mark Deadline tasks", "error", err)
		return
	}
	if count > 0 {
		slog.Info("marked tasks deadline", "count", count)
	}
}
