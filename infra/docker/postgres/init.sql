CREATE TABLE IF NOT EXISTS communities (
                                           id UUID PRIMARY KEY,
                                           name TEXT NOT NULL,
                                           description TEXT NOT NULL,
                                           created_at TIMESTAMP NOT NULL
);