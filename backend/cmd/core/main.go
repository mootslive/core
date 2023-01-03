package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend"
	"github.com/mootslive/mono/backend/db"
	"github.com/mootslive/mono/proto/mootslive/v1/mootslivepbv1connect"
	"golang.org/x/exp/slog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

func run(out io.Writer) error {
	ctx := context.Background()
	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	log := slog.New(logOpts.NewJSONHandler(out))
	log.Info("starting mootslive backend")

	pgxCfg, err := pgxpool.ParseConfig("postgres://mootslive:mootslive@localhost:5432/mootslive")
	if err != nil {
		return fmt.Errorf("parsing cfg: %w", err)
	}

	conn, err := pgxpool.ConnectConfig(ctx, pgxCfg)
	if err != nil {
		return fmt.Errorf("connecting to pgx: %w", err)
	}
	defer conn.Close()

	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		poller := &backend.SpotifyPoller{
			DB:      conn,
			Queries: db.New(conn),
			Log:     log.WithGroup("poller"),
		}

		if err := poller.Run(gctx); err != nil {
			return fmt.Errorf("polling: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		mux := http.NewServeMux()
		mux.Handle(mootslivepbv1connect.NewAdminServiceHandler(&backend.AdminService{}))
		mux.Handle(mootslivepbv1connect.NewUserServiceHandler(backend.NewUserService(db.New(conn))))
		return http.ListenAndServe("localhost:8080", h2c.NewHandler(mux, &http2.Server{}))
	})

	return eg.Wait()
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Println("Exiting with fatal err: ", err)
	}
}
