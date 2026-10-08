package store

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetAllContent(t *testing.T) {
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
	// Keep the temporary tables on the same connection for every query.
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	pg := &Postgres{pool: pool}
	// Temporary tables isolate the fixtures from any existing application data.
	if _, err := pool.Exec(ctx, `
		CREATE TEMP TABLE posts (content TEXT NOT NULL, published BOOLEAN NOT NULL);
		CREATE TEMP TABLE about (content TEXT NOT NULL);
	`); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name      string
		published []string
		drafts    []string
		about     []string
	}{
		{name: "empty site"},
		{name: "about only", about: []string{"![portrait](/img/images/about.png)"}},
		{
			name:      "posts without about",
			published: []string{"![post](/img/images/post.png)"},
			drafts:    []string{"![draft](/img/images/draft.png)"},
		},
		{
			name:      "posts drafts and about",
			published: []string{"![post](/img/images/post.png)"},
			drafts:    []string{"![draft](/img/images/draft.png)"},
			about:     []string{"![portrait](/img/images/about.png)"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := pool.Exec(ctx, "TRUNCATE pg_temp.posts, pg_temp.about"); err != nil {
				t.Fatal(err)
			}
			for _, posts := range []struct {
				contents  []string
				published bool
			}{{tc.published, true}, {tc.drafts, false}} {
				for _, content := range posts.contents {
					if _, err := pool.Exec(ctx, "INSERT INTO pg_temp.posts VALUES ($1, $2)", content, posts.published); err != nil {
						t.Fatal(err)
					}
				}
			}
			for _, content := range tc.about {
				if _, err := pool.Exec(ctx, "INSERT INTO pg_temp.about VALUES ($1)", content); err != nil {
					t.Fatal(err)
				}
			}

			got, err := pg.GetAllContent(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := slices.Concat(tc.published, tc.drafts, tc.about)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("content = %q, want %q", got, want)
			}
		})
	}
}
