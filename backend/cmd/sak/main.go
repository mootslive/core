package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/mootslive/mono/backend/db"
	"github.com/segmentio/ksuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout))
	log.Info(ksuid.New().String())

	pgxCfg, err := pgx.ParseConfig("postgres://mootslive:mootslive@localhost:5432/mootslive")
	if err != nil {
		panic(err)
	}

	conn, err := pgx.ConnectConfig(ctx, pgxCfg)
	if err != nil {
		panic(err)
	}

	queries := db.New(conn)

	state := ksuid.New().String()
	sa := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadRecentlyPlayed))
	ch := make(chan *oauth2.Token)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := sa.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Error("failed to get token", err)
			return
		}

		ch <- tok
	})
	go func() {
		err := http.ListenAndServe(":8080", mux)
		if err != nil {
			log.Error("failed to listen", err)
		}
	}()

	url := sa.AuthURL(state)
	log.Info("please log in to Spotify by visiting the following page in your browser:", url)

	tok := <-ch

	client := spotify.New(sa.Client(ctx, tok))
	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		panic(err)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		panic(err)
	}

	queries = queries.WithTx(tx)
	userID := ksuid.New().String()
	err = queries.CreateUser(ctx, db.CreateUserParams{
		ID:        userID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		panic(err)
	}
	err = queries.CreateSpotifyAccount(ctx, db.CreateSpotifyAccountParams{
		SpotifyUserID: currentUser.ID,
		AccessToken:   tok.AccessToken,
		RefreshToken:  tok.RefreshToken,
		UserID:        userID,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		panic(err)
	}
	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
}
