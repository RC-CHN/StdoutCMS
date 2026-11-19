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

func postgresWithTempTables(t *testing.T, ctx context.Context) *Postgres {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
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
	t.Cleanup(pool.Close)
	return &Postgres{pool: pool}
}

func TestListPostSummaries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pg := postgresWithTempTables(t, ctx)
	// Intentionally leave out content so a query that loads post bodies fails.
	if _, err := pg.pool.Exec(ctx, `
		CREATE TEMP TABLE posts (slug text, title text, excerpt text, tags text[], created_at timestamptz, published boolean);
		INSERT INTO posts VALUES
			('old', 'Old', '', '{}', '2020-01-01', true),
			('new', 'New', 'Excerpt', '{tag}', '2020-01-02', true),
			('draft', 'Draft', '', '{}', '2020-01-03', false);
	`); err != nil {
		t.Fatal(err)
	}
	posts, total, err := pg.ListPostSummaries(ctx, 1, 1)
	if err != nil || total != 2 || len(posts) != 1 {
		t.Fatalf("first page = %v, total = %d, err = %v", posts, total, err)
	}
	post := posts[0]
	if post.Slug != "new" || post.Title != "New" || post.Excerpt != "Excerpt" || !slices.Equal(post.Tags, []string{"tag"}) || post.CreatedAt.IsZero() {
		t.Fatalf("unexpected summary: %+v", post)
	}
	posts, total, err = pg.ListPostSummaries(ctx, 3, 1)
	if err != nil || total != 2 || posts == nil || len(posts) != 0 {
		t.Fatalf("empty page = %v, total = %d, err = %v", posts, total, err)
	}
}

func TestImageCleanupCandidates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pg := postgresWithTempTables(t, ctx)
	if _, err := pg.pool.Exec(ctx, `
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
