package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend/db"
	"io"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/mootslive/mono/backend"
	"github.com/mootslive/mono/proto/mootslive/v1/mootslivepbv1connect"
	"github.com/rs/cors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/exp/slog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

func setupTraceExporting() error {
	jaegerExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(),
	)
	if err != nil {
		return fmt.Errorf("creating jaeger exporter: %w", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(jaegerExporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNamespaceKey.String("mootslive"),
			semconv.ServiceNameKey.String("core"),
		)),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func run(out io.Writer) error {
	ctx := context.Background()
	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	log := slog.New(logOpts.NewJSONHandler(out))
	log.Info("starting mootslive backend")

	if err := setupTraceExporting(); err != nil {
		return fmt.Errorf("setting up trace exporting: %w", err)
	}

	pgxCfg, err := pgxpool.ParseConfig("postgres://mootslive:mootslive@localhost:5432/mootslive")
	if err != nil {
		return fmt.Errorf("parsing cfg: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, pgxCfg)
	if err != nil {
		return fmt.Errorf("connecting to pgx: %w", err)
	}
	defer pool.Close()

	queries := db.NewQueries(pool)

	authEngine := backend.NewAuthEngine(
		[]byte("PULLMEFROMACONFIG"),
		queries,
	)

	adminService := &backend.AdminServerHandler{}
	userService := backend.NewUserServiceHandler(queries, log, authEngine)

	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := backend.NewSpotifyPoller(log, queries).Run(gctx); err != nil {
			return fmt.Errorf("polling: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		loggingInterceptor := NewLoggingUnaryInteceptor(log)
		telemetryInterceptor := otelconnect.NewInterceptor()

		mux := http.NewServeMux()
		reflector := grpcreflect.NewStaticReflector(
			mootslivepbv1connect.AdminServiceName,
			mootslivepbv1connect.UserServiceName,
		)
		mux.Handle(grpcreflect.NewHandlerV1(
			reflector,
			connect.WithInterceptors(telemetryInterceptor, loggingInterceptor),
		))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(
			reflector,
			connect.WithInterceptors(telemetryInterceptor, loggingInterceptor),
		))

		mux.Handle(mootslivepbv1connect.NewAdminServiceHandler(
			adminService,
			connect.WithInterceptors(telemetryInterceptor, loggingInterceptor),
		))
		mux.Handle(mootslivepbv1connect.NewUserServiceHandler(
			userService,
			connect.WithInterceptors(telemetryInterceptor, loggingInterceptor),
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
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			log.Info("received unary request", slog.String("procedure", req.Spec().Procedure))
			return next(ctx, req)
		}
	}

	return f
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Println("Exiting with fatal err: ", err)
	}
}
