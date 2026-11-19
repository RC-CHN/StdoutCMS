package task

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
)

type cleanupQueue []string

func (q cleanupQueue) ListImageCleanupCandidates(context.Context) ([]string, error) {
	return q, nil
}

func (q cleanupQueue) DeleteImageIfUnreferenced(ctx context.Context, key string, remove func(context.Context, string) error) (bool, error) {
	err := remove(ctx, key)
	return err == nil, err
}

func TestCleanupProcessesRemainingCandidatesAfterFailure(t *testing.T) {
	queue := cleanupQueue{"images/first.png", "images/second.png"}
	var attempted []string
	storageErr := errors.New("object storage unavailable")
	remove := func(ctx context.Context, key string) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("deletion must have a deadline")
		}
		attempted = append(attempted, key)
		if key == queue[0] {
			return storageErr
		}
		return nil
	}

	err := CleanupOrphanImages(slog.Default(), queue, remove)(context.Background())
	if !errors.Is(err, storageErr) || !slices.Equal(attempted, []string(queue)) {
		t.Fatalf("attempted %v, got error %v", attempted, err)
	}
}

func TestCleanupEmptyQueuePreservesDraftUploads(t *testing.T) {
	remove := func(_ context.Context, key string) error {
		t.Fatalf("unexpected deletion of %s", key)
		return nil
	}
	if err := CleanupOrphanImages(slog.Default(), cleanupQueue{}, remove)(context.Background()); err != nil {
		t.Fatal(err)
	}
}
