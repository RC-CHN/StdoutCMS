package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a query matches or affects no rows.
// Handlers map it to HTTP 404; anything else is a genuine storage error.
var ErrNotFound = errors.New("record not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(dsn string) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

// ---- Stats ----

func (p *Postgres) CountPosts(ctx context.Context) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM posts WHERE published = true").Scan(&n)
	return n, err
}

func (p *Postgres) CountProjects(ctx context.Context) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM projects").Scan(&n)
	return n, err
}

func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// ---- Posts ----

func (p *Postgres) ListPosts(ctx context.Context, page, size int) ([]Post, int, error) {
	var total int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM posts WHERE published = true").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := p.pool.Query(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, published, created_at, updated_at
		FROM posts
		WHERE published = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var po Post
		err := rows.Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
			&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.Published, &po.CreatedAt, &po.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, po)
	}
	if posts == nil {
		posts = make([]Post, 0)
	}
	return posts, total, rows.Err()
}

func (p *Postgres) GetPostBySlug(ctx context.Context, slug string) (*Post, error) {
	var po Post
	err := p.pool.QueryRow(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, published, created_at, updated_at
		FROM posts
		WHERE slug = $1 AND published = true
	`, slug).Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
		&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.Published, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &po, nil
}

func (p *Postgres) GetPostBySlugAdmin(ctx context.Context, slug string) (*Post, error) {
	var po Post
	err := p.pool.QueryRow(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, published, created_at, updated_at
		FROM posts
		WHERE slug = $1
	`, slug).Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
		&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.Published, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &po, nil
}

func (p *Postgres) ListPostsAdmin(ctx context.Context, page, size int) ([]Post, int, error) {
	var total int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM posts").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := p.pool.Query(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, published, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var po Post
		err := rows.Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
			&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.Published, &po.CreatedAt, &po.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, po)
	}
	if posts == nil {
		posts = make([]Post, 0)
	}
	return posts, total, rows.Err()
}

func (p *Postgres) CreatePost(ctx context.Context, po *Post) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO posts (slug, title, content, excerpt, tags, author, word_count, read_time, published)
		VALUES ($1,$2,$3,$4,$5::text[],$6,$7,$8,$9)
	`, po.Slug, po.Title, po.Content, po.Excerpt, po.Tags, po.Author, po.WordCount, po.ReadTime, po.Published)
	return err
}

func (p *Postgres) UpdatePost(ctx context.Context, slug string, po *Post) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE posts
		SET title=$1, content=$2, excerpt=$3, tags=$4::text[], word_count=$5, read_time=$6, published=$7, updated_at=now()
		WHERE slug=$8
	`, po.Title, po.Content, po.Excerpt, po.Tags, po.WordCount, po.ReadTime, po.Published, slug)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeletePost(ctx context.Context, slug string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM posts WHERE slug = $1`, slug)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Projects ----

func (p *Postgres) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, name, description, lang, status, url, sort_order
		FROM projects
		ORDER BY sort_order, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var pr Project
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.Description, &pr.Lang, &pr.Status, &pr.URL, &pr.SortOrder); err != nil {
			return nil, err
		}
		projects = append(projects, pr)
	}
	if projects == nil {
		projects = make([]Project, 0)
	}
	return projects, rows.Err()
}

// ListProjectsAdmin returns all projects for the admin panel.
func (p *Postgres) ListProjectsAdmin(ctx context.Context) ([]Project, error) {
	return p.ListProjects(ctx)
}

func (p *Postgres) GetProject(ctx context.Context, id int) (*Project, error) {
	var pr Project
	err := p.pool.QueryRow(ctx, `
		SELECT id, name, description, lang, status, url, sort_order
		FROM projects WHERE id=$1
	`, id).Scan(&pr.ID, &pr.Name, &pr.Description, &pr.Lang, &pr.Status, &pr.URL, &pr.SortOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &pr, nil
}

func (p *Postgres) CreateProject(ctx context.Context, pr *Project) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO projects (name, description, lang, status, url, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, pr.Name, pr.Description, pr.Lang, pr.Status, pr.URL, pr.SortOrder)
	return err
}

func (p *Postgres) UpdateProject(ctx context.Context, id int, pr *Project) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE projects SET name=$1, description=$2, lang=$3, status=$4, url=$5, sort_order=$6
		WHERE id=$7
	`, pr.Name, pr.Description, pr.Lang, pr.Status, pr.URL, pr.SortOrder, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeleteProject(ctx context.Context, id int) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAllContent returns the content of every post (including drafts).
// Used by orphan-image cleanup to check which images are still referenced.
func (p *Postgres) GetAllContent(ctx context.Context) ([]string, error) {
	rows, err := p.pool.Query(ctx, `SELECT content FROM posts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		contents = append(contents, c)
	}
	return contents, rows.Err()
}

// ---- Admin ----

func (p *Postgres) GetAdminByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := p.pool.QueryRow(ctx, `
		SELECT id, username, password_hash FROM admins WHERE username = $1
	`, username).Scan(&a.ID, &a.Username, &a.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *Postgres) SeedAdmin(ctx context.Context, username, passwordHash string) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`, username, passwordHash)
	return err
}

// ExecMigration runs SQL only if the posts table doesn't exist yet.
func (p *Postgres) ExecMigration(ctx context.Context, sql string) error {
	var exists bool
	err := p.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'posts'
		)
	`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = p.pool.Exec(ctx, sql)
	return err
}

// ExecRaw runs SQL unconditionally (used for subsequent migrations).
func (p *Postgres) ExecRaw(ctx context.Context, sql string) error {
	_, err := p.pool.Exec(ctx, sql)
	return err
}

// ---- About ----

func (p *Postgres) GetAbout(ctx context.Context) (*About, error) {
	var a About
	err := p.pool.QueryRow(ctx, `
		SELECT title, content, updated_at FROM about ORDER BY id LIMIT 1
	`).Scan(&a.Title, &a.Content, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (p *Postgres) UpdateAbout(ctx context.Context, title, content string) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO about (id, title, content, updated_at)
		VALUES (1, $1, $2, now())
		ON CONFLICT (id) DO UPDATE SET title = $1, content = $2, updated_at = now()
	`, title, content)
	return err
}
