CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    username varchar(200) UNIQUE NOT NULL,
    password bytea NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,

    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);