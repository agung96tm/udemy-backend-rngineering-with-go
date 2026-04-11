ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_id_fkey;

ALTER TABLE users
    DROP COLUMN IF EXISTS role_id;