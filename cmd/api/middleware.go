package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("missing authorization header"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		jwtToken, err := app.authenticator.ValidateToken(parts[1])
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)
		userId, err := parseSubjectClaim(claims["sub"])
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization claims"))
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, userId)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization claims"))
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid authorization header"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid authorization header"))
				return
			}

			username := "admin"
			password := "admin"

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid authorization header"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseSubjectClaim(sub any) (int64, error) {
	switch v := sub.(type) {
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("subject claim is not an integer")
		}
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported subject claim type: %T", sub)
	}
}
