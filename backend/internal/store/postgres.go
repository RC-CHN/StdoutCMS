package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (p *Postgres) CountPosts() (int, error) {
	ctx := context.Background()
	var n int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM postsWHERE published = true").Scan(&n)
	return n, err
}

func (p *Postgres) CountProjects() (int, error) {
	ctx := context.Background()
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

func (p *Postgres) ListPosts(page, size int) ([]Post, int, error) {
	ctx := context.Background()

	var total int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM postsWHERE published = true").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := p.pool.Query(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, created_at
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
			&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.CreatedAt)
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

func (p *Postgres) GetPostBySlug(slug string) (*Post, error) {
	ctx := context.Background()
	var po Post
	err := p.pool.QueryRow(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, created_at, updated_at
		FROM posts
		WHERE slug = $1 AND published = true
	`, slug).Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
		&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (p *Postgres) GetPostBySlugAdmin(slug string) (*Post, error) {
	ctx := context.Background()
	var po Post
	err := p.pool.QueryRow(ctx, `
		SELECT slug, title, content, excerpt, tags, author, word_count, read_time, published, created_at, updated_at
		FROM posts
		WHERE slug = $1
	`, slug).Scan(&po.Slug, &po.Title, &po.Content, &po.Excerpt,
		&po.Tags, &po.Author, &po.WordCount, &po.ReadTime, &po.Published, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (p *Postgres) ListPostsAdmin(page, size int) ([]Post, int, error) {
	ctx := context.Background()

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

func (p *Postgres) CreatePost(po *Post) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		INSERT INTO posts (slug, title, content, excerpt, tags, author, word_count, read_time, published)
		VALUES ($1,$2,$3,$4,$5::text[],$6,$7,$8,$9)
	`, po.Slug, po.Title, po.Content, po.Excerpt, po.Tags, po.Author, po.WordCount, po.ReadTime, po.Published)
	return err
}

func (p *Postgres) UpdatePost(slug string, po *Post) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		UPDATE posts
		SET title=$1, content=$2, excerpt=$3, tags=$4::text[], word_count=$5, read_time=$6, published=$7, updated_at=now()
		WHERE slug=$8
	`, po.Title, po.Content, po.Excerpt, po.Tags, po.WordCount, po.ReadTime, po.Published, slug)
	return err
}

func (p *Postgres) DeletePost(slug string) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `DELETE FROM posts WHERE slug = $1`, slug)
	return err
}

// ---- Projects ----

func (p *Postgres) ListProjects() ([]Project, error) {
	ctx := context.Background()
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
func (p *Postgres) ListProjectsAdmin() ([]Project, error) {
	return p.ListProjects()
}

func (p *Postgres) GetProject(id int) (*Project, error) {
	ctx := context.Background()
	var pr Project
	err := p.pool.QueryRow(ctx, `
		SELECT id, name, description, lang, status, url, sort_order
		FROM projects WHERE id=$1
	`, id).Scan(&pr.ID, &pr.Name, &pr.Description, &pr.Lang, &pr.Status, &pr.URL, &pr.SortOrder)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (p *Postgres) CreateProject(pr *Project) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		INSERT INTO projects (name, description, lang, status, url, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, pr.Name, pr.Description, pr.Lang, pr.Status, pr.URL, pr.SortOrder)
	return err
}

func (p *Postgres) UpdateProject(id int, pr *Project) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		UPDATE projects SET name=$1, description=$2, lang=$3, status=$4, url=$5, sort_order=$6
		WHERE id=$7
	`, pr.Name, pr.Description, pr.Lang, pr.Status, pr.URL, pr.SortOrder, id)
	return err
}

func (p *Postgres) DeleteProject(id int) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return err
}

// GetAllContent returns the content of every post (including drafts).
// Used by orphan-image cleanup to check which images are still referenced.
func (p *Postgres) GetAllContent() ([]string, error) {
	ctx := context.Background()
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

func (p *Postgres) GetAdminByUsername(username string) (*Admin, error) {
	ctx := context.Background()
	var a Admin
	err := p.pool.QueryRow(ctx, `
		SELECT id, username, password_hash FROM admins WHERE username = $1
	`, username).Scan(&a.ID, &a.Username, &a.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *Postgres) SeedAdmin(username, passwordHash string) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = $2
	`, username, passwordHash)
	return err
}

// ExecMigration runs SQL only if the posts table doesn't exist yet.
func (p *Postgres) ExecMigration(sql string) error {
	ctx := context.Background()
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
func (p *Postgres) ExecRaw(sql string) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, sql)
	return err
}

// ---- About ----

func (p *Postgres) GetAbout() (*About, error) {
	ctx := context.Background()
	var a About
	err := p.pool.QueryRow(ctx, `
		SELECT title, content, updated_at FROM about ORDER BY id LIMIT 1
	`).Scan(&a.Title, &a.Content, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *Postgres) UpdateAbout(title, content string) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx, `
		INSERT INTO about (id, title, content, updated_at)
		VALUES (1, $1, $2, now())
		ON CONFLICT (id) DO UPDATE SET title = $1, content = $2, updated_at = now()
	`, title, content)
	return err
}
