-- name: GetMenu :many
SELECT * FROM menuitem
WHERE restaurant_id=$1;

-- name: CreateMenuItem :one
INSERT INTO menuitem (restaurant_id, name, price, description, available)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
)
RETURNING *;