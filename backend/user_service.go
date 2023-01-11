package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v4"
	"github.com/mootslive/mono/backend/db"
	mootslivepbv1 "github.com/mootslive/mono/proto/mootslive/v1"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	queries    *db.TracedQueries
	log        *slog.Logger
	twitterCfg *oauth2.Config
	authEngine *authEngine
}

func NewUserService(
	queries *db.TracedQueries,
	log *slog.Logger,
	authEngine *authEngine,
) *UserService {
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
			Scopes: []string{
				// generates refresh token
				"offline.access",
				// allows us to post
				"tweet.write",
				// for some reason, the /me endpoint does not work without the
				// inclusion of these scopes.
				"users.read",
				"tweet.read",
			},
		},
		authEngine: authEngine,
	}
}

func (us *UserService) GetMe(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.GetMeRequest],
) (*connect.Response[mootslivepbv1.GetMeResponse], error) {
	authCtx, err := us.authEngine.handleReq(
		ctx, req, handleReqOpts{},
	)
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	res := connect.NewResponse(&mootslivepbv1.GetMeResponse{
		Id:        authCtx.user.ID,
		CreatedAt: timestamppb.New(authCtx.user.CreatedAt),
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
	_, err := us.authEngine.handleReq(ctx, req, handleReqOpts{
		noAuth: true,
	})
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

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

// TODO: Pull this out into a twitter package :D
// TwitterMeResponse is the structure of the response from
// https://api.twitter.com/2/users/me
//
//	{
//	  "data": {
//	    "id": "2244994945",
//	    "name": "TwitterDev",
//	    "username": "Twitter Dev"
//	  }
//	}
type TwitterMeResponse struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}

func (us *UserService) FinishTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.FinishTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.FinishTwitterAuthResponse], error) {
	_, err := us.authEngine.handleReq(
		ctx, req, handleReqOpts{
			noAuth: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	if req.Msg.State.State != req.Msg.ReceivedState {
		return nil, fmt.Errorf(
			"state received from twitter did not match initial state",
		)
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
	me := TwitterMeResponse{}
	if err := json.Unmarshal(bytes, &me); err != nil {
		return nil, fmt.Errorf("unmarshalling response json: %w", err)
	}

	acct, err := us.queries.GetTwitterAccount(ctx, me.Data.ID)
	if err != nil {
		// TODO: This is fucking horrible. Refactor this.
		// This is a registration if does not already exist.
		if errors.Is(err, pgx.ErrNoRows) {
			commit, rollback, queries, err := us.queries.Tx(ctx)
			if err != nil {
				return nil, fmt.Errorf("opening tx: %w", err)
			}
			defer func() {
				if err := rollback(context.Background()); err != nil {
					if !errors.Is(err, pgx.ErrTxClosed) {
						us.log.Error("failed to rollback", err)
					}
				}
			}()

			userId := ksuid.New().String()
			now := time.Now()
			err = queries.CreateUser(ctx, db.CreateUserParams{
				ID:        userId,
				CreatedAt: now,
			})
			if err != nil {
				return nil, fmt.Errorf("creating user: %w", err)
			}

			err = queries.CreateTwitterAccount(ctx, db.CreateTwitterAccountParams{
				TwitterUserID: me.Data.ID,
				UserID:        userId,
				OauthToken:    db.OAuth2Token(*tok),
				CreatedAt:     now,
			})
			if err != nil {
				return nil, fmt.Errorf("creating twitter account: %w", err)
			}

			if err := commit(ctx); err != nil {
				return nil, fmt.Errorf("committing transaction: %w", err)
			}

			idToken, err := us.authEngine.createIDToken(ctx, userId)
			if err != nil {
				return nil, fmt.Errorf("creating id token: %w", err)
			}

			res := connect.NewResponse(&mootslivepbv1.FinishTwitterAuthResponse{
				IdToken: idToken,
			})
			return res, nil
		} else {
			return nil, fmt.Errorf("fetching twitter account: %w", err)
		}
	}
	idToken, err := us.authEngine.createIDToken(ctx, acct.UserID)
	if err != nil {
		return nil, fmt.Errorf("creating id token: %w", err)
	}

	res := connect.NewResponse(&mootslivepbv1.FinishTwitterAuthResponse{
		IdToken: idToken,
	})
	return res, nil
}

func (us *UserService) ListListens(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.ListListensRequest],
) (*connect.Response[mootslivepbv1.ListListensResponse], error) {
	authCtx, err := us.authEngine.handleReq(
		ctx, req, handleReqOpts{},
	)
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	listens, err := us.queries.ListListensForUser(ctx, authCtx.user.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	res := &mootslivepbv1.ListListensResponse{
		Listens: make([]*mootslivepbv1.Listen, 0, len(listens)),
	}
	for _, listen := range listens {
		res.Listens = append(res.Listens, &mootslivepbv1.Listen{
			Id:         listen.ID,
			CreatedAt:  timestamppb.New(listen.CreatedAt),
			Source:     listen.Source,
			Isrc:       listen.Isrc,
			ListenedAt: timestamppb.New(listen.ListenedAt),
		})
	}

	return connect.NewResponse(res), nil
}
