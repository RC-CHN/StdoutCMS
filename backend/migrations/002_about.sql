-- STDOUT_CMS_ELF — 002: standalone About table
-- 从 posts 表中分离 about 页面到独立表

DO $$
BEGIN
    -- Serialize concurrent startups before checking whether 002 already ran.
    PERFORM pg_advisory_xact_lock(hashtext('stdoutcms.002_about'));
    IF to_regclass('about') IS NOT NULL THEN
        RETURN;
    END IF;

    CREATE TABLE about (
        id         SERIAL PRIMARY KEY,
        title      VARCHAR(500) NOT NULL DEFAULT '',
        content    TEXT NOT NULL DEFAULT '',
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

    -- Only the pre-002 schema used an ordinary post for the About page.
    WITH legacy_page AS (
        DELETE FROM posts WHERE slug = 'about'
        RETURNING title, content, updated_at
    )
    INSERT INTO about (title, content, updated_at)
    SELECT title, content, updated_at FROM legacy_page;
END;
$$;
