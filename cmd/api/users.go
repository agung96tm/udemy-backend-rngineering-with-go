package main

import (
	"context"
	"errors"
	"net/http"
	"socialv2/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtxKey userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)
	_ = app.writeJSON(w, http.StatusOK, user)
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.notFoundError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Followers.Follow(ctx, followerUser.ID, followedID)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusNoContent, nil)
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.notFoundError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Followers.Unfollow(ctx, followerUser.ID, unfollowedID)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusNoContent, nil)
}

func (app *application) getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtxKey).(*store.User)
	return user
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.writeJSON(w, http.StatusNoContent, nil)
}
