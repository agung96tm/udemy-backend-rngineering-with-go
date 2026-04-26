CREATE TABLE IF NOT EXISTS user_invitations (
    id SERIAL PRIMARY KEY,
    token BYTEA NOT NULL,
    user_id bigint NOT NULL REFERENCES users(id)
);
