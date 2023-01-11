package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/mootslive/mono/backend/trace"
)

// Maybe switch to pgx/v5 and hook into that instead

// TODO: Not sure I like this :(
type DBTXer interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type QueriesWrapper struct {
	Queries
}

func (q *QueriesWrapper) SelectSpotifyAccountForUpdate(ctx context.Context, db DBTX, spotifyUserID string) (SpotifyAccount, error) {
	ctx, span := trace.Start(ctx, "backend/db/QueriesWrapper.SelectSpotifyAccountForUpdate")
	defer span.End()
	return q.Queries.SelectSpotifyAccountForUpdate(ctx, db, spotifyUserID)
}

func (q *QueriesWrapper) GetUser(ctx context.Context, db DBTX, id string) (User, error) {
	ctx, span := trace.Start(ctx, "backend/db/QueriesWrapper.GetUser")
	defer span.End()
	return q.Queries.GetUser(ctx, db, id)
}
