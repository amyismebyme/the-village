DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'resources'
          AND column_name = 'url'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'resources'
          AND column_name = 'slug'
    ) THEN
ALTER TABLE resources
    RENAME COLUMN url TO slug;
END IF;
END
$$;