-- name: CreateUser :one
INSERT INTO users (email, hash_password, user_role, phone, created_at, address, user_name)
VALUES (
        $1,
        $2,
        $3,
        $4,
        NOW(),
        $5,
        $6
)
RETURNING *;


-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email=$1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id=$1;

-- name: GetUsersByRole :many
SELECT * FROM users
WHERE user_role=$1;

-- name: GetRestaurantAndMenuByID :many
SELECT
    users.id AS restaurant_id,
    users.user_name AS restaurant_name,
    users.address AS restaurant_address,
    users.phone AS restaurant_phone,

    menuitem.id AS menu_item_ID,
    menuitem.name AS menu_item_name,
    menuitem.price,
    menuitem.description,
    menuitem.available

FROM users
JOIN menuitem ON users.id = menuitem.restaurant_id
WHERE users.id = $1 AND users.user_role = 'restaurant';


-- name: GetNameByID :one
SELECT user_name FROM users
WHERE id=$1;

