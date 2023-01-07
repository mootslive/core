-- name: CreateUser :exec
INSERT INTO users (
    id, created_at
) VALUES ($1, $2);

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;