package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *Postgres) ListImageCleanupCandidates(ctx context.Context) ([]string, error) {
	rows, err := p.pool.Query(ctx, "SELECT object_key FROM image_cleanup_candidates ORDER BY object_key")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// DeleteImageIfUnreferenced consumes a queued key. If any content still
// references it, removing that reference later will enqueue the key again.
func (p *Postgres) DeleteImageIfUnreferenced(ctx context.Context, key string, remove func(context.Context, string) error) (bool, error) {
	var removed bool
	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		// Writers lock content before their triggers enqueue keys. Take the
		// same lock order and keep references stable until storage replies.
		if _, err := tx.Exec(ctx, "LOCK TABLE posts, about IN SHARE MODE"); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, "DELETE FROM image_cleanup_candidates WHERE object_key = $1", key)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}

		var referenced bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM posts WHERE strpos(content, $1) > 0)
			    OR EXISTS (SELECT 1 FROM about WHERE strpos(content, $1) > 0)
		`, key).Scan(&referenced)
		if err != nil || referenced {
			return err
		}
		if err := remove(ctx, key); err != nil {
			return err
		}
		removed = true
		return nil
	})
	return removed && err == nil, err
}
