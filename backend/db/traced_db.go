package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend/trace"
)

// TODO: Not sure I like this :(
// Maybe switch to pgx/v5 and hook into that instead

type DBTXer interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewTracedQueries(conn *pgxpool.Pool) *TracedQueries {
	return &TracedQueries{
		Queries: New(conn),
	}
}

type TracedQueries struct {
	conn *pgxpool.Pool
	*Queries
}

func (q *TracedQueries) Tx(ctx context.Context) (commit func(ctx context.Context) error, rollback func(ctx context.Context) error, queries *TracedQueries, err error) {
	if q.conn == nil {
		return nil, nil, nil, fmt.Errorf("subtransactions not supported")
	}

	tx, err := q.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open tx: %w", err)
	}

	return tx.Commit, tx.Rollback, &TracedQueries{
		Queries: q.Queries.WithTx(tx),
	}, nil
}

func (q *TracedQueries) SelectSpotifyAccountForUpdate(ctx context.Context, spotifyUserID string) (SpotifyAccount, error) {
	ctx, span := trace.Start(ctx, "backend/db/TracedQueries.SelectSpotifyAccountForUpdate")
	defer span.End()
	return q.Queries.SelectSpotifyAccountForUpdate(ctx, spotifyUserID)
}

func (q *TracedQueries) GetUser(ctx context.Context, id string) (User, error) {
	ctx, span := trace.Start(ctx, "backend/db/TracedQueries.GetUser")
	defer span.End()
	return q.Queries.GetUser(ctx, id)
}
