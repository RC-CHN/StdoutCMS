package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type imageCleanupStore interface {
	ListImageCleanupCandidates(context.Context) ([]string, error)
	DeleteImageIfUnreferenced(context.Context, string, func(context.Context, string) error) (bool, error)
}

// Only images removed from server-saved content are candidates. An upload may
// belong to a browser draft for any length of time before its first server save.
func CleanupOrphanImages(logger *slog.Logger, pg imageCleanupStore, remove func(context.Context, string) error) func(context.Context) error {
	return func(ctx context.Context) error {
		keys, err := pg.ListImageCleanupCandidates(ctx)
		if err != nil {
			return fmt.Errorf("list image cleanup candidates: %w", err)
		}

		var failures []error
		deleted := 0
		for _, key := range keys {
			if err := ctx.Err(); err != nil {
				return err
			}
			// An object-store outage must not hold up content writes indefinitely.
			attempt, cancel := context.WithTimeout(ctx, 5*time.Second)
			removed, err := pg.DeleteImageIfUnreferenced(attempt, key, remove)
			cancel()
			if err != nil {
				failures = append(failures, fmt.Errorf("clean up %s: %w", key, err))
			} else if removed {
				deleted++
			}
		}
		logger.Info("orphan-images: cleanup finished", "candidates", len(keys), "deleted", deleted, "failed", len(failures))
		return errors.Join(failures...)
	}
}
