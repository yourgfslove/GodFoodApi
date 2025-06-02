-- +goose Up
ALTER TABLE users
ALTER COLUMN user_role
SET DEFAULT 'customer';

-- +goose Down
ALTER TABLE users
ALTER COLUMN user_role
DROP DEFAULT;
