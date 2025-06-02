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