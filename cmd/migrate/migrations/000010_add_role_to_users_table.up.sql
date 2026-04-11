ALTER TABLE users
    ADD COLUMN role_id INT;

UPDATE users u
SET role_id = r.id
FROM roles r
WHERE r.name = 'user';

ALTER TABLE users
    ADD CONSTRAINT users_role_id_fkey
        FOREIGN KEY (role_id) REFERENCES roles(id);

ALTER TABLE users
    ALTER COLUMN role_id SET NOT NULL;