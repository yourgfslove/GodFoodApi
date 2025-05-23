-- +goose Up

ALTER TABLE users
ADD COLUMN phone TEXT NOT NULL default 'inset';

-- +goose Down

ALTER TABLE users
DROP COLUMN phone;