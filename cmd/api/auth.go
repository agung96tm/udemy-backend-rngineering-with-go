package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"socialv3/internal/mailer"
	"socialv3/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validated:"required,max=50,min=4"`
	Email    string `json:"email" validated:"required,email"`
	Password string `json:"password" validated:"required,min=8,max=32"`
}

type UserWithToken struct {
	*store.User
	token string
}

// authRegisterHandler godoc
//
//	@Summary		Register user
//	@Description	Mendaftarkan user baru. Mengirim username, email, password. Setelah itu token invite dikirim via email (jika mailing diaktifkan).
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterUserPayload	true	"Username (4-50), email, password (8-32)"
//	@Success		201		{object}	UserWithToken		"User yang baru dibuat beserta token aktivasi"
//	@Failure		400		{object}	object				"Request body tidak valid"
//	@Failure		500		{object}	object				"Server error"
//	@Router			/authentication/user [post]
func (app *application) authRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])
	if err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	// mailing
	userWithToken := UserWithToken{
		User:  user,
		token: plainToken,
	}
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, hashToken)
	isProd := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProd)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerError(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusCreated, userWithToken)
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validated:"required,email"`
	Password string `json:"password" validated:"required,min=6,max=32"`
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
	if user.Password.Compare(payload.Password) != nil {
		app.unauthorizedError(w, r, fmt.Errorf("password is invalid"))
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.jwt.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.jwt.issuer,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
