-- +goose Up
CREATE SCHEMA IF NOT EXISTS posts;

CREATE TABLE posts.posts (
    id            UUID        PRIMARY KEY,
    author_id     UUID        NOT NULL,
    author_username    VARCHAR(255) NOT NULL,
    author_display_name VARCHAR(255) NOT NULL,
    author_avatar_url  TEXT        NOT NULL DEFAULT '',
    author_verified    BOOLEAN     NOT NULL DEFAULT FALSE,
    body          TEXT        NOT NULL,
    parent_id     UUID        REFERENCES posts.posts (id) ON DELETE SET NULL,
    repost_of_id  UUID        REFERENCES posts.posts (id) ON DELETE SET NULL,
    like_count    INTEGER     NOT NULL DEFAULT 0,
    repost_count  INTEGER     NOT NULL DEFAULT 0,
    reply_count   INTEGER     NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_posts_author_id  ON posts.posts (author_id, created_at DESC);
CREATE INDEX idx_posts_parent_id  ON posts.posts (parent_id, created_at ASC) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_posts_created_at ON posts.posts (created_at DESC);

CREATE TABLE posts.media (
    id         UUID    PRIMARY KEY,
    post_id    UUID    NOT NULL REFERENCES posts.posts (id) ON DELETE CASCADE,
    media_url  TEXT    NOT NULL,
    media_type VARCHAR(50) NOT NULL DEFAULT '',
    position   INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_media_post_id ON posts.media (post_id, position);

CREATE TABLE posts.link_previews (
    id          UUID PRIMARY KEY,
    post_id     UUID NOT NULL UNIQUE REFERENCES posts.posts (id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    title       TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    image_url   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE posts.hashtags (
    post_id UUID        NOT NULL REFERENCES posts.posts (id) ON DELETE CASCADE,
    tag     VARCHAR(255) NOT NULL,
    PRIMARY KEY (post_id, tag)
);

CREATE INDEX idx_hashtags_tag ON posts.hashtags (lower(tag));

CREATE TABLE posts.likes (
    post_id    UUID        NOT NULL REFERENCES posts.posts (id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_likes_user_id ON posts.likes (user_id, created_at DESC);

-- +goose Down
DROP SCHEMA IF EXISTS posts CASCADE;
