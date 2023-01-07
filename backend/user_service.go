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
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mootslive/mono/backend/db"
	mootslivepbv1 "github.com/mootslive/mono/proto/mootslive/v1"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	queries    *db.Queries
	log        *slog.Logger
	twitterCfg *oauth2.Config
	db         *pgxpool.Pool
}

func NewUserService(queries *db.Queries, log *slog.Logger, db *pgxpool.Pool) *UserService {
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
				"offline.access",
				"tweet.write",
				"users.read",
				"tweet.read",
			},
		},

		db: db,
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

const thisIsVeryBadJWTSigningKey = "ahaha-this-wont-last-long"

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
	me := TwitterMeResponse{}
	if err := json.Unmarshal(bytes, &me); err != nil {
		return nil, fmt.Errorf("unmarshalling response json: %w", err)
	}

	acct, err := us.queries.GetTwitterAccount(ctx, me.Data.ID)
	if err != nil {
		// TODO: This is fucking horrible. Refactor this.
		// This is a registration if does not already exist.
		if errors.Is(err, pgx.ErrNoRows) {
			tx, err := us.db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return nil, fmt.Errorf("opening tx: %w", err)
			}
			queries := us.queries.WithTx(tx)

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

			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("committing transaction: %w", err)
			}

			idToken, err := createIDToken(userId)
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
	idToken, err := createIDToken(acct.UserID)
	if err != nil {
		return nil, fmt.Errorf("creating id token: %w", err)
	}

	res := connect.NewResponse(&mootslivepbv1.FinishTwitterAuthResponse{
		IdToken: idToken,
	})
	return res, nil
}

func createIDToken(userID string) (string, error) {
	idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// TODO: Sort out expiry and issuer to match service hostname
			Issuer:    "https://moots.live",
			Audience:  []string{"https://moots.live"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second * -5)),
			// TODO: Determine sane token expiry
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
			Subject:   userID,
			ID:        ksuid.New().String(),
		},
	})

	tok, err := idToken.SignedString([]byte(thisIsVeryBadJWTSigningKey))
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return tok, nil
}

type IDTokenClaims struct {
	jwt.RegisteredClaims
}
