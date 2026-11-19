package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestAboutMigration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	migration, err := os.ReadFile("../../migrations/002_about.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name      string
		wantPages int
	}{
		{"fresh installation", 0},
		{"upgrade with legacy About post", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := conn.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			exec := func(query string) {
				t.Helper()
				if _, err := tx.Exec(ctx, query); err != nil {
					t.Fatal(err)
				}
			}

			exec(`CREATE TEMP TABLE posts (
				slug text PRIMARY KEY, title text, content text, updated_at timestamptz
			);
			SET LOCAL search_path TO pg_temp;`)
			if tc.wantPages == 1 {
				exec(`INSERT INTO posts VALUES ('about', 'Legacy', 'Original text', '2020-01-01T12:00:00Z')`)
			}
			exec(string(migration))

			// After either upgrade, "about" is a valid slug for a new article.
			exec(`INSERT INTO posts VALUES ('about', 'Article', 'Keep this article', now())`)
			exec(string(migration))

			var article string
			var pages, legacyPages int
			err = tx.QueryRow(ctx, `SELECT
				(SELECT content FROM posts WHERE slug = 'about'),
				count(*),
				count(*) FILTER (WHERE id = 1 AND title = 'Legacy' AND content = 'Original text'
					AND updated_at = '2020-01-01T12:00:00Z')
				FROM about`).Scan(&article, &pages, &legacyPages)
			if err != nil {
				t.Fatal(err)
			}
			if article != "Keep this article" || pages != tc.wantPages || legacyPages != tc.wantPages {
				t.Errorf("after restart: article = %q, About pages = %d, legacy copies = %d",
					article, pages, legacyPages)
			}
		})
	}
}
