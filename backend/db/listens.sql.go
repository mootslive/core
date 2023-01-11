// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: listens.sql

package db

import (
	"context"
	"time"
)

const createListen = `-- name: CreateListen :exec
INSERT INTO listens (
    id,
    user_id,
    created_at,
    source,
    isrc,
    listened_at
) VALUES ($1, $2, $3, $4, $5, $6)
`

type CreateListenParams struct {
	ID         string
	UserID     string
	CreatedAt  time.Time
	Source     string
	Isrc       string
	ListenedAt time.Time
}

func (q *Queries) CreateListen(ctx context.Context, db DBTX, arg CreateListenParams) error {
	_, err := db.Exec(ctx, createListen,
		arg.ID,
		arg.UserID,
		arg.CreatedAt,
		arg.Source,
		arg.Isrc,
		arg.ListenedAt,
	)
	return err
}

const listListensForUser = `-- name: ListListensForUser :many
SELECT id, user_id, created_at, listened_at, isrc, source FROM listens WHERE user_id = $1
`

func (q *Queries) ListListensForUser(ctx context.Context, db DBTX, userID string) ([]Listen, error) {
	rows, err := db.Query(ctx, listListensForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Listen
	for rows.Next() {
		var i Listen
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.CreatedAt,
			&i.ListenedAt,
			&i.Isrc,
			&i.Source,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
