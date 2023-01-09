package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mootslive/mono/backend/db"
	"github.com/segmentio/ksuid"
)

// authEngine manages users and enforcing auth
type authEngine struct {
	signingKey []byte
	issuer     string
	queries    *db.Queries
}

func NewAuthEngine(signingKey []byte, queries *db.Queries) *authEngine {
	return &authEngine{
		signingKey: signingKey,
		// TODO: Pass in real URI of service
		issuer:  "https://api.moots.live",
		queries: queries,
	}
}

type idTokenClaims struct {
	jwt.RegisteredClaims
}

func (ae *authEngine) createIDToken(userID string) (string, error) {
	idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, idTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: ae.issuer,
			Audience: []string{
				ae.issuer,
			},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second * -5)),
			// TODO: Determine sane token expiry
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
			Subject:   userID,
			ID:        ksuid.New().String(),
		},
	})

	tok, err := idToken.SignedString(ae.signingKey)
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return tok, nil
}

func (ae *authEngine) validateIDToken(idToken string) (*idTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		idToken,
		&idTokenClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v", t.Header["alg"],
				)
			}
			return ae.signingKey, nil
		})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*idTokenClaims)
	if !ok {
		panic("unexpected token claims content")
	}

	return claims, nil
}

type handleReqOpts struct {
	// noAuth enforces that the user should not be authenticated to access this
	// endpoint
	noAuth bool
}

type authCtx struct {
	user *db.User
}

func (ae *authEngine) handleReq(
	ctx context.Context, req connect.AnyRequest, opt handleReqOpts,
) (*authCtx, error) {
	headers := req.Header()
	authHeader := headers.Get("Authorization")
	if opt.noAuth {
		if authHeader != "" {
			return nil, fmt.Errorf(
				"authorization header provided on no auth endpoint",
			)
		}
		return nil, nil
	} else if authHeader == "" {
		return nil, fmt.Errorf("no authorization header provided")
	}

	splitAuthHeader := strings.Split(authHeader, " ")
	if len(splitAuthHeader) != 2 {
		return nil, fmt.Errorf(
			"did not receive two parts in authorization header",
		)
	}

	if splitAuthHeader[0] != "Bearer" {
		return nil, fmt.Errorf("received non-bearer authorization header")
	}

	claims, err := ae.validateIDToken(splitAuthHeader[1])
	if err != nil {
		return nil, fmt.Errorf("validating id token: %w", err)
	}

	user, err := ae.queries.GetUser(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return &authCtx{
		user: &user,
	}, nil
}
