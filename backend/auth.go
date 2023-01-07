package backend

import (
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/segmentio/ksuid"
)

// TODO: Replace this either Google KMS or a secret pulled from env or
// secrets manager :))
const thisIsVeryBadJWTSigningKey = "ahaha-this-wont-last-long"

type IDTokenClaims struct {
	jwt.RegisteredClaims
}

func createIDToken(userID string) (string, error) {
	idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// TODO: Sort out audience and issuer to match service hostname
			Issuer: "https://api.moots.live",
			Audience: []string{
				"https://api.moots.live",
			},
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

type authOptions struct {
	// noAuth enforces that the user should not be authenticated to access this
	// endpoint
	noAuth bool
}

type authCtx struct {
	userID string
	// TODO: Maybe fetch user and insert into auth ctx?
}

func auth(req connect.AnyRequest, opt authOptions) (*authCtx, error) {
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

	token, err := jwt.ParseWithClaims(
		splitAuthHeader[1],
		&IDTokenClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v", t.Header["alg"],
				)
			}
			return []byte(thisIsVeryBadJWTSigningKey), nil
		})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*IDTokenClaims)
	if !ok {
		panic("unexpected token claims content")
	}

	return &authCtx{
		userID: claims.Subject,
	}, nil
}
