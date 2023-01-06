-- name: GetSpotifyAccountsForScanning :many
SELECT * FROM spotify_accounts;

-- name: CreateSpotifyAccount :exec
INSERT INTO spotify_accounts (
    spotify_user_id,
    user_id,
    oauth_token,
    created_at
)  VALUES ($1, $2, $3, $4);

-- name: SelectSpotifyAccountForUpdate :one
SELECT * FROM spotify_accounts WHERE spotify_user_id = $1 FOR UPDATE;

-- name: UpdateSpotifyAccountListenedAt :exec
UPDATE spotify_accounts SET last_listened_at = $1 WHERE spotify_user_id = $2;