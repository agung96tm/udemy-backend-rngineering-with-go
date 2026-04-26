CREATE TABLE IF NOT EXISTS followers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_follow UNIQUE (user_id, follower_id),
    CONSTRAINT no_self_follow CHECK (user_id <> follower_id)
)