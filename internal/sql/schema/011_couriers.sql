-- +goose Up
CREATE TABLE IF NOT EXISTS CouriersStats (
    id int NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    orderCount int DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS  CouriersStats;