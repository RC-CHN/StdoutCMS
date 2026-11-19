package store

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestImageCleanupCandidates(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	// Keep all fixtures isolated in temporary tables on the same connection.
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	pg := &Postgres{pool: pool}
	if _, err := pool.Exec(ctx, `
		CREATE TEMP TABLE posts (content text NOT NULL, published boolean NOT NULL);
		CREATE TEMP TABLE about (content text NOT NULL);
		CREATE TEMP TABLE image_cleanup_candidates (object_key text PRIMARY KEY);
		INSERT INTO posts VALUES ('![photo](/img/images/post.png)', true), ('![photo](/img/images/draft.png)', false);
		INSERT INTO about VALUES ('![photo](/img/images/about.png)');
		INSERT INTO image_cleanup_candidates VALUES ('images/post.png'), ('images/draft.png'), ('images/about.png'), ('images/removed.png');
	`); err != nil {
		t.Fatal(err)
	}

	checkCandidates := func(want ...string) {
		t.Helper()
		keys, err := pg.ListImageCleanupCandidates(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(keys, want) {
			t.Fatalf("queued keys = %q, want %q", keys, want)
		}
	}
	checkCandidates("images/about.png", "images/draft.png", "images/post.png", "images/removed.png")

	unexpectedRemoval := func(_ context.Context, key string) error {
		t.Errorf("unexpected storage deletion for %q", key)
		return nil
	}
	for _, key := range []string{"images/local-only.png", "images/post.png", "images/draft.png", "images/about.png"} {
		if removed, err := pg.DeleteImageIfUnreferenced(ctx, key, unexpectedRemoval); err != nil || removed {
			t.Fatalf("cleanup %q: removed = %t, error = %v", key, removed, err)
		}
	}
	const orphan = "images/removed.png"
	checkCandidates(orphan)

	storageError := errors.New("storage temporarily unavailable")
	removed, err := pg.DeleteImageIfUnreferenced(ctx, orphan, func(context.Context, string) error { return storageError })
	if removed || !errors.Is(err, storageError) {
		t.Fatalf("failed deletion: removed = %t, error = %v", removed, err)
	}
	checkCandidates(orphan)

	calls := 0
	removed, err = pg.DeleteImageIfUnreferenced(ctx, orphan, func(_ context.Context, key string) error {
		calls++
		if key != orphan {
			t.Errorf("requested storage deletion for %q, want %q", key, orphan)
		}
		return nil
	})
	if !removed || err != nil || calls != 1 {
		t.Fatalf("retried deletion: removed = %t, calls = %d, error = %v", removed, calls, err)
	}
	if removed, err := pg.DeleteImageIfUnreferenced(ctx, orphan, unexpectedRemoval); err != nil || removed {
		t.Fatalf("already consumed: removed = %t, error = %v", removed, err)
	}
	checkCandidates()
}
