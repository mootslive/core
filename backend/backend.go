package backend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend/db"
	"github.com/segmentio/ksuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type SpotifyPoller struct {
	DB      *pgxpool.Pool
	Queries *db.Queries
	Log     *slog.Logger
}

func (sp *SpotifyPoller) Run(ctx context.Context) error {
	sp.Log.Info("starting poller")

	for {
		sp.Log.Info("running account scan")
		accounts, err := sp.Queries.GetSpotifyAccountsForScanning(ctx)
		if err != nil {
			return fmt.Errorf("fetching accounts: %w", err)
		}

		for _, account := range accounts {
			if err := sp.ScanAccount(ctx, account.SpotifyUserID); err != nil {
				return fmt.Errorf("scanning account: %w", err)
			}
		}

		select {
		// Simple ten second backoff
		case <-time.After(time.Second * 10):
			continue
		case <-ctx.Done():
			sp.Log.Info("context cancelled, stopping poller")
			return ctx.Err()
		}
	}
}

const (
	sourceSpotify = "spotify"
)

func (sp *SpotifyPoller) ScanAccount(
	ctx context.Context, spotifyUserID string,
) error {
	start := time.Now()
	tx, err := sp.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("opening tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(context.Background()); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				sp.Log.Error("failed to rollback", err)
			}
		}
	}()
	queries := sp.Queries.WithTx(tx)

	account, err := queries.SelectSpotifyAccountForUpdate(ctx, spotifyUserID)
	if err != nil {
		return fmt.Errorf("locking account: %w", err)
	}

	var afterEpochMs int64 = 0
	if account.LastListenedAt.Valid {
		afterEpochMs = (account.LastListenedAt.Time.Add(time.Second).Unix()) * 1000
	}

	client := clientForSpotifyAccount(ctx, account)
	played, err := client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{
		Limit:        50,
		AfterEpochMs: afterEpochMs,
	})
	if err != nil {
		return fmt.Errorf("fetching recently played: %w", err)
	}

	var listenedAt *time.Time
	for _, track := range played {
		track := track
		sp.Log.Debug("recording listen", "user_id", account.UserID, "track_name", track.Track.Name, "listened_at", track.PlayedAt)
		err := queries.CreateListen(ctx, db.CreateListenParams{
			ID:         ksuid.New().String(),
			UserID:     account.UserID,
			CreatedAt:  time.Now(),
			Source:     sourceSpotify,
			Isrc:       track.Track.ExternalIDs.ISRC,
			ListenedAt: track.PlayedAt,
		})
		if err != nil {
			return fmt.Errorf("recording listen: %w", err)
		}
		if listenedAt == nil {
			listenedAt = &track.PlayedAt
		}
	}

	if listenedAt != nil {
		err := queries.UpdateSpotifyAccountListenedAt(ctx, db.UpdateSpotifyAccountListenedAtParams{
			SpotifyUserID: account.SpotifyUserID,
			LastListenedAt: sql.NullTime{
				Valid: true,
				Time:  *listenedAt,
			},
		})
		if err != nil {
			return fmt.Errorf("updating listened at: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	sp.Log.Info("recorded listens for user",
		slog.String("user_id", account.UserID),
		slog.Int("count", len(played)),
		slog.String("duration", time.Since(start).String()),
	)

	return nil
}

func clientForSpotifyAccount(
	ctx context.Context, account db.SpotifyAccount,
) *spotify.Client {
	token := oauth2.Token(account.OauthToken)
	httpClient := spotifyauth.New().Client(ctx, &token)
	client := spotify.New(httpClient)
	return client
}
