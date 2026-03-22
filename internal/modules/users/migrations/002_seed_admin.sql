-- +goose Up

-- Seed the dev admin account.
-- Credentials: email=admin@soapbox.dev, password=admin123 (bcrypt cost 10)
--
-- The INSERT ... ON CONFLICT DO NOTHING guards against re-runs in environments
-- where migrations are applied to a pre-populated database (e.g. snapshot restores).
--
-- IMPORTANT: Change this password before deploying to any non-development environment.

DO $$
DECLARE
    v_user_id uuid := gen_random_uuid();
BEGIN
    -- Create the profile for the admin user.
    INSERT INTO users.profiles (id, username, display_name, bio, verified)
    VALUES (v_user_id, 'admin', 'Admin', 'Soapbox administrator', true)
    ON CONFLICT DO NOTHING;

    -- Resolve the actual id in case the profile already existed.
    -- (The ON CONFLICT above means v_user_id may not be the real id.)
    SELECT id INTO v_user_id
    FROM users.profiles
    WHERE lower(username) = 'admin';

    -- Attach email/password credentials.
    INSERT INTO users.credentials (user_id, email, password_hash)
    VALUES (
        v_user_id,
        'admin@soapbox.dev',
        '$2b$10$6FJatNYVsjxi2MFkZGjJj.MYcPdz9oBGFfy80EkwvUvlE2fezHxfG'
    )
    ON CONFLICT (user_id) DO NOTHING;

    -- Grant admin role.
    INSERT INTO users.roles (user_id, role)
    VALUES (v_user_id, 'admin')
    ON CONFLICT (user_id) DO NOTHING;
END;
$$;

-- +goose Down

DO $$
DECLARE
    v_user_id uuid;
BEGIN
    SELECT id INTO v_user_id
    FROM users.profiles
    WHERE lower(username) = 'admin';

    IF v_user_id IS NOT NULL THEN
        -- Cascade handles credentials, sessions, oauth_links, roles, follows.
        DELETE FROM users.profiles WHERE id = v_user_id;
    END IF;
END;
$$;
