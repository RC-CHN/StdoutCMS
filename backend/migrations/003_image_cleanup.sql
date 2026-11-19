-- An upload may exist only in a browser draft. Only enqueue keys that an
-- author has removed from content previously saved on this server.
SELECT pg_advisory_xact_lock(hashtext('stdoutcms.003_image_cleanup'));

CREATE TABLE IF NOT EXISTS image_cleanup_candidates (
    object_key text PRIMARY KEY
);

CREATE OR REPLACE FUNCTION queue_removed_images() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
    remaining text := '';
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF NEW.content = OLD.content THEN
            RETURN NULL;
        END IF;
        remaining := NEW.content;
    END IF;

    INSERT INTO image_cleanup_candidates (object_key)
    SELECT matched[1]
    -- Some legacy MIME types yielded a key without a file extension.
    -- Require a link or text delimiter so a longer path cannot be truncated.
    FROM regexp_matches(OLD.content,
        $key$(?<![[:alnum:]_.-])images/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}(?:\.[a-zA-Z0-9]+)?(?=$|[[:space:]?#"'<>)]|\])$key$,
        'g') AS keys(matched)
    WHERE strpos(remaining, matched[1]) = 0
    ORDER BY 1
    ON CONFLICT DO NOTHING;
    RETURN NULL;
END;
$$;

CREATE OR REPLACE TRIGGER queue_removed_post_images
AFTER UPDATE OF content OR DELETE ON posts
FOR EACH ROW EXECUTE FUNCTION queue_removed_images();

CREATE OR REPLACE TRIGGER queue_removed_about_images
AFTER UPDATE OF content OR DELETE ON about
FOR EACH ROW EXECUTE FUNCTION queue_removed_images();
