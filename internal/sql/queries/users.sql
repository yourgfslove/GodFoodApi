-- name: CreateUser :one
INSERT INTO users (email, hash_password, user_role, phone, created_at)
VALUES (
        $1,
        $2,
        $3,
        $4,
        NOW()
)
RETURNING *;


-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email=$1;