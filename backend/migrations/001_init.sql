-- SYS_BLOG.EXE — 初始建表
-- 回滚: DROP TABLE IF EXISTS posts, projects, admins CASCADE;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 文章表
CREATE TABLE IF NOT EXISTS posts (
    id         SERIAL PRIMARY KEY,
    slug       VARCHAR(255) UNIQUE NOT NULL,
    title      VARCHAR(500) NOT NULL,
    content    TEXT NOT NULL,
    excerpt    TEXT NOT NULL DEFAULT '',
    tags       TEXT[] NOT NULL DEFAULT '{}',
    author     VARCHAR(100) NOT NULL DEFAULT 'root',
    word_count INT NOT NULL DEFAULT 0,
    read_time  VARCHAR(20) NOT NULL DEFAULT '',
    published  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_posts_slug      ON posts (slug);
CREATE INDEX IF NOT EXISTS idx_posts_published ON posts (published) WHERE published = true;
CREATE INDEX IF NOT EXISTS idx_posts_created   ON posts (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_tags      ON posts USING gin (tags);

-- 全文搜索（中英文混合）
CREATE INDEX IF NOT EXISTS idx_posts_content_trgm ON posts USING gin (content gin_trgm_ops);

-- 项目表
CREATE TABLE IF NOT EXISTS projects (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    lang        VARCHAR(100) NOT NULL DEFAULT '',
    status      VARCHAR(50) NOT NULL DEFAULT 'active',
    url         VARCHAR(500) NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_sort ON projects (sort_order);

-- admin 用户表
CREATE TABLE IF NOT EXISTS admins (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
