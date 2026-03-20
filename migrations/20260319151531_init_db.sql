-- +goose Up
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS links (
short_url CHAR(10) PRIMARY KEY,
original_url TEXT NOT NULL UNIQUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS links;