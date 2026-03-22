-- +goose Up
-- +goose StatementBegin

-- Seed development accounts: admin, moderator, and regular user.
-- Guarded by a runtime check: only executes when current_setting('app.env')
-- is 'development' (the default). Production deployments must set this to
-- 'production' to skip seeding.
--
-- The INSERT ... ON CONFLICT DO NOTHING guards against re-runs.

DO $$
DECLARE
    v_env text;
BEGIN
    -- Check environment. Default to 'development' if not set.
    BEGIN
        v_env := current_setting('app.env');
    EXCEPTION WHEN undefined_object THEN
        v_env := 'development';
    END;

    IF v_env <> 'development' THEN
        RAISE NOTICE '002_seed_dev_accounts: skipping seed (app.env=%)', v_env;
        RETURN;
    END IF;

    -- Admin: admin@soapbox.dev / admin123
    INSERT INTO users.profiles (id, username, display_name, bio, verified)
    VALUES ('a0000000-0000-7000-8000-000000000001', 'admin', 'Admin', 'Soapbox administrator', true)
    ON CONFLICT DO NOTHING;

    INSERT INTO users.credentials (id, user_id, email, password_hash)
    VALUES ('a0000000-0000-7000-8000-000000000002', 'a0000000-0000-7000-8000-000000000001',
            'admin@soapbox.dev', '$2a$10$/QjsfM9VeDP9heFtTTnZOu1QDiaF7V.JbmH80wf8LJHWWiBN9rAha')
    ON CONFLICT (user_id) DO NOTHING;

    INSERT INTO users.roles (id, user_id, role)
    VALUES ('a0000000-0000-7000-8000-000000000003', 'a0000000-0000-7000-8000-000000000001', 'admin')
    ON CONFLICT (user_id) DO NOTHING;

    -- Moderator: mod@soapbox.dev / mod12345
    INSERT INTO users.profiles (id, username, display_name, bio, verified)
    VALUES ('a0000000-0000-7000-8000-000000000010', 'moderator', 'Moderator', 'Community moderator', false)
    ON CONFLICT DO NOTHING;

    INSERT INTO users.credentials (id, user_id, email, password_hash)
    VALUES ('a0000000-0000-7000-8000-000000000011', 'a0000000-0000-7000-8000-000000000010',
            'mod@soapbox.dev', '$2a$10$g5EyqOeihyWroqhlL5WAROL4hAwbe7oX4GxZmeNejPPK/82OGjpTu')
    ON CONFLICT (user_id) DO NOTHING;

    INSERT INTO users.roles (id, user_id, role)
    VALUES ('a0000000-0000-7000-8000-000000000012', 'a0000000-0000-7000-8000-000000000010', 'moderator')
    ON CONFLICT (user_id) DO NOTHING;

    -- Regular user: user@soapbox.dev / user1234
    INSERT INTO users.profiles (id, username, display_name, bio, verified)
    VALUES ('a0000000-0000-7000-8000-000000000020', 'testuser', 'Test User', 'Just a regular user', false)
    ON CONFLICT DO NOTHING;

    INSERT INTO users.credentials (id, user_id, email, password_hash)
    VALUES ('a0000000-0000-7000-8000-000000000021', 'a0000000-0000-7000-8000-000000000020',
            'user@soapbox.dev', '$2a$10$flUXZulV9BVVeuMHUFtK9et8.gDiOWAZKuHx6FU3q9Fb79TzHKLiS')
    ON CONFLICT (user_id) DO NOTHING;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
BEGIN
    DELETE FROM users.profiles WHERE username IN ('admin', 'moderator', 'testuser');
END;
$$;
-- +goose StatementEnd
