package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TestAuthenticator struct{}

const secret = "secret"

var testClaim = jwt.MapClaims{
	"sub": "99",
	"exp": time.Now().Add(time.Hour).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
	"iss": "test-99",
	"aud": "test-99",
}

func (ta TestAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaim)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func (ta TestAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
