-- +goose Up
CREATE SCHEMA IF NOT EXISTS notifications;

CREATE TABLE notifications.notifications (
    id         UUID        NOT NULL PRIMARY KEY,
    user_id    UUID        NOT NULL,
    type       TEXT        NOT NULL,
    actor_id   UUID        NOT NULL,
    post_id    UUID,
    read       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_notifications_user_created ON notifications.notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread ON notifications.notifications (user_id) WHERE read = FALSE;

-- +goose Down
DROP SCHEMA IF EXISTS notifications CASCADE;
