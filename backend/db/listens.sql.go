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

func (q *Queries) CreateListen(ctx context.Context, arg CreateListenParams) error {
	_, err := q.db.Exec(ctx, createListen,
		arg.ID,
		arg.UserID,
		arg.CreatedAt,
		arg.Source,
		arg.Isrc,
		arg.ListenedAt,
	)
	return err
}
