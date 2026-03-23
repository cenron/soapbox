-- +goose Up
CREATE SCHEMA IF NOT EXISTS feed;

CREATE TABLE feed.timelines (
    user_id    UUID        NOT NULL,
    post_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX idx_timelines_user_timeline ON feed.timelines (user_id, created_at DESC);

-- +goose Down
DROP SCHEMA IF EXISTS feed CASCADE;
