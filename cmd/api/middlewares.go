package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"socialv3/internal/store"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimiterExceedError(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
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

			payload, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, err)
				return
			}

			username := app.config.auth.basic.username
			pass := app.config.auth.basic.password

			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) != 2 || pair[0] != username || pair[1] != pass {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			//user, pass, ok := r.BasicAuth()
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) authJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedBasicError(w, r, fmt.Errorf("missing authorization header"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization header"))
			return
		}
		if parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		user, err := app.getUserForMiddleware(r.Context(), userID)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post, _ := app.getPostFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenError(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, requiredRole string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, requiredRole)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUserForMiddleware(ctx context.Context, userID int64) (*store.User, error) {
	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err == nil {
		return user, nil
	}

	user, err = app.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
