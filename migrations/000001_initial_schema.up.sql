CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100)
);

CREATE INDEX idx_user_phone_number ON users (phone_number);

CREATE TABLE user_otps (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) NOT NULL,
    user_id BIGINT NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_user_otps_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    secret_key VARCHAR(255) NOT NULL,
    access_token_expire BIGINT NOT NULL
);

INSERT INTO settings (secret_key, access_token_expire) VALUES ('', 1440);
