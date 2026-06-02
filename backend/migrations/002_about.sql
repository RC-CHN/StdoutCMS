-- STDOUT_CMS_ELF — 002: standalone About table
-- 从 posts 表中分离 about 页面到独立表

CREATE TABLE IF NOT EXISTS about (
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(500) NOT NULL DEFAULT '',
    content    TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 从旧 posts 迁移数据（仅当 about 表为空时）
INSERT INTO about (title, content, updated_at)
SELECT title, content, updated_at
FROM posts
WHERE slug = 'about'
  AND NOT EXISTS (SELECT 1 FROM about);

-- 清理 posts 中旧的 about 行
DELETE FROM posts WHERE slug = 'about';
