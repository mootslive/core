package backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/mootslive/mono/backend/twitter"
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

type UserServiceHandler struct {
	queries    db.TXQuerier
	log        *slog.Logger
	twitterCfg *oauth2.Config
	authEngine *authEngine
}

func NewUserServiceHandler(
	queries db.TXQuerier,
	log *slog.Logger,
	authEngine *authEngine,
) *UserServiceHandler {
	return &UserServiceHandler{
		log:     log,
		queries: queries,

		twitterCfg: twitter.OAuthConfig(),
		authEngine: authEngine,
	}
}

func (us *UserServiceHandler) GetMe(
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

func (us *UserServiceHandler) BeginTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.BeginTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.BeginTwitterAuthResponse], error) {
	_, err := us.authEngine.handleReq(ctx, req, handleReqOpts{
		noAuth: true,
	})
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	state, pkceCodeVerifier, redirect, err := twitter.BeginTwitterAuth()
	if err != nil {
		return nil, fmt.Errorf("starting twitter auth: %w", err)
	}

	res := connect.NewResponse(&mootslivepbv1.BeginTwitterAuthResponse{
		RedirectUrl: redirect,
		State: &mootslivepbv1.OAuth2State{
			State:            state,
			PkceCodeVerifier: pkceCodeVerifier,
		},
	})
	return res, nil
}

func (us *UserServiceHandler) FinishTwitterAuth(
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

	tok, err := twitter.FinishTwitterAuth(
		ctx,
		req.Msg.ReceivedState,
		req.Msg.State.State,
		req.Msg.State.PkceCodeVerifier,
		req.Msg.ReceivedCode,
	)
	if err != nil {
		return nil, fmt.Errorf("twitter auth: %w", err)
	}

	client := twitter.NewClient(ctx, tok)
	me, err := client.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("requesting me: %w", err)
	}

	acct, err := us.queries.GetTwitterAccount(ctx, me.Data.ID)
	isNotFound := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !isNotFound {
		return nil, fmt.Errorf("fetching twitter account: %w", err)
	}

	// TODO: Probably pull most of this out into a registration function
	if isNotFound {
		commit, rollback, tx, err := us.queries.BeginTx(ctx)
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
		err = tx.CreateUser(ctx, db.CreateUserParams{
			ID:        userId,
			CreatedAt: now,
		})
		if err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}

		err = tx.CreateTwitterAccount(ctx, db.CreateTwitterAccountParams{
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

func (us *UserServiceHandler) ListListens(
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
