package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"socialv3/internal/store"

	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validated:"required,max=50,min=4"`
	Email    string `json:"email" validated:"required,email"`
	Password string `json:"password" validated:"required,min=8,max=32"`
}

// authRegisterHandler godoc
//
//	@Summary		Register user
//	@Description	Mendaftarkan user baru. Mengirim username, email, password. Setelah itu token invite dikirim via email (jika mailing diaktifkan).
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterUserPayload		true	"Username (4-50), email, password (8-32)"
//	@Success		201		{object}	map[string]interface{}	"message: user registered"
//	@Failure		400		{object}	object					"Request body tidak valid"
//	@Failure		500		{object}	object					"Server error"
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

	_ = app.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user registered",
	})
}
