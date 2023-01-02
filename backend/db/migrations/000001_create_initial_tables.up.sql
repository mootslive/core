CREATE TABLE listens (
    id CHAR(27) PRIMARY KEY,
    user_id CHAR(27) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    isrc CHAR(12) NOT NULL
);

CREATE TABLE users (
    id CHAR(27) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE spotify_accounts (
    spotify_user_id VARCHAR(256) PRIMARY KEY,
    user_id CHAR(27) NOT NULL,
    access_token VARCHAR(256) NOT NULL,
    refresh_token VARCHAR(256) NOT NULL,
    last_scanned TIMESTAMPTZ
);