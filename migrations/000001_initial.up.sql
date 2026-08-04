CREATE TABLE communities (

                             id BIGSERIAL PRIMARY KEY,

                             name TEXT NOT NULL,

                             description TEXT NOT NULL,

                             slug TEXT NOT NULL,

                             external_source TEXT NOT NULL DEFAULT 'internal',

                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);;

CREATE INDEX idx_communities_name
    ON communities(name);



CREATE TABLE resources (

                           id BIGSERIAL PRIMARY KEY,

                           title TEXT NOT NULL,

                           description TEXT NOT NULL,

                           slug TEXT NOT NULL,

                           category TEXT NOT NULL,

                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resources_category
    ON resources(category);