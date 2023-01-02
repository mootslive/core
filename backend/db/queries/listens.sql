-- name: CreateListen :exec
INSERT INTO listens (
    id,
    user_id,
    created_at,
    source,
    isrc
) VALUES ($1, $2, $3, $4, $5);

-- name: GetSpotifyAccountsForScanning :many
SELECT * FROM spotify_accounts;

-- name: CreateSpotifyAccount :exec
INSERT INTO spotify_accounts (
    spotify_user_id,
    user_id,
    access_token,
    refresh_token,
    created_at
)  VALUES ($1, $2, $3, $4, $5);
