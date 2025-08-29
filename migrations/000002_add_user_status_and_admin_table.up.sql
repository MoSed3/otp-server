ALTER TABLE users ADD COLUMN status INT NOT NULL DEFAULT 1;

CREATE TABLE admins (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    username VARCHAR(255) UNIQUE NOT NULL,
    role INT NOT NULL DEFAULT 3,
    hashed_password VARCHAR(255) NOT NULL,
    password_reset_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_admins_username ON admins (username);
