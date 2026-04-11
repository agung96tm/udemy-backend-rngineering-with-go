CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description TEXT
);

INSERT INTO roles (name, level, description)
VALUES
    ('user', 1, 'A user can create posts and comments'),
    ('moderator', 2, 'A moderator can update other user posts'),
    ('admin', 3, 'An Admin can update and delete other user posts');

