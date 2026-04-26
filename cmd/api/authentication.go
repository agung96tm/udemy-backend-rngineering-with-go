package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"socialv2/internal/mailer"
	"socialv2/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type RegisterUserPayload struct {
	Name     string `json:"name" validate:"required,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// has password and store
	user := &store.User{
		Name:  payload.Name,
		Email: payload.Email,
	}
	err := user.Password.Set(payload.Password)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])
	err = app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	vars := struct {
		Name          string
		ActivationURL string
	}{
		Name:          user.Name,
		ActivationURL: fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, hashToken),
	}
	err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Name, user.Email, vars, false)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)
		app.internalServerError(w, r, err)
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	if err := app.jsonResponse(w, http.StatusOK, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=80"`
}

func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.unauthorizedError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	token, err := app.authenticator.GenerateToken(jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, map[string]any{
		"token": token,
	}); err != nil {
		app.internalServerError(w, r, err)
	}
}
