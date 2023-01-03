package backend

import (
	"context"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/mootslive/mono/backend/db"
	mootslivepbv1 "github.com/mootslive/mono/proto/mootslive/v1"
	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	queries *db.Queries

	twitterCfg *oauth2.Config
}

func NewUserService(queries *db.Queries) *UserService {
	return &UserService{
		queries: queries,

		twitterCfg: &oauth2.Config{
			ClientID:     os.Getenv("TWITTER_CLIENT_ID"),
			ClientSecret: os.Getenv("TWITTER_CLIENT_SECRET"),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://twitter.com/i/oauth2/authorize",
				TokenURL: "https://api.twitter.com/2/oauth2/token",
			},
			RedirectURL: "http://localhost:8080/auth/callback/twitter",
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

func (as *UserService) BeginTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.BeginTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.BeginTwitterAuthResponse], error) {
	res := connect.NewResponse(&mootslivepbv1.BeginTwitterAuthResponse{
		RedirectUrl: as.twitterCfg.AuthCodeURL(
			ksuid.New().String(),
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", "123"),
			// TODO: Dont use plain
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
		),
	})
	return res, nil
}

func (as *UserService) FinishTwitterAuth(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.FinishTwitterAuthRequest],
) (*connect.Response[mootslivepbv1.FinishTwitterAuthResponse], error) {
	res := connect.NewResponse(&mootslivepbv1.FinishTwitterAuthResponse{})
	return res, nil
}
