CREATE TABLE IF NOT EXISTS user_invitations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token bytea NOT NULL UNIQUE,

    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
