package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/mootslive/mono/backend/db"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type SpotifyPoller struct {
	DB  *db.Queries
	Log *slog.Logger
}

func (sp *SpotifyPoller) Run(ctx context.Context) error {
	sp.Log.Info("starting poller")

	for {
		sp.Log.Info("running account scan")
		accounts, err := sp.DB.GetSpotifyAccountsForScanning(ctx)
		if err != nil {
			return fmt.Errorf("fetching accounts: %w", err)
		}

		for _, account := range accounts {
			if err := sp.ScanAccount(ctx, account); err != nil {
				return fmt.Errorf("scanning account: %w", err)
			}
		}

		select {
		case <-time.After(time.Second * 10):
			continue
		case <-ctx.Done():
			sp.Log.Info("context cancelled, stopping poller")
			return ctx.Err()
		}
	}
}

func (sp *SpotifyPoller) ScanAccount(ctx context.Context, account db.SpotifyAccount) error {
	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	played, err := client.PlayerRecentlyPlayed(ctx)
	if err != nil {
		return fmt.Errorf("fetching recently played: %w", err)
	}

	for _, track := range played {
		sp.Log.Info("track", "val", track)
	}
	return nil
}
