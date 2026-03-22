-- +goose Up

-- profiles is the canonical user identity record.
-- All other tables reference it via user_id.
-- Created first because every other table holds a FK to it.
CREATE TABLE users.profiles (
    id           uuid        PRIMARY KEY,
    username     text        NOT NULL,
    display_name text        NOT NULL DEFAULT '',
    bio          text        NOT NULL DEFAULT '',
    avatar_url   text        NOT NULL DEFAULT '',
    verified     boolean     NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT profiles_username_len      CHECK (char_length(username)     BETWEEN 1 AND 50),
    CONSTRAINT profiles_display_name_len  CHECK (char_length(display_name) <= 100),
    CONSTRAINT profiles_bio_len           CHECK (char_length(bio)          <= 500),
    CONSTRAINT profiles_username_format   CHECK (username ~ '^[A-Za-z0-9_]+$')
);

CREATE UNIQUE INDEX profiles_username_idx ON users.profiles (lower(username));

-- credentials holds the email/password pair for local auth.
-- One credential row per user (unique user_id).
CREATE TABLE users.credentials (
    id            uuid        PRIMARY KEY,
    user_id       uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    email         text        NOT NULL,
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT credentials_email_len     CHECK (char_length(email) <= 254),
    CONSTRAINT credentials_user_id_unique UNIQUE (user_id)
);

CREATE UNIQUE INDEX credentials_email_idx ON users.credentials (lower(email));

-- oauth_links stores third-party OAuth identities bound to a profile.
-- A user may link multiple providers; a given provider+provider_id pair
-- is globally unique.
CREATE TABLE users.oauth_links (
    id          uuid        PRIMARY KEY,
    user_id     uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    provider    text        NOT NULL,
    provider_id text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT oauth_links_provider_not_empty    CHECK (char_length(provider)    > 0),
    CONSTRAINT oauth_links_provider_id_not_empty CHECK (char_length(provider_id) > 0),
    CONSTRAINT oauth_links_provider_unique UNIQUE (provider, provider_id)
);

CREATE INDEX oauth_links_user_id_idx ON users.oauth_links (user_id);

-- sessions tracks refresh token hashes for the token-rotation auth flow.
-- Access tokens are short-lived JWTs and are not stored here.
-- Raw tokens are never persisted; only SHA-256 hashes are stored.
CREATE TABLE users.sessions (
    id                 uuid        PRIMARY KEY,
    user_id            uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    refresh_token_hash text        NOT NULL,
    expires_at         timestamptz NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT sessions_refresh_token_hash_unique UNIQUE (refresh_token_hash)
);

CREATE INDEX sessions_user_id_idx   ON users.sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON users.sessions (expires_at);

-- roles assigns an elevated role to a user.
-- Users with no row here are ordinary users (lowest tier).
-- Hierarchy: (no row) < moderator < admin
-- Only one role row per user is permitted (unique user_id).
CREATE TABLE users.roles (
    id         uuid        PRIMARY KEY,
    user_id    uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    role       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT roles_role_valid     CHECK (role IN ('moderator', 'admin')),
    CONSTRAINT roles_user_id_unique UNIQUE (user_id)
);

CREATE INDEX roles_role_idx ON users.roles (role);

-- follows records directed follow relationships between users.
-- The composite PK prevents duplicate follows.
-- A user cannot follow themselves.
CREATE TABLE users.follows (
    follower_id  uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    following_id uuid        NOT NULL REFERENCES users.profiles (id) ON DELETE CASCADE,
    created_at   timestamptz NOT NULL DEFAULT now(),

    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT follows_no_self_follow CHECK (follower_id <> following_id)
);

-- Lookup "who does this user follow?" (feed fan-out, following list)
CREATE INDEX follows_follower_id_idx  ON users.follows (follower_id);
-- Lookup "who follows this user?" (follower list, notification targets)
CREATE INDEX follows_following_id_idx ON users.follows (following_id);

-- +goose Down

DROP TABLE IF EXISTS users.follows;
DROP TABLE IF EXISTS users.roles;
DROP TABLE IF EXISTS users.sessions;
DROP TABLE IF EXISTS users.oauth_links;
DROP TABLE IF EXISTS users.credentials;
DROP TABLE IF EXISTS users.profiles;
