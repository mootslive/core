-- name: CreateListen :exec
INSERT INTO listens (
    id,
    user_id,
    created_at,
    source,
    isrc,
    listened_at
) VALUES ($1, $2, $3, $4, $5, $6);

