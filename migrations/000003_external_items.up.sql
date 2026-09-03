CREATE TABLE external_items (
                                id BIGSERIAL PRIMARY KEY,

                                source TEXT NOT NULL,

                                external_id TEXT NOT NULL,

                                title TEXT NOT NULL,

                                description TEXT NOT NULL DEFAULT '',

                                url TEXT NOT NULL DEFAULT '',

                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                CONSTRAINT external_items_identity_unique
                                    UNIQUE (source, external_id)
);

CREATE INDEX idx_external_items_source
    ON external_items(source);