-- +goose Up
ALTER TABLE users
ADD COLUMN user_name TEXT;

-- +goose Down
ALTER TABLE users
DROP COLUMN user_name;