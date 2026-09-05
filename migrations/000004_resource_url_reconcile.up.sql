DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'resources'
          AND column_name = 'slug'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'resources'
          AND column_name = 'url'
    ) THEN
ALTER TABLE resources
    RENAME COLUMN slug TO url;
END IF;
END
$$;