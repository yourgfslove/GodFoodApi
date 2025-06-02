-- +goose Up
ALTER TABLE orders
ADD COLUMN address TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE orders
DROP COLUMN address;