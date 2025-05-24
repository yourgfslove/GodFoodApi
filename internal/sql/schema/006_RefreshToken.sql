-- +goose Up

CREATE TABLE IF NOT EXISTS refreshTokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP
);

-- +goose Down

DROP TABLE IF EXISTS refreshTokens;