package store

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestImageCleanupMigration(t *testing.T) {
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
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	schema := fmt.Sprintf("test_image_cleanup_%d", conn.PgConn().PID())
	exec("CREATE SCHEMA " + schema)
	exec("SET LOCAL search_path TO " + schema)
	exec(`CREATE TABLE posts (id integer PRIMARY KEY, content text NOT NULL);
		CREATE TABLE about (id integer PRIMARY KEY, content text NOT NULL)`)
	migration, err := os.ReadFile("../../migrations/003_image_cleanup.sql")
	if err != nil {
		t.Fatal(err)
	}
	const (
		removed       = "images/00000000-0000-4000-8000-000000000001.png"
		shared        = "images/00000000-0000-4000-8000-000000000002.jpg"
		kept          = "images/00000000-0000-4000-8000-000000000003.webp"
		newKey        = "images/00000000-0000-4000-8000-000000000004.gif"
		extensionless = "images/00000000-0000-4000-8000-000000000006"
		invalidBase   = "images/00000000-0000-4000-8000-000000000007"
	)
	post := "![](/img/" + shared + ") ![](https://cdn.example/" + removed + "?v=1) " +
		removed + " " + kept + " images/custom.png other-images/00000000-0000-4000-8000-000000000005.png"
	exec("INSERT INTO posts VALUES (1, $1)", post)
	exec("INSERT INTO about VALUES (1, $1)", shared)
	exec(string(migration))

	assertCandidates := func(want ...string) {
		t.Helper()
		rows, err := tx.Query(ctx, "SELECT object_key FROM image_cleanup_candidates ORDER BY object_key")
		if err != nil {
			t.Fatal(err)
		}
		got, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("image cleanup candidates = %v, want %v", got, want)
		}
	}
	exec("INSERT INTO posts VALUES (2, $1)", newKey)
	assertCandidates() // Neither installing nor adding an image makes it a candidate.
	exec("UPDATE posts SET content = $1 WHERE id = 1", "kept "+kept)
	assertCandidates(removed, shared) // The cleanup worker checks shared references before deletion.
	exec(string(migration))
	assertCandidates(removed, shared)
	// A worker may already have dismissed the shared image while About still uses it.
	exec("DELETE FROM image_cleanup_candidates WHERE object_key = $1", shared)
	exec("SAVEPOINT before_about_edit")
	exec("UPDATE about SET content = ''")
	assertCandidates(removed, shared)
	exec("ROLLBACK TO before_about_edit")
	exec("UPDATE about SET content = $1", "Same picture: "+shared)
	assertCandidates(removed)
	exec("DELETE FROM about")
	assertCandidates(removed, shared)
	exec("SAVEPOINT before_edit")
	exec("UPDATE posts SET content = '' WHERE id = 2")
	assertCandidates(removed, shared, newKey)
	exec("ROLLBACK TO before_edit")
	assertCandidates(removed, shared)
	exec("DELETE FROM posts WHERE id = 1")
	assertCandidates(removed, shared, kept)
	oldPaths := "![](/img/" + extensionless + ") " +
		invalidBase + "/nested.png " + invalidBase + ".png/thumb " +
		invalidBase + ".x-custom " + invalidBase + ".a.b " + invalidBase + "%2Ffile"
	exec("UPDATE posts SET content = $1 WHERE id = 2", oldPaths)
	assertCandidates(removed, shared, kept, newKey)
	exec("DELETE FROM posts WHERE id = 2")
	assertCandidates(removed, shared, kept, newKey, extensionless)
}
