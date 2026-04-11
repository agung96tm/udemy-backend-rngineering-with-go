package main

import (
	"context"
	"errors"
	"net/http"
	"socialv3/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// getUserHandler godoc
//
//	@Summary		Ambil detail user
//	@Description	Mengambil satu user berdasarkan ID.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		int			true	"ID user"
//	@Success		200		{object}	store.User	"Detail user"
//	@Failure		400		{object}	object		"ID tidak valid"
//	@Failure		404		{object}	object		"User tidak ditemukan"
//	@Failure		500		{object}	object		"Server error"
//	@Router			/users/{userID} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	_ = app.jsonResponse(w, http.StatusOK, user)
}

type FollowUserRequest struct {
	UserID int64 `json:"user_id" validate:"required,min=1"`
}

// followUserHandler godoc
//
//	@Summary		Follow user
//	@Description	User saat ini (dari context) mem-follow user dengan user_id dari body.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int						true	"ID user (follower / yang mem-follow)"
//	@Param			body	body		FollowUserRequest		true	"ID user yang akan di-follow"
//	@Success		200		{object}	map[string]interface{}	"message: user followed successfully"
//	@Failure		400		{object}	object					"Request body tidak valid"
//	@Failure		404		{object}	object					"User tidak ditemukan"
//	@Failure		500		{object}	object					"Server error"
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Follower.Follow(ctx, followerUser.ID, followedID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "user followed successfully",
	})
}

// unfollowUserHandler godoc
//
//	@Summary		Unfollow user
//	@Description	User saat ini (dari context) unfollow user dengan user_id dari body.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int						true	"ID user (yang unfollow)"
//	@Param			body	body		FollowUserRequest		true	"ID user yang akan di-unfollow"
//	@Success		200		{object}	map[string]interface{}	"message: user unfollowed successfully"
//	@Failure		400		{object}	object					"Request body tidak valid"
//	@Failure		404		{object}	object					"User tidak ditemukan"
//	@Failure		500		{object}	object					"Server error"
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Follower.Unfollow(ctx, followerUser.ID, unfollowedID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "user unfollowed successfully",
	})
}

// activateUserHandler godoc
//
//	@Summary		Aktivasi user
//	@Description	Mengaktivasi user dengan token invite yang dikirim via email saat registrasi.
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string					true	"Token aktivasi (dari email invite)"
//	@Success		200		{object}	map[string]interface{}	"message: user activated successfully"
//	@Failure		400		{object}	object					"Token tidak valid atau kadaluarsa"
//	@Failure		404		{object}	object					"Token tidak ditemukan"
//	@Failure		500		{object}	object					"Server error"
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "user activated successfully",
	})
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := app.readID(r, "userID")
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
		ctx = context.WithValue(ctx, userCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	ctx := r.Context()
	user, ok := ctx.Value(userCtx).(*store.User)
	if !ok {
		return nil
	}
	return user
}
