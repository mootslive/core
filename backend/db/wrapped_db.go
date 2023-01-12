package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/mootslive/mono/backend/trace"
)

// What you observe here is pretty horrible.
// This wrapper serves two purposes:
// - Add understandable tracing spans to queries
// - Get rid of the WithTx method, and replace it with something reasonably
//   easier to unit-test.
//
// I've been held back by the limitations of pgx, sqlc and Go itself here, so,
// whilst I'd rather not define an interface at the implementation, this has
// really become my only option until I can invest more time in this and
// potentially merge a fix to upstream sqlc.

type dbtxer interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type TXQuerier interface {
	Querier
	BeginTx(ctx context.Context) (func(ctx context.Context) error, func(ctx context.Context) error, TXQuerier, error)
}

type queriesWrapper struct {
	dbtx    dbtxer
	queries Queries
}

func NewQueries(db dbtxer) TXQuerier {
	return &queriesWrapper{
		dbtx:    db,
		queries: Queries{db: db},
	}
}

// BeginTx opens a transaction, returning commit, rollback and a queries object
// that uses that transaction.
func (q *queriesWrapper) BeginTx(ctx context.Context) (func(ctx context.Context) error, func(ctx context.Context) error, TXQuerier, error) {
	tx, err := q.dbtx.Begin(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("beginning transaction: %w")
	}

	return tx.Commit, tx.Rollback, NewQueries(tx), nil
}

// Traced methods wrapped below
// TODO: Generate these, or implement tracing at a lower level :)

func (q *queriesWrapper) SelectSpotifyAccountForUpdate(ctx context.Context, spotifyUserID string) (SpotifyAccount, error) {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.SelectSpotifyAccountForUpdate")
	defer span.End()
	return q.queries.SelectSpotifyAccountForUpdate(ctx, spotifyUserID)
}

func (q *queriesWrapper) GetUser(ctx context.Context, id string) (User, error) {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.GetUser")
	defer span.End()
	return q.queries.GetUser(ctx, id)
}
func (q *queriesWrapper) CreateListen(ctx context.Context, arg CreateListenParams) error {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.CreateListen")
	defer span.End()
	return q.queries.CreateListen(ctx, arg)
}

func (q *queriesWrapper) CreateSpotifyAccount(ctx context.Context, arg CreateSpotifyAccountParams) error {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.CreateSpotifyAccount")
	defer span.End()
	return q.queries.CreateSpotifyAccount(ctx, arg)
}

func (q *queriesWrapper) CreateTwitterAccount(ctx context.Context, arg CreateTwitterAccountParams) error {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.CreateTwitterAccount")
	defer span.End()
	return q.queries.CreateTwitterAccount(ctx, arg)
}

func (q *queriesWrapper) CreateUser(ctx context.Context, arg CreateUserParams) error {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.CreateUser")
	defer span.End()
	return q.queries.CreateUser(ctx, arg)
}

func (q *queriesWrapper) GetSpotifyAccountsForScanning(ctx context.Context) ([]SpotifyAccount, error) {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.GetSpotifyAccountsForScanning")
	defer span.End()
	return q.queries.GetSpotifyAccountsForScanning(ctx)
}

func (q *queriesWrapper) GetTwitterAccount(ctx context.Context, twitterUserID string) (TwitterAccount, error) {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.GetTwitterAccount")
	defer span.End()
	return q.queries.GetTwitterAccount(ctx, twitterUserID)
}

func (q *queriesWrapper) ListListensForUser(ctx context.Context, userID string) ([]Listen, error) {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.ListListensForUser")
	defer span.End()
	return q.queries.ListListensForUser(ctx, userID)
}

func (q *queriesWrapper) UpdateSpotifyAccountListenedAt(ctx context.Context, arg UpdateSpotifyAccountListenedAtParams) error {
	ctx, span := trace.Start(ctx, "backend/db/queriesWrapper.UpdateSpotifyAccountListenedAt")
	defer span.End()
	return q.queries.UpdateSpotifyAccountListenedAt(ctx, arg)
}
