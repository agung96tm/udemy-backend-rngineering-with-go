package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = app.writeJSONError(w, http.StatusInternalServerError, "Internal server error")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = app.writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = app.writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	_ = app.writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized basic token error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)
	_ = app.writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
}
