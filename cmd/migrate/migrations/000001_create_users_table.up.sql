CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     name VARCHAR(255) NOT NULL,
                                     email VARCHAR(255) UNIQUE NOT NULL,
                                     password TEXT NOT NULL,
                                     is_active boolean default false,
                                     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
