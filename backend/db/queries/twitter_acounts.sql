-- name: GetTwitterAccount :one
SELECT * FROM twitter_accounts WHERE twitter_user_id = $1 LIMIT 1;

-- name: CreateTwitterAccount :exec
INSERT INTO twitter_accounts (
    twitter_user_id,
    user_id,
    oauth_token,
    created_at
)  VALUES ($1, $2, $3, $4);