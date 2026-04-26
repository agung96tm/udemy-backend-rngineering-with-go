package main

import (
	"net/http"
	"net/http/httptest"
	"socialv2/internal/auth"
	"socialv2/internal/store"
	"testing"

	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockAuthenticator := &auth.TestAuthenticator{}

	return &application{
		logger:        logger,
		store:         mockStore,
		authenticator: mockAuthenticator,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
