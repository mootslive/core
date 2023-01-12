package twitter

import "golang.org/x/oauth2"

var Endpoint = oauth2.Endpoint{
	AuthStyle: oauth2.AuthStyleInHeader,
	AuthURL:   "https://twitter.com/i/oauth2/authorize",
	TokenURL:  "https://api.twitter.com/2/oauth2/token",
}

var (
	ScopeOfflineAccess = "offline.access"
	ScopeTweetWrite    = "tweet.write"
	ScopeUsersRead     = "users.read"
	TweetRead          = "tweet.read"
)

func DefaultScopes() []string {
	return []string{
		// generates refresh token
		"offline.access",
		// allows us to post
		"tweet.write",
		// for some reason, the /me endpoint does not work without the
		// inclusion of these two following scopes.
		"users.read",
		"tweet.read",
	}
}

// UsersMeResponse is the structure of the response from
// https://api.twitter.com/2/users/me
//
//	{
//	  "data": {
//	    "id": "2244994945",
//	    "name": "TwitterDev",
//	    "username": "Twitter Dev"
//	  }
//	}
type UsersMeResponse struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}
