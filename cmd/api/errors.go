package main

import (
	"fmt"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = app.writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
}

func (app *application) forbiddenError(w http.ResponseWriter, r *http.Request) {
	app.logger.Errorw("forbidden", "method", r.Method, "path", r.URL.Path, "error")
	_ = app.writeJSONError(w, http.StatusForbidden, "Forbidden error")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request error", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = app.writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = app.writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized", "method", r.Method, "path", r.URL.Path, "error", err)
	_ = app.writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized", "method", r.Method, "path", r.URL.Path, "error", err)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	_ = app.writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) rateLimiterExceedError(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	_ = app.writeJSONError(w, http.StatusTooManyRequests, fmt.Sprintf("Rate limit exceeded: %s", retryAfter))
}
