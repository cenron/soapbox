-- +goose Up
CREATE TABLE media.uploads (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL,
    file_key     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size         BIGINT NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_uploads_user_id ON media.uploads (user_id);
CREATE INDEX idx_uploads_status ON media.uploads (status);

-- +goose Down
DROP TABLE IF EXISTS media.uploads;
