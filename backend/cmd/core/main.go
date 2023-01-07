package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend"
	"github.com/mootslive/mono/backend/db"
	"github.com/mootslive/mono/proto/mootslive/v1/mootslivepbv1connect"
	"github.com/rs/cors"
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

	authEngine := backend.NewAuthEngine([]byte("PULLMEFROMACONFIG"))

	adminService := &backend.AdminService{}
	userService := backend.NewUserService(db.New(conn), log, conn, authEngine)

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
		loggingInterceptor := NewLoggingUnaryInteceptor(log)

		mux := http.NewServeMux()
		reflector := grpcreflect.NewStaticReflector(
			mootslivepbv1connect.AdminServiceName,
			mootslivepbv1connect.UserServiceName,
		)
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

		mux.Handle(mootslivepbv1connect.NewAdminServiceHandler(
			adminService,
			connect.WithInterceptors(loggingInterceptor),
		))
		mux.Handle(mootslivepbv1connect.NewUserServiceHandler(
			userService,
			connect.WithInterceptors(loggingInterceptor),
		))

		return http.ListenAndServe(
			"localhost:9000",
			h2c.NewHandler(cors.AllowAll().Handler(mux), &http2.Server{}),
		)
	})

	return eg.Wait()
}

func NewLoggingUnaryInteceptor(log *slog.Logger) connect.UnaryInterceptorFunc {
	f := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			log.Info("received unary request", slog.String("procedure", req.Spec().Procedure))
			return next(ctx, req)
		})
	}

	return connect.UnaryInterceptorFunc(f)
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Println("Exiting with fatal err: ", err)
	}
}
