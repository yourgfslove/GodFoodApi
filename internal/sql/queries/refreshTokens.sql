-- name: CreateToken :one
INSERT INTO refreshTokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1,
        NOW(),
        NOW(),
        $2,
        NOW() + INTERVAL '10 days'
)
RETURNING token, expires_at;


-- name: GetTokensByUser :many
SELECT * FROM refreshtokens
WHERE user_id = $1;


-- name: GetUserByToken :one
SELECT user_id FROM refreshtokens
WHERE token = $1;
