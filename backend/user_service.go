package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/mootslive/mono/backend/db"
	mootslivepbv1 "github.com/mootslive/mono/proto/mootslive/v1"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	queries    *db.Queries
	log        *slog.Logger
	twitterCfg *oauth2.Config
}

func NewUserService(queries *db.Queries, log *slog.Logger) *UserService {
	return &UserService{
		log:     log,
		queries: queries,

		twitterCfg: &oauth2.Config{
			ClientID:     os.Getenv("TWITTER_CLIENT_ID"),
			ClientSecret: os.Getenv("TWITTER_CLIENT_SECRET"),
			Endpoint: oauth2.Endpoint{
				AuthStyle: oauth2.AuthStyleInHeader,
				AuthURL:   "https://twitter.com/i/oauth2/authorize",
				TokenURL:  "https://api.twitter.com/2/oauth2/token",
			},
			RedirectURL: "http://localhost:3000/auth/twitter/callback",
			Scopes:      []string{"offline.access", "tweet.write"},
		},
	}
}

func (us *UserService) GetMe(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.GetMeRequest],
) (*connect.Response[mootslivepbv1.GetMeResponse], error) {
	// TODO: Check auth
	// TODO: Fetch user
	res := connect.NewResponse(&mootslivepbv1.GetMeResponse{
		Id:        "foo",
		CreatedAt: timestamppb.Now(),
	})
	return res, nil
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (us *UserService) BeginTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.BeginTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.BeginTwitterAuthResponse], error) {
	state, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}
	pkceCodeVerifier, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	res := connect.NewResponse(&mootslivepbv1.BeginTwitterAuthResponse{
		RedirectUrl: us.twitterCfg.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", pkceCodeVerifier),
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
		),
		State: &mootslivepbv1.OAuth2State{
			State:            state,
			PkceCodeVerifier: pkceCodeVerifier,
		},
	})
	return res, nil
}

func (us *UserService) FinishTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.FinishTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.FinishTwitterAuthResponse], error) {
	if req.Msg.State.State != req.Msg.ReceivedState {
		return nil, fmt.Errorf("state received from twitter did not match initial state")
	}

	tok, err := us.twitterCfg.Exchange(
		ctx,
		req.Msg.ReceivedCode,
		oauth2.SetAuthURLParam("code_verifier", req.Msg.State.PkceCodeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	client := us.twitterCfg.Client(ctx, tok)
	resp, err := client.Get("https://api.twitter.com/2/users/me")
	if err != nil {
		return nil, fmt.Errorf("requesting me: %w", err)
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO: persist tok
	res := connect.NewResponse(&mootslivepbv1.FinishTwitterAuthResponse{
		Me: string(bytes),
	})
	return res, nil
}
