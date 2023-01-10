package db

import (
	"context"
	"github.com/mootslive/mono/backend/trace"
)

// TODO: Not sure I like this :(
// Maybe switch to pgx/v5 and hook into that instead

type TracedQueries struct {
	*Queries
}

func (q *TracedQueries) GetUser(ctx context.Context, id string) (User, error) {
	ctx, span := trace.Start(ctx, "backend/db/TracedQueries.GetUser")
	defer span.End()
	return q.Queries.GetUser(ctx, id)
}
