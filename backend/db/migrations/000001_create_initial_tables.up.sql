CREATE TABLE users (
    id CHAR(27) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE listens (
    id CHAR(27) PRIMARY KEY,
    user_id CHAR(27) NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    listened_at TIMESTAMPTZ NOT NULL,
    isrc CHAR(12) NOT NULL,
    source VARCHAR(32) NOT NULL
);

CREATE TABLE spotify_accounts (
    spotify_user_id VARCHAR(256) PRIMARY KEY,
    user_id CHAR(27) NOT NULL REFERENCES users ON DELETE CASCADE,
    access_token VARCHAR(256) NOT NULL,
    refresh_token VARCHAR(256) NOT NULL,
    last_listened_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE twitter_accounts (
    twitter_user_id VARCHAR(32) PRIMARY KEY,
    user_id CHAR(27) NOT NULL REFERENCES users ON DELETE CASCADE,
    oauth_token JSON NOT NULL,
    created_at TIMESTAMPTZ
);