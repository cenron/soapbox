-- +goose Up
CREATE INDEX idx_posts_body_fts ON posts.posts USING GIN (to_tsvector('english', body));

-- +goose Down
DROP INDEX IF EXISTS posts.idx_posts_body_fts;
