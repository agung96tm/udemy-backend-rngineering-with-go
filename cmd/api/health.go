package main

import (
	"net/http"
)

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"message": "OK",
		"env":     app.config.env,
		"version": app.config.version,
	}

	if err := app.writeJSON(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
