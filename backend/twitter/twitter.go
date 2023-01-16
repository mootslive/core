package twitter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"os"
)

var Endpoint = oauth2.Endpoint{
	AuthStyle: oauth2.AuthStyleInHeader,
	AuthURL:   "https://twitter.com/i/oauth2/authorize",
	TokenURL:  "https://api.twitter.com/2/oauth2/token",
}

var (
	ScopeOfflineAccess = "offline.access"
	ScopeTweetWrite    = "tweet.write"
	ScopeUsersRead     = "users.read"
	ScopeTweetRead     = "tweet.read"
)

func DefaultScopes() []string {
	return []string{
		// generates refresh token
		ScopeOfflineAccess,
		// allows us to post
		ScopeTweetWrite,
		// for some reason, the /me endpoint does not work without the
		// inclusion of these two following scopes.
		ScopeUsersRead,
		ScopeTweetRead,
	}
}

func OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("TWITTER_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITTER_CLIENT_SECRET"),
		Endpoint:     Endpoint,
		RedirectURL:  "http://localhost:3000/auth/twitter/callback",
		Scopes:       DefaultScopes(),
	}
}

type Client struct {
	http *http.Client
}

func NewClient(ctx context.Context, tok *oauth2.Token) *Client {
	httpClient := OAuthConfig().Client(ctx, tok)
	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)
	return &Client{
		http: httpClient,
	}
}

// GetMeResponse is the structure of the response from
// https://api.twitter.com/2/users/me
//
//	{
//	  "data": {
//	    "id": "2244994945",
//	    "name": "TwitterDev",
//	    "username": "Twitter Dev"
//	  }
//	}
type GetMeResponse struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}

func (c *Client) GetMe(ctx context.Context) (*GetMeResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitter.com/2/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	obj := GetMeResponse{}
	if err := json.Unmarshal(bytes, &obj); err != nil {
		return nil, fmt.Errorf("unmarshalling response: %w", err)
	}

	return &obj, nil
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func BeginTwitterAuth() (state string, pkceCodeVerifier string, redirect string, err error) {
	state, err = generateRandomString(32)
	if err != nil {
		return "", "", "", err
	}
	pkceCodeVerifier, err = generateRandomString(32)
	if err != nil {
		return "", "", "", err
	}

	redirect = OAuthConfig().AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceCodeVerifier),
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)

	return state, pkceCodeVerifier, redirect, err
}

func FinishTwitterAuth(ctx context.Context, state, prevState, pkceCodeVerifier, code string) (*oauth2.Token, error) {
	if state != prevState {
		return nil, fmt.Errorf(
			"state received from twitter did not match initial state",
		)
	}

	tok, err := OAuthConfig().Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", pkceCodeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	return tok, nil
}
